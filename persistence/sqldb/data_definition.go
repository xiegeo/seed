package sqldb

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
	err = db.doTransaction(ctx, func(txc txContext) error {
		return createDomainTx(txc, domainInfo)
	})
	if err != nil {
		return err
	}
	if db.defaultDomain.Name == "" {
		db.defaultDomain = domainInfo // the first domain added is the default domain
	}
	db.domains[d.Name] = domainInfo
	return nil
}

func (db *DB) generateTableName(ctx context.Context, d *seed.Domain, ob *seed.Object) (string, error) {
	err := ctx.Err() // hide unparam lint for future usage
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_ob_%s", d.Name, ob.Name), nil
}

func createDomainTx(txc txContext, domain domainInfo) error {
	for _, obj := range domain.objectMap {
		err := createObjectTx(txc, obj)
		if err != nil {
			return seederrors.WithMessagef(err, "in object %s", obj.Name)
		}
	}
	return nil
}

func createObjectTx(txc txContext, obj *objectInfo) error {
	err := createTableTx(txc, obj.mainTable)
	if err != nil {
		return seederrors.WithMessagef(err, "in main table %s", obj.mainTable.Name)
	}
	for _, table := range obj.helperTables {
		err = createTableTx(txc, table)
		if err != nil {
			return seederrors.WithMessagef(err, "in helper table %s", table.Name)
		}
	}
	return nil
}

func createTableTx(txc txContext, table Table) error {
	sql := &strings.Builder{}
	_, err := CreateTable(table).WriteTo(sql)
	if err != nil {
		return err
	}
	_, err = txc.Exec(sql.String())
	if err != nil {
		return err
	}
	return nil
}
