package main

import (
	"context"
	"encoding/json"
	"github.com/LeakIX/mastosearch/indexer"
	"github.com/LeakIX/mastosearch/models"
	"github.com/nsqio/go-nsq"
	"github.com/olivere/elastic/v7"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var err error
	esClient, err = elastic.NewClient(elastic.SetURL(getConfig("ES_SERVER", "http://127.0.0.1:9200")),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false))
	if err != nil {
		panic(err)
	}
	bulkIndexer, err = indexer.NewBulkService(esClient, 200)
	if err != nil {
		panic(err)
	}
	config := nsq.NewConfig()

	consumer, err := nsq.NewConsumer("updates", "indexer", config)
	if err != nil {
		log.Fatal(err)
	}
	consumer.AddHandler(&Indexer{})
	err = consumer.ConnectToNSQD(getConfig("NSQD_SERVER", "127.0.0.1:4150"))
	if err != nil {
		log.Fatal(err)
	}

	deleter, err := nsq.NewConsumer("deletes", "indexer", config)
	if err != nil {
		log.Fatal(err)
	}
	deleter.AddHandler(&Deleter{})
	err = deleter.ConnectToNSQD(getConfig("NSQD_SERVER", "127.0.0.1:4150"))
	if err != nil {
		log.Fatal(err)
	}

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Gracefully stop the consumer.
	consumer.Stop()
}

type Deleter struct{}

// HandleMessage implements the Handler interface.
func (h *Deleter) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	// do whatever actual message processing is desired
	var deleteRequest models.DeleteRequest
	err := json.Unmarshal(m.Body, &deleteRequest)
	if err != nil {
		return err
	}
	_, err = esClient.DeleteByQuery().Index("updates").Query(
		elastic.NewBoolQuery().Must(
			elastic.NewMatchQuery("id", deleteRequest.Id),
			elastic.NewMatchQuery("account.server", deleteRequest.Server),
		)).Do(context.Background())

	return err
}

type Indexer struct{}

// HandleMessage implements the Handler interface.
func (h *Indexer) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	// do whatever actual message processing is desired
	var update models.Update
	err := json.Unmarshal(m.Body, &update)
	if err != nil {
		return err
	}
	bulkIndexer.InputChannel <- update

	return nil
}

var bulkIndexer *indexer.BulkService
var esClient *elastic.Client

func getConfig(configKey string, defaultValue string) string {
	if value, found := os.LookupEnv(configKey); found {
		return value
	}
	return defaultValue
}
