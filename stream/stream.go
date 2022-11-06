package stream

import (
	"encoding/json"
	"github.com/LeakIX/mastosearch/models"
	"github.com/LeakIX/mastosearch/net"
	"log"
)

type PublicStream struct {
	url           string
	domain        string
	ssre          *net.ServerSentEventReader
	ssreChannel   chan net.ServerSentEvent
	outputChannel chan models.Update
	deleteChannel chan models.DeleteRequest
}

func NewPublicStream(server models.Server, outputChannel chan models.Update, deleteChannel chan models.DeleteRequest) (ps *PublicStream, err error) {
	ps = &PublicStream{
		url:           "https://" + server.Domain + "/api/v1/streaming/public/local",
		domain:        server.Domain,
		ssreChannel:   make(chan net.ServerSentEvent),
		outputChannel: outputChannel,
		deleteChannel: deleteChannel,
	}
	ps.ssre, err = net.NewServerSentEventReader(ps.url, ps.ssreChannel)
	if err != nil {
		return nil, err
	}
	go ps.translateEvents()
	return ps, nil
}

func (ps *PublicStream) translateEvents() {
	for serverEvent := range ps.ssreChannel {
		switch serverEvent.Event {
		case "update":
			var event models.Update
			err := json.Unmarshal(serverEvent.Data, &event)
			if err != nil {
				log.Println(err)
				continue
			}
			ps.outputChannel <- event
		case "delete":
			ps.deleteChannel <- models.DeleteRequest{
				Server: ps.domain,
				Id:     string(serverEvent.Data),
			}
		}

	}
}
