---
title: Agentic Migration
description: Migrasi project CodeIgniter 3 ke Goigniter dengan AI agent.
sidebar:
  order: 11
---

Goigniter menyertakan skill agentic untuk migrasi otomatis dari CodeIgniter 3. Skill ini dipasang sebagai file prompt yang bisa diikuti AI coding agent (Claude Code, Codex, Antigravity) untuk memigrasi project CI3.

## Instalasi

Pasang skill untuk AI agent kamu:

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

Tambah file skill ke direktori `.antigravity/skills/` project kamu.

Setelah instalasi, minta AI agent kamu:
> "Run the CI3 to Goigniter migration skill"

## Workflow Migrasi

Skill mengikuti 6 fase:

| Fase | Deskripsi |
|------|-----------|
| 1. Analyze | Scan struktur kode CI3 |
| 2. Plan | Petakan setiap komponen CI3 ke Goigniter |
| 3. Config | Migrasi database, environment, dan konfigurasi |
| 4. Models | Konversi model CI3 ke GORM models |
| 5. Controllers | Konversi controller CI3 dengan routes |
| 6. Views | Konversi template PHP ke Go templates |

## Pemetaan Migrasi

### Controllers

| CI3 | Goigniter |
|-----|-----------|
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

## CLI Tool

Repository menyertakan CLI migration:

```bash
go run setup/migration/main.go --source=/path/to/ci3-project --output=./migrated-app
```

## Catatan Penting

- Migration assistant membantu translasi kode, tapi review manual tetap penting
- Logika autentikasi dan otorisasi mungkin perlu penyesuaian
- Library CI3 kustom perlu translasi manual
- Uji aplikasi hasil migrasi secara menyeluruh sebelum deploy

---

Migrasi project lama adalah kesempatan untuk modernisasi. Goigniter menjembatani pattern familiar dengan performa Go.