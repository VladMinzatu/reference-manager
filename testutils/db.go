package testutils

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

func SetupTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	require.NoError(t, err)

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	goose.SetDialect("sqlite3")
	err = goose.Up(db, "../db/migrations")
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}
