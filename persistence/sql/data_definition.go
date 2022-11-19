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
	// todo
	return null
}
