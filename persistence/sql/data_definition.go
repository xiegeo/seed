package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

// AddDomain adds the domain and readies the database to serve this domain.
func (db *DB) AddDomain(ctx context.Context, d *seed.Domain) error {
	if _, ok := db.domains[d.Name]; ok {
		return seederrors.NewCodeNameExistsError(d.Name, seederrors.ThingTypeDomain, "")
	}
	domainInfo, err := db.domainInfoFromDomain(ctx, d)
	if err != nil {
		return err
	}
	db.domains[d.Name] = domainInfo
	return nil
}

func (db *DB) generateTableName(ctx context.Context, d *seed.Domain, ob *seed.Object) (string, error) {
	return fmt.Sprintf("%s_ob_%s", d.Name, ob.Name), nil
}

func (db *DB) createTable(txc txContext, d *seed.Domain, ob *seed.Object) (err error) {
	tableName, err := db.generateTableName(txc, d, ob)
	if err != nil {
		return err
	}
	var preHooks, postHooks []func(tx UseTx) error
	var fieldDefinitions []string
	var tableOption TableOption
	if db.option.TableDefinition != nil {
		td, err := db.option.TableDefinition(ob)
		if err != nil {
			return err
		}
		preHooks = append(preHooks, td.PreHook)
		tableOption = td.Option
		postHooks = append(postHooks, td.PostHook)
	}
	for i := range ob.Fields {
		f := &ob.Fields[i]
		fd, err := db.option.FieldDefinition(f)
		if err != nil {
			return err
		}
		preHooks = append(preHooks, fd.PreHook)
		fieldDefinitions = append(fieldDefinitions, fd.Fields...)
		postHooks = append(postHooks, fd.PostHook)
	}

	for _, h := range preHooks {
		if err := h(txc); err != nil {
			return err
		}
	}
	sql := fmt.Sprintf("CREAT TABLE %s (\n\t%s\n) %s", tableName, strings.Join(fieldDefinitions, ",\n\t"), tableOption)
	_, err = txc.Exec(sql)
	if err != nil {
		return seederrors.WithMessagef(err, `can not create table for object "%s" with SQL "%s"`, ob.Name, sql)
	}
	for _, h := range postHooks {
		if err := h(txc); err != nil {
			return err
		}
	}
	return nil
}
