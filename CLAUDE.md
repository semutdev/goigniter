# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Run the application
go run main.go

# Build the application
go build -o goigniter

# Download dependencies
go mod tidy
```

## Environment Setup

Create a `.env` file with:
- `DB_DSN` - MySQL connection string (GORM DSN format)
- `APP_PORT` - Server port (defaults to `:6789` if not set)

## Architecture

This is a Go web application using the Echo framework with a CodeIgniter-inspired auto-routing system.

### Auto-Routing System (libs/autoroute.go)

The `AutoRoute` function uses reflection to automatically map URLs to controller methods:
- URL pattern: `/:controller/:method` or `/:controller/:method/:id`
- Controllers are registered in a map in `main.go`
- Method names are title-cased from the URL (e.g., `/users/index` → `Users.Index()`)

### Key Patterns

**Controllers** (`controllers/`): Struct-based controllers with methods that accept `echo.Context`. Register new controllers in the `registry` map in `main.go`.

**Models** (`models/`): GORM models with auto-migration enabled in `main.go`.

**Views** (`views/`): Go HTML templates rendered via Echo's template system. Templates use `{{define "name"}}` blocks.

**Database**: Global `config.DB` instance (GORM with MySQL driver). Auto-migration runs on startup.

### Frontend

Uses HTMX for dynamic updates with template partials. Static files served from `/static` → `public/`.
