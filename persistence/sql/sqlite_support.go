package sql

import (
	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

const (
	sqliteTableDefinition       = "STRICT"
	sqliteHelperTableDefinition = "STRICT, WITHOUT ROWID"
)

func sqliteColumnType(f *seed.Field, c *Column) error {
	switch f.FieldType {
	case seed.String:
		c.Type = "TEXT"
	case seed.Binary:
		c.Type = "BLOB"
	case seed.Boolean:
		c.Type = "INTEGER" // 0: false, 1: true
	case seed.TimeStamp:
		setting, err := seed.GetFieldTypeSetting[seed.TimeStampSetting](f)
		if err != nil {
			return err
		}
		if setting.WithTimeZone {
			return seederrors.NewFieldNotSupportedError(f.FieldType.String(), f.Name, "WithTimeZone")
		}
		c.Type = "TEXT"
	case seed.Integer:
		c.Type = "INTEGER"
	case seed.Real:
		c.Type = "REAL" // float64
	default:
		return seederrors.NewFieldNotSupportedError(f.FieldType.String(), f.Name)
	}
	return nil
}
