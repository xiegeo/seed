package sqldb_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/demo/testdomain"
	"github.com/xiegeo/seed/persistence/sqldb"

	_ "github.com/mattn/go-sqlite3"
)

func TestInsertsSqlite3(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rawDB, err := sql.Open("sqlite3", ":memory:")
	// rawDB, err := sql.Open("sqlite3", "testdata.sqlite3") // use this to inspect the database after
	require.NoError(t, err)
	defer func() {
		require.NoError(t, rawDB.Close())
	}()
	db, err := sqldb.New(rawDB, sqldb.Sqlite)
	require.NoError(t, err)

	domain := testdomain.DomainLevel0base()
	msg := "sub test is required to succeed"
	require.True(t, testAddDomain(t, ctx, db, domain), msg)
	require.True(t, testInserts(t, ctx, db), msg)
}

func testAddDomain(t *testing.T, ctx context.Context, db *sqldb.DB, domain seed.DomainGetter) (success bool) {
	t.Run("create "+string(domain.GetName()), func(t *testing.T) {
		require.NoError(t, db.AddDomain(ctx, domain))
		success = true
	})
	return success
}

func testInserts(t *testing.T, ctx context.Context, db *sqldb.DB) (success bool) {
	domain := db.DefaultDomain()
	t.Run("inserts "+string(domain.GetName()), func(t *testing.T) {
		success = true
	})
	return success
}
