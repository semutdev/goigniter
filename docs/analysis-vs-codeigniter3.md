# Analisis GoIgniter vs CodeIgniter 3

> Dokumen analisis feature-parity antara **GoIgniter v0.2.0** dan **CodeIgniter 3**,
> beserta saran fitur selanjutnya dalam bentuk checklist.

---

## 📊 Ringkasan Persentase

| Kategori | Persentase | Keterangan |
|----------|-----------:|------------|
| **Routing** | 🟢 95% | 7 HTTP method, route group, param, wildcard, AutoRoute, static |
| **Controller** | 🟢 85% | Base controller, AutoRoute reflection, middleware per-method |
| **Model / ORM** | 🟡 50% | Query builder + struct scanning, tapi bukan full ORM |
| **View / Template** | 🟢 80% | html/template, hot reload, layout/partial |
| **Form Validation** | 🔴 0% | belum ada |
| **Session** | 🟢 80% | HMAC-SHA256 + AES-256-GCM, flash, cookie config |
| **Database / Query Builder** | 🟡 70% | Builder lengkap, MySQL+SQLite, transaksi, tanpa migration/seeder core |
| **File Upload** | 🟢 90% | MIME validation, blocklist, path traversal, double extension |
| **Image Manipulation** | 🟡 75% | resize, crop, fit, fill, rotate, thumbnail |
| **Pagination** | 🔴 10% | hanya template partial, no library |
| **Security** | 🟢 85% | CSRF, security headers, rate limit, body limit, upload safety |
| **Middleware** | 🟢 90% | logger, recovery, CORS, CSRF, ratelimit, security, auth |
| **Config / Environment** | 🟡 40% | env vars, no central config loader di core |
| **Helpers** | 🟡 40% | url, template, debug saja |
| **Email / Mailer** | 🔴 0% | belum ada |
| **Caching** | 🔴 0% | belum ada |
| **Encryption Library** | 🟡 60% | session-only, no standalone library |
| **CLI Tooling** | 🟡 30% | bash setup wizard, no Go CLI binary |
| **Testing Helpers** | 🟡 50% | core test, no app test helper |
| **Documentation** | 🟡 65% | README, security.md, 10 guide pages, docs plans |
| **i18n / Language** | 🔴 0% | belum ada |
| **Profiler / Benchmark** | 🔴 0% | belum ada |
| **DB Drivers** | 🟡 33% | MySQL + SQLite (CI3 punya ±6 driver) |
| **Logging (structured)** | 🟡 40% | logger middleware, no structured log library |

### 🎯 Estimasi Total: **± 52%** dibanding CodeIgniter 3

GoIgniter sudah memiliki **core foundation yang solid** (routing, controller, template, query builder, session, middleware keamanan) tetapi masih kekurangan banyak *utility library* yang membuat CI3 populer: form validation, email, pagination, caching, profiler, dan i18n.

---

## 📦 Tabel Fitur yang Sudah Ada

### 1. Core System (`system/core/`)

| Fitur | Status | File | Catatan |
|-------|:------:|------|---------|
| Application (net/http stdlib) | ✅ | `goigniter.go`, `application.go` | Zero external HTTP dep |
| Radix-tree Router custom | ✅ | `router.go`, `internal/radix/` | Static > param priority, backtracking |
| HTTP GET/POST/PUT/DELETE | ✅ | `application.go` | Semua 7 method |
| HTTP PATCH/OPTIONS/HEAD | ✅ | `application.go` | Commit 6f68d7f |
| Route Group + Nesting | ✅ | `application.go` | `g.Group(...)` nested |
| Path Parameters `:id` | ✅ | `radix.go` | Multi-param |
| Wildcard `*filepath` | ✅ | `radix.go` | Static file serving |
| Static File Serving | ✅ | `application.go` | `Static(prefix, root)` |
| AutoRoute (reflection) | ✅ | `registry.go` | `/{controller}/{method}`, `Routes()`, `AllowedMethods()`, `MiddlewareFor()` |
| Context Pool (`sync.Pool`) | ✅ | `application.go` | Re-use Context |
| Context Helpers | ✅ | `context.go` | JSON/HTML/String/Redirect/File/Blob/NoContent |
| Input Helpers | ✅ | `context.go` | Param/Query/Form/QueryInt/Bind/BodyWithLimit |
| Bind (JSON) + Body Limit | ✅ | `context.go` | `BindWithLimit`, `BodyWithLimit` |
| Cookie get/set | ✅ | `context.go` | |
| IP detection (X-Forwarded-For) | ✅ | `context.go` | |
| View/Render to string (partial) | ✅ | `context.go` | `View`, `ViewWithCode`, `Render` |
| Base Controller + Loader | ✅ | `controller.go` | `Load.Model/Library/View/Helper` |
| ASCII Banner | ✅ | `banner.go` | |
| Version info | ✅ | `version.go` | v0.2.0 |

