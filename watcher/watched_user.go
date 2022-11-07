package watcher

import (
	"github.com/LeakIX/mastosearch/models"
	"time"
)

type WatchedUser struct {
	Id                string `gorm:"primaryKey"`
	LastUpdateId      string
	PostCount         int
	NoIndex           bool
	LastNoIndexUpdate time.Time
	AccountDetails    models.Account `gorm:"-"`
}
