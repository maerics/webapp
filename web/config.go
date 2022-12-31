package web

import "io/fs"

type Config struct {
	Environment  string    `json:"env"`
	BaseURL      string    `json:"base_url"`
	Build        BuildInfo `json:"build"`
	PublicAssets fs.FS     `json:"-"`
}

type BuildInfo struct {
	Branch    string `json:"branch"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}
