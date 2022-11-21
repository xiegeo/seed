package sqldb

import (
	"context"
	"database/sql"

	"github.com/xiegeo/seed/seederrors"
)

type txContext struct {
	context.Context //nolint:containedctx
	UseTx
}

// UseTx supports everything in *sql.Tx except Commit and Rollback
type UseTx interface {
	// Commit() error
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	// Rollback() error
	Stmt(stmt *sql.Stmt) *sql.Stmt
	StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
}

func (db *DB) doTransaction(ctx context.Context, f func(txc txContext) error) (err error) {
	var success bool
	tx, err := db.sqldb.BeginTx(ctx, nil)
	if err != nil {
		seederrors.WithMessagef(err, "BeginTx in createTable")
	}
	defer func() {
		if !success {
			err = seederrors.CombineErrors(err, tx.Rollback())
		}
	}()

	err = f(txContext{
		Context: ctx,
		UseTx:   tx,
	})
	if err != nil {
		return err
	}

	err = tx.Commit()
	success = err == nil
	return err
}
