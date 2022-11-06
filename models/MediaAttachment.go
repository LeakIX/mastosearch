package models

type MediaAttachment struct {
	Id               string    `json:"id"`
	Type             string    `json:"type"`
	Url              string    `json:"url"`
	PreviewUrl       string    `json:"preview_url"`
	RemoteUrl        string    `json:"remote_url"`
	PreviewRemoteUrl string    `json:"preview_remote_url"`
	TextUrl          string    `json:"text_url"`
	Meta             MediaMeta `json:"meta"`
	Description      string    `json:"description"`
	Blurhash         string    `json:"blurhash"`
}

type MediaSize struct {
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Size   string  `json:"size"`
	Aspect float64 `json:"aspect"`
}

type MediaMeta struct {
	Original MediaSize `json:"original"`
	Small    MediaSize `json:"small"`
}
