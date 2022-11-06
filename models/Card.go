package models

type Card struct {
	Url          string `json:"url"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	AuthorName   string `json:"author_name"`
	AuthorUrl    string `json:"author_url"`
	ProviderName string `json:"provider_name"`
	ProviderUrl  string `json:"provider_url"`
	Html         string `json:"html"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Image        string `json:"image"`
	EmbedUrl     string `json:"embed_url"`
	Blurhash     string `json:"blurhash"`
}
