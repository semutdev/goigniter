# GoIgniter

<p align="center">
  <img src="logo.png" alt="GoIgniter Logo" width="200">
</p>

<p align="center">
  <strong>A lightweight Go web framework inspired by CodeIgniter, built on net/http stdlib.</strong>
</p>

> **Under Development** - This framework is still in active development. API may change without notice. Not recommended for production use yet.

## Features

- Zero external HTTP dependencies (net/http stdlib)
- CodeIgniter-inspired routing & controllers
- Built-in middleware (logger, recovery, CORS, rate limit, auth)
- Template engine with hot reload
- Session management (cookie-based)
- Auto-routing support
- **Setup wizard** - Create new projects with one command

---

## Create New Project

The fastest way to create a new GoIgniter project is using the setup script:

### Quick Start (Copy & Paste)

```bash
# Interactive mode - will prompt for project details
curl -sSL https://raw.githubusercontent.com/semutdev/goigniter/main/setup/setup.sh | bash
```

```bash
# Non-interactive mode - create project directly
curl -sSL https://raw.githubusercontent.com/semutdev/goigniter/main/setup/setup.sh | bash -s -- --name=myapp --db=sqlite
```

```bash
# With MySQL database
curl -sSL https://raw.githubusercontent.com/semutdev/goigniter/main/setup/setup.sh | bash -s -- --name=myapp --db=mysql --db-user=root --db-pass=password --db-name=myapp
```

### Setup Options

| Flag | Description | Default |
|------|-------------|---------|
| `--name` | Project name | (prompt) |
| `--db` | Database type (sqlite/mysql) | sqlite |
| `--db-host` | MySQL host | 127.0.0.1 |
| `--db-port` | MySQL port | 3306 |
| `--db-user` | MySQL user | root |
| `--db-pass` | MySQL password | (empty) |
| `--db-name` | MySQL database name | (project name) |

### Generated Project Includes

- **Login System** - Session-based authentication
- **Admin Dashboard** - Stats, user management (CRUD), settings
- **HTMX + Minimal CSS** - Modern UI without page reloads
- **SQLite/MySQL** - Choose your database

### After Setup

```bash
cd myapp

# Add local goigniter dependency (for development)
echo 'replace github.com/semutdev/goigniter => ../' >> go.mod

go mod tidy

# First run: create tables and admin user
DB_SEED=true go run main.go
```

**Default login:** `admin@admin.com` / `password`

---

## Manual Quick Start

If you prefer to build manually without the setup script:

```go
package main

import (
    "goigniter/system/core"
    "goigniter/system/middleware"
)

func main() {
    app := core.New()

    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())

    app.GET("/", func(c *core.Context) error {
        return c.JSON(200, core.Map{"message": "Hello GoIgniter!"})
    })

    app.Run(":8080")
}
```

---

## Examples

| Example | Description | Run |
|---------|-------------|-----|
| `examples/simple` | Basic routing & JSON response | `go run examples/simple/main.go` |
| `examples/autoroute` | Controller auto-routing | `go run examples/autoroute/main.go` |
| `examples/views` | Template rendering | `cd examples/views && go run main.go` |
| `examples/database` | Query builder (SQLite) | `go run examples/database/main.go` |
| `examples/full-crud` | Full CRUD with database (GORM) | `cd examples/full-crud && go run main.go` |

## Documentation

See [examples/](./examples) for usage patterns.

## Project Structure

```
goigniter/
├── system/
│   ├── core/           # Core framework (router, context, controller)
│   ├── middleware/     # Built-in middleware
│   ├── libraries/      # Session, database query builder
│   └── helpers/        # URL helpers, template funcs
├── examples/           # Usage examples
├── starter/            # Quick start template
└── setup/              # Project setup wizard
    ├── setup.sh        # Entry point (curl downloads this)
    ├── installer.sh    # Template processor
    └── templates/      # All generated file templates
```

## License

MIT