### 2. Middleware (`system/middleware/`)

| Fitur | Status | File | Catatan vs CI3 |
|-------|:------:|------|----------------|
| Logger (colored) | ✅ | `logger.go` | CI3: hooks, logging library |
| Recovery (panic) | ✅ | `recovery.go` | CI3: no built-in |
| CORS | ✅ | `cors.go` | CI3: manual |
| CSRF Protection | ✅ | `csrf.go` | Double-submit, constant-time compare |
| Rate Limiting | ✅ | `ratelimit.go` | Sliding window, blok race-safe |
| Security Headers | ✅ | `security.go` | XSS/NoSniff/FrameGuard/HSTS/CSP/Referrer-Policy |
| Auth (Basic + Bearer) | ✅ | `auth.go` | Custom validator, redirect |

### 3. Libraries (`system/libraries/`)

| Fitur | Status | File | CI3 Equivalent |
|-------|:------:|------|---------------|
| Database (DB, Builder, Result) | ✅ | `database/database.go`, `builder.go`, `result.go` | CI3 Active Record |
| SELECT + columns | ✅ | `builder.go` | ✅ |
| WHERE/OR/IN/NOT IN/NULL/Raw | ✅ | `builder.go` | ✅ |
| JOIN (Inner/Left/Right) | ✅ | `builder.go` | ✅ |
| ORDER BY / LIMIT / OFFSET | ✅ | `builder.go` | ✅ |
| GROUP BY / HAVING | ✅ | `builder.go` | ✅ |
| Aggregates (Count/Sum/Avg/Min/Max) | ✅ | `builder.go` | ✅ |
| INSERT / InsertGetID / InsertStruct | ✅ | `builder.go`, `result.go` | ✅ |
| UPDATE / UpdateStruct / DELETE | ✅ | `builder.go`, `result.go` | ✅ |
| Raw Query | ✅ | `result.go` | ✅ |
| Transactions (Begin/Commit/Rollback/Transaction(fn)) | ✅ | `database.go` | ✅ |
| ToSQL (debug) | ✅ | `builder.go` | ✅ |
| Struct scanning via reflection (db/json tags) | ✅ | `result.go` | Embedded struct support |
| Driver Registry | ✅ | `drivers/driver.go` | |
| MySQL Driver | ✅ | `drivers/mysql.go` | go-sql-driver/mysql |
| SQLite Driver (pure Go, no CGO) | ✅ | `drivers/sqlite.go` | modernc.org/sqlite |
| Session (cookie, HMAC-SHA256 + AES-256-GCM) | ✅ | `session/session.go` | CI3 Session |
| Flash Messages | ✅ | `session/session.go` | ✅ |
| Session Key Generator | ✅ | `session/key.go` | |
| Upload (secure, MIME detection) | ✅ | `upload/upload.go` | CI3 Upload |
| Dangerous extension blocklist | ✅ | `upload/upload.go` | ⬆️ more strict |
| Double extension / null-byte prevention | ✅ | `upload/upload.go` | ⬆️ |
| Path traversal prevention | ✅ | `upload/upload.go` | |
| Image Processing (resize/crop/fit/fill/rotate) | ✅ | `upload/image.go` | CI3 Image_lib (GD/IM/NetPBM) |
| Thumbnail generation | ✅ | `upload/image.go` | |
| JPEG/PNG/GIF support | ✅ | `upload/image.go` | |

### 4. Helpers (`system/helpers/`)

