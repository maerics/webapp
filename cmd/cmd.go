package cmd

import (
	"io/fs"
	"runtime/debug"

	log "github.com/maerics/golog"
	cobra "github.com/spf13/cobra"
)

const (
	Env_ENV          = "ENV"          // The name of the environment, e.g. "production", "development", etc.
	Env_DATABASE_URL = "DATABASE_URL" // The database connection string.
)

var PublicAssets fs.FS

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
