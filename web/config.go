package web

import (
	"io/fs"
	"runtime/debug"

	"golang.org/x/crypto/acme/autocert"
)

var Branch string

type Config struct {
	Environment string    `json:"env"`
	BaseURL     string    `json:"base_url"`
	Build       BuildInfo `json:"build"`

	PublicAssets    fs.FS             `json:"-"`
	AutoCertManager *autocert.Manager `json:"-"`
}

type BuildInfo struct {
	Timestamp string `json:"timestamp,omitempty"`
	Version   string `json:"version,omitempty"`
	Branch    string `json:"branch,omitempty"`
	Dirty     bool   `json:"dirty,omitempty"`
	GoVersion string `json:"goversion,omitempty"`
}

var (
	cachedBuildInfo BuildInfo
	loadedBuildInfo bool
)

func GetBuildInfo() BuildInfo {
	if !loadedBuildInfo {
		cachedBuildInfo = getBuildInfo()
		loadedBuildInfo = true
	}
	return cachedBuildInfo
}

func getBuildInfo() BuildInfo {
	var bi BuildInfo
	buildinfo, ok := debug.ReadBuildInfo()
	if ok {
		bi.GoVersion = buildinfo.GoVersion

		settings := map[string]string{}
		for _, setting := range buildinfo.Settings {
			settings[setting.Key] = setting.Value
		}
		bi.Dirty = settings["vcs.modified"] == "true"
		bi.Version = settings["vcs.revision"]
		bi.Timestamp = settings["vcs.time"]

		if Branch != "" {
			bi.Branch = Branch
		}
	}
	return bi
}
