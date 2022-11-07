package watcher

import (
	"github.com/LeakIX/mastosearch/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	watchedServer := NewWatchedServer(server, w.userDb, w.outputChannel, w.deleteChannel)
	w.servers[server.Domain] = watchedServer
	return
}
