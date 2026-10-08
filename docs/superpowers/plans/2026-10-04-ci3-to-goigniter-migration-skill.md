# CI3 to GoIgniter Migration Skill Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a complete, cross-agent skill (`ci3-to-goigniter`) equipped with embedded CLI automation scripts to migrate CodeIgniter 3 PHP applications into GoIgniter with minimum token usage and zero hallucinations.

**Architecture:** A standalone skill folder under `skills/ci3-to-goigniter/` featuring `SKILL.md` (agent SOP), `references/mapping-cheatsheet.md` (API dictionary), and `scripts/` containing zero-token deterministic Go tools (`inspect.go`, `sql2struct.go`, `view2gotpl.go`).

**Tech Stack:** Go standard library (no external dependencies), regex parsing, SQL DDL tokenizer, HTML/template transformation.

---

### Task 1: Reference Mapping Cheatsheet

**Files:**
- Create: `skills/ci3-to-goigniter/references/mapping-cheatsheet.md`

- [ ] **Step 1: Write `mapping-cheatsheet.md`**

Create `skills/ci3-to-goigniter/references/mapping-cheatsheet.md` covering all CodeIgniter 3 APIs mapped to GoIgniter:
- Input handling (`$this->input->post/get/server` $\rightarrow$ `c.Form`, `c.Query`, `c.QueryInt`, `c.Header`)
- Response handling (`$this->output`, `echo json_encode` $\rightarrow$ `c.JSON`, `c.String`, `c.HTML`, `c.Redirect`)
- Session management (`$this->session->userdata/set_userdata/set_flashdata` $\rightarrow$ `session.Get`, `session.Set`, `session.SetFlash`, `session.GetFlash`)
- Query Builder (`$this->db->get/where/insert/update/delete/join/order_by/limit` $\rightarrow$ GoIgniter DB builder)
- Views and layout rendering (`$this->load->view` $\rightarrow$ `c.HTML`)
- Upload and file validation (`$this->load->library('upload')` $\rightarrow$ `c.FormFile`, `upload.New`)

- [ ] **Step 2: Commit**

```bash
git add skills/ci3-to-goigniter/references/mapping-cheatsheet.md
git commit -m "feat(skills): add CI3 to GoIgniter API mapping cheatsheet"
```

---

### Task 2: Project Inspection Script (`scripts/inspect.go`)

**Files:**
- Create: `skills/ci3-to-goigniter/scripts/inspect.go`
- Create: `skills/ci3-to-goigniter/scripts/inspect_test.go`

- [ ] **Step 1: Write the failing test for `inspect.go`**

Create `skills/ci3-to-goigniter/scripts/inspect_test.go` testing scanning of a dummy CI3 directory with `controllers/`, `models/`, `views/`, and `config/routes.php`.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./skills/ci3-to-goigniter/scripts -run TestInspect`
Expected: FAIL (compilation error, functions not defined)

- [ ] **Step 3: Implement `scripts/inspect.go`**

Implement scanner in `skills/ci3-to-goigniter/scripts/inspect.go` using `os`, `path/filepath`, and `regexp` to scan:
- `application/controllers/*.php` for `function action_name(...)`
- `application/models/*.php` for model names
- `application/views/**/*.php` for view templates
- `application/config/routes.php` for `$route['...']`
- Output Markdown report or JSON manifest.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./skills/ci3-to-goigniter/scripts -run TestInspect`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add skills/ci3-to-goigniter/scripts/inspect.go skills/ci3-to-goigniter/scripts/inspect_test.go
git commit -m "feat(skills): add CI3 codebase inspect tool and tests"
```

---

### Task 3: SQL to Struct Generator (`scripts/sql2struct.go`)

**Files:**
- Create: `skills/ci3-to-goigniter/scripts/sql2struct.go`
- Create: `skills/ci3-to-goigniter/scripts/sql2struct_test.go`

- [ ] **Step 1: Write the failing test for `sql2struct.go`**

Create `skills/ci3-to-goigniter/scripts/sql2struct_test.go` providing sample SQL `CREATE TABLE users (...)` and `CREATE TABLE posts (...)`, testing conversion into Go struct definitions with `json` and `db` tags.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./skills/ci3-to-goigniter/scripts -run TestSQL2Struct`
Expected: FAIL (functions not defined)

- [ ] **Step 3: Implement `scripts/sql2struct.go`**

