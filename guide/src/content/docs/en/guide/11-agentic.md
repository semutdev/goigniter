---
title: Agentic Migration
description: Migrate CodeIgniter 3 projects to Goigniter using AI agents.
sidebar:
  order: 11
---

Goigniter includes an agentic skill for automated migration from CodeIgniter 3. This skill is installed as a prompt file that AI coding agents (Claude Code, Codex, Antigravity) can follow to migrate CI3 projects.

## Installation

Install the skill for your AI agent:

### Claude Code

```bash
mkdir -p ~/.claude/skills/
curl -o ~/.claude/skills/ci3-to-goigniter.md https://raw.githubusercontent.com/semutdev/goigniter/main/skills/ci3-to-goigniter/SKILL.md
```

### Codex

```bash
mkdir -p .codex/skills/
curl -o .codex/skills/ci3-to-goigniter.md https://raw.githubusercontent.com/semutdev/goigniter/main/skills/ci3-to-goigniter/SKILL.md
```

### Antigravity

Add the skill file to your project's `.antigravity/skills/` directory.

After installation, ask your AI agent:
> "Run the CI3 to Goigniter migration skill"

## Migration Workflow

The skill follows a 6-phase workflow:

| Phase | Description |
|-------|-------------|
| 1. Analyze | Scan the CI3 codebase structure |
| 2. Plan | Map each CI3 component to Goigniter |
| 3. Config | Migrate database, environment, and config |
| 4. Models | Convert CI3 models to GORM models |
| 5. Controllers | Convert CI3 controllers with routes |
| 6. Views | Convert PHP templates to Go templates |

## Migration Mapping

### Controllers

| CI3 | Goigniter |
|-----|-----------|
| `application/controllers/` | `application/controllers/` |
| `$this->load->view()` | `w.Ctx.View()` |
| `$this->input->post()` | `w.Ctx.FormValue()` |
| `$this->input->get()` | `w.Ctx.Query()` |
| `$this->uri->segment()` | `w.Ctx.Param()` |
| `$this->session->userdata()` | `helpers.SessionGet()` |
| `redirect()` | `w.Ctx.Redirect()` |

### Models

| CI3 | Goigniter |
|-----|-----------|
| `$this->db->get()` | `config.DB.Find()` |
| `$this->db->get_where()` | `config.DB.Where().Find()` |
| `$this->db->insert()` | `config.DB.Create()` |
| `$this->db->update()` | `config.DB.Save()` |
| `$this->db->delete()` | `config.DB.Delete()` |
| `$this->db->query()` | `config.DB.Raw()` |

### Views

| CI3 | Goigniter |
|-----|-----------|
| `<?= $title ?>` | `{{ .Title }}` |
| `<?php foreach(): ?>` | `{{ range .Items }}` |
| `<?php if(): ?>` | `{{ if .Cond }}` |
| `<?= base_url() ?>` | `helper "base_url"` |
| `<?= site_url() ?>` | `helper "site_url"` |

## Automated CLI Tool
The repository includes a CLI migration tool:

```bash
go run setup/migration/main.go --source=/path/to/ci3-project --output=./migrated-app
```

Flags:
- `--source` — path to CI3 project directory
- `--output` — output directory for the Goigniter project
- `--db-dsn` — database DSN for the new project

## Important Notes

- The migration assistant helps with code translation, but manual review is essential
- Authentication and authorization logic may need adjustments
- Custom CI3 libraries need manual translation
- Test the migrated application thoroughly before deploying

---

Migrating a legacy project is a great opportunity to modernize. Goigniter bridges the gap between familiar patterns and Go's performance.
