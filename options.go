package txmanager

import (
	"log"
)

type Option func(*contextManager)

func WithLogging(logger *log.Logger) Option {
	return func(cm *contextManager) {
		cm.logger = logger
	}
}

func WithPgxAdapter(conn PgxConn) Option {
	return func(cm *contextManager) {
		cm.adapter = newPgx(conn)
	}
}
