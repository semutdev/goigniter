// Package database provides a query builder for database operations.
package database

import (
	"database/sql"
	"fmt"

	"goigniter/system/libraries/database/drivers"
)

// DB represents a database connection with query builder capabilities.
type DB struct {
	conn       *sql.DB
	driver     drivers.Driver
	lastError  error
	tx         *sql.Tx
	inTransaction bool
}

// global default instance
var defaultDB *DB

// Open creates a new database connection.
func Open(driverName, dsn string) (*DB, error) {
	driver, ok := drivers.Get(driverName)
	if !ok {
		return nil, fmt.Errorf("database: unknown driver %q", driverName)
	}

	conn, err := driver.Open(dsn)
	if err != nil {
		return nil, err
	}

	return &DB{
		conn:   conn,
		driver: driver,
	}, nil
}

// SetDefault sets the global default database instance.
func SetDefault(db *DB) {
	defaultDB = db
}

// Default returns the global default database instance.
func Default() *DB {
	return defaultDB
}

// Table starts a new query builder for the given table.
func (db *DB) Table(name string) *Builder {
	return newBuilder(db, name)
}

// Table starts a new query builder using the default database.
func Table(name string) *Builder {
	if defaultDB == nil {
		panic("database: no default database set, call SetDefault() first")
	}
	return defaultDB.Table(name)
}

// Query executes a raw SQL query.
func (db *DB) Query(query string, args ...any) *RawResult {
	return &RawResult{
		db:    db,
		query: query,
		args:  args,
	}
}

// Query executes a raw SQL query using the default database.
func Query(query string, args ...any) *RawResult {
	if defaultDB == nil {
		panic("database: no default database set, call SetDefault() first")
	}
	return defaultDB.Query(query, args...)
}

// Exec executes a raw SQL statement (INSERT, UPDATE, DELETE).
func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	if db.inTransaction && db.tx != nil {
		return db.tx.Exec(query, args...)
	}
	return db.conn.Exec(query, args...)
}

// Exec executes a raw SQL statement using the default database.
func Exec(query string, args ...any) (sql.Result, error) {
	if defaultDB == nil {
		panic("database: no default database set, call SetDefault() first")
	}
	return defaultDB.Exec(query, args...)
}

// Begin starts a new transaction.
func (db *DB) Begin() (*DB, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("database: failed to begin transaction: %w", err)
	}

	return &DB{
		conn:          db.conn,
		driver:        db.driver,
		tx:            tx,
		inTransaction: true,
	}, nil
}

// Commit commits the current transaction.
func (db *DB) Commit() error {
	if !db.inTransaction || db.tx == nil {
		return fmt.Errorf("database: not in transaction")
	}
	err := db.tx.Commit()
	db.inTransaction = false
	db.tx = nil
	return err
}

// Rollback aborts the current transaction.
func (db *DB) Rollback() error {
	if !db.inTransaction || db.tx == nil {
		return fmt.Errorf("database: not in transaction")
	}
	err := db.tx.Rollback()
	db.inTransaction = false
	db.tx = nil
	return err
}

// Transaction executes a function within a transaction.
func (db *DB) Transaction(fn func(tx *DB) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Transaction executes a function within a transaction using the default database.
func Transaction(fn func(tx *DB) error) error {
	if defaultDB == nil {
		panic("database: no default database set, call SetDefault() first")
	}
	return defaultDB.Transaction(fn)
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Error returns the last error.
func (db *DB) Error() error {
	return db.lastError
}

// Conn returns the underlying sql.DB connection.
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Driver returns the current driver.
func (db *DB) Driver() drivers.Driver {
	return db.driver
}

// execer returns the appropriate executor (tx or conn)
func (db *DB) execer() interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
} {
	if db.inTransaction && db.tx != nil {
		return db.tx
	}
	return db.conn
}
