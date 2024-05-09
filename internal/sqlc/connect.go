package sqlc

import (
	"context"
	"database/sql"
	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var ddl string

func Connect(create bool) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "companies.db")
	if err != nil {
		return nil, err
	}

	if create {
		// create tables
		if _, err = db.ExecContext(context.Background(), ddl); err != nil {
			return nil, err
		}
	}

	return db, nil
}
