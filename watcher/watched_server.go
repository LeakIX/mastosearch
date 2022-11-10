package watcher

import (
	"context"
	"encoding/json"
	"github.com/LeakIX/mastosearch/models"
	"github.com/LeakIX/mastosearch/stream"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strings"
	"time"
)

// TODO: clean me, I just download user timelines

type WatchedServer struct {
	server            models.Server
	status            string
	publicStream      *stream.PublicStream
	userQueue         chan models.Update
	outputChannel     chan models.Update
	deleteChannel     chan models.DeleteRequest
	proxyChannel      chan models.Update
	userDb            *gorm.DB
	httpClient        *http.Client
	httpClientLimiter *rate.Limiter
}

func NewWatchedServer(server models.Server, userDb *gorm.DB, outputChannel chan models.Update, deleteChannel chan models.DeleteRequest) *WatchedServer {
	var err error
	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 15 * time.Second,
	}
	ws := &WatchedServer{
		server:        server,
		userDb:        userDb,
		outputChannel: outputChannel,
		deleteChannel: deleteChannel,
		proxyChannel:  make(chan models.Update),
		userQueue:     make(chan models.Update, 256),
		httpClient:    httpClient,
		// default mastodon limits :
		httpClientLimiter: rate.NewLimiter(rate.Every(5*time.Minute), 300),
	}
	ws.publicStream, err = stream.NewPublicStream(ws.server, ws.proxyChannel, ws.deleteChannel)
	if err != nil {
		log.Println(ws.server.Domain, err)
	}
	go ws.runProxy()
	go ws.watchUser()
	return ws
}

func (ws *WatchedServer) runProxy() {
	for event := range ws.proxyChannel {
		// If the user posts this status we will remove all prior content from the index :
		for _, tag := range event.Tags {
			if tag.Name == "RemoveMyContentFromSearchEngines" {
				deleteRequest := models.DeleteRequest{
					Server: ws.server.Domain,
					UserId: event.Account.Id,
				}
				ws.deleteChannel <- deleteRequest
				trueVal := true
				event.Account.NoIndex = &trueVal
			}
		}
		if event.Account.NoIndex == nil {
			// NoIndex wasn't provided in update stream, lookup in the cache first,
			watchedUser := ws.getWatchedUser(event)
			if watchedUser == nil || time.Since(watchedUser.LastNoIndexUpdate) > 60*time.Minute {
				watchedUser = &WatchedUser{
					Id:                event.Account.Url,
					NoIndex:           ws.legacyNoIndexStatus(event.Account),
					LastNoIndexUpdate: time.Now(),
				}
				// cache in local database
				ws.userDb.Save(&watchedUser)
			}
			event.Account.NoIndex = &watchedUser.NoIndex
		}
		if !*event.Account.NoIndex {
			// noindex is false, indexing
			ws.dispatchEvent(event)
			continue
		}

	}
}

func (ws *WatchedServer) dispatchEvent(event models.Update) {
	select {
	case ws.userQueue <- event:
		// Watched server is now processing user account
	default:
		// Do nothing, server account check is full
	}
	ws.outputChannel <- event
}

func (ws *WatchedServer) getWatchedUser(update models.Update) *WatchedUser {
	var watchedUser WatchedUser
	tx := ws.userDb.First(&watchedUser, WatchedUser{Id: update.Account.Url})
	if tx.Error == gorm.ErrRecordNotFound {
		return nil
	}
	if tx.Error != nil {
		// Couldn't access the database, panic mode
		panic(tx.Error)
	}
	return &watchedUser
}

func (ws *WatchedServer) legacyNoIndexStatus(account models.Account) bool {
	req, err := http.NewRequest(http.MethodGet, account.Url, nil)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("User-Agent", "Mastosearch/0.1 (+https://github.com/LeakIX/mastosearch)")
	resp, err := ws.HttpDo(req)
	if err != nil || resp.StatusCode != 200 {
		log.Println(err)
		// in doubt, no index
		return true
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println(err)
		return true
	}
	noIndex := false
	doc.Find("meta").Each(func(i int, selection *goquery.Selection) {
		if name, _ := selection.Attr("name"); name == "robots" {
			if content, _ := selection.Attr("content"); strings.Contains(content, "noindex") {
				noIndex = true
			}
		}
	})
	return noIndex
}

func (ws *WatchedServer) watchUser() {
	for update := range ws.userQueue {
		ws.downloadUserPosts(update)
	}
}

func (ws *WatchedServer) downloadUserPosts(update models.Update) {
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

	ws.getAllPosts(update, watchedUser)

	watchedUser.PostCount = update.Account.StatusesCount
	watchedUser.LastUpdateId = update.Id
}
func (ws *WatchedServer) getAllPosts(lastUpdate models.Update, watchedUser WatchedUser) {
	accountId := lastUpdate.Account.Id
	maxId := lastUpdate.Id
	for {
		statusUrl := "https://" + ws.server.Domain + "/api/v1/accounts/" + accountId + "/statuses?limit=80&max_id=" + maxId
		if len(watchedUser.LastUpdateId) > 0 {
			statusUrl += "&since_id=" + watchedUser.LastUpdateId
		}
		req, _ := http.NewRequest(http.MethodGet, statusUrl, nil)
		resp, err := ws.HttpDo(req)
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
			ws.outputChannel <- status
		}
	}

}

func (ws *WatchedServer) HttpDo(req *http.Request) (*http.Response, error) {
	err := ws.httpClientLimiter.Wait(context.Background()) // This is a blocking call. Honors the rate limit
	if err != nil {
		return nil, err
	}
	resp, err := ws.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
