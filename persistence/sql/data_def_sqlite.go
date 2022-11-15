package sql

import (
	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

func sqliteTableDefinition(*seed.Object) (ObjectDefinition, error) {
	return ObjectDefinition{
		Option: "STRICT",
	}, nil
}

func sqliteHelperTableDefinition(*seed.Object) (ObjectDefinition, error) {
	return ObjectDefinition{
		Option: "STRICT, WITHOUT ROWID",
	}, nil
}

func sqliteFieldType(f *seed.Field) (string, error) {
	switch f.FieldType {
	case seed.String:
		return "TEXT", nil
	case seed.Binary:
		return "BLOB", nil
	case seed.Boolean:
		return "INTEGER", nil // 0: false, 1: true
	case seed.TimeStamp:
		return "TEXT", nil
	case seed.Integer:
		return "INTEGER", nil
	case seed.Real:
		return "REAL", nil
	default:
		return "", seederrors.NewFieldNotSupportedError(f.FieldType.String(), f.Name)
	}
}
