package watcher

import (
	"github.com/LeakIX/mastosearch/models"
	"github.com/LeakIX/mastosearch/stream"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"sync"
)

type Watcher struct {
	serverLock    sync.RWMutex
	servers       map[string]*WatchedServer
	outputChannel chan models.Update
	deleteChannel chan models.DeleteRequest
	userDb        *gorm.DB
}

func NewWatcher(outputChannel chan models.Update, deleteChannel chan models.DeleteRequest) *Watcher {
	db, err := gorm.Open(sqlite.Open("watcher.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(WatchedUser{})
	if err != nil {
		panic(err)
	}
	return &Watcher{
		servers:       make(map[string]*WatchedServer),
		outputChannel: outputChannel,
		deleteChannel: deleteChannel,
		userDb:        db,
	}
}

func (w *Watcher) AddServer(server models.Server) {
	w.serverLock.Lock()
	defer w.serverLock.Unlock()
	_, found := w.servers[server.Domain]
	if found {
		return
	}
	watchedServer := &WatchedServer{
		server:    server,
		status:    "disconnected",
		userQueue: make(chan models.Update, 256),
		userDb:    w.userDb,
	}
	w.servers[server.Domain] = watchedServer
	go w.watchServer(watchedServer)
	go watchedServer.DownloadUserPosts(w.outputChannel)
	return
}

func (w *Watcher) watchServer(watchedServer *WatchedServer) {
	proxyChannel := make(chan models.Update)
	_, err := stream.NewPublicStream(watchedServer.server, proxyChannel, w.deleteChannel)
	if err != nil {
		log.Println(watchedServer.server.Domain, err)
	}
	for event := range proxyChannel {
		select {
		case watchedServer.userQueue <- event:
			// Watched server is now processing user account
		default:
			// Do nothing, server account check is full
		}
		w.outputChannel <- event
	}
}