| Fitur | Status | File |
|-------|:------:|------|
| URL helpers (BaseURL/SiteURL/AssetURL) | ✅ | `url.go` |
| Template Funcs (base_url/site_url/asset_url) | ✅ | `url.go`, `template.go` |
| String funcs (safe/upper/lower/title/trim/contains/replace/split/join) | ✅ | `template.go` |
| Conditional helpers (default/eq/ne) | ✅ | `template.go` |
| Debug (JSON pretty / reflection) | ✅ | `debug.go` |

### 5. Template Engine

| Fitur | Status | Catatan |
|-------|:------:|---------|
| html/template parsing | ✅ | `application.go` TemplateEngine |
| Hot Reload | ✅ | `Reload: true` re-parses per render |
| Custom FuncMap | ✅ | `LoadTemplatesWithFuncs` |
| Layout / Partial pattern | ✅ | `Render(name, data)` to string |
| Multiple layouts | ✅ | Examples + starter demos |

### 6. Setup Wizard / CLI Tooling

| Fitur | Status | Catatan |
|-------|:------:|---------|
| Bash setup wizard (interactive + non-interactive) | ✅ | `setup/setup.sh` |
| Template processor | ✅ | `setup/installer.sh` (25 .tmpl) |
| GitHub-curl installer | ✅ | `curl -sSL ... \| bash` |
| Generated admin app (login + dashboard + CRUD) | ✅ | Default: admin@admin.com / password |
| DB choice (SQLite / MySQL) | ✅ | `--db` flag |
| APP_KEY auto-generate | ✅ | `openssl rand` |

### 7. Testing

| Fitur | Status | Catatan |
|-------|:------:|---------|
| Core routing tests | ✅ | `application_test.go` |
| Context tests | ✅ | `context_test.go` |
| Radix tree tests | ✅ | `radix_test.go` |
| Middleware tests | ✅ | `middleware_test.go`, `security_test.go` |
| In-memory SQLite DB tests | ✅ | `:memory:` |
| httptest integration | ✅ | |

### 8. Documentation

| Fitur | Status | Catatan |
|-------|:------:|---------|
| README | ✅ | ID + EN mix |
| Security Guide | ✅ | `docs/security.md` |
| Astro Star doc site (10 pages) | ✅ | `guide/` |
| Design Plans (11 docs) | ✅ | `docs/plans/` |
| Setup Tool design spec | ✅ | `docs/superpowers/` |

### 9. Examples

| Contoh | Status | Desc |
|--------|:------:|------|
| `examples/simple` | ✅ | JSON/HTML/param/auth |
| `examples/autoroute` | ✅ | Reflection CRUD |
| `examples/views` | ✅ | Layout + static |
| `examples/database` | ✅ | Builder demo |
| `examples/full-crud` | ✅ | Ion Auth-style, admin panel, HTMX |

---

## 📋 Status Kemudahan (Developer Experience) vs CI3

| Aspek | CI3 | GoIgniter | Selisih |
|-------|-----|-----------|---------|
| One-command install | ✅ (zip) | ✅ (curl setup.sh) | setara |
| Auto-routing | ✅ | ✅ AutoRoute | setara |
| Controller inheritance | ✅ | ✅ base Controller | setara |
| Migration builder | ⚠️ manual SQL | ❌ | belom |
| Seeder | ✅ | ⚠️ bash `DB_SEED=true` | basic |
| Config file load | ✅ `config.php` | ⚠️ env only | kurang |
| `form_open()` helper | ✅ | ❌ | belom |
| Validation set_rules | ✅ | ❌ | besar |
| Flashdata | ✅ | ✅ | setara |
| Active Record chain | ✅ | ✅ Builder | setara |
| Template parser | ✅ Parser Class | ⚠️ html/template | beda style |
| Make command / CLI | ❌ native | ❌ bash only | setara-rendah |
| Scaffold CRUD | manual | ✅ setup wizard | ⬆️ lebih baik |
| Profiler toolbar | ✅ | ❌ | belom |

---

## 🚀 Saran Fitur Selanjutnya — Checklist

Berikut roadmap prioritas fitur yang sebaiknya dikembangkan untuk mendekati / melampaui feature-parity CI3.

### Prioritas Tinggi (High Priority)

- [ ] **Form Validation Library** (`system/libraries/validation/`)
  - [ ] `set_rules(field, label, rules)` style CI3
  - [ ] Rule: required, min_length, max_length, exact_length, valid_email, regex_match, matches, is_unique, alpha, alpha_numeric, alpha_dash, numeric, integer, decimal, in_list
  - [ ] Custom callback rule
  - [ ] Error messages per field + multi-bahasa
  - [ ] `validation_errors()` / `form_error(field)` helper
  - [ ] Integration dengan Context `Bind()` & template engine
