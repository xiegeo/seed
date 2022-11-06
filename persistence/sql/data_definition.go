package sql

import (
	"context"
	"fmt"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

// AddDomain adds the domain and readies the database to serve this domain.
func (db *DB) AddDomain(ctx context.Context, d *seed.Domain) error {
	if _, ok := db.domains[d.Name]; ok {
		return seederrors.NewCodeNameExistsError(d.Name, seederrors.ThingTypeDomain, "")
	}
	db.domains[d.Name] = domainInfoFromDomain(*d)
	return nil
}

func (db *DB) generateTableName(ctx context.Context, d *seed.Domain, ob *seed.Object) (string, error) {
	return fmt.Sprint("%s_ob_%s", d.Name, ob.Name), nil
}

func (db *DB) createTable(txc txContext, d *seed.Domain, ob *seed.Object) (err error) {
	tableName, err := db.generateTableName(txc, d, ob)
	if err != nil {
		return err
	}
	var preHooks, postHooks []func(tx UseTx) error
	var fieldDefinitions []string
	for _, field := range ob.Fields {
		db.
	}
}

// FieldsToSQL implements GetDefinition
type FieldsToSQL struct {
	I18N   *FieldsToSQL
	ByType [seed.FieldTypeMax + 1]func(*seed.Field) (FieldDefinition, error)
}

func (toSQL *FieldsToSQL) GetDefinition(f *seed.Field) (FieldDefinition, error) {
	typed := toSQL.ByType[f.FieldType]
	if typed == nil {
		return  FieldDefinition{}, seederrors.NewFieldNotSupportedError(f.FieldType.String(),f.Name)
	}

}