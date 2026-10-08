---
name: ci3-to-goigniter
description: Migrate existing PHP CodeIgniter 3 projects to GoIgniter framework systematically with zero hallucination and minimal tokens
---

# CodeIgniter 3 to GoIgniter Migration Skill

A deterministic, production-grade migration system for migrating legacy PHP CodeIgniter 3 (CI3) applications to [GoIgniter](https://github.com/semutdev/goigniter) — the modern, idiomatic Go framework built with CodeIgniter-familiar developer experience (MVC, Active Record / Query Builder, flash sessions, template engine, form validation, and CLI commands).

---

## 1. Overview & Core Philosophy

Migrating dynamic PHP codebases to a statically typed, compiled language like Go can quickly consume enormous context windows and cause hallucination if handled naively by prompting an LLM to rewrite thousands of lines of PHP code.

This skill follows a **Hybrid Automation Architecture**:

1. **Zero-Token Mechanical Tooling**: Repetitive, deterministic parsing and conversion tasks are offloaded to specialized local Go CLI tools:
   - `inspect`: Scans the legacy CI3 directory, extracts controllers, models, views, libraries, helpers, config, routes, database tables, and produces an actionable migration inventory checklist.
   - `sql2struct`: Parses MySQL/PostgreSQL/SQLite DDL schemas and generates Go structs with struct tags (`db`, `json`), Active Record model definitions, and CRUD boilerplate.
   - `view2gotpl`: AST-aware transpiler converting PHP view files (`.php`) into Go `html/template` files (`.html`), rewriting PHP echo statements, conditionals, loops, CI3 helpers (`site_url`, `base_url`, `form_open`), and view partial includes (`$this->load->view`).

2. **Guided Agent Reasoning for Business Logic**: Agent intelligence is focused exclusively where human judgment is needed:
   - Porting controller action workflows and business rules.
   - Structuring relational queries and transaction boundaries (`database.Transaction(func(tx *database.DB) error { ... })`).
   - Translating dynamic PHP types and nullable values to strict Go types.
   - Adapting third-party libraries and authentication layers (e.g. bcrypt-compatible password verification).
   - Establishing middleware pipelines (CSRF, Sessions, Logging, Recovery, Security Headers).

3. **Zero Hallucination with Definitive Reference**:
   - Every API mapping is strictly governed by the exhaustive [`references/mapping-cheatsheet.md`](references/mapping-cheatsheet.md), providing exact side-by-side equivalents for CI3 and GoIgniter.

---

## 2. Skill Directory Structure

```
skills/ci3-to-goigniter/
├── SKILL.md                          # Primary agent SOP & migration guide (this document)
├── main.go                           # Migration toolkit CLI entrypoint
├── scripts/
│   ├── inspect.go                    # Phase 1: CI3 project analyzer & checklist generator
│   ├── inspect_test.go               # Unit tests for inspect
│   ├── sql2struct.go                 # Phase 2: SQL DDL to Go model struct generator
│   ├── sql2struct_test.go            # Unit tests for sql2struct
│   ├── view2gotpl.go                 # Phase 3: PHP view to Go template transpiler
│   └── view2gotpl_test.go            # Unit tests for view2gotpl
└── references/
    └── mapping-cheatsheet.md         # Exhaustive CI3 vs GoIgniter API dictionary
```

---

## 3. Installation & Agent Setup

Install this skill into your preferred AI agent environment to make it discoverable across projects:

### Claude Code
Symlink or copy the skill directory into your Claude Code skills directory:
```bash
# Global installation
mkdir -p ~/.claude/skills
ln -s /path/to/goigniter/skills/ci3-to-goigniter ~/.claude/skills/ci3-to-goigniter

# Or local project installation
mkdir -p .claude/skills
cp -r /path/to/goigniter/skills/ci3-to-goigniter .claude/skills/
```

### OpenAI Codex / Agent Workspaces
```bash
mkdir -p ~/.agents/skills
ln -s /path/to/goigniter/skills/ci3-to-goigniter ~/.agents/skills/ci3-to-goigniter
```

### Antigravity / Gemini CLI
Symlink or copy into the Antigravity skills directory:
```bash
# User-level installation
mkdir -p ~/.gemini/antigravity-cli/skills
ln -s /path/to/goigniter/skills/ci3-to-goigniter ~/.gemini/antigravity-cli/skills/ci3-to-goigniter

# Or use directly within the GoIgniter repository under skills/ci3-to-goigniter
```

---

## 4. The 6-Phase Migration Workflow (SOP)

Follow this standardized 6-phase procedure sequentially. Never skip phases.

```dot
digraph migration_workflow {
    rankdir=LR;
    node [shape=box, style=rounded];
    P1 [label="Phase 1\nDiscovery & Inventory"];
    P2 [label="Phase 2\nSchema & Models"];
    P3 [label="Phase 3\nViews & Assets"];
    P4 [label="Phase 4\nControllers & Logic"];
    P5 [label="Phase 5\nRoutes & Middleware"];
    P6 [label="Phase 6\nVerification & Test"];

    P1 -> P2 -> P3 -> P4 -> P5 -> P6;
}
```

---

### Phase 1: Project Discovery & Inventory

Before writing any code, scan the legacy CodeIgniter 3 project to understand its scope, architecture, and dependencies.

#### Step 1.1: Run `inspect`
Run the automated discovery CLI against the legacy CI3 codebase:
```bash
go run ./skills/ci3-to-goigniter inspect -src /path/to/ci3 -out ./migration-inventory.md
```
Or for machine-readable JSON:
```bash
go run ./skills/ci3-to-goigniter inspect -src /path/to/ci3 -format json -out ./migration-inventory.json
```

#### Step 1.2: Analyze the Inventory Report
The generated `migration-inventory.md` categorizes every component:
- **Controllers & Methods**: List of all classes extending `CI_Controller` and their action methods.
- **Models & Table References**: Models extending `CI_Model` and database tables accessed.
- **Views & Partials**: Header, footer, layout files, and subdirectories.
- **Custom Libraries & Helpers**: User-defined libraries in `application/libraries` and helper scripts in `application/helpers`.
- **Config & Routing**: Autoloaded packages, custom config items, database connection drivers, and routing rules in `application/config/routes.php`.
- **Assets & Static Files**: CSS, JS, fonts, uploads, and media folders.

Keep `migration-inventory.md` in your project root as a checklist. Mark items as completed as you progress.

---

### Phase 2: Database Schema & Models Scaffolding

Set up the database connection and generate strongly typed model structs from your database schema.

#### Step 2.1: Obtain SQL DDL Schema
Export the schema from your legacy database (without data inserts):
```bash
# MySQL
mysqldump -u root -p --no-data my_database > schema.sql

# PostgreSQL
pg_dump -U postgres --schema-only my_database > schema.sql

# SQLite
sqlite3 my_database.db .schema > schema.sql
```

#### Step 2.2: Run `sql2struct`
Run the schema generator to produce Go model files with table definitions, JSON/DB tags, and Active Record helper methods:
```bash
go run ./skills/ci3-to-goigniter sql2struct -sql schema.sql -out ./app/models -pkg models
```
*Note: If your target project uses `application/models`, point `-out ./application/models` instead.*

When `-out` points to a directory, `sql2struct` generates individual files per table (e.g. `users.go`, `posts.go`, `orders.go`). Each generated model includes:
- Go struct with exported fields matching column names (e.g. `user_id` -> `UserID`).
- Exact type mappings: `INT` -> `int64`, `VARCHAR/TEXT` -> `string`, `DECIMAL/FLOAT` -> `float64`, `DATETIME/TIMESTAMP` -> `time.Time`, `TINYINT(1)/BOOLEAN` -> `bool`, nullable columns -> `*Type`.
- CRUD helper methods on the model struct (`Find(id)`, `All()`, `Create(data)`, `Update(id, data)`, `Delete(id)`).

#### Step 2.3: Configure Database Connection
Configure environment variables in `.env`:
```env
APP_ENV=development
APP_PORT=:8080
APP_KEY=your-secret-key-32-bytes-length!!

DB_DRIVER=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=secret
DB_NAME=my_database
```

Verify `app/config/database.go` (or `application/config/database.go`) initializes the connection and sets it as the default instance:
```go
package config

import (
    "fmt"
    "log"
    "os"

    "github.com/semutdev/goigniter/system/libraries/database"
    _ "github.com/go-sql-driver/mysql" // or _ "modernc.org/sqlite"
)

var DB *database.DB

func ConnectDB() {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_NAME"),
    )

    var err error
    DB, err = database.Open(os.Getenv("DB_DRIVER"), dsn)
    if err != nil {
        log.Fatalf("Database connection failed: %v", err)
    }

    // Set as global default so models generated by sql2struct can use database.Table(...) directly!
    database.SetDefault(DB)
}
```

> [!IMPORTANT]
> Always call `database.SetDefault(DB)`. This enables models generated by `sql2struct` (which import `"github.com/semutdev/goigniter/system/libraries/database"`) as well as controllers to invoke `database.Table(...)` directly without passing around database handles!

---

### Phase 3: View Templates Transpilation & Assets

Migrate PHP view files to Go HTML templates and arrange static assets.

#### Step 3.1: Run `view2gotpl`
Transpile the entire CI3 views directory to Go templates:
```bash
go run ./skills/ci3-to-goigniter view2gotpl -src /path/to/ci3/application/views -out ./app/views
```
Or transpile an individual file:
```bash
go run ./skills/ci3-to-goigniter view2gotpl -file /path/to/ci3/application/views/welcome_message.php -out ./app/views
```

#### Transpilation Conversions Performed Automatically:
| CodeIgniter 3 PHP View | Transpiled Go Template | Notes |
|---|---|---|
| `<?= $title ?>` | `{{ .Title }}` | Variables capitalized for struct export |
| `<?php echo htmlspecialchars($body); ?>` | `{{ .Body }}` | Go `html/template` auto-escapes safely |
| `<?php if ($user): ?>...<?php endif; ?>` | `{{ if .User }}...{{ end }}` | Conditionals |
| `<?php if (!empty($items)): ?>` | `{{ if .Items }}...{{ end }}` | `empty()` / `isset()` normalization |
| `<?php foreach ($users as $u): ?>` | `{{ range .Users }}...{{ end }}` | Loops |
| `<?php echo base_url(); ?>/assets/css/app.css` | `{{ base_url }}/assets/css/app.css` | Asset URLs with helper |
| `<?= base_url('assets/css/app.css') ?>` | `{{ base_url }}/assets/css/app.css` | Normalized without double slashes |
| `<?= site_url('users/edit/' . $id) ?>` | `{{ site_url }}/users/edit/{{ .ID }}` | Concatenation with model variables |
| `<?= form_open('login') ?>` | `<form action="{{ site_url }}/login" method="POST">` | Form helper with site_url |
| `<?php $this->load->view('header'); ?>` | `{{ template "header.html" . }}` | Partials / layouts (in templates use `"name.html"`; in controllers `c.View()` omits `.html`) |

#### Step 3.2: Static Assets Migration
1. Copy all static files (`css`, `js`, `images`, `fonts`, `uploads`) from the CI3 web root or `assets/` to `public/`:
   ```bash
   cp -r /path/to/ci3/assets ./public/assets
   ```
2. Configure static routing in `main.go`:
   ```go
   app.Static("/static/", "./public")
   // or
   app.Static("/assets/", "./public/assets")
   ```
3. Load templates with helper functions:
   ```go
   helpers.Init("http://localhost:8080")
   if err := app.LoadTemplatesWithFuncs("./app/views", true, helpers.AllTemplateFuncs()); err != nil {
       log.Fatalf("Template load error: %v", err)
   }
   ```

---

### Phase 4: Controller & Business Logic Porting

Port each controller identified in Phase 1 methodically. Refer to [`references/mapping-cheatsheet.md`](references/mapping-cheatsheet.md) for full API translations.

#### Step 4.1: Controller Anatomy in GoIgniter
CI3 controllers extend `CI_Controller`. In GoIgniter, controllers embed `core.Controller`.

GoIgniter supports two routing controller paradigms:
1. **Explicit Route Handlers**: Handlers accept `c *core.Context` and return an `error`:
   `func (u *UserController) Index(c *core.Context) error`
2. **AutoRoute Handlers**: Controllers embed `core.Controller`, and methods take no arguments `func (u *UserController) Index() error` (or `void`), accessing the request context via `u.Ctx`.

```go
package controllers

import (
    "net/http"
    "github.com/semutdev/goigniter/system/core"
    "myapp/app/models"
)

type UserController struct {
    core.Controller
    userModel *models.UserModel
}

func NewUserController() *UserController {
    return &UserController{
        userModel: models.NewUserModel(),
    }
}

// 1. Explicit Route Handler (registered via app.GET("/users", userController.Index)):
func (u *UserController) Index(c *core.Context) error {
    users, err := u.userModel.FindAll()
    if err != nil {
        return c.String(http.StatusInternalServerError, err.Error())
    }
    // Template names omit the ".html" extension
    return c.View("users/index", core.Map{
        "Title": "All Users",
        "Users": users,
    })
}

// 2. AutoRoute Handler (registered via core.Register(&UserController{}) and app.AutoRoute()):
// Method takes no parameters, accessing request state via u.Ctx:
func (u *UserController) Detail() error {
    id, err := u.Ctx.ParamInt("id")
    if err != nil {
        return u.Ctx.String(http.StatusBadRequest, "Invalid ID")
    }

    user, err := u.userModel.Find(int64(id))
    if err != nil {
        return u.Ctx.String(http.StatusNotFound, "User not found")
    }

    return u.Ctx.View("users/detail", core.Map{
        "Title": "User Detail",
        "User":  user,
    })
}
```

#### Step 4.2: Input Handling & Sanitization
| CI3 PHP | GoIgniter | Notes |
|---|---|---|
| `$this->input->post('email')` | `c.Form("email")` | POST form data (or `c.FormValue`) |
| `$this->input->get('page')` | `c.Query("page")` | Query string param (`c.QueryDefault("page", "1")`) |
| `(int)$this->input->get('page')`| `c.QueryInt("page")` | Parsed int query parameter |
| `$this->uri->segment(3)` | `c.Param("id")` | Route parameter (`c.ParamInt("id")`) |
| `$this->input->post(NULL, TRUE)` | `c.Bind(&req)` | Binds JSON or form body automatically into struct |
| `$this->input->is_ajax_request()` | `c.Header("X-Requested-With") == "XMLHttpRequest"` | Checks AJAX request header |
| `$this->input->ip_address()` | `c.IP()` | Client remote IP (handles `X-Forwarded-For`) |
| `$this->input->server('HTTP_HOST')`| `c.Header("Host")` | Incoming request Host header |
| `$this->upload->do_upload('pic')` | `file, hdr, err := c.Request.FormFile("pic")` | Standard form file or `upload.New(cfg).Do("pic", c.Request)` |

#### Step 4.3: Database Queries & Active Record
Use the GoIgniter Active Record Query Builder (`github.com/semutdev/goigniter/system/libraries/database`):

```go
import (
    "github.com/semutdev/goigniter/system/libraries/database"
    "myapp/app/models"
)

// CI3: $this->db->get_where('users', ['status' => 'active'])->result_array();
// GoIgniter: Where("col", val) and Get(&users) scans into slice of structs
var users []models.User
err := database.Table("users").
    Where("status", "active").
    Get(&users)

// CI3: $this->db->select('id, name, email')->from('users')->where('id', $id)->limit(1)->get()->row_array();
// GoIgniter: First(&user) scans single record into destination struct
var user models.User
err := database.Table("users").
    Select("id", "name", "email").
    Where("id", id).
    First(&user)

// CI3: $this->db->insert('users', $data);
// GoIgniter: Insert(data) returns only error
err := database.Table("users").Insert(map[string]any{
    "name":  req.Name,
    "email": req.Email,
})

// CI3: $this->db->insert('users', $data); $insert_id = $this->db->insert_id();
// GoIgniter: InsertGetId returns auto-increment ID (int64, error)
id, err := database.Table("users").InsertGetId(map[string]any{
    "name":  req.Name,
    "email": req.Email,
})

// CI3: $this->db->where('id', $id)->update('users', $data);
// GoIgniter: Update(data) returns only error
err := database.Table("users").
    Where("id", id).
    Update(map[string]any{"name": req.Name})

// CI3: $this->db->where('id', $id)->delete('users');
// GoIgniter: Delete() returns only error
err := database.Table("users").
    Where("id", id).
    Delete()

// CI3 Transactions: $this->db->trans_start(); ... $this->db->trans_complete();
// GoIgniter Transactions: closure runner with automatic commit/rollback on error
err := database.Transaction(func(tx *database.DB) error {
    if err := tx.Table("accounts").Where("id", from).Update(map[string]any{"balance": fromBalance}); err != nil {
        return err // Triggers automatic rollback
    }
    if err := tx.Table("accounts").Where("id", to).Update(map[string]any{"balance": toBalance}); err != nil {
        return err // Triggers automatic rollback
    }
    return nil // Commits transaction
})
```

> [!NOTE]
> In GoIgniter's query builder, `Where("col", val)` defaults to equality (`col = ?`). For comparisons, pass three arguments: `Where("col", ">", val)` or `Where("age", ">=", 18)`. Do not pass raw SQL fragments like `"col = ?"`.

#### Step 4.4: Sessions, Flash Data, and Auth
In GoIgniter, sessions use HMAC-signed (and optionally AES-256-GCM encrypted) cookie sessions managed via `"github.com/semutdev/goigniter/system/libraries/session"`:

```go
import (
    "os"
    "github.com/semutdev/goigniter/system/core"
    "github.com/semutdev/goigniter/system/libraries/session"
)

// In main.go initialization:
session.Init(session.Config{
    Secret:     os.Getenv("APP_KEY"), // HMAC signing key (min 32 bytes)
    CookieName: "goigniter_session",
    MaxAge:     86400 * 7,            // 7 days
})

// In controllers:
func (u *UserController) Login(c *core.Context) error {
    sess := session.Get(c)
    sess.Set("user_id", user.ID)
    sess.Save(c)

    // Flash data (automatically saved to cookie)
    session.SetFlash(c, "success", "Profile updated!")
    return c.Redirect(302, "/dashboard")
}

func (u *UserController) Dashboard(c *core.Context) error {
    sess := session.Get(c)
    userID := sess.Get("user_id") // or sess.GetInt("user_id")

    // Consume flash data
    msg := session.GetFlash(c, "success")

    return c.View("dashboard/index", core.Map{
        "UserID":  userID,
        "Message": msg,
    })
}

func (u *UserController) Logout(c *core.Context) error {
    sess := session.Get(c)
    sess.Destroy(c)
    return c.Redirect(302, "/login")
}
```

#### Step 4.5: Password Verification (CI3 Bcrypt Compatibility)
CI3 passwords hashed with `password_hash($pwd, PASSWORD_BCRYPT)` are 100% compatible with Go's `golang.org/x/crypto/bcrypt`:

```go
import "golang.org/x/crypto/bcrypt"

// Verify legacy password
err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(passwordAttempt))
if err != nil {
    // Invalid password
}

// Hash new password
hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
```

#### Step 4.6: Response Handling & Views
```go
// Render View (CI3: $this->load->view('users/index', $data))
// Note: Template names omit the ".html" extension (e.g. "users/index")
return c.View("users/index", core.Map{
    "Title": "All Users",
    "Users": users,
})

// Or render view with explicit HTTP status code:
// return c.ViewWithCode(http.StatusOK, "users/index", core.Map{...})

// JSON Response (CI3: echo json_encode($data))
return c.JSON(http.StatusOK, core.Map{
    "status": "success",
    "data":   users,
})

// Redirect (CI3: redirect('users/login'))
return c.Redirect(http.StatusFound, "/users/login")
```

---

### Phase 5: Route Registration & Middleware Setup

Map legacy routing rules from `application/config/routes.php` and set up the production middleware chain.

#### Step 5.1: Route Registration
In GoIgniter, routes can be registered explicitly or through reflection-based auto-routing:

```go
func RegisterRoutes(app *core.Application) {
    // 1. Explicit Routes (translating $route['users/(:num)'] = 'user/detail/$1')
    app.GET("/", welcomeHandler)
    app.GET("/users", userController.Index)
    app.GET("/users/create", userController.Create)
    app.POST("/users", userController.Store)
    app.GET("/users/:id", userController.Detail)
    app.POST("/users/:id/update", userController.Update)
    app.POST("/users/:id/delete", userController.Delete)

    // 2. Route Groups for API or Admin
    api := app.Group("/api/v1")
    {
        api.GET("/status", apiController.Status)
        api.POST("/auth/login", apiController.Login)
    }

    // 3. Auto-Routing (GoIgniter's signature feature matching CI3 convention)
    // Automatically strips "Controller" suffix, supports prefix namespaces,
    // registers index aliases (/admin/dashboard and /admin/dashboard/index),
    // and maps Go PascalCase methods to CI3 snake_case routes (FilterSummary -> /admin/dashboard/filter_summary)
    app.Register(&controllers.WelcomeController{})
    app.Register(&controllers.UserController{})
    app.Register(&controllers.DashboardController{}, "admin") // prefix "admin" -> /admin/dashboard
    app.AutoRoute()
}
```

#### Step 5.2: Middleware Pipeline
Configure sessions and required middleware in `main.go`:
```go
import (
    "os"
    "github.com/semutdev/goigniter/system/middleware"
    "github.com/semutdev/goigniter/system/libraries/session"
)

// 1. Session initialization (system/libraries/session)
session.Init(session.Config{
    Secret:     os.Getenv("APP_KEY"),
    CookieName: "goigniter_session",
    MaxAge:     86400 * 7,
})

// 2. Global Middleware Pipeline
// Panic recovery
app.Use(middleware.Recovery())

// HTTP request logger
app.Use(middleware.Logger())

// Security headers (X-Content-Type-Options, X-Frame-Options, X-XSS-Protection)
app.Use(middleware.SecurityHeaders())

// CSRF Protection (matching CI3 $config['csrf_protection'] = TRUE)
// Use default configuration:
app.Use(middleware.CSRF())

// Or customize via CSRFWithConfig:
app.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
    HeaderName:    "X-CSRF-Token",
    FormFieldName: "csrf_token",
    CookieName:    "csrf_token",
    Secure:        false, // true in production with HTTPS
}))
```

---

### Phase 6: Verification, Compilation & Testing

Ensure zero regression before declaring migration complete.

#### Step 6.1: Compile Codebase
```bash
go build ./...
```
Fix any type errors or missing package imports. Go's strict compiler guarantees zero syntax or missing-method bugs.

#### Step 6.2: Run Static Analysis
```bash
go vet ./...
```

#### Step 6.3: Run Unit & Integration Tests
```bash
go test -v ./...
```

#### Step 6.4: Smoke Test Checklist
- [ ] Application starts cleanly (`go run main.go`).
- [ ] Static assets (`/assets/...` or `/static/...`) return HTTP 200.
- [ ] Database queries succeed with no syntax errors.
- [ ] Existing CI3 user passwords authenticate successfully using bcrypt.
- [ ] Flash messages display on next page view and do not leak into subsequent requests.
- [ ] Form validation triggers appropriate error messages.
- [ ] CSRF tokens validate correctly on POST/PUT requests.

---

## 5. Quick Reference Cheat Table

| Functionality | CodeIgniter 3 (PHP) | GoIgniter (Go) | Notes |
|---|---|---|---|
| **Base Controller** | `class Home extends CI_Controller` | `type Home struct { core.Controller }` | Controller struct embedding |
| **Action (Explicit)** | `public function index()` | `func (h *Home) Index(c *core.Context) error` | Explicit route handler |
| **Action (AutoRoute)**| `public function index()` | `func (h *Home) Index() error` | Accesses `h.Ctx` |
| **Get Query Param** | `$this->input->get('q')` | `c.Query("q")` | Fallback: `c.QueryDefault("q", "1")` |
| **Get POST Param** | `$this->input->post('name')` | `c.Form("name")` | Or `c.FormValue("name")` |
| **Route Param** | `$id` in `function view($id)` | `c.Param("id")` | Parsed int: `c.ParamInt("id")` |
| **Bind Payload** | `json_decode(file_get_contents('php://input'))` | `c.Bind(&structVar)` | Binds JSON or form bodies |
| **AJAX Request** | `$this->input->is_ajax_request()` | `c.Header("X-Requested-With") == "XMLHttpRequest"` | Request header check |
| **Database Select** | `$this->db->get('users')->result()` | `var users []User; database.Table("users").Get(&users)` | Scans into slice of structs |
| **Database First** | `$this->db->get('users')->row()` | `var user User; database.Table("users").First(&user)` | Scans into single struct |
| **Database Where** | `$this->db->where('id', $id)` | `database.Table("users").Where("id", id)` | Operator: `Where("age", ">", 18)` |
| **Database Insert** | `$this->db->insert('users', $data)` | `database.Table("users").Insert(data)` | Returns `error` |
| **Database Insert ID**| `$this->db->insert_id()` | `id, err := database.Table("users").InsertGetId(data)` | Returns `(int64, error)` |
| **Database Update** | `$this->db->update('users', $data)` | `database.Table("users").Where("id", id).Update(data)` | Returns `error` |
| **Database Delete** | `$this->db->delete('users')` | `database.Table("users").Where("id", id).Delete()` | Returns `error` |
| **Transactions** | `$this->db->trans_start(); ...` | `database.Transaction(func(tx *database.DB) error { ... })` | Auto rollback on error |
| **Session Init** | `$autoload['libraries'] = ['session']` | `session.Init(session.Config{...})` | In `main.go` startup |
| **Set Session** | `$this->session->set_userdata('k', $v)` | `sess := session.Get(c); sess.Set("k", v); sess.Save(c)` | Session cookie save |
| **Get Session** | `$this->session->userdata('k')` | `sess := session.Get(c); sess.Get("k")` | `sess.GetString("k")`, `GetInt("k")` |
| **Flash Message** | `$this->session->set_flashdata('m', $v)`| `session.SetFlash(c, "m", v)` | Auto-saves cookie |
| **Get Flash** | `$this->session->flashdata('m')` | `session.GetFlash(c, "m")` | Consumes & clears flash |
| **Destroy Session**| `$this->session->sess_destroy()` | `sess := session.Get(c); sess.Destroy(c)` | Clears session cookie |
| **CSRF Middleware**| `$config['csrf_protection'] = TRUE;` | `app.Use(middleware.CSRF())` | Or `CSRFWithConfig(cfg)` |
| **Render View** | `$this->load->view('home', $data)` | `c.View("home", data)` | Omit `.html` extension |
| **Render with Code**| `$this->output->set_status_header(404); ...` | `c.ViewWithCode(404, "errors/404", data)` | Custom HTTP status code |
| **JSON Response** | `echo json_encode($data)` | `c.JSON(http.StatusOK, data)` | Auto Content-Type header |
| **Redirect** | `redirect('home/index')` | `c.Redirect(http.StatusFound, "/home/index")` | StatusFound (302) |
| **Base URL** | `base_url('css/style.css')` | `helpers.BaseURL("css/style.css")` | System helper |
| **Site URL** | `site_url('user/login')` | `helpers.SiteURL("user/login")` | Alias to BaseURL |
| **Password Verify**| `password_verify($pwd, $hash)` | `bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))` | Bcrypt binary-compatible |
| **File Upload** | `$this->upload->do_upload('pic')` | `file, hdr, err := c.Request.FormFile("pic")` | Or `upload.New(cfg).Do("pic", c.Request)` |

---

## 6. Critical Migration Gotchas & Anti-Patterns

### 1. Go Struct Field Exportability
In Go templates (`html/template`) and JSON encoding (`json.Marshal`), **unexported (lowercase) fields are invisible**.
- ❌ **Wrong**: `type User struct { name string; email string }` -> Templates cannot read `{{ .name }}`!
- ✅ **Correct**: `type User struct { Name string; Email string }` -> Templates read `{{ .Name }}`.

### 2. State & Concurrency Safety
PHP runs one process per HTTP request (`share-nothing` model). Go runs requests concurrently across lightweight goroutines in a single persistent process.
- ❌ **Never** store request-scoped data in global variables or controller struct pointer fields.
- ✅ **Always** pass request state through `*core.Context` or local method variables.

### 3. Loose Types vs Strict Types
In PHP, `0`, `""`, `NULL`, and `FALSE` are often checked interchangeably (`empty()`, `if ($val)`).
In Go, types are explicit. When a database column is nullable, represent it as:
- A pointer: `*string`, `*int64`, `*time.Time`.
- Or SQL null types: `sql.NullString`, `sql.NullInt64`.
- Check `if val != nil` before dereferencing!

### 4. Template Extension Matching
In CI3, `$this->load->view('header')` looks for `header.php`. In GoIgniter:
- The transpiled template files on disk have the `.html` extension (e.g. `app/views/header.html`, `app/views/users/index.html`).
- Reference partial templates within Go templates using the filename: `{{ template "header.html" . }}`.
- Reference templates in controllers using `c.View()` or `c.ViewWithCode()` **omitting the `.html` extension** (e.g. `c.View("header", data)` or `c.View("users/index", data)`).

### 5. Bcrypt Hashes are Binary-Compatible
Do not reset user passwords! CI3's `password_hash()` generates standard modular crypt format bcrypt strings (`$2y$...`). Go's `golang.org/x/crypto/bcrypt` can verify `$2y$` and `$2a$` hashes natively.
