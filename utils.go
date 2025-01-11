package txmanager

import (
	"context"
)

func getTxFromCtx(ctx context.Context) Tx {
	tx, ok := ctx.Value(txContextKey).(Tx)
	if !ok {
		return nil
	}
	return tx
}
