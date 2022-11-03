package sql

import (
	"database/sql"
	"regexp"
	"strconv"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

// DB supports reading and writing data values to a sql database
type DB struct {
	sqldb *sql.DB

	translateStatment func(string) string // translate ? in statements to a format the database understands.

	domains map[seed.CodeName]*seed.Domain
}

func Open(driverName, dataSourceName string) (*DB, error) {
	sqldb, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, seederrors.WithMessagef(err, "open sql database driverName=%s", driverName)
	}
	return New(driverName, sqldb)
}

func New(driverFamily string, sqldb *sql.DB) (*DB, error) {
	db := DB{
		sqldb: sqldb,
	}

	switch driverFamily { // not fully tested, idea taken from github.com/bradfitz/go-sql-test
	case "postgres", "pgx":
		db.translateStatment = generateTranslateStatementFunc("$")
	case "oracle", "goracle":
		db.translateStatment = generateTranslateStatementFunc(":")
	default:
		db.translateStatment = func(s string) string { return s }
	}

	return &db, nil
}

var qrx = regexp.MustCompile(`\?`)

// generateTranslateStatementFunc converts "?" characters to $1, $2, $n on postgres, :1, :2, :n on Oracle
func generateTranslateStatementFunc(pref string) func(string) string {
	n := 0
	return func(sql string) string {
		return qrx.ReplaceAllStringFunc(sql, func(string) string {
			n++
			return pref + strconv.Itoa(n)
		})
	}
}
