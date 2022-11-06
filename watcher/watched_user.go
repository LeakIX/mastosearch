package watcher

import "github.com/LeakIX/mastosearch/models"

type WatchedUser struct {
	Id             string `gorm:"primaryKey"`
	LastUpdateId   string
	PostCount      int
	AccountDetails models.Account `gorm:"-"`
}
