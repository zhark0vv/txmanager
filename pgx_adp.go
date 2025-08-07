package txmanager

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5"
)

type PgxConn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
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

func (a *PgxAdapter) Select(ctx context.Context, dest interface{}, sql string, args ...interface{}) error {
	tx := getTxFromCtx(ctx)
	if tx != nil {
		if pgxTx, ok := tx.(*PgxTx); ok {
			if err := pgxscan.Select(ctx, pgxTx.raw, dest, sql, args...); err != nil {
				return fmt.Errorf("pgxscan select in transaction: %w", err)
			}
			return nil
		}
	}

	if err := pgxscan.Select(ctx, a.conn, dest, sql, args...); err != nil {
		return fmt.Errorf("pgxscan select: %w", err)
	}
	return nil
}

func (a *PgxAdapter) Get(ctx context.Context, dest interface{}, sql string, args ...interface{}) error {
	tx := getTxFromCtx(ctx)
	if tx != nil {
		if pgxTx, ok := tx.(*PgxTx); ok {
			if err := pgxscan.Get(ctx, pgxTx.raw, dest, sql, args...); err != nil {
				return fmt.Errorf("pgxscan get in transaction: %w", err)
			}
			return nil
		}
	}

	if err := pgxscan.Get(ctx, a.conn, dest, sql, args...); err != nil {
		return fmt.Errorf("pgxscan get: %w", err)
	}
	return nil
}
