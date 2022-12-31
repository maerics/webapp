package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"runtime/debug"
	"webapp/web"

	log "github.com/maerics/golog"
	cobra "github.com/spf13/cobra"
)

const (
	Env_ENV          = "ENV"          // The name of the environment, e.g. "production", "development", etc.
	Env_DATABASE_URL = "DATABASE_URL" // The database connection string.
)

var (
	BuildBranch    string
	BuildVersion   string
	BuildTimestamp string

	PublicAssets fs.FS
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func Run() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("recovered %#v (%v)\n%s", r, r, string(debug.Stack()))
		}
	}()

	log.Must(rootCmd.Execute())
}

var rootCmd = &cobra.Command{
	Use:   "webapp",
	Short: "A database connected web application.",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"ver", "v"},
	Short:   "Print build and version information.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(string(log.Must1(json.MarshalIndent(web.BuildInfo{
			Branch:    BuildBranch,
			Version:   BuildVersion,
			Timestamp: BuildTimestamp,
		}, "", "  "))))
	},
}
