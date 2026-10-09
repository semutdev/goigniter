---
title: Installation
description: Install Goigniter and start a new project.
sidebar:
  order: 2
---

## Prerequisites

Make sure Go is installed:

```bash
go version
```

## Quick Start

### Option 1: Setup Wizard (Recommended)

Run a single command to scaffold a new project:

```bash
curl -fsSL https://raw.githubusercontent.com/semutdev/goigniter/main/setup/setup.sh | bash
```

The wizard will ask for:
- Project name
- Database type (SQLite / MySQL)
- MySQL credentials (if MySQL is selected)

After completion:

```bash
cd yourproject
go mod tidy
go run main.go
```

### Option 2: Download Starter

Clone the repository and use the starter template:

```bash
git clone https://github.com/semutdev/goigniter
cd goigniter/starter
```

### Copy Environment

```bash
cp .env.example .env
```

### Install Dependencies

```bash
go mod tidy
```

### Run

```bash
go run main.go
```

Open http://localhost:8080 — you should see the welcome page.

## Project Structure

```
myapp/
├── application/
│   ├── controllers/    # Your controllers
│   └── views/          # HTML templates
├── public/             # Static files (CSS, JS, images)
├── go.mod
└── main.go             # Entry point
```

| Directory | Purpose |
|-----------|---------|
| `application/controllers/` | HTTP request handlers |
| `application/views/` | HTML templates |
| `public/` | Static assets |
| `main.go` | Application entry point |

## Hello World

Add a new route in `main.go`:

```go
app.GET("/hello", func(c *core.Context) error {
    return c.JSON(200, core.Map{
        "message": "Hello World!",
    })
})
```

Restart the server and visit http://localhost:8080/hello.

## Hot Reload (Optional)

Use [Air](https://github.com/cosmtrek/air) for auto-restart on file changes:

```bash
go install github.com/cosmtrek/air@latest
air
```

---

Next: [Routing](/en/guide/03-routing) to learn how to define routes.

Migrating from CodeIgniter 3? Check [Agentic Migration](/en/guide/11-agentic) for automated migration with AI assistance.
