package txmanager

import (
	"context"
	"errors"
	"fmt"
	"log"
)

// contextKey используется для хранения данных в контексте
type contextKey struct {
	name string
}

type Rows interface {
	Scan(dest ...interface{}) error
	Next() bool
	Close()
}

// Tx - interface for transaction
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Adapter представляет интерфейс для начала транзакций
type Adapter interface {
	Begin(ctx context.Context) (Tx, error)
	Query(ctx context.Context, sql string, args ...interface{}) (Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) error
}

// ContextManager управляет транзакциями на уровне сервисного слоя
type ContextManager interface {
	Start(ctx context.Context) (context.Context, error)
	Finish(ctx context.Context, err error) error
	Query(ctx context.Context, sql string, args ...interface{}) (Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) error
}

// contextManager — конкретная реализация ContextManager
type contextManager struct {
	adapter Adapter
	logger  *log.Logger
}

func New(opts ...Option) ContextManager {
	m := &contextManager{
		logger: log.Default(),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

var txContextKey = &contextKey{"tx"}

func (cm *contextManager) Start(ctx context.Context) (context.Context, error) {
	tx, err := cm.adapter.Begin(ctx)
	if err != nil {
		return nil, errors.New("cannot start transaction: " + err.Error())
	}
	return context.WithValue(ctx, txContextKey, tx), nil
}

func (cm *contextManager) Finish(ctx context.Context, err error) error {
	tx := getTxFromCtx(ctx)
	if tx == nil {
		cm.logger.Println("no transaction found in context")
		return nil
	}

	if err != nil {
		rbErr := cm.rollback(ctx, tx, err)
		if rbErr != nil {
			cm.logger.Printf("failed to rollback transaction: %v", rbErr)
		}
		return nil
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}
	cm.logger.Println("transaction committed successfully")
	return nil
}

func (cm *contextManager) rollback(ctx context.Context, tx Tx, err error) error {
	cm.logger.Printf("error occurred, rolling back transaction: %v", err)
	if rbErr := tx.Rollback(ctx); rbErr != nil {
		return fmt.Errorf("failed to rollback transaction: %w", rbErr)
	}
	cm.logger.Println("transaction rolled back successfully")
	return err
}

func (cm *contextManager) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	return cm.adapter.Query(ctx, sql, args...)
}

func (cm *contextManager) Exec(ctx context.Context, sql string, args ...interface{}) error {
	return cm.adapter.Exec(ctx, sql, args...)
}
