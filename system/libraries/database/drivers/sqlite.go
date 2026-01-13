package drivers

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // Pure Go SQLite driver (no CGO required)
)

// SQLiteDriver implements the Driver interface for SQLite.
type SQLiteDriver struct{}

func init() {
	Register("sqlite", &SQLiteDriver{})
}

// Open creates a new SQLite database connection.
func (d *SQLiteDriver) Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("sqlite: failed to ping database: %w", err)
	}

	return db, nil
}

// Placeholder returns "?" for SQLite.
func (d *SQLiteDriver) Placeholder(index int) string {
	return "?"
}

// Name returns the driver name.
func (d *SQLiteDriver) Name() string {
	return "sqlite"
}
