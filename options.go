package txmanager

import (
	"database/sql"
	"log"
)

type Option func(*contextManager)

func WithLogging(logger *log.Logger) Option {
	return func(cm *contextManager) {
		cm.logger = logger
	}
}

func WithSQLAdapter(conn *sql.DB) Option {
	return func(cm *contextManager) {
		cm.adapter = newSQL(conn)
	}
}

func WithPgxAdapter(conn PgxConn) Option {
	return func(cm *contextManager) {
		cm.adapter = newPgx(conn)
	}
}

func WithAdapter(adapter Adapter) Option {
	return func(cm *contextManager) {
		cm.adapter = adapter
	}
}
