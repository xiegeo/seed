package sqldb

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

	defaultDomain domainInfo
	domains       map[seed.CodeName]domainInfo
}

type DBOption struct {
	TranslateStatement func(string) string // translate ? in statements to a format the database understands.
	ColumnFeatures                         // describes column type support
	TableOption
}

func newDefaultOption() *DBOption {
	return &DBOption{
		TranslateStatement: func(s string) string { return s },
	}
}

func New(sqldb *sql.DB, options ...func(*DBOption) error) (*DB, error) {
	db := DB{
		sqldb:   sqldb,
		option:  newDefaultOption(),
		domains: make(map[seed.CodeName]domainInfo),
	}
	for _, op := range options {
		err := op(db.option)
		if err != nil {
			return nil, err
		}
	}
	return &db, nil
}

var qrx = regexp.MustCompile(`\?`)

// generateTranslateStatementFunc converts "?" characters to $1, $2, $n on postgres, :1, :2, :n on Oracle
// case "postgres", "pgx":
//
//	db.TranslateStatement = generateTranslateStatementFunc("$")
//
// case "oracle", "goracle":
//
//	db.TranslateStatement = generateTranslateStatementFunc(":")
func generateTranslateStatementFunc(pref string) func(string) string {
	n := 0
	return func(sql string) string {
		return qrx.ReplaceAllStringFunc(sql, func(string) string {
			n++
			return pref + strconv.Itoa(n)
		})
	}
}
