package main

import (
	"encoding/json"
	"github.com/LeakIX/mastosearch/models"
	"github.com/LeakIX/mastosearch/watcher"
	"github.com/nsqio/go-nsq"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	channel := make(chan models.Update)
	deleteChannel := make(chan models.DeleteRequest)
	msWatcher := watcher.NewWatcher(channel, deleteChannel)
	server := getConfig("MASTODON_SERVER", "")
	if server == "" {
		var postBody = "{\"query\":\"{nodes{domain,signup}}\\n    \",\"variables\":null}"
		resp, err := http.DefaultClient.Post("https://api.fediverse.observer/", "application/json", strings.NewReader(postBody))
		if err != nil {
			panic(err)
		}
		decoder := json.NewDecoder(resp.Body)
		var nodeList FediverseObsServer
		err = decoder.Decode(&nodeList)
		if err != nil {
			panic(err)
		}
		for _, node := range nodeList.Data.Nodes {
			mastodonServer := models.Server{
				Domain:           node.Domain,
				ApprovalRequired: node.Signup,
			}
			msWatcher.AddServer(mastodonServer)
		}
	} else {
		msWatcher.AddServer(models.Server{
			Domain: server,
		})
	}
	log.Println("Watcher is running")
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer(getConfig("NSQD_SERVER", "127.0.0.1:4150"), config)
	if err != nil {
		log.Fatal(err)
	}
	go func(deleteChannel chan models.DeleteRequest) {
		for deleteRequest := range deleteChannel {
			payload, _ := json.Marshal(&deleteRequest)
			producer.Publish("deletes", payload)
		}
	}(deleteChannel)
	for update := range channel {
		payload, _ := json.Marshal(&update)
		producer.Publish("updates", payload)
	}
}

func getConfig(configKey string, defaultValue string) string {
	if value, found := os.LookupEnv(configKey); found {
		return value
	}
	return defaultValue
}

type FediverseObsServer struct {
	Data struct {
		Nodes []struct {
			Domain string `json:"domain"`
			Signup bool   `json:"signup"`
		} `json:"nodes"`
	} `json:"data"`
}
