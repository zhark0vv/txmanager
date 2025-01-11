package manager

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgxAdapter(t *testing.T) {
	ctx := context.Background()

	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(ctx)

	m := New(WithPgxAdapter(mock))

	t.Run("Query", func(t *testing.T) {
		mock.ExpectBegin()
		tx, err := m.Start(ctx)
		require.NoError(t, err)
		require.NotNil(t, tx)

		mock.ExpectQuery("SELECT 1").WillReturnRows(mock.NewRows([]string{"number"}).AddRow(1))
		rows, err := m.Query(ctx, "SELECT 1")
		require.NoError(t, err)
		defer rows.Close()

		require.True(t, rows.Next())

		var number int
		err = rows.Scan(&number)
		require.NoError(t, err)
		require.Equal(t, 1, number)
	})

	t.Run("Exec", func(t *testing.T) {
		mock.ExpectBegin()
		ctx, err = m.Start(ctx)
		require.NoError(t, err)

		mock.ExpectExec(`UPDATE users SET name = \$1 WHERE id = \$2`).
			WithArgs("John Doe", 1).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := m.Exec(ctx, "UPDATE users SET name = $1 WHERE id = $2", "John Doe", 1)
		require.NoError(t, err)
	})

	require.NoError(t, mock.ExpectationsWereMet())

	t.Run("Commit", func(t *testing.T) {
		mock.ExpectBegin()
		ctx, err = m.Start(ctx)
		require.NoError(t, err)

		mock.ExpectCommit()

		err := m.Finish(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("Rollback", func(t *testing.T) {
		mock.ExpectBegin()
		ctx, err = m.Start(ctx)
		require.NoError(t, err)

		mock.ExpectRollback()

		err = m.Finish(ctx, assert.AnError)
		require.NoError(t, err)
	})
}
