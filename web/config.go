package web

import "io/fs"

type Config struct {
	Environment  string    `json:"env"`
	BaseURL      string    `json:"base_url"`
	Build        BuildInfo `json:"build"`
	PublicAssets fs.FS     `json:"-"`
}

type BuildInfo struct {
	Dirty     string `json:"dirty,omitempty"`
	Branch    string `json:"branch,omitempty"`
	Version   string `json:"version,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}