- [ ] **Pagination Library** (`system/libraries/pagination/`)
  - [ ] `Pagination` struct dengan config (PerPage, TotalRows, BaseURL, Segment)
  - [ ] `CreateLinks()` returns HTML
  - [ ] Template helper `{{ .pagination | safe }}`
  - [ ] Query builder integration (`Builder.Paginate(page, perPage)`)
- [ ] **Email Library** (`system/libraries/email/`)
  - [ ] SMTP / sendmail / mail driver
  - [ ] HTML + plain text body, attachments, CC/BCC
  - [ ] Template-based mail (`Mail.SendTemplate("welcome", data)`)
- [ ] **Central Config Loader** (`system/core/config.go`)
  - [ ] Load multi config file (config.go, database.go, app.go)
  - [ ] Env override (dev / prod / test)
  - [ ] `Config.Get("key", "default")` accessor

### Prioritas Sedang (Medium Priority)

- [ ] **Caching Library** (`system/libraries/cache/`)
  - [ ] Driver interface: File, Memory (sync.Map), Redis
  - [ ] `Cache.Set/Get/Delete/Remember/Flush` API
  - [ ] TTL support
- [ ] **Standalone Encryption Library** (`system/libraries/encryption/`)
  - [ ] `Encrypt()` / `Decrypt()` generic
  - [ ] Driver: AES-256-GCM (default),libsodium / Fernet
- [ ] **Go-based CLI Binary** (`cmd/goigniter/`)
  - [ ] `goigniter new <name>` generate project
  - [ ] `goigniter serve` dev server dengan auto-rebuild (air/watch)
  - [ ] `goigniter make:controller|model|migration|middleware`
  - [ ] `goigniter migrate` / `goigniter migrate:rollback`
  - [ ] `goigniter db:seed`
- [ ] **Migration System** (`system/libraries/migration/`)
  - [ ] Schema migration files timestamped
  - [ ] Up/down, status, refresh
  - [ ] Builder Schema (CreateTable, AddColumn, etc.)
- [ ] **Logger Library (structured)** (`system/libraries/log/`)
  - [ ] Severity level (DEBUG/INFO/WARN/ERROR)
  - [ ] Driver: file (rotating), stdout, syslog
  - [ ] Structured fields & JSON output mode
- [ ] **Pagination + Form helpers** (`system/helpers/`)
  - [ ] `form_open(action, attrs)`, `form_input`, `form_dropdown`, `form_checkbox`
  - [ ] `anchor(url, text, attrs)`, `ul()`, `ol()`
  - [ ] `set_value(field)`, `set_select`, `set_checkbox` (sticky form)
  - [ ] `heading()`, `br()`, `nbs()`
- [ ] **Database Driver tambahan** 
  - [ ] PostgreSQL driver
  - [ ] SQL Server (mssql) driver
  - [ ] MongoDB (optional NoSQL)

### Prioritas Rendah (Low Priority / Nice-to-have)

- [ ] **Profiler Toolbar** (`system/libraries/profiler/`)
  - [ ] Endpoint / query timing
  - [ ] Memory usage
  - [ ] Toggle via `APP_ENV=development`
  - [ ] Browser toolbar (CI3-style bottom bar)
- [ ] **Benchmark Library**
  - [ ] `Benchmark.Mark/Elapsed` untuk profiling manual
- [ ] **i18n / Language Library** (`system/libraries/language/`)
  - [ ] `lang.Load("form_validation", "english")`
  - [ ] `lang.Line("required")` lookup
  - [ ] Multi-bahasa file per-bundle
- [ ] **Captcha Library** (`system/libraries/captcha/`)
  - [ ] Image captcha (math/text)
  - [ ] Audio captcha
- [ ] **Zip / Archive Library** (`system/libraries/zip/`)
  - [ ] Create + extract ZIP
- [ ] **FTP Library** (`system/libraries/ftp/`)
- [ ] **Calendar Library** (`system/libraries/calendar/`)
- [ ] **Unit Testing Helper** (`system/libraries/unit_test/`)
  - [ ] `Test()` registration, `report()` HTML summary
  - [ ] Integration dengan `go test`
