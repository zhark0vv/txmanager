package manager

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type PgxConn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

type PgxTx struct {
	raw pgx.Tx
}

func (t *PgxTx) Commit(ctx context.Context) error {
	return t.raw.Commit(ctx)
}

func (t *PgxTx) Rollback(ctx context.Context) error {
	return t.raw.Rollback(ctx)
}

type PgxAdapter struct {
	conn PgxConn
}

func newPgx(conn PgxConn) *PgxAdapter {
	return &PgxAdapter{conn: conn}
}

func (a *PgxAdapter) Begin(ctx context.Context) (Tx, error) {
	tx, err := a.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}

	return &PgxTx{raw: tx}, nil
}

func (a *PgxAdapter) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	tx := getTxFromCtx(ctx)
	if tx != nil {
		if pgxTx, ok := tx.(*PgxTx); ok {
			rows, err := pgxTx.raw.Query(ctx, sql, args...)
			if err != nil {
				return nil, fmt.Errorf("cannot query in transaction: %w", err)
			}
			return rows, nil
		}
	}

	rows, err := a.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("cannot query: %w", err)
	}
	return rows, nil
}

func (a *PgxAdapter) Exec(ctx context.Context, sql string, args ...interface{}) error {
	tx := getTxFromCtx(ctx)
	if tx != nil {
		if pgxTx, ok := tx.(*PgxTx); ok {
			_, err := pgxTx.raw.Exec(ctx, sql, args...)
			if err != nil {
				return fmt.Errorf("cannot exec in transaction: %w", err)
			}
			return nil
		}
	}

	_, err := a.conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("cannot exec: %w", err)
	}
	return nil
}
