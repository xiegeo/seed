package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/persistence/sqldb"

	_ "github.com/mattn/go-sqlite3"
)

func timeWithMinute(t *testing.T, value string) time.Time {
	out, err := time.Parse("2006-01-02T15:04", value)
	require.NoError(t, err)
	return out
}

func TestAddDomain(t *testing.T) {
	ctx := context.TODO()
	rawDB, err := sql.Open("sqlite3", ":memory:")
	// rawDB, err := sql.Open("sqlite3", "testdata.sqlite3") // use this to inspect the database after
	require.NoError(t, err)
	defer func() {
		require.NoError(t, rawDB.Close())
	}()
	db, err := sqldb.New(rawDB, sqldb.Sqlite)
	require.NoError(t, err)
	guestbook := Domain()
	err = db.AddDomain(ctx, &guestbook)
	require.NoError(t, err)
	err = db.InsertObjects(ctx, map[seed.CodeName]any{
		Event().Name: []map[seed.CodeName]any{{
			StartTimeField().Name:         timeWithMinute(t, "2006-01-02T15:00"),
			EndTimeField().Name:           timeWithMinute(t, "2006-01-02T15:00"),
			PublishField().Name:           true,
			MaxNumberOfGuestsField().Name: 100,
		}},
	})
	require.NoError(t, err)
}
