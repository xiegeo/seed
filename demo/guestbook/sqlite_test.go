package main

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/xiegeo/seed/persistence/sqldb"

	_ "github.com/mattn/go-sqlite3"
)

func TestAddDomain(t *testing.T) {
	rawDB, err := sql.Open("sqlite3", ":memory:")
	// rawDB, err := sql.Open("sqlite3", "testdata") // use this to inspect the database after
	require.NoError(t, err)
	db, err := sqldb.New(rawDB, sqldb.Sqlite)
	require.NoError(t, err)
	guestbook := Domain()
	err = db.AddDomain(context.Background(), &guestbook)
	require.NoError(t, err)
}
