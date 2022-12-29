package db

import (
	"embed"
	"fmt"
	"path"
	"sort"

	log "github.com/maerics/golog"
)

// Migration files MUST be idempotent!
const MigrationsDirname = "migrations"

//go:embed migrations/*
var migrationsfs embed.FS

func (db *DB) Migrate() error {
	// Load the migration files and sort by name.
	entries, err := migrationsfs.ReadDir(MigrationsDirname)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	// Run each migration in sequence.
	for _, entry := range entries {
		log.Debugf("running migration %q", entry.Name())
		filename := path.Join(MigrationsDirname, entry.Name())
		query, err := migrationsfs.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("reading %q: %w",
				path.Join(MigrationsDirname, entry.Name()), err)
		}
		if _, err := db.DB.Exec(string(query)); err != nil {
			return fmt.Errorf("executing %q: %w",
				path.Join(MigrationsDirname, entry.Name()), err)
		}
	}
	log.Printf("successfully ran %v database migration(s)", len(entries))

	return nil
}
