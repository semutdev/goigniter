package drivers

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// MySQLDriver implements the Driver interface for MySQL.
type MySQLDriver struct{}

func init() {
	Register("mysql", &MySQLDriver{})
}

// Open creates a new MySQL database connection.
// DSN format: user:password@tcp(host:port)/dbname?parseTime=true
func (d *MySQLDriver) Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql: failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql: failed to ping database: %w", err)
	}

	return db, nil
}

// Placeholder returns "?" for MySQL.
func (d *MySQLDriver) Placeholder(index int) string {
	return "?"
}

// Name returns the driver name.
func (d *MySQLDriver) Name() string {
	return "mysql"
}
