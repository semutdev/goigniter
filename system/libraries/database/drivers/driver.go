// Package drivers provides database driver interfaces and implementations.
package drivers

import "database/sql"

// Driver defines the interface that all database drivers must implement.
type Driver interface {
	// Open creates a new database connection.
	Open(dsn string) (*sql.DB, error)

	// Placeholder returns the placeholder format for the driver.
	// SQLite/MySQL use "?", PostgreSQL uses "$1", "$2", etc.
	Placeholder(index int) string

	// Name returns the driver name.
	Name() string
}

// registry holds registered drivers
var registry = make(map[string]Driver)

// Register adds a driver to the registry.
func Register(name string, driver Driver) {
	registry[name] = driver
}

// Get retrieves a driver from the registry.
func Get(name string) (Driver, bool) {
	driver, ok := registry[name]
	return driver, ok
}
