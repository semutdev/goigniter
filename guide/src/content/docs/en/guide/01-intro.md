---
title: Why Goigniter?
description: A lightweight MVC framework for Go with CodeIgniter-like simplicity.
sidebar:
  order: 1
---

Go is a fantastic language for building web applications. It compiles to a single binary, handles thousands of concurrent requests with ease, and its standard library (`net/http`) covers most HTTP needs out of the box.

But building a web application from scratch in Go often means writing the same boilerplate over and over — request routing, parameter parsing, template rendering, database queries.

Goigniter handles these common tasks so you can focus on building your application.

## The Problem Goigniter Solves

### Repetitive Boilerplate

Every web app needs the same things:

- Routing requests to handlers
- Parsing form data, query parameters, JSON bodies
- Rendering HTML templates
- Connecting to a database
- Managing user sessions

Goigniter provides these as a cohesive toolkit — no assembly required.

### Unfamiliar Patterns

Many Go frameworks follow patterns that feel foreign if you come from PHP, Python, or Ruby. Goigniter uses an **MVC structure** that developers already know:

- `controllers/` — handle HTTP requests
- `models/` — interact with the database
- `views/` — render HTML templates

## What Goigniter Offers

- **Auto-routing** — register a controller, and its methods become routes automatically
- **Template engine** — Go's `html/template` with added helpers (`base_url`, `site_url`, formatting)
- **Query builder** — method-chaining SQL builder similar to ActiveRecord
- **Session management** — signed and encrypted cookie sessions
- **Database drivers** — SQLite and MySQL support out of the box
- **Middleware pipeline** — global, group, and controller-level middleware
- **File upload** — with validation and image processing (resize, crop, thumbnail)
- **Security features** — CSRF protection, security headers, encrypted sessions
- **CLI setup wizard** — `curl` one-liner to scaffold a new project
- **Agentic migration skill** — migrate CodeIgniter 3 projects to Goigniter with AI assistance

## Who Is It For?

Goigniter is a good fit if:
- You want to build web applications in Go without writing boilerplate
- You prefer MVC structure over micro-frameworks
- You are familiar with CodeIgniter, Laravel, or similar frameworks
- You need a lightweight framework — not a full-featured enterprise platform

Goigniter might not be for you if:
- You prefer idiomatic Go patterns with minimal abstraction
- You need advanced enterprise features (service mesh, event sourcing, CQRS)

---

Ready to start? Head over to [Installation](/en/guide/02-installation).
