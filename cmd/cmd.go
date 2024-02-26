package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"webapp/web"

	log "github.com/maerics/golog"
	util "github.com/maerics/goutil"
	cobra "github.com/spf13/cobra"
)

const (
	Env_MODE                   = "MODE"         // The deployment mode, e.g. debug, release, test
	Env_DATABASE_URL           = "DATABASE_URL" // The database connection string.
	Env_COOKIE_ENCRYPTION_KEYS = "COOKIE_ENCRYPTION_KEYS"
)

var (
	BuildDirty     string
	BuildBranch    string
	BuildVersion   string
	BuildTimestamp string
)

var (
	optCmdShowVersion bool
)

func init() {
	rootCmd.AddCommand(versionCmd)

	rootCmd.Flags().BoolVarP(&optCmdShowVersion,
		"version", "v", false, "show version and build information")
}

func Run() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("recovered %#v (%v)\n%s", r, r, string(debug.Stack()))
		}
	}()

	must(rootCmd.Execute())
}

var rootCmd = &cobra.Command{
	Use:   "webapp",
	Short: "A database connected web application",
	Run: func(cmd *cobra.Command, args []string) {
		if optCmdShowVersion {
			fmt.Println(mustGetVersionString())
			os.Exit(0)
		}
		cmd.Help()
	},
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"ver", "v"},
	Short:   "Show version and build information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(mustGetVersionString())
	},
}

func mustGetVersionString() string {
	return util.MustJson(web.GetBuildInfo(), true)
}

func must(err error) {
	log.Must(err)
}

func must1[T any](t T, err error) T {
	must(err)
	return t
}
