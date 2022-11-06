package models

import "time"

type Update struct {
	Id                 string    `json:"id"`
	CreatedAt          time.Time `json:"created_at"`
	InReplyToId        string    `json:"in_reply_to_id"`
	InReplyToAccountId string    `json:"in_reply_to_account_id"`
	Sensitive          bool      `json:"sensitive"`
	Visibility         string    `json:"visibility"`
	Language           string    `json:"language"`
	Uri                string    `json:"uri"`
	Url                string    `json:"url"`
	RepliesCount       int       `json:"replies_count"`
	ReblogsCount       int       `json:"reblogs_count"`
	FavouritesCount    int       `json:"favourites_count"`
	EditedAt           time.Time `json:"edited_at"`
	Content            string    `json:"content"`
	//Reblog             string            `json:"reblog"`
	Application      Application       `json:"application"`
	Account          Account           `json:"account"`
	Tags             []Tag             `json:"tags"`
	MediaAttachments []MediaAttachment `json:"media_attachments"`
	Card             Card              `json:"card"`
}

type Tag struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Application struct {
	Name    string `json:"name"`
	Website string `json:"website"`
}
