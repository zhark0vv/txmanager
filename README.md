
# txmanager

`txmanager` is a Go library designed to simplify transaction management by providing a flexible and extensible interface for working with different database drivers. It supports custom adapters, allowing developers to integrate it with any database system that implements the required interfaces.

## Installation

To install `txmanager`, run:

```bash
go get github.com/zhark0vv/txmanager
```

## Key Features

- Simple and flexible transaction management.
- Support for custom database adapters.
- Driver-agnostic interface for `Query`, `Exec`, and transaction handling.
- Error-safe transactional workflow with automatic rollback on failure.

## Interfaces

### `Adapter`

The `Adapter` interface defines the methods required for a database adapter:

```go
type Adapter interface {
    Begin(ctx context.Context) (Tx, error)                // Start a new transaction
    Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) // Execute a query
    Exec(ctx context.Context, sql string, args ...interface{}) error          // Execute a command
}
```

### `Tx`

The `Tx` interface represents a transaction:

```go
type Tx interface {
    Commit(ctx context.Context) error   // Commit the transaction
    Rollback(ctx context.Context) error // Rollback the transaction
}
```

### `Rows`

The `Rows` interface abstracts the result set of a query:

```go
type Rows interface {
    Close()                   // Close the result set
    Next() bool               // Advance to the next row
    Scan(dest ...interface{}) error // Scan the current row into destination variables
}
```

## Usage

### Example: Using `txmanager` with the `pgx` Driver

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jackc/pgx/v4"
    "github.com/zhark0vv/txmanager"
)

func main() {
    ctx := context.Background()

    // Connect to the PostgreSQL database
    conn, err := pgx.Connect(ctx, "postgres://user:password@localhost:5432/db")
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer conn.Close(ctx)

    // Initialize transaction manager with pgx adapter
    manager := txmanager.New(txmanager.WithPgxAdapter(conn))

    // Start a transaction
    ctx, err = manager.Start(ctx)
    if err != nil {
        log.Fatalf("Failed to start transaction: %v", err)
    }

    // Ensure transaction is properly finished
    var txErr error
    defer func() {
        if err := manager.Finish(ctx, txErr); err != nil {
            log.Fatalf("Failed to finish transaction: %v", err)
        }
    }()

    // Execute a query within the transaction
    rows, txErr := manager.Query(ctx, "SELECT id, name FROM users")
    if txErr != nil {
        return
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        var name string
        if err := rows.Scan(&id, &name); err != nil {
            log.Fatalf("Failed to scan row: %v", err)
        }
        fmt.Printf("User: %d, %s\n", id, name)
    }

    // Execute a command within the transaction
    txErr = manager.Exec(ctx, "UPDATE users SET active = $1 WHERE id = $2", true, 1)
    if txErr != nil {
        log.Fatalf("Failed to execute command: %v", txErr)
    }
}
```

## Custom Adapters

You can implement your own adapter to work with a custom database or driver. A custom adapter must implement the `Adapter` interface.

### Example: Custom Adapter for a Hypothetical Database

```go
package customadapter

import (
    "context"
    "fmt"
    "pgxctx/pkg/txmanager"
)

// HypotheticalConn represents a connection to a custom database
type HypotheticalConn struct {}

// HypotheticalTx represents a transaction in the custom database
type HypotheticalTx struct {}

func (tx *HypotheticalTx) Commit(ctx context.Context) error {
    fmt.Println("Transaction committed")
    return nil
}

func (tx *HypotheticalTx) Rollback(ctx context.Context) error {
    fmt.Println("Transaction rolled back")
    return nil
}

type CustomAdapter struct {
    conn *HypotheticalConn
}

func NewCustomAdapter(conn *HypotheticalConn) *CustomAdapter {
    return &CustomAdapter{conn: conn}
}

func (a *CustomAdapter) Begin(ctx context.Context) (txmanager.Tx, error) {
    fmt.Println("Transaction started")
    return &HypotheticalTx{}, nil
}

func (a *CustomAdapter) Query(ctx context.Context, sql string, args ...interface{}) (txmanager.Rows, error) {
    fmt.Printf("Executing query: %s\n", sql)
    return nil, nil
}

func (a *CustomAdapter) Exec(ctx context.Context, sql string, args ...interface{}) error {
    fmt.Printf("Executing command: %s\n", sql)
    return nil
}
```

## Advanced Configuration

```go
manager := txmanager.New(
    txmanager.WithAdapter(customAdapter), // Use a custom adapter
)
```

## License

`txmanager` is licensed under the MIT License. See the [LICENSE](./LICENSE) file for more details.
