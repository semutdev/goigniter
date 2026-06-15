# GoIgniter Setup Tool Design

**Date:** 2026-06-15  
**Status:** Approved

## Overview

Tool untuk membuat project GoIgniter baru dengan satu command via curl + bash. User dapat memilih nama aplikasi, database, dan fitur yang diinginkan melalui interactive wizard atau command-line flags.

## Usage

### Quick Start (Interactive)
```bash
curl -sSL https://raw.githubusercontent.com/semutdev/goigniter/main/setup.sh | bash
```

### With Flags (Non-interactive)
```bash
curl -sSL https://raw.githubusercontent.com/semutdev/goigniter/main/setup.sh | bash -s -- --name=myapp --db=mysql --features=login,admin
```

### Available Flags
| Flag | Description | Default |
|------|-------------|---------|
| `--name` | Project name | (prompt) |
| `--db` | Database type (sqlite/mysql) | sqlite |
| `--db-host` | MySQL host | 127.0.0.1 |
| `--db-port` | MySQL port | 3306 |
| `--db-user` | MySQL user | root |
| `--db-pass` | MySQL password | (empty) |
| `--db-name` | MySQL database name | (project name) |
| `--features` | Comma-separated: login,admin | login,admin |

## Interactive Flow

```
🚀 GoIgniter Setup Wizard
━━━━━━━━━━━━━━━━━━━━━━━━━

? Project name: myapp
? Database type (sqlite/mysql) [sqlite]: 
? MySQL host [127.0.0.1]: (only if mysql)
? MySQL port [3306]: (only if mysql)
? MySQL user [root]: (only if mysql)
? MySQL password: (only if mysql)
? MySQL database name [myapp]: (only if mysql)

✓ Creating project structure...
✓ Generating configuration...
✓ Generating auth controllers...
✓ Generating admin dashboard...
✓ Done!

Run: cd myapp && go mod tidy && go run main.go
```

## Generated Project Structure

```
myapp/
├── main.go                    # Entry point
├── .env                       # Environment config
├── .env.example               # Example env for git
├── go.mod                     # Go module
├── application/
│   ├── config/
│   │   └── database.go        # DB connection (SQLite/MySQL)
│   ├── controllers/
│   │   ├── auth.go            # Login/logout handlers
│   │   └── admin/
│   │       ├── dashboard.go   # Dashboard homepage
│   │       ├── users.go       # User management CRUD
│   │       └── settings.go   # App settings CRUD
│   ├── models/
│   │   ├── user.go            # User model
│   │   └── setting.go         # Setting model
│   ├── views/
│   │   ├── layouts/
│   │   │   └── main.html      # Base layout
│   │   ├── auth/
│   │   │   └── login.html     # Login form
│   │   ├── admin/
│   │   │   ├── dashboard.html
│   │   │   ├── users/
│   │   │   │   ├── index.html
│   │   │   │   ├── _form.html
│   │   │   │   └── _row.html
│   │   │   ├── settings/
│   │   │   │   └── index.html
│   │   │   └── partials/
│   │   │       ├── _header.html
│   │   │       ├── _sidebar.html
│   │   │       └── _footer.html
│   │   └── partials/
│   │       ├── _messages.html
│   │       └── _pagination.html
│   └── libs/
│       └── auth.go            # Auth helpers (login, logout, check)
├── public/
│   └── static/
│       └── css/
│           └── style.css      # Minimal CSS
└── database/
    └── seed.go                # Migrations & seed data
```

## Features

### 1. Login System
- Login page with email/password form
- Session management via cookies
- Logout functionality
- Protected routes (middleware)
- Default admin account: `admin@admin.com` / `password`

### 2. Admin Dashboard
- **Dashboard Homepage**: Stats cards (total users, etc), recent activity
- **User Management**: 
  - List users with pagination
  - Create new user
  - Edit user (inline via HTMX)
  - Delete user (with confirmation)
- **Settings**:
  - Site name
  - Tagline
  - Logo URL
  - Stored in database, loaded on startup

### 3. UI/Styling
- **HTMX** for interactivity (no page reload on CRUD operations)
- **Minimal CSS** (~200-300 lines)
  - Clean, modern look
  - Responsive sidebar layout
  - Form styling
  - Table styling
  - Button styling
  - Loading indicators (htmx-indicator class)
- No build step required (HTMX via CDN, CSS inline)

## Setup Script Architecture

```
setup/
├── setup.sh              # Main entry point (curl downloads this)
├── installer.sh          # Download templates, validate, generate
└── templates/
    ├── main.go.tmpl
    ├── env.tmpl
    ├── database.go.tmpl
    ├── user.go.tmpl
    ├── setting.go.tmpl
    ├── auth.go.tmpl
    ├── auth_lib.go.tmpl
    ├── dashboard.go.tmpl
    ├── users.go.tmpl
    ├── settings.go.tmpl
    ├── seed.go.tmpl
    ├── main.html.tmpl
    ├── login.html.tmpl
    ├── dashboard.html.tmpl
    ├── users_index.html.tmpl
    ├── users_form.html.tmpl
    ├── users_row.html.tmpl
    ├── settings_index.html.tmpl
    ├── header.html.tmpl
    ├── sidebar.html.tmpl
    ├── footer.html.tmpl
    ├── messages.html.tmpl
    ├── pagination.html.tmpl
    └── style.css.tmpl
```

## Template Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `{{.AppName}}` | Project name | myapp |
| `{{.ModulePath}}` | Go module path | github.com/user/myapp |
| `{{.DbType}}` | Database type | sqlite / mysql |
| `{{.DbDsn}}` | Connection string | ./app.db or user:pass@tcp(host:port)/db |
| `{{.AppKey}}` | Session secret | (generated 32 chars) |

## Technical Details

### Database Schema

**Users Table:**
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(20) DEFAULT 'user',
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME,
    updated_at DATETIME
);
```

**Settings Table:**
```sql
CREATE TABLE settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT,
    created_at DATETIME,
    updated_at DATETIME
);
```

### Session Management
- Cookie-based sessions using existing `system/libraries/session`
- Session secret from `APP_KEY` in `.env`
- Session expires in 24 hours (configurable)

### Middleware
- `middleware.Logger()` - Request logging
- `middleware.Recovery()` - Panic recovery
- Custom `middleware.AuthRequired()` - Protect admin routes

## Dependencies

From existing goigniter:
- `github.com/semutdev/goigniter/system/core`
- `github.com/semutdev/goigniter/system/middleware`
- `github.com/semutdev/goigniter/system/helpers`
- `github.com/semutdev/goigniter/system/libraries/session`
- `github.com/joho/godotenv` (for .env loading)

For SQLite:
- `modernc.org/sqlite` (pure Go, no CGO)

For MySQL:
- `github.com/go-sql-driver/mysql`

## Success Criteria

1. User can create new project with single curl command
2. Generated project runs immediately with `go mod tidy && go run main.go`
3. Login works with default admin credentials
4. Admin dashboard accessible after login
5. CRUD operations work without page reload (HTMX)
6. Settings persist in database
7. Works on macOS and Linux