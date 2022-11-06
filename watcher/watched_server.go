package watcher

import (
	"encoding/json"
	"github.com/LeakIX/mastosearch/models"
	"github.com/LeakIX/mastosearch/stream"
	"gorm.io/gorm"
	"log"
	"net/http"
)

// TODO: clean me, I just download user timelines

type WatchedServer struct {
	server       models.Server
	status       string
	publicStream *stream.PublicStream
	userQueue    chan models.Update
	userDb       *gorm.DB
}

func (ws *WatchedServer) processUpdate(update models.Update, outputChannel chan models.Update) {
	var watchedUser WatchedUser
	tx := ws.userDb.First(&watchedUser, WatchedUser{Id: update.Account.Url})
	if tx.Error != nil {
		watchedUser.Id = update.Account.Url
	}
	defer ws.userDb.Save(&watchedUser)
	watchedUser.PostCount++
	if watchedUser.PostCount >= update.Account.StatusesCount {
		watchedUser.LastUpdateId = update.Id
		return
	}
	log.Printf("downloading posts for user %s", update.Account.Url)

	ws.getAllPosts(update, watchedUser, outputChannel)

	watchedUser.PostCount = update.Account.StatusesCount
	watchedUser.LastUpdateId = update.Id
}
func (ws *WatchedServer) getAllPosts(lastUpdate models.Update, watchedUser WatchedUser, outputChannel chan models.Update) {
	accountId := lastUpdate.Account.Id
	maxId := lastUpdate.Id
	for {
		statusUrl := "https://" + ws.server.Domain + "/api/v1/accounts/" + accountId + "/statuses?limit=80&max_id=" + maxId
		if len(watchedUser.LastUpdateId) > 0 {
			statusUrl += "&since_id=" + watchedUser.LastUpdateId
		}

		resp, err := http.DefaultClient.Get(statusUrl)
		if err != nil {
			log.Println("failed user statuses", err)
			return
		}
		jsonDecoder := json.NewDecoder(resp.Body)
		var statuses []models.Update
		err = jsonDecoder.Decode(&statuses)
		if err != nil {
			log.Println("failed user statuses", err)
			return
		}
		if len(statuses) < 1 {
			break
		}
		for _, status := range statuses {
			maxId = status.Id
			outputChannel <- status
		}
	}

}

func (ws *WatchedServer) DownloadUserPosts(outputChannel chan models.Update) {
	for update := range ws.userQueue {
		ws.processUpdate(update, outputChannel)
	}
}
