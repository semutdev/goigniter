# GoIgniter

A lightweight Go web framework inspired by CodeIgniter, built on net/http stdlib.

> **Under Development** - This framework is still in active development. API may change without notice. Not recommended for production use yet.

## Features

- Zero external HTTP dependencies (net/http stdlib)
- CodeIgniter-inspired routing & controllers
- Built-in middleware (logger, recovery, CORS, rate limit, auth)
- Template engine with hot reload
- Session management (cookie-based)
- Auto-routing support

## Quick Start

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
└── starter/            # Quick start template
```

## License

MIT
