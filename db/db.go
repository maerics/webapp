package db

import (
	"net/url"

	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	log "github.com/maerics/golog"

	// Other supported database drivers can go here
	_ "github.com/mattn/go-sqlite3"
)

type DB struct{ *sqlx.DB }

func Connect(dburl string) (*DB, error) {
	u, err := url.Parse(dburl)
	if err != nil {
		return nil, err
	}

	driver := u.Scheme
	connstr, connstrSafe := dburl, u.Redacted()
	switch driver {
	case "sqlite3":
		connstr, connstrSafe = u.Host, u.Host
	}
	log.Debugf("database driver=%q, connection=%q", driver, connstrSafe)

	sqlxdb, err := sqlx.Connect(driver, connstr)
	if err != nil {
		return nil, err
	}

	sqlxdb.MapperFunc(strcase.ToSnake)

	log.Printf("connected to database at %q", u.Redacted())
	return &DB{sqlxdb}, nil
}