- [ ] **Template Parser Class** (CI3-style pseudo-tag)
  - [ ] `{username}` → `{{ .username }}`
  - [ ] For designer / non-Go users
- [ ] **Hooks System**
  - [ ] Pre-system / post-controller hook points
  - [ ] Auto-load user hook bundles from `application/hooks/`
- [ ] **OWASP / Web Application Firewall middleware**
  - [ ] SQLi / XSS pattern detection
  - [ ] User-agent filter
- [ ] **WebSocket / SSE library** (modern feature di luar CI3)
  - [ ] Real-time pub-sub helper
- [ ] **API toolkit** (di luar CI3, modern)
  - [ ] REST resource controller (`Index/Show/Create/Update/Delete`)
  - [ ] Built-in JSON:API / OpenAPI/docs
  - [ ] API versioning helper (v1/v2 route group)
- [ ] **ORM (optional, alternative ke Builder)**
  - [ ] Model relation (HasMany / BelongsTo)
  - [ ] Eager/lazy loading
  - [ ] Auto-migrate from struct tags
- [ ] **Task Scheduler / Cron** (di luar CI3)
  - [ ] `goigniter schedule:run` CLI command
  - [ ] Background job dispatcher

### Kategori "Polish & Quality"

- [ ] Update `README.md` (CI3 sudah lengkap; goigniter masih 'Under Development')
- [ ] Perbaiki `CLAUDE.md` (masih nyebut Echo / GORM — stale)
- [ ] Tambah `CONTRIBUTING.md` + code style guide
- [ ] Tambah CI/CD pipeline (GitHub Actions: lint, test, build)
- [ ] Tambah `.golangci.yml` lint config
- [ ] Tambah test coverage report (codecov)
- [ ] Tambah `CHANGELOG.md` (Keep a Changelog format)
- [ ] Tag release semver (`v0.3.0`, `v0.4.0`, dst.)
- [ ] Godoc publik di pkg.go.dev
- [ ] Playground / demo website live

---

## 📈 Estimasi Pencapaian ke 80% Parity

Bila **high priority** (Form Validation, Pagination, Email, Config) selesai → naik ke **± 65%**.
Bila **medium priority** (Caching, Encryption, CLI binary, Migration, Logger, Driver) selesai → **± 78%**.
Bila **low priority** (Profiler, i18n, additional helpers) selesai → **± 85-90%** parity CodeIgniter 3.

### Milestone yang disarankan

| Versi Target | Fokus | Estimasi Pencapaian |
|--------------|-------|---------------------|
| `v0.3.0` | Form Validation + Pagination + Config Loader | ~60% |
| `v0.4.0` | Email + Caching + CLI Binary (new/serve/make) | ~70% |
| `v0.5.0` | Migration System + Logger + additional DB Drivers | ~75% |
| `v0.6.0` | Scheduler + CLI `migrate` + Profiler | ~78% |
| `v0.7.0` | i18n + Form/HTML helpers + Encryption lib | ~82% |
| `v1.0.0` | Polishing, full docs, examples real-world, release readme | ~85-90% |

---

## 📝 Catatan

- **Gerak cepat**: GoIgniter sudah unggul dibanding CI3 dalam beberapa aspek:
  - AutoRoute reflection (CI3 harus konfigurasi manual)
  - Setup wizard 1 command menghasilkan admin panel lengkap
  - XSS/CSRF/Security headers built-in (CI3 manual)
  - Hot reload template
  - Image processing berbasis GPU-friendly `golang.org/x/image/draw` (CatmullRom)
  - Zero CGO SQLite (modernc.org/sqlite)
- **Gap terbesar**: **Form Validation** (alasan utama banyak orang pilih CI3), **Email**, dan
  **Pagination**. Ketiganya adalah "killer feature" CI3 yang wajib diprioritaskan.
- **Modern addition** yang tidak ada di CI3 tetapi sebaiknya dipertahankan / ditambah:
  WebSocket/SSE, API toolkit, scheduler — ini membedakan goigniter dari sekadar "CI3 di Go".

---

*Dokumen ini dibuat otomatis berdasarkan analisa source code GoIgniter v0.2.0 (commit 6f68d7f) per tanggal 11 Juli 2026.*
