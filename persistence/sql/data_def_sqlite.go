package sql

import "github.com/xiegeo/seed"

func sqliteTableDefinition(*seed.Object) (ObjectDefinition, error) {
	return ObjectDefinition{
		Option: "STRICT, WITHOUT ROWID",
	}, nil
}

func sqliteStringField(seed.StringSetting) (string, error) {
}
