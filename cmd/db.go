package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"
	"webapp/db"
	"webapp/models"
	"webapp/web"

	"github.com/iancoleman/strcase"
	log "github.com/maerics/golog"
	util "github.com/maerics/goutil"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dbCmd)

	dbCmd.AddCommand(selectCmd)
	dbCmd.AddCommand(executeCmd)
	dbCmd.AddCommand(generateCmd)
	dbCmd.AddCommand(migrateCmd)
	dbCmd.AddCommand(seedCmd)

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
		must1(io.Copy(buf, os.Stdin))
		query := strings.TrimSpace(buf.String())

		// Execute the query.
		dburl := util.MustEnv(Env_DATABASE_URL)
		db := must1(db.Connect(dburl))
		tx := db.MustBegin()
		log.Printf("executing query:\n\n    %v\n\n", query)
		t0 := time.Now()

		// Print the affected rows and last insert id, if applicable.
		result := must1(tx.Exec(query))
		if n, err := result.RowsAffected(); err == nil {
			log.Printf("query affected %v row(s)", n)
		}
		if n, err := result.LastInsertId(); err == nil {
			log.Printf("last insert id=%v", n)
		}

		// Commit or rollback.
		if optDbExecuteCommit {
			must(tx.Commit())
			log.Printf("committed transaction in %v", time.Since(t0))
		} else {
			must(tx.Rollback())
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
		must1(io.Copy(buf, os.Stdin))
		query := strings.TrimSpace(buf.String())

		// Execute the query.
		dburl := util.MustEnv(Env_DATABASE_URL)
		db := must1(db.Connect(dburl))
		log.Printf("executing query:\n\n    %v\n\n", query)
		rows := must1(db.Query(query))

		// Inspect the result set column types.
		columns := must1(rows.Columns())
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
			must(w.Write(columns))
		}

		// Print to stdout.
		count := 0
		enc := json.NewEncoder(os.Stdout)
		for rows.Next() {
			count++
			must(rows.Scan(scanArgs...))

			// Print the JSON or CSV result.
			if outputcsv {
				row := make([]string, len(values))
				for i, v := range values {
					if v != nil {
						row[i] = fmt.Sprintf("%v", v)
					}
				}
				must(w.Write(row))
			} else {
				must(enc.Encode(util.OrderedJsonObj{Keys: columns, Values: values}))
			}
		}

		if outputcsv {
			w.Flush()
			must(w.Error())
		}
		log.Printf("query returned %v row(s)", count)
	},
}

var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen", "g"},
	Short:   "Generate models from the existing database structure", // TODO: also api crud routes?
	Run: func(cmd *cobra.Command, args []string) {
		dburl := util.MustEnv(Env_DATABASE_URL)
		db := must1(db.Connect(dburl))
		// TODO: cowardly refuse for some reason?

		// Select the public table names.
		var tableNames []string
		must(db.Select(&tableNames, `SELECT table_name FROM information_schema.tables WHERE table_schema='public'`))

		// Select the column info for each table.
		for _, tableName := range tableNames {
			var tableInfos []tableInfo
			query := fmt.Sprintf(`SELECT column_name, data_type, (is_nullable='YES') as nullable FROM information_schema.columns WHERE table_name = '%s' ORDER BY ordinal_position`, tableName)
			must(db.Select(&tableInfos, query))

			// Generate a Go struct model for each table.
			filename := fmt.Sprintf("./models/%s.go", filenameFor(tableName))
			var modelGoCode = bytes.Buffer{}
			tmpl := must1(template.New("model").Parse(modelGoCodeTemplate))
			must(tmpl.Execute(&modelGoCode, map[string]any{
				"Name": strcase.ToCamel(filenameFor(tableName)),
				"Columns": Map(tableInfos, func(ti tableInfo) map[string]any {
					return map[string]any{
						"Name":     strcase.ToCamel(ti.Name),
						"Type":     typeFor(ti.Type),
						"Nullable": ti.Nullable,
						"Annotations": "`" + strings.Join(Map([]string{"json", "db"}, func(key string) string {
							return fmt.Sprintf("%s:%q", key, ti.Name)
						}), " ") + "`",
					}
				}),
			}))
			// fmt.Println(modelGoCode.String())
			must(os.WriteFile(filename, modelGoCode.Bytes(), os.FileMode(0o644)))
			log.Printf("wrote %q", filename)
		}
	},
}

type tableInfo struct {
	Name     string `db:"column_name"`
	Type     string `db:"data_type"`
	Nullable bool   `db:"nullable"`
}

func typeFor(postgresType string) string {
	switch postgresType {
	case "integer":
		return "int"
	case "text":
		return "string"
	}

	if strings.HasPrefix(postgresType, "time") {
		return "*time.Time"
	}

	panic(fmt.Errorf("unhandled postgres type %q", postgresType))
}

// TODO: move to util
func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i, t := range ts {
		us[i] = f(t)
	}
	return us
}

const modelGoCodeTemplate = `package models

import "time"

type {{ .Name }} struct { {{range .Columns}}
	{{ .Name }} {{if .Nullable}}*{{end}}{{ .Type }} {{ .Annotations }} {{end}}
}
`

func filenameFor(s string) string {
	if strings.HasSuffix(s, "ies") {
		return s[:len(s)-3] + "y"
	} else if strings.HasSuffix(s, "s") {
		return s[:len(s)-1]
	}
	panic(fmt.Errorf("unhandled table singularlization %q", s))
}

var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Aliases: []string{"m"},
	Short:   "Run the database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		dburl := util.MustEnv(Env_DATABASE_URL)
		db := must1(db.Connect(dburl))
		must(db.Migrate())
	},
}

var seedCmd = &cobra.Command{
	Use:     "seed",
	Aliases: []string{"sd"},
	Short:   "Seed the database with example data",
	Run: func(cmd *cobra.Command, args []string) {
		dburl := util.MustEnv(Env_DATABASE_URL)
		db := must1(db.Connect(dburl))
		must(db.Migrate())
		password := "secret"
		user := models.User{
			Email:    "hello@example.com",
			Password: password,
		}
		log.Printf("creating user %v:%v", user.Email, user.Password)
		query := "INSERT INTO users (email, password) VALUES ($1, $2)"
		if _, err := db.Exec(query, user.Email, web.BCryptPassword(user.Password)); err != nil {
			log.Fatalf("%v", err)
		}
		log.Printf("successfully seeded database")
	},
}
