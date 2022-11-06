package models

type DeleteRequest struct {
	Server string `json:"server"`
	Id     string `json:"id"`
}
