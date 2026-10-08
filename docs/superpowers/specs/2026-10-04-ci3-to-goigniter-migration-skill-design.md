# Design Document: CodeIgniter 3 to GoIgniter Migration Skill (`ci3-to-goigniter`)

**Date:** 2026-10-04  
**Status:** Approved  
**Target:** Claude Code, Codex, Antigravity CLI, and AI Agent environments  

---

## 1. Overview

Migrating a CodeIgniter 3 (PHP) application to GoIgniter (Golang) manually or via unguided AI prompts consumes tens of thousands of LLM tokens and frequently results in broken assumptions, hallucinated dependencies (e.g., using Gin/Echo instead of GoIgniter), and inconsistent typing.

This design introduces **`ci3-to-goigniter`**, an **All-in-One Cross-Agent Skill** that pairs a structured **Standard Operating Procedure (SKILL.md)** with **embedded zero-token CLI helper scripts** in `scripts/`. Repetitive and mechanical tasks (directory scaffolding, SQL DDL parsing to Go structs, and PHP template view conversion) are executed deterministically by the embedded scripts without wasting LLM tokens. The AI agent focuses purely on reasoning tasks: business logic translation, complex query mapping, session/auth flow, and compilation verification.

---

## 2. Architecture & File Layout

The skill will live directly inside the GoIgniter repository under `skills/ci3-to-goigniter/` and can be copied or symlinked to any agent skill directory:

```
skills/ci3-to-goigniter/
├── SKILL.md                          # Primary agent SOP & workflow instructions
├── scripts/
│   ├── inspect.go                    # Scans CI3 project & outputs inventory report
│   ├── sql2struct.go                 # Converts SQL CREATE TABLE to Go structs & models
│   └── view2gotpl.go                 # Transpiles PHP view templates to Go html/template
└── references/
    └── mapping-cheatsheet.md         # Comprehensive CI3 vs GoIgniter API dictionary
```

### Installation Targets
* **Claude Code:** `~/.claude/skills/ci3-to-goigniter/`
* **Codex / Agents:** `~/.agents/skills/ci3-to-goigniter/`
* **Antigravity / Gemini:** `~/.gemini/antigravity-cli/skills/ci3-to-goigniter/` or workspace-relative.

---

## 3. Component Specifications

### 3.1 `scripts/inspect.go` (Inventory & Discovery)
* **Goal:** Zero-token scanning of the legacy CI3 codebase.
* **Input:** Path to CI3 project root.
* **Output:** Markdown summary file (`migration-inventory.md`) and JSON manifest containing:
  * List of Controllers and their public action methods.
  * List of Models and referenced database tables.
  * List of Views and partial layout inclusions.
  * Custom routes defined in `application/config/routes.php`.
  * External libraries or helpers loaded in `application/config/autoload.php`.
* **Execution:**
  ```bash
  go run ./scripts/inspect.go -src /path/to/ci3_app -out ./migration-inventory.md
  ```

### 3.2 `scripts/sql2struct.go` (Deterministic Schema to Structs)
* **Goal:** Automatically create type-safe Go structs and base models from SQL DDL with zero LLM tokens.
* **Input:** SQL DDL file (`.sql`) containing `CREATE TABLE` statements.
* **Output:** Generated model files in `app/models/<table_name>.go`.
* **Type Mapping:**
  * `INT`, `TINYINT`, `SMALLINT`, `MEDIUMINT` $\rightarrow$ `int`
  * `BIGINT` $\rightarrow$ `int64`
  * `VARCHAR`, `CHAR`, `TEXT`, `LONGTEXT` $\rightarrow$ `string`
  * `DATETIME`, `TIMESTAMP`, `DATE` $\rightarrow$ `time.Time`
  * `BOOLEAN`, `TINYINT(1)` $\rightarrow$ `bool`
  * `FLOAT`, `DOUBLE`, `DECIMAL` $\rightarrow$ `float64`
  * Nullable fields $\rightarrow$ Go pointer types (`*string`, `*int`, etc.) or standard types.
* **Generated Code Pattern:**
  * Struct definition with `json:"..."` and `db:"..."` struct tags.
  * Base CRUD methods conforming to GoIgniter's database library.
* **Execution:**
  ```bash
  go run ./scripts/sql2struct.go -sql ./schema.sql -out ./app/models/
  ```

