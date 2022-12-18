package sqldb

import (
	"context"
	"strings"

	"github.com/xiegeo/seed"
	"github.com/xiegeo/seed/seederrors"
)

// AddDomain adds the domain and readies the database to serve this domain.
func (db *DB) AddDomain(ctx context.Context, d seed.DomainGetter) error {
	if _, ok := db.domains[d.GetName()]; ok {
		return seederrors.NewCodeNameExistsError(d.GetName(), seederrors.ThingTypeDomain, "")
	}
	domainInfo, err := db.domainInfoFromDomain(d)
	if err != nil {
		return err
	}
	err = db.doTransaction(ctx, func(txc txContext) error {
		return createDomainTx(txc, domainInfo)
	})
	if err != nil {
		return err
	}
	if db.defaultDomain == nil {
		db.defaultDomain = domainInfo // the first domain added is the default domain
	}
	db.domains[d.GetName()] = domainInfo
	return nil
}

func createDomainTx(txc txContext, domain *domainInfo) error {
	return domain.objectMap.RangeLogical(func(cn seed.CodeName, obj *objectInfo) error {
		err := createObjectTx(txc, obj)
		if err != nil {
			return seederrors.WithMessagef(err, "in object %s", obj.Name)
		}
		return nil
	})
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

func createTableTx(txc txContext, table *Table) error {
	sql := &strings.Builder{}
	_, err := MakeCreateTable(table).WriteTo(sql)
	if err != nil {
		return err
	}
	_, err = txc.Exec(sql.String())
	if err != nil {
		return err
	}
	return nil
}
