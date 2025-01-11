package txmanager

import (
	"context"
	"database/sql"
	"fmt"
)

type SqlRows struct {
	rows *sql.Rows
}

func (r *SqlRows) Close() {
	r.rows.Close()
}

func (r *SqlRows) Next() bool {
	return r.rows.Next()
}

func (r *SqlRows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

type SqlTx struct {
	tx *sql.Tx
}

func (t *SqlTx) Commit(_ context.Context) error {
	return t.tx.Commit()
}

func (t *SqlTx) Rollback(_ context.Context) error {
	return t.tx.Rollback()
}

type SqlAdapter struct {
	db *sql.DB
}

func newSQL(db *sql.DB) *SqlAdapter {
	return &SqlAdapter{db: db}
}

func (a *SqlAdapter) Begin(ctx context.Context) (Tx, error) {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}
	return &SqlTx{tx: tx}, nil
}

func (a *SqlAdapter) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	tx := getTxFromCtx(ctx)
	if tx != nil {
		if sqlTx, ok := tx.(*SqlTx); ok {
			rows, err := sqlTx.tx.QueryContext(ctx, sql, args...)
			if err != nil {
				return nil, fmt.Errorf("cannot query in transaction: %w", err)
			}
			return &SqlRows{rows: rows}, nil
		}
	}

	rows, err := a.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("cannot query: %w", err)
	}
	return &SqlRows{rows: rows}, nil
}

func (a *SqlAdapter) Exec(ctx context.Context, sql string, args ...interface{}) error {
	tx := getTxFromCtx(ctx)
	if tx != nil {
		if sqlTx, ok := tx.(*SqlTx); ok {
			_, err := sqlTx.tx.ExecContext(ctx, sql, args...)
			if err != nil {
				return fmt.Errorf("cannot exec in transaction: %w", err)
			}
			return nil
		}
	}

	_, err := a.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("cannot exec: %w", err)
	}
	return nil
}
