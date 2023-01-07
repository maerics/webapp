package web

import (
	"io/fs"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"

	log "github.com/maerics/golog"
)

type Config struct {
	Environment  string    `json:"env"`
	BaseURL      string    `json:"base_url"`
	Build        BuildInfo `json:"build"`
	PublicAssets fs.FS     `json:"-"`
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

		if bi.Version != "" {
			cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "head")
			if branchName, err := cmd.Output(); err == nil {
				bi.Branch = strings.TrimSpace(string(branchName))
			} else {
				log.Errorf("failed to determine git branch name: %v", err)
				bi.Branch = os.Getenv("GIT_BRANCH")
			}
		}
	}
	return bi
}
