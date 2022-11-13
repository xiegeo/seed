package sql

import (
	"database/sql"
	"regexp"
	"strconv"

	"github.com/xiegeo/seed"
)

// DB supports reading and writing data values to a sql database
type DB struct {
	sqldb *sql.DB

	option *DBOption

	domains map[seed.CodeName]domainInfo
}

type DBOption struct {
	TranslateStatement func(string) string // translate ? in statements to a format the database understands.
	FieldDefinition    func(*seed.Field) (FieldDefinition, error)
}

// FieldDefinition list 0 to many Fields.
// Use hooks to apply additional operations outside field definition.
type FieldDefinition struct {
	PreHook  func(tx UseTx) error // dialect specific pre-hock, such as preparing helper tables
	Fields   []string             // dialect specific field definition
	PostHook func(tx UseTx) error // dialect specific post-hook
}

func newDefaultOption() *DBOption {
	return &DBOption{
		TranslateStatement: func(s string) string { return s },
	}
}

func New(sqldb *sql.DB, options ...func(*DBOption) error) (*DB, error) {
	db := DB{
		sqldb:  sqldb,
		option: newDefaultOption(),
	}
	for _, op := range options {
		err := op(db.option)
		if err != nil {
			return nil, err
		}
	}
	return &db, nil
}

/*
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
		db.TranslateStatement = generateTranslateStatementFunc("$")
	case "oracle", "goracle":
		db.TranslateStatement = generateTranslateStatementFunc(":")
	default:
		db.TranslateStatement = func(s string) string { return s }
	}

	return &db, nil
}
*/

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
