package models

type Server struct {
	Domain           string `json:"domain"`
	ApprovalRequired bool   `json:"approval_required"`
}