### 3.3 `scripts/view2gotpl.go` (Deterministic View Transpiler)
* **Goal:** Convert 80–90% of PHP view template markup to Go's `html/template` syntax instantly.
* **Transformations:**
  * Echo variables: `<?= $var ?>` or `<?php echo $var; ?>` $\rightarrow$ `{{ .Var }}`
  * Object properties: `<?= $user->name ?>` $\rightarrow$ `{{ .User.Name }}`
  * Array indexing: `<?= $user['name'] ?>` $\rightarrow$ `{{ .User.Name }}`
  * Conditionals: `<?php if ($cond): ?> ... <?php endif; ?>` $\rightarrow$ `{{ if .Cond }} ... {{ end }}`
  * Loops: `<?php foreach ($users as $user): ?> ... <?php endforeach; ?>` $\rightarrow$ `{{ range .Users }} ... {{ end }}`
  * Static URLs: `<?= base_url('assets/css/style.css') ?>` $\rightarrow$ `/assets/css/style.css`
  * Complex unconvertible PHP $\rightarrow$ commented out safely: `{{/* TODO_MIGRATE: <?php ... ?> */}}`
* **Execution:**
  ```bash
  go run ./scripts/view2gotpl.go -src /path/to/ci3_app/application/views -out ./app/views
  ```

### 3.4 `references/mapping-cheatsheet.md` (API Dictionary)
Provides the AI agent with a definitive reference table mapping CI3 features to GoIgniter:

| CI3 Feature | CI3 Code (PHP) | GoIgniter Equivalent (Go) |
|---|---|---|
| **Route GET** | `$route['users'] = 'user/index';` | `app.Get("/users", userCtrl.Index)` |
| **Route Param** | `$route['users/(:num)'] = 'user/detail/$1';` | `app.Get("/users/:id", userCtrl.Detail)` |
| **Form Input** | `$this->input->post('username')` | `c.Form("username")` |
| **Query Param** | `$this->input->get('page')` | `c.Query("page")` / `c.QueryInt("page")` |
| **JSON Response**| `$this->output->set_content_type('application/json')->set_output(json_encode($data));` | `c.JSON(http.StatusOK, data)` |
| **Render View** | `$this->load->view('users/index', $data);` | `c.HTML(http.StatusOK, "users/index", data)` |
| **Redirect** | `redirect('/login');` | `c.Redirect(http.StatusFound, "/login")` |
| **Session Set** | `$this->session->set_userdata('id', 1);` | `session.Set(c, "id", 1)` |
| **Session Get** | `$this->session->userdata('id');` | `session.Get(c, "id")` |
| **Flash Data** | `$this->session->set_flashdata('msg', 'ok');` | `session.SetFlash(c, "msg", "ok")` |
| **DB Get All** | `$this->db->get('users')->result();` | `db.Table("users").Get(&users)` |
| **DB Where** | `$this->db->where('id', $id)->get('users')->row();` | `db.Table("users").Where("id", id).First(&user)` |
| **DB Insert** | `$this->db->insert('users', $data);` | `db.Table("users").Insert(user)` |
| **DB Update** | `$this->db->where('id', $id)->update('users', $data);` | `db.Table("users").Where("id", id).Update(data)` |
| **DB Delete** | `$this->db->where('id', $id)->delete('users');` | `db.Table("users").Where("id", id).Delete()` |

---

## 4. End-to-End Migration Workflow (Agent SOP in `SKILL.md`)

When an agent is invoked to migrate a CI3 project, it must execute the following 6 phases sequentially:

1. **Phase 1: Project Inspection & Scaffolding**
   * Run `scripts/inspect.go` to generate the migration inventory.
   * Scaffold GoIgniter project layout (`main.go`, `go.mod`, `.env`, `app/`).
2. **Phase 2: Database Models & Structs**
   * Parse the SQL schema via `scripts/sql2struct.go` to populate `app/models/`.
   * Configure database connection in `.env` and `app/config/`.
3. **Phase 3: Views & Asset Migration**
   * Run `scripts/view2gotpl.go` to transpile views to `app/views/*.html`.
   * Copy static assets (`css`, `js`, `images`) to `public/` or `static/`.
4. **Phase 4: Controller & Business Logic Porting**
   * For each controller in the inventory, create `app/controllers/<name>.go`.
   * Translate input reading, validation, model calls, and view rendering using `mapping-cheatsheet.md`.
5. **Phase 5: Route Registration & Middleware Setup**
   * Register converted routes in `main.go` or `app/config/routes.go`.
   * Apply necessary middleware (Session, Recovery, Logger, CSRF, Security Headers).
6. **Phase 6: Compilation & Verification**
   * Run `go build ./...` and fix any type mismatches.
   * Verify test suites via `go test ./...`.
   * Confirm the application starts up cleanly.

---

## 5. Testing & Verification Plan

1. **Unit Testing of Scripts:**
   * Test `inspect.go` against a sample fixture CI3 directory.
   * Test `sql2struct.go` with various SQL dialects and column types (nullable, timestamps, primary keys).
   * Test `view2gotpl.go` with sample PHP view files containing standard tags, nested loops, and edge cases.
2. **End-to-End Test Scenario:**
   * Include a minimal sample CI3 application (`examples/fixtures/ci3_sample_app/`).
   * Perform migration using the skill and automated scripts.
   * Verify the resulting GoIgniter application compiles with `go build` and passes HTTP endpoint tests.
