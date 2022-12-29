package cmd

import (
	"webapp/db"

	log "github.com/maerics/golog"
	util "github.com/maerics/goutil"
	cobra "github.com/spf13/cobra"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	rootCmd.AddCommand(dbCmd)

	dbCmd.AddCommand(migrateCmd)
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage the database.",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run the database migrations.",
	Run: func(cmd *cobra.Command, args []string) {
		dburl := util.MustEnv(Env_DATABASE_URL)
		db := log.Must1(db.Connect(dburl))
		log.Must(db.Migrate())
	},
}
