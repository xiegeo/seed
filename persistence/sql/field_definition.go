package sql

import "github.com/xiegeo/seed"

// generateFieldDefinition support a seed defined field with 0 to many columns and 0 to many tables.
// of both columns and tables are none, an error must be returned.
func (db *DB) generateFieldDefinition(f *seed.Field) ([]Column, []Table, error) {
	cf, found := db.option.ColumnFeatures.Match(f)
	if found {
	}
}
