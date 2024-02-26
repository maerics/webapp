package cmd

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"
	"webapp/db"

	"github.com/iancoleman/strcase"
	log "github.com/maerics/golog"
	util "github.com/maerics/goutil"
	cobra "github.com/spf13/cobra"
)

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
				"Columns": util.Map(tableInfos, func(ti tableInfo) map[string]any {
					return map[string]any{
						"Name":     strcase.ToCamel(ti.Name),
						"Type":     typeFor(ti.Type),
						"Nullable": ti.Nullable,
						"Annotations": "`" + strings.Join(util.Map([]string{"json", "db"}, func(key string) string {
							return fmt.Sprintf("%s:%q", key, ti.Name)
						}), " ") + "`",
					}
				}),
			}))
			// fmt.Println(modelGoCode.String())
			gocode := must1(format.Source(modelGoCode.Bytes()))
			must(os.WriteFile(filename, gocode, os.FileMode(0o644)))
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

const modelGoCodeTemplate = `package models

import "time"

type {{ .Name }} struct { {{- range .Columns }}
	{{ .Name }} {{if .Nullable}}*{{end}}{{ .Type }} {{ .Annotations }}{{end}}
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
