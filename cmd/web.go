package cmd

import (
	"os"
	"strings"
	"webapp/db"
	"webapp/web"

	"github.com/gin-gonic/gin"
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
	Short:   "Start the web server",
	Run: func(cmd *cobra.Command, args []string) {
		var dbh *db.DB = nil
		if dburl := strings.TrimSpace(os.Getenv(Env_DATABASE_URL)); dburl != "" {
			dbh = must1(db.Connect(dburl))
		} else {
			log.Printf("skipping database, set %q to connect", Env_DATABASE_URL)
		}

		config := web.Config{
			Mode:                 util.Getenv(Env_GIN_MODE, gin.DebugMode),
			Build:                web.GetBuildInfo(),
			CookieEncryptionKeys: cookieEncryptionKeysFromEnv(),
		}

		server := must1(web.NewServer(config, dbh))
		must(server.Run())
	},
}

func cookieEncryptionKeysFromEnv() [][]byte {
	keyStrings := strings.Split(util.Getenv(Env_COOKIE_ENCRYPTION_KEYS, "TODO:webapp-env-secret"), ",")
	keys := make([][]byte, len(keyStrings))
	for i, k := range keyStrings {
		keys[i] = []byte(k)
	}
	return keys
}
