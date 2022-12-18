package sqldb_test

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xiegeo/must"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/demo/testdomain"
	"github.com/xiegeo/seed/persistence/sqldb"
	"github.com/xiegeo/seed/seedfake"

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
	require.NoError(t, domain.Objects.AddValue(testdomain.ObjLevel0Identities()))
	require.NoError(t, domain.Objects.AddValue(testdomain.ObjLevel0List()))
	msg := "sub test is required to succeed"
	require.True(t, testAddDomain(t, ctx, db, domain), msg)
	counter := successCounter{
		"TestInsertsSqlite3/inserts_test_level_0_base/level_0":      85,
		"TestInsertsSqlite3/inserts_test_level_0_base/level_0_ids":  58,
		"TestInsertsSqlite3/inserts_test_level_0_base/level_0_list": 0, // not supported yet
	}
	require.True(t, testInserts(t, ctx, db, counter), msg)
}

func testAddDomain(t *testing.T, ctx context.Context, db *sqldb.DB, domain seed.DomainGetter) (success bool) {
	t.Run("create "+string(domain.GetName()), func(t *testing.T) {
		require.NoError(t, db.AddDomain(ctx, domain))
		success = true
	})
	return success
}

type successCounter map[string]int

func testInserts(t *testing.T, ctx context.Context, db *sqldb.DB, counter successCounter) (success bool) {
	success = true
	domain := db.DefaultDomain()
	t.Run("inserts "+string(domain.GetName()), func(t *testing.T) {
		gen := seedfake.NewValueGen(seedfake.NewMinMaxFlat(rand.NewSource(0), 1, 1, 5))
		require.NoError(t, domain.GetObjects().RangeLogical(func(obName seed.CodeName, ob seed.ObjectGetter) error {
			t.Run(string(obName), func(t *testing.T) {
				var errs []error
				total := 100
				for i := total; i > 0; i-- {
					err := db.InsertObjects(ctx, map[seed.CodeName]any{
						obName: must.VT(gen.ValuesForObject(ob, 1))(t),
					})
					if err != nil {
						errs = append(errs, err)
					}
				}
				added := total - len(errs)
				if assert.Equal(t, counter[t.Name()], added, "successes not eq") { // allow some rows to fail quietly on random data because of constraints
					t.Logf("added %s:%d expected map[error:count]=%v", obName, added, countedStringSet(errs))
				} else {
					success = false
					t.Errorf("not enough rows inserted for %s, expected=%d successes=%d, map[error:count]=%v", obName, counter[t.Name()], added, countedStringSet(errs))
				}
			})
			return nil
		}))
	})
	return success
}

func countedStringSet[T any](vs []T) map[string]int {
	out := make(map[string]int)
	for _, v := range vs {
		s := fmt.Sprint(v)
		out[s]++
	}
	return out
}
