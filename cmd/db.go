package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"webapp/db"

	log "github.com/maerics/golog"
	util "github.com/maerics/goutil"
	cobra "github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dbCmd)

	dbCmd.AddCommand(selectCmd)
	dbCmd.AddCommand(executeCmd)
	dbCmd.AddCommand(migrateCmd)

	executeCmd.Flags().BoolVarP(&optDbExecuteCommit,
		"commit", "", false, "commit the transaction instead of rolling back")

	selectCmd.Flags().BoolVarP(&optDbSelectCsvOutput,
		"csv", "c", false, "format result set as CSV instead of JSON")
	selectCmd.Flags().StringVarP(&optDbSelectCsvSep,
		"sep", "s", ",", "separator to use for CSV output")
}

var (
	optDbExecuteCommit   = false
	optDbSelectCsvOutput = false
	optDbSelectCsvSep    = ","
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage the database",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var executeCmd = &cobra.Command{
	Use:     "execute",
	Aliases: []string{"exec", "e"},
	Short:   "Execute SQL commands from STDIN",
	Run: func(cmd *cobra.Command, args []string) {
		// Read the query from STDIN.
		buf := &bytes.Buffer{}
		log.Must1(io.Copy(buf, os.Stdin))
		query := strings.TrimSpace(buf.String())

		// Execute the query.
		dburl := util.MustEnv(Env_DATABASE_URL)
		db := log.Must1(db.Connect(dburl))
		tx := db.MustBegin()
		log.Printf("executing query:\n\n    %v\n\n", query)
		t0 := time.Now()

		// Print the affected rows and last insert id, if applicable.
		result := log.Must1(tx.Exec(query))
		if n, err := result.RowsAffected(); err == nil {
			log.Printf("query affected %v row(s)", n)
		}
		if n, err := result.LastInsertId(); err == nil {
			log.Printf("last insert id=%v", n)
		}

		// Commit or rollback.
		if optDbExecuteCommit {
			log.Must(tx.Commit())
			log.Printf("committed transaction in %v", time.Since(t0))
		} else {
			log.Must(tx.Rollback())
			log.Printf("rolled back transaction in %v (see help for details)", time.Since(t0))
		}
	},
}

var selectCmd = &cobra.Command{
	Use:     "select",
	Aliases: []string{"sel", "s"},
	Short:   "Print the results of a database query from STDIN to STDOUT",
	Run: func(cmd *cobra.Command, args []string) {
		outputcsv := optDbSelectCsvOutput
		if outputcsv {
			if len(optDbSelectCsvSep) != 1 {
				log.Fatalf("CSV separator must be one byte long, got %q", optDbSelectCsvSep)
			}
		}

		// Read the query from STDIN.
		buf := &bytes.Buffer{}
		log.Must1(io.Copy(buf, os.Stdin))
		query := strings.TrimSpace(buf.String())

		// Execute the query.
		dburl := util.MustEnv(Env_DATABASE_URL)
		db := log.Must1(db.Connect(dburl))
		log.Printf("executing query:\n\n    %v\n\n", query)
		rows := log.Must1(db.Query(query))

		// Inspect the result set column types.
		columns := log.Must1(rows.Columns())
		values := make([]any, len(columns))
		scanArgs := make([]any, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		// Setup the optional CSV writer.
		var w *csv.Writer
		if outputcsv {
			w = csv.NewWriter(os.Stdout)
			w.Comma = rune(optDbSelectCsvSep[0])
			log.Must(w.Write(columns))
		}

		// Print to stdout.
		count := 0
		enc := json.NewEncoder(os.Stdout)
		for rows.Next() {
			count++
			log.Must(rows.Scan(scanArgs...))

			// Print the JSON or CSV result.
			if outputcsv {
				row := make([]string, len(values))
				for i, v := range values {
					if v != nil {
						row[i] = fmt.Sprintf("%v", v)
					}
				}
				log.Must(w.Write(row))
			} else {
				log.Must(enc.Encode(util.OrderedJsonObj{Keys: columns, Values: values}))
			}
		}

		if outputcsv {
			w.Flush()
			log.Must(w.Error())
		}
		log.Printf("query returned %v row(s)", count)
	},
}

var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Aliases: []string{"m"},
	Short:   "Run the database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		dburl := util.MustEnv(Env_DATABASE_URL)
		db := log.Must1(db.Connect(dburl))
		log.Must(db.Migrate())
	},
}
