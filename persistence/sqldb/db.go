package sqldb

import (
	"database/sql"
	"regexp"
	"strconv"

	"github.com/xiegeo/must"

	"github.com/xiegeo/seed"
)

// DB supports reading and writing data values to a sql database
type DB struct {
	sqldb *sql.DB

	option *DBOption

	defaultDomain *domainInfo
	domains       map[seed.CodeName]*domainInfo
}

type DBOption struct {
	TranslateStatement func(string) string // translate ? in statements to a format the database understands.
	ColumnFeatures                         // describes column type support

	// PrimaryKeys defines the primary keys for a table, it can return one of three value types.
	//
	//  - int: the index of identity to use as primary key. (-1 for do not use any)
	//  - ColumnType: the type definition of auto increment primary key.
	//    (empty string if auto increment should not be used)
	//  - error: something went wrong.
	PrimaryKeys func(ob seed.FieldGroupGetter) (int, ColumnType, error)

	TableOption         string // Default table option
	TableOptionNoAutoID string // The table option to use in addition if PrimaryKeys does not use auto increment
}

func newDefaultOption() *DBOption {
	return &DBOption{
		TranslateStatement: func(s string) string { return s },
		PrimaryKeys: func(ob seed.FieldGroupGetter) (int, ColumnType, error) {
			return -1, "INTEGER", nil
		},
	}
}

func (op *DBOption) getPrimaryKeys(ob seed.FieldGroupGetter) (int, ColumnType, error) {
	index, columnType, err := op.PrimaryKeys(ob)
	if err != nil {
		return -1, "", err
	}
	if index >= 0 {
		id := ob.GetIdentities()[index]
		for _, fieldName := range id.Fields {
			field, _ := must.B2(ob.GetFields().Get(fieldName))(must.Any, true)
			if err = CheckPrimaryKey(field); err != nil {
				return -1, "", err
			}
		}
	}
	return index, columnType, err
}

func New(sqldb *sql.DB, options ...func(*DBOption) error) (*DB, error) {
	db := DB{
		sqldb:   sqldb,
		option:  newDefaultOption(),
		domains: make(map[seed.CodeName]*domainInfo),
	}
	for _, op := range options {
		err := op(db.option)
		if err != nil {
			return nil, err
		}
	}
	return &db, nil
}

func (db *DB) DefaultDomain() seed.DomainGetter {
	return db.defaultDomain
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
