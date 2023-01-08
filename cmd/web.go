package cmd

import (
	"os"
	"strings"
	"webapp/db"
	"webapp/web"

	log "github.com/maerics/golog"
	util "github.com/maerics/goutil"
	cobra "github.com/spf13/cobra"
	"golang.org/x/crypto/acme/autocert"
)

func init() {
	rootCmd.AddCommand(webCmd)

	webCmd.Flags().StringVarP(&optWebAutotlsEmail,
		"autotls-email", "", "", "email address for problems with issued certs")
	webCmd.Flags().StringArrayVarP(&optWebAutotlsHosts,
		"autotls-hosts", "", nil, "list of hostnames for autotls")
	webCmd.Flags().StringVarP(&optWebAutotlsDircache,
		"autotls-dir", "", "", "autotls dircache location")
}

var (
	optWebAutotlsEmail    string
	optWebAutotlsHosts    []string
	optWebAutotlsDircache string
)

var webCmd = &cobra.Command{
	Use:     "web",
	Aliases: []string{"w"},
	Short:   "Start the web server",
	Run: func(cmd *cobra.Command, args []string) {
		var dbh *db.DB = nil
		if dburl := strings.TrimSpace(os.Getenv(Env_DATABASE_URL)); dburl != "" {
			dbh = log.Must1(db.Connect(dburl))
		} else {
			log.Printf("skipping database, set %q to connect", Env_DATABASE_URL)
		}

		config := web.Config{
			Environment:  util.Getenv(Env_ENV, "development"),
			Build:        web.GetBuildInfo(),
			PublicAssets: PublicAssets,
		}

		if optWebAutotlsEmail != "" || len(optWebAutotlsHosts) > 0 || optWebAutotlsDircache != "" {
			config.AutoCertManager = &autocert.Manager{Prompt: autocert.AcceptTOS}
			if optWebAutotlsEmail != "" {
				config.AutoCertManager.Email = optWebAutotlsEmail
				log.Debugf("autotls-email=%#v", optWebAutotlsEmail)
			}
			if len(optWebAutotlsHosts) > 0 {
				config.AutoCertManager.HostPolicy = autocert.HostWhitelist(optWebAutotlsHosts...)
				log.Debugf("autotls-hosts=%#v", optWebAutotlsHosts)
			}
			if optWebAutotlsDircache != "" {
				log.Must(os.MkdirAll(optWebAutotlsDircache, os.FileMode(0o755)))
				config.AutoCertManager.Cache = autocert.DirCache(optWebAutotlsDircache)
				log.Debugf("autotls-dir=%#v", optWebAutotlsDircache)
			}
		}

		server := log.Must1(web.NewServer(config, dbh))
		log.Must(server.Run())
	},
}
