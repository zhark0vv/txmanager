package txmanager

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqlAdapter(t *testing.T) {
	ctx := context.Background()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := New(WithSQLAdapter(db))

	t.Run("Query", func(t *testing.T) {
		mock.ExpectBegin()
		ctx, err = manager.Start(ctx)
		require.NoError(t, err)

		mock.ExpectQuery("SELECT 1").
			WillReturnRows(sqlmock.NewRows([]string{"number"}).
				AddRow(1))

		rows, err := manager.Query(ctx, "SELECT 1")
		require.NoError(t, err)
		defer rows.Close()

		require.True(t, rows.Next())

		var number int
		err = rows.Scan(&number)
		require.NoError(t, err)
		require.Equal(t, 1, number)

		mock.ExpectCommit()
		err = manager.Finish(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("Exec", func(t *testing.T) {
		mock.ExpectBegin()
		ctx, err := manager.Start(ctx)
		require.NoError(t, err)

		mock.ExpectExec("UPDATE users SET name = \\? WHERE id = \\?").
			WithArgs("John Doe", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = manager.Exec(ctx, "UPDATE users SET name = ? WHERE id = ?", "John Doe", 1)
		require.NoError(t, err)

		mock.ExpectCommit()
		err = manager.Finish(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("Commit", func(t *testing.T) {
		mock.ExpectBegin()
		ctx, err := manager.Start(ctx)
		require.NoError(t, err)

		mock.ExpectCommit()
		err = manager.Finish(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("Rollback", func(t *testing.T) {
		mock.ExpectBegin()
		ctx, err := manager.Start(ctx)
		require.NoError(t, err)

		mock.ExpectRollback()
		err = manager.Finish(ctx, assert.AnError)
		require.NoError(t, err)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