Implement SQL parser and Go code generator in `skills/ci3-to-goigniter/scripts/sql2struct.go`:
- Parse `CREATE TABLE [IF NOT EXISTS] <name> (...)`
- Extract column name, SQL data type, nullability
- Map SQL data types to Go types (`int`, `int64`, `string`, `time.Time`, `bool`, `float64`)
- Generate Go struct with tags and model constructor.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./skills/ci3-to-goigniter/scripts -run TestSQL2Struct`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add skills/ci3-to-goigniter/scripts/sql2struct.go skills/ci3-to-goigniter/scripts/sql2struct_test.go
git commit -m "feat(skills): add SQL to Go struct generator and tests"
```

---

### Task 4: View Transpiler Script (`scripts/view2gotpl.go`)

**Files:**
- Create: `skills/ci3-to-goigniter/scripts/view2gotpl.go`
- Create: `skills/ci3-to-goigniter/scripts/view2gotpl_test.go`

- [ ] **Step 1: Write the failing test for `view2gotpl.go`**

Create `skills/ci3-to-goigniter/scripts/view2gotpl_test.go` testing conversion of:
- `<?= $title ?>` $\rightarrow$ `{{ .Title }}`
- `<?= $user->email ?>` $\rightarrow$ `{{ .User.Email }}`
- `<?php if ($logged_in): ?> ... <?php endif; ?>` $\rightarrow$ `{{ if .LoggedIn }} ... {{ end }}`
- `<?php foreach ($users as $user): ?> ... <?php endforeach; ?>` $\rightarrow$ `{{ range .Users }} ... {{ end }}`
- `<?= base_url('css/main.css') ?>` $\rightarrow$ `/css/main.css`

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./skills/ci3-to-goigniter/scripts -run TestView2GoTpl`
Expected: FAIL (functions not defined)

- [ ] **Step 3: Implement `scripts/view2gotpl.go`**

Implement regex-based transpiler in `skills/ci3-to-goigniter/scripts/view2gotpl.go`:
- Variable echo replacement
- Conditional conversion (`if`, `else`, `endif`)
- Loop conversion (`foreach`, `endforeach`)
- Base URL substitution
- Complex unconvertible logic wrapping (`{{/* TODO_MIGRATE: ... */}}`)

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./skills/ci3-to-goigniter/scripts -run TestView2GoTpl`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add skills/ci3-to-goigniter/scripts/view2gotpl.go skills/ci3-to-goigniter/scripts/view2gotpl_test.go
git commit -m "feat(skills): add PHP view to Go template transpiler and tests"
```

---

### Task 5: Primary Skill Definition (`SKILL.md`)

**Files:**
- Create: `skills/ci3-to-goigniter/SKILL.md`

- [ ] **Step 1: Write `SKILL.md`**

Create `skills/ci3-to-goigniter/SKILL.md` with:
- YAML frontmatter (`name: ci3-to-goigniter`, `description: Migrate existing PHP CodeIgniter 3 projects to GoIgniter framework systematically with zero hallucination and minimal tokens`)
- Clear agent SOP:
  - Step 1: Run `inspect.go` to get the inventory.
  - Step 2: Run `sql2struct.go` on DB schema.
  - Step 3: Run `view2gotpl.go` on views.
  - Step 4: Port controllers method-by-method with reference to `mapping-cheatsheet.md`.
  - Step 5: Port routes and configuration.
  - Step 6: Verify build with `go build` and `go test ./...`.
- Examples and edge case handling.

- [ ] **Step 2: Commit**

```bash
git add skills/ci3-to-goigniter/SKILL.md
git commit -m "feat(skills): add ci3-to-goigniter main SKILL.md definition"
```

---

### Task 6: End-to-End Verification Fixture

**Files:**
- Create: `skills/ci3-to-goigniter/testdata/ci3_fixture/` (sample miniature CI3 project with routes, controller, model, view, schema.sql)
- Create: `skills/ci3-to-goigniter/scripts/e2e_test.go`

- [ ] **Step 1: Write fixture and e2e test**

Create a miniature CI3 project in `testdata/ci3_fixture` and an end-to-end test in `scripts/e2e_test.go` that runs `inspect`, `sql2struct`, and `view2gotpl` against it and verifies the output files match expected GoIgniter formats.

- [ ] **Step 2: Run the end-to-end test**

Run: `go test -v ./skills/ci3-to-goigniter/scripts -run TestE2E`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add skills/ci3-to-goigniter/testdata skills/ci3-to-goigniter/scripts/e2e_test.go
git commit -m "test(skills): add end-to-end migration fixture and test"
```
