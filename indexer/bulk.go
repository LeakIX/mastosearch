package indexer

import (
	"context"
	"github.com/LeakIX/mastosearch/models"
	"github.com/olivere/elastic/v7"
	"log"
	"time"
)

type BulkService struct {
	esClient     *elastic.Client
	bulkService  *elastic.BulkService
	InputChannel chan models.Update
	bulkSize     int
}

func NewBulkService(esClient *elastic.Client, bulkSize int) (*BulkService, error) {
	bs := &BulkService{
		esClient:     esClient,
		bulkService:  nil,
		InputChannel: make(chan models.Update),
		bulkSize:     bulkSize,
	}
	err := bs.initBulkService()
	if err != nil {
		return nil, err
	}
	go bs.runBulkService()
	return bs, nil
}

func (bs *BulkService) initBulkService() (err error) {
	bs.bulkService = bs.esClient.Bulk()
	bs.bulkService.Retrier(elastic.NewBackoffRetrier(elastic.NewConstantBackoff(10 * time.Second)))
	return err
}

func (bs *BulkService) runBulkService() {
	for {
		event, isOpen := <-bs.InputChannel
		if !isOpen {
			bs.commitBulkActions()
			break
		}

		bs.bulkService.Add(
			elastic.NewBulkIndexRequest().Index("updates").Id(event.Url).RetryOnConflict(10).Doc(event))
		bs.bulkService.Add(
			elastic.NewBulkIndexRequest().Index("users").Id(event.Account.Id).RetryOnConflict(10).Doc(event.Account))
		if bs.bulkService.NumberOfActions() >= bs.bulkSize {
			bs.commitBulkActions()
		}
	}
}

func (bs *BulkService) commitBulkActions() {
	log.Printf("Storing %d items", bs.bulkService.NumberOfActions())
	bulkResponses, err := bs.bulkService.Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if bulkResponses.Errors {
		for _, bulkResponse := range bulkResponses.Failed() {
			log.Println(bulkResponse.Error.Reason)
		}
		panic("Bulk failed")
	}
	bs.bulkService.Reset()
	log.Println("Stored")
}
