package cmd

import (
	"strings"
	"webapp/db"
	"webapp/web"

	log "github.com/maerics/golog"
	util "github.com/maerics/goutil"
	cobra "github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(webCmd)
}

var webCmd = &cobra.Command{
	Use:     "web",
	Aliases: []string{"w"},
	Short:   "Start the web server.",
	Run: func(cmd *cobra.Command, args []string) {
		config := web.Config{
			Environment: util.Getenv(Env_ENV, "development"),
			Build: web.BuildInfo{
				Dirty:     strings.TrimRight(BuildDirty, ","),
				Branch:    BuildBranch,
				Version:   BuildVersion,
				Timestamp: BuildTimestamp,
			},
			PublicAssets: PublicAssets,
		}
		db := log.Must1(db.Connect(util.MustEnv(Env_DATABASE_URL)))
		server := log.Must1(web.NewServer(config, db))
		log.Must(server.Run())
	},
}
