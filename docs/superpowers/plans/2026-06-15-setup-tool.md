# GoIgniter Setup Tool Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a curl-based setup tool that generates GoIgniter projects with login system and admin dashboard.

**Architecture:** Bash script entry point that downloads templates and generates files using simple placeholder replacement. Templates stored as .tmpl files with `{{.Variable}}` syntax.

**Tech Stack:** Bash, Go templates via sed replacement, GoIgniter framework, HTMX, minimal CSS

---

## File Structure

### Setup Tool Files (to create)
```
setup/
├── setup.sh              # Entry point, interactive wizard
├── installer.sh          # Template processing and file generation
└── templates/
    ├── main.go.tmpl
    ├── env.tmpl
    ├── env.example.tmpl
    ├── gomod.tmpl
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

### Generated Project Files (output)
```
myapp/
├── main.go
├── .env
├── .env.example
├── go.mod
├── application/
│   ├── config/database.go
│   ├── controllers/auth.go
│   ├── controllers/admin/dashboard.go
│   ├── controllers/admin/users.go
│   ├── controllers/admin/settings.go
│   ├── models/user.go
│   ├── models/setting.go
│   ├── views/layouts/main.html
│   ├── views/auth/login.html
│   ├── views/admin/dashboard.html
│   ├── views/admin/users/index.html
│   ├── views/admin/users/_form.html
│   ├── views/admin/users/_row.html
│   ├── views/admin/settings/index.html
│   ├── views/admin/partials/_header.html
│   ├── views/admin/partials/_sidebar.html
│   ├── views/admin/partials/_footer.html
│   ├── views/partials/_messages.html
│   ├── views/partials/_pagination.html
│   └── libs/auth.go
├── public/static/css/style.css
└── database/seed.go
```

---

## Task 1: Setup Script Entry Point

**Files:**
- Create: `setup/setup.sh`

- [ ] **Step 1: Create setup.sh with interactive wizard**

```bash
#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Banner
echo -e "${BLUE}"
echo "🚀 GoIgniter Setup Wizard"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${NC}"

# Defaults
APP_NAME=""
DB_TYPE="sqlite"
DB_HOST="127.0.0.1"
DB_PORT="3306"
DB_USER="root"
DB_PASS=""
DB_NAME=""
FEATURES="login,admin"

# Parse flags
while [[ $# -gt 0 ]]; do
    case $1 in
        --name=*) APP_NAME="${1#*=}" ;;
        --db=*) DB_TYPE="${1#*=}" ;;
        --db-host=*) DB_HOST="${1#*=}" ;;
        --db-port=*) DB_PORT="${1#*=}" ;;
        --db-user=*) DB_USER="${1#*=}" ;;
        --db-pass=*) DB_PASS="${1#*=}" ;;
        --db-name=*) DB_NAME="${1#*=}" ;;
        --features=*) FEATURES="${1#*=}" ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
    shift
done

# Interactive prompts if no name provided
if [ -z "$APP_NAME" ]; then
    read -p "? Project name: " APP_NAME
fi

# Validate app name
if [ -z "$APP_NAME" ]; then
    echo -e "${RED}Error: Project name is required${NC}"
    exit 1
fi

# Set default db name
if [ -z "$DB_NAME" ]; then
    DB_NAME="$APP_NAME"
fi

# Ask for db type if not provided via flag
if [ "$DB_TYPE" == "sqlite" ] && [ -z "$DB_TYPE_PROVIDED" ]; then
    read -p "? Database type (sqlite/mysql) [sqlite]: " DB_INPUT
    if [ "$DB_INPUT" == "mysql" ]; then
        DB_TYPE="mysql"
    fi
fi

# MySQL specific prompts
if [ "$DB_TYPE" == "mysql" ]; then
    read -p "? MySQL host [127.0.0.1]: " DB_HOST_INPUT
    [ -n "$DB_HOST_INPUT" ] && DB_HOST="$DB_HOST_INPUT"
    
    read -p "? MySQL port [3306]: " DB_PORT_INPUT
    [ -n "$DB_PORT_INPUT" ] && DB_PORT="$DB_PORT_INPUT"
    
    read -p "? MySQL user [root]: " DB_USER_INPUT
    [ -n "$DB_USER_INPUT" ] && DB_USER="$DB_USER_INPUT"
    
    read -s -p "? MySQL password: " DB_PASS
    echo
    
    read -p "? MySQL database name [$DB_NAME]: " DB_NAME_INPUT
    [ -n "$DB_NAME_INPUT" ] && DB_NAME="$DB_NAME_INPUT"
fi

# Generate APP_KEY (32 chars)
APP_KEY=$(openssl rand -base64 32 | tr -d '\n' | cut -c1-32)

# Module path
MODULE_PATH="github.com/$(whoami)/$APP_NAME"

# DB DSN
if [ "$DB_TYPE" == "sqlite" ]; then
    DB_DSN="./app.db"
else
    if [ -n "$DB_PASS" ]; then
        DB_DSN="$DB_USER:$DB_PASS@tcp($DB_HOST:$DB_PORT)/$DB_NAME?charset=utf8mb4&parseTime=True&loc=Local"
    else
        DB_DSN="$DB_USER@tcp($DB_HOST:$DB_PORT)/$DB_NAME?charset=utf8mb4&parseTime=True&loc=Local"
    fi
fi

echo ""
echo -e "${GREEN}Configuration:${NC}"
echo "  Project: $APP_NAME"
echo "  Module: $MODULE_PATH"
echo "  Database: $DB_TYPE"
[ "$DB_TYPE" == "mysql" ] && echo "  DSN: $DB_DSN"
echo "  Features: $FEATURES"
echo ""

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATES_DIR="$SCRIPT_DIR/templates"

# Create project directory
if [ -d "$APP_NAME" ]; then
    echo -e "${RED}Error: Directory '$APP_NAME' already exists${NC}"
    exit 1
fi

mkdir -p "$APP_NAME"

# Source installer
source "$SCRIPT_DIR/installer.sh"

echo ""
echo -e "${GREEN}✓ Done!${NC}"
echo ""
echo "Run: cd $APP_NAME && go mod tidy && go run main.go"
echo ""
```

- [ ] **Step 2: Make setup.sh executable**

```bash
chmod +x setup/setup.sh
```

- [ ] **Step 3: Commit setup entry point**

```bash
git add setup/setup.sh
git commit -m "feat(setup): add setup.sh entry point with interactive wizard"
```

---

## Task 2: Installer Script

**Files:**
- Create: `setup/installer.sh`

- [ ] **Step 1: Create installer.sh with template processing**

```bash
#!/bin/bash

# Installer functions for GoIgniter setup
# This file is sourced by setup.sh

# Replace template variables
replace_vars() {
    local content="$1"
    content="${content//\{\{\.AppName\}\}/$APP_NAME}"
    content="${content//\{\{\.ModulePath\}\}/$MODULE_PATH}"
    content="${content//\{\{\.DbType\}\}/$DB_TYPE}"
    content="${content//\{\{\.DbDsn\}\}/$DB_DSN}"
    content="${content//\{\{\.DbHost\}\}/$DB_HOST}"
    content="${content//\{\{\.DbPort\}\}/$DB_PORT}"
    content="${content//\{\{\.DbUser\}\}/$DB_USER}"
    content="${content//\{\{\.DbName\}\}/$DB_NAME}"
    content="${content//\{\{\.AppKey\}\}/$APP_KEY}"
    echo "$content"
}

# Generate file from template
generate_file() {
    local template="$1"
    local output="$2"
    local content
    
    content=$(cat "$template")
    content=$(replace_vars "$content")
    
    mkdir -p "$(dirname "$output")"
    echo "$content" > "$output"
}

echo -e "${YELLOW}Creating project structure...${NC}"

# Create directories
mkdir -p "$APP_NAME/application/config"
mkdir -p "$APP_NAME/application/controllers/admin"
mkdir -p "$APP_NAME/application/models"
mkdir -p "$APP_NAME/application/views/layouts"
mkdir -p "$APP_NAME/application/views/auth"
mkdir -p "$APP_NAME/application/views/admin/users"
mkdir -p "$APP_NAME/application/views/admin/settings"
mkdir -p "$APP_NAME/application/views/admin/partials"
mkdir -p "$APP_NAME/application/views/partials"
mkdir -p "$APP_NAME/application/libs"
mkdir -p "$APP_NAME/public/static/css"
mkdir -p "$APP_NAME/database"

echo -e "${YELLOW}Generating configuration files...${NC}"

# Generate core files
generate_file "$TEMPLATES_DIR/gomod.tmpl" "$APP_NAME/go.mod"
generate_file "$TEMPLATES_DIR/env.tmpl" "$APP_NAME/.env"
generate_file "$TEMPLATES_DIR/env.example.tmpl" "$APP_NAME/.env.example"
generate_file "$TEMPLATES_DIR/main.go.tmpl" "$APP_NAME/main.go"

echo -e "${YELLOW}Generating application files...${NC}"

# Config
generate_file "$TEMPLATES_DIR/database.go.tmpl" "$APP_NAME/application/config/database.go"

# Models
generate_file "$TEMPLATES_DIR/user.go.tmpl" "$APP_NAME/application/models/user.go"
generate_file "$TEMPLATES_DIR/setting.go.tmpl" "$APP_NAME/application/models/setting.go"

# Controllers
generate_file "$TEMPLATES_DIR/auth.go.tmpl" "$APP_NAME/application/controllers/auth.go"
generate_file "$TEMPLATES_DIR/dashboard.go.tmpl" "$APP_NAME/application/controllers/admin/dashboard.go"
generate_file "$TEMPLATES_DIR/users.go.tmpl" "$APP_NAME/application/controllers/admin/users.go"
generate_file "$TEMPLATES_DIR/settings.go.tmpl" "$APP_NAME/application/controllers/admin/settings.go"

# Libs
generate_file "$TEMPLATES_DIR/auth_lib.go.tmpl" "$APP_NAME/application/libs/auth.go"

# Database
generate_file "$TEMPLATES_DIR/seed.go.tmpl" "$APP_NAME/database/seed.go"

echo -e "${YELLOW}Generating views...${NC}"

# Layouts
generate_file "$TEMPLATES_DIR/main.html.tmpl" "$APP_NAME/application/views/layouts/main.html"

# Auth views
generate_file "$TEMPLATES_DIR/login.html.tmpl" "$APP_NAME/application/views/auth/login.html"

# Admin views
generate_file "$TEMPLATES_DIR/dashboard.html.tmpl" "$APP_NAME/application/views/admin/dashboard.html"
generate_file "$TEMPLATES_DIR/users_index.html.tmpl" "$APP_NAME/application/views/admin/users/index.html"
generate_file "$TEMPLATES_DIR/users_form.html.tmpl" "$APP_NAME/application/views/admin/users/_form.html"
generate_file "$TEMPLATES_DIR/users_row.html.tmpl" "$APP_NAME/application/views/admin/users/_row.html"
generate_file "$TEMPLATES_DIR/settings_index.html.tmpl" "$APP_NAME/application/views/admin/settings/index.html"

# Partials
generate_file "$TEMPLATES_DIR/header.html.tmpl" "$APP_NAME/application/views/admin/partials/_header.html"
generate_file "$TEMPLATES_DIR/sidebar.html.tmpl" "$APP_NAME/application/views/admin/partials/_sidebar.html"
generate_file "$TEMPLATES_DIR/footer.html.tmpl" "$APP_NAME/application/views/admin/partials/_footer.html"
generate_file "$TEMPLATES_DIR/messages.html.tmpl" "$APP_NAME/application/views/partials/_messages.html"
generate_file "$TEMPLATES_DIR/pagination.html.tmpl" "$APP_NAME/application/views/partials/_pagination.html"

echo -e "${YELLOW}Generating static files...${NC}"

# CSS
generate_file "$TEMPLATES_DIR/style.css.tmpl" "$APP_NAME/public/static/css/style.css"

echo -e "${GREEN}✓ All files generated${NC}"
```

- [ ] **Step 2: Commit installer script**

```bash
git add setup/installer.sh
git commit -m "feat(setup): add installer.sh with template processing"
```

---

## Task 3: Core Templates (main.go, go.mod, .env)

**Files:**
- Create: `setup/templates/main.go.tmpl`
- Create: `setup/templates/gomod.tmpl`
- Create: `setup/templates/env.tmpl`
- Create: `setup/templates/env.example.tmpl`

- [ ] **Step 1: Create main.go template**

```go
package main

import (
	"log"
	"os"

	"{{.ModulePath}}/application/config"
	"{{.ModulePath}}/application/controllers"
	_ "{{.ModulePath}}/application/controllers/admin"
	"{{.ModulePath}}/application/libs"
	"{{.ModulePath}}/database"

	"github.com/joho/godotenv"
	"github.com/semutdev/goigniter/system/core"
	"github.com/semutdev/goigniter/system/helpers"
	"github.com/semutdev/goigniter/system/libraries/session"
	"github.com/semutdev/goigniter/system/middleware"
)

func main() {
	// Load .env file
	godotenv.Load()

	// Connect to database
	config.ConnectDB()

	// Run seeder if DB_SEED=true
	if os.Getenv("DB_SEED") == "true" {
		database.Seed(config.DB)
	}

	// Create application
	app := core.New()

	// Initialize helpers
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":8080"
	}
	helpers.Init("http://localhost" + port)

	// Initialize session
	sessionSecret := os.Getenv("APP_KEY")
	if sessionSecret == "" {
		sessionSecret = "{{.AppKey}}"
	}
	session.Init(session.Config{
		Secret: sessionSecret,
		MaxAge: 86400,
	})

	// Load templates
	if err := app.LoadTemplatesWithFuncs("./application/views", true, helpers.AllTemplateFuncs()); err != nil {
		log.Printf("Warning: Could not load templates: %v", err)
	}

	// Global middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// Static files
	app.Static("/static/", "./public")

	// Auth middleware for admin routes
	app.Use(func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			// Skip auth for login and static routes
			path := c.Request.URL.Path
			if path == "/auth/login" || path == "/auth/dologin" || 
			   len(path) >= 7 && path[:7] == "/static" {
				return next(c)
			}
			
			// Require auth for admin routes
			if len(path) >= 6 && path[:6] == "/admin" {
				if !libs.IsLoggedIn(c) {
					return c.Redirect(302, "/auth/login")
				}
			}
			return next(c)
		}
	})

	// Auto-route from registered controllers
	app.AutoRoute()

	// Root redirect
	app.GET("/", func(c *core.Context) error {
		return c.Redirect(302, "/admin/dashboard")
	})

	// Start server
	log.Println("Server starting on " + port)
	log.Println("Default login: admin@admin.com / password")
	log.Fatal(app.Run(port))
}
```

- [ ] **Step 2: Create go.mod template**

```
module {{.ModulePath}}

go 1.24

require (
	github.com/joho/godotenv v1.5.1
	github.com/semutdev/goigniter v0.0.0
)

// Replace with local goigniter for development
// replace github.com/semutdev/goigniter => ../
```

- [ ] **Step 3: Create .env template**

```env
# Application
APP_ENV=local
APP_PORT=:8080
APP_URL=http://localhost:8080
APP_KEY={{.AppKey}}

# Database
DB_DRIVER={{.DbType}}
{{if eq .DbType "sqlite"}}DB_DSN=./app.db{{end}}{{if eq .DbType "mysql"}}DB_DSN={{.DbUser}}:{{.DbPass}}@tcp({{.DbHost}}:{{.DbPort}})/{{.DbName}}?charset=utf8mb4&parseTime=True&loc=Local{{end}}

# Seeder (set to true once to create tables and admin user)
DB_SEED=false
```

- [ ] **Step 4: Create .env.example template**

```env
# Application
APP_ENV=local
APP_PORT=:8080
APP_URL=http://localhost:8080
APP_KEY=your-secret-key-32-characters-long

# Database (sqlite)
DB_DRIVER=sqlite
DB_DSN=./app.db

# Database (mysql)
# DB_DRIVER=mysql
# DB_DSN=user:password@tcp(127.0.0.1:3306/dbname?charset=utf8mb4&parseTime=True&loc=Local

# Seeder (set to true once to create tables and admin user)
DB_SEED=false
```

- [ ] **Step 5: Commit core templates**

```bash
git add setup/templates/main.go.tmpl setup/templates/gomod.tmpl setup/templates/env.tmpl setup/templates/env.example.tmpl
git commit -m "feat(setup): add core templates (main.go, go.mod, .env)"
```

---

## Task 4: Database and Model Templates

**Files:**
- Create: `setup/templates/database.go.tmpl`
- Create: `setup/templates/user.go.tmpl`
- Create: `setup/templates/setting.go.tmpl`

- [ ] **Step 1: Create database.go template**

```go
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"{{.ModulePath}}/application/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	godotenv.Load()

	driver := os.Getenv("DB_DRIVER")
	dsn := os.Getenv("DB_DSN")

	var err error
	var dialector gorm.Dialector

	switch driver {
	case "mysql":
		dialector = mysql.Open(dsn)
	default:
		dialector = sqlite.Open(dsn)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	// Auto migrate
	DB.AutoMigrate(&models.User{}, &models.Setting{})
}
```

- [ ] **Step 2: Create user.go model template**

```go
package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	FirstName string         `gorm:"size:100" json:"first_name"`
	LastName  string         `gorm:"size:100" json:"last_name"`
	Role      string         `gorm:"size:20;default:user" json:"role"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) FullName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.Email
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}
```

- [ ] **Step 3: Create setting.go model template**

```go
package models

import (
	"time"

	"gorm.io/gorm"
)

type Setting struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Key       string         `gorm:"uniqueIndex;size:100;not null" json:"key"`
	Value     string         `gorm:"type:text" json:"value"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GetSetting retrieves a setting by key
func GetSetting(key string) string {
	var setting Setting
	result := DB.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		return ""
	}
	return setting.Value
}

// SetSetting sets a setting value
func SetSetting(key, value string) error {
	var setting Setting
	result := DB.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		// Create new
		setting = Setting{Key: key, Value: value}
		return DB.Create(&setting).Error
	}
	// Update existing
	setting.Value = value
	return DB.Save(&setting).Error
}
```

- [ ] **Step 4: Commit database and model templates**

```bash
git add setup/templates/database.go.tmpl setup/templates/user.go.tmpl setup/templates/setting.go.tmpl
git commit -m "feat(setup): add database config and model templates"
```

---

## Task 5: Auth Library Template

**Files:**
- Create: `setup/templates/auth_lib.go.tmpl`

- [ ] **Step 1: Create auth library template**

```go
package libs

import (
	"{{.ModulePath}}/application/models"
	"{{.ModulePath}}/application/config"

	"github.com/semutdev/goigniter/system/core"
	"github.com/semutdev/goigniter/system/libraries/session"
)

const (
	SessionKeyUserID    = "user_id"
	SessionKeyUserEmail = "user_email"
	SessionKeyUserRole   = "user_role"
)

// Login authenticates a user and returns the user if successful
func Login(email, password string) (*models.User, error) {
	var user models.User
	result := config.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if !user.CheckPassword(password) {
		return nil, ErrInvalidPassword
	}

	return &user, nil
}

// SetSession stores user info in session
func SetSession(c *core.Context, user *models.User) {
	sess := session.Get(c)
	sess.Set(SessionKeyUserID, user.ID)
	sess.Set(SessionKeyUserEmail, user.Email)
	sess.Set(SessionKeyUserRole, user.Role)
	sess.Save()
}

// ClearSession removes user info from session
func ClearSession(c *core.Context) {
	sess := session.Get(c)
	sess.Delete(SessionKeyUserID)
	sess.Delete(SessionKeyUserEmail)
	sess.Delete(SessionKeyUserRole)
	sess.Save()
}

// IsLoggedIn checks if user is logged in
func IsLoggedIn(c *core.Context) bool {
	sess := session.Get(c)
	userID := sess.Get(SessionKeyUserID)
	return userID != nil
}

// GetCurrentUser returns the current logged in user
func GetCurrentUser(c *core.Context) *models.User {
	sess := session.Get(c)
	userID := sess.Get(SessionKeyUserID)
	if userID == nil {
		return nil
	}

	var user models.User
	result := config.DB.First(&user, userID)
	if result.Error != nil {
		return nil
	}
	return &user
}

// IsAdmin checks if current user is admin
func IsAdmin(c *core.Context) bool {
	sess := session.Get(c)
	role, ok := sess.Get(SessionKeyUserRole).(string)
	if !ok {
		return false
	}
	return role == "admin"
}

// CreateUser creates a new user
func CreateUser(email, password, firstName, lastName, role string) (*models.User, error) {
	user := &models.User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		IsActive:  true,
	}
	
	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	if err := config.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// Error types
var (
	ErrUserInactive    = &AuthError{Message: "User account is inactive"}
	ErrInvalidPassword = &AuthError{Message: "Invalid email or password"}
)

type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
```

- [ ] **Step 2: Commit auth library template**

```bash
git add setup/templates/auth_lib.go.tmpl
git commit -m "feat(setup): add auth library template"
```

---

## Task 6: Controller Templates

**Files:**
- Create: `setup/templates/auth.go.tmpl`
- Create: `setup/templates/dashboard.go.tmpl`
- Create: `setup/templates/users.go.tmpl`
- Create: `setup/templates/settings.go.tmpl`
- Create: `setup/templates/seed.go.tmpl`

- [ ] **Step 1: Create auth controller template**

```go
package controllers

import (
	"net/http"

	"{{.ModulePath}}/application/libs"

	"github.com/semutdev/goigniter/system/core"
)

func init() {
	core.Register(&Auth{})
}

type Auth struct {
	core.Controller
}

// Login displays the login form
func (a *Auth) Login() {
	if libs.IsLoggedIn(a.Ctx) {
		a.Ctx.Redirect(http.StatusFound, "/admin/dashboard")
		return
	}

	a.Ctx.View("auth/login", core.Map{
		"Title": "Login",
		"Error": a.Ctx.Query("error"),
	})
}

// Dologin processes the login
func (a *Auth) Dologin() {
	email := a.Ctx.FormValue("email")
	password := a.Ctx.FormValue("password")

	if email == "" || password == "" {
		a.Ctx.View("auth/login", core.Map{
			"Title": "Login",
			"Error": "Email and password are required",
		})
		return
	}

	user, err := libs.Login(email, password)
	if err != nil {
		a.Ctx.View("auth/login", core.Map{
			"Title": "Login",
			"Error": "Invalid email or password",
		})
		return
	}

	libs.SetSession(a.Ctx, user)
	a.Ctx.Redirect(http.StatusFound, "/admin/dashboard")
}

// Logout clears the session
func (a *Auth) Logout() {
	libs.ClearSession(a.Ctx)
	a.Ctx.Redirect(http.StatusFound, "/auth/login")
}
```

- [ ] **Step 2: Create dashboard controller template**

```go
package admin

import (
	"{{.ModulePath}}/application/config"
	"{{.ModulePath}}/application/libs"
	"{{.ModulePath}}/application/models"

	"github.com/semutdev/goigniter/system/core"
)

func init() {
	core.Register(&Dashboard{})
}

type Dashboard struct {
	core.Controller
}

// Index displays the admin dashboard
func (d *Dashboard) Index() {
	user := libs.GetCurrentUser(d.Ctx)
	
	// Get stats
	var userCount int64
	config.DB.Model(&models.User{}).Count(&userCount)

	d.Ctx.View("admin/dashboard", core.Map{
		"Title":     "Dashboard",
		"User":      user,
		"UserCount": userCount,
	})
}
```

- [ ] **Step 3: Create users controller template**

```go
package admin

import (
	"net/http"
	"strconv"

	"{{.ModulePath}}/application/config"
	"{{.ModulePath}}/application/libs"
	"{{.ModulePath}}/application/models"

	"github.com/semutdev/goigniter/system/core"
)

func init() {
	core.Register(&Users{})
}

type Users struct {
	core.Controller
}

// Index displays user list
func (u *Users) Index() {
	page, _ := strconv.Atoi(u.Ctx.Query("page"))
	if page < 1 {
		page = 1
	}
	perPage := 10

	var users []models.User
	var total int64
	config.DB.Model(&models.User{}).Count(&total)
	config.DB.Offset((page - 1) * perPage).Limit(perPage).Find(&users)

	user := libs.GetCurrentUser(u.Ctx)

	u.Ctx.View("admin/users/index", core.Map{
		"Title":    "Users",
		"Users":    users,
		"User":     user,
		"Page":     page,
		"Total":    total,
		"PerPage":  perPage,
		"HasPages": total > int64(perPage),
	})
}

// Create displays create user form
func (u *Users) Create() {
	user := libs.GetCurrentUser(u.Ctx)
	u.Ctx.View("admin/users/_form", core.Map{
		"Title": "Create User",
		"User":  user,
		"EditUser": &models.User{},
	})
}

// Store saves new user
func (u *Users) Store() {
	email := u.Ctx.FormValue("email")
	password := u.Ctx.FormValue("password")
	firstName := u.Ctx.FormValue("first_name")
	lastName := u.Ctx.FormValue("last_name")
	role := u.Ctx.FormValue("role")

	user, err := libs.CreateUser(email, password, firstName, lastName, role)
	if err != nil {
		u.Ctx.View("admin/users/_form", core.Map{
			"Title":    "Create User",
			"EditUser": &models.User{Email: email, FirstName: firstName, LastName: lastName, Role: role},
			"Error":    "Failed to create user: " + err.Error(),
		})
		return
	}

	_ = user
	u.Ctx.Redirect(http.StatusFound, "/admin/users/index")
}

// Edit displays edit user form
func (u *Users) Edit() {
	id := u.Ctx.Param("id")
	var editUser models.User
	if err := config.DB.First(&editUser, id).Error; err != nil {
		u.Ctx.Redirect(http.StatusFound, "/admin/users/index")
		return
	}

	user := libs.GetCurrentUser(u.Ctx)
	u.Ctx.View("admin/users/_form", core.Map{
		"Title":    "Edit User",
		"User":     user,
		"EditUser": &editUser,
	})
}

// Update saves user changes
func (u *Users) Update() {
	id := u.Ctx.Param("id")
	var editUser models.User
	if err := config.DB.First(&editUser, id).Error; err != nil {
		u.Ctx.Redirect(http.StatusFound, "/admin/users/index")
		return
	}

	editUser.Email = u.Ctx.FormValue("email")
	editUser.FirstName = u.Ctx.FormValue("first_name")
	editUser.LastName = u.Ctx.FormValue("last_name")
	editUser.Role = u.Ctx.FormValue("role")
	editUser.IsActive = u.Ctx.FormValue("is_active") == "on"

	password := u.Ctx.FormValue("password")
	if password != "" {
		editUser.SetPassword(password)
	}

	config.DB.Save(&editUser)
	u.Ctx.Redirect(http.StatusFound, "/admin/users/index")
}

// Delete removes a user
func (u *Users) Delete() {
	id := u.Ctx.Param("id")
	config.DB.Delete(&models.User{}, id)
	u.Ctx.Redirect(http.StatusFound, "/admin/users/index")
}
```

- [ ] **Step 4: Create settings controller template**

```go
package admin

import (
	"{{.ModulePath}}/application/libs"
	"{{.ModulePath}}/application/models"

	"github.com/semutdev/goigniter/system/core"
)

func init() {
	core.Register(&Settings{})
}

type Settings struct {
	core.Controller
}

// Index displays settings form
func (s *Settings) Index() {
	user := libs.GetCurrentUser(s.Ctx)

	s.Ctx.View("admin/settings/index", core.Map{
		"Title":        "Settings",
		"User":         user,
		"SiteName":     models.GetSetting("site_name"),
		"SiteTagline":  models.GetSetting("site_tagline"),
		"SiteLogo":     models.GetSetting("site_logo"),
		"Success":      s.Ctx.Query("success"),
	})
}

// Update saves settings
func (s *Settings) Update() {
	models.SetSetting("site_name", s.Ctx.FormValue("site_name"))
	models.SetSetting("site_tagline", s.Ctx.FormValue("site_tagline"))
	models.SetSetting("site_logo", s.Ctx.FormValue("site_logo"))

	s.Ctx.Redirect(http.StatusFound, "/admin/settings/index?success=1")
}
```

- [ ] **Step 5: Create seed.go template**

```go
package database

import (
	"{{.ModulePath}}/application/models"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	// Auto migrate
	if err := db.AutoMigrate(&models.User{}, &models.Setting{}); err != nil {
		return err
	}

	// Create admin user if not exists
	var count int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count == 0 {
		admin := &models.User{
			Email:     "admin@admin.com",
			FirstName: "Admin",
			LastName:  "User",
			Role:      "admin",
			IsActive:  true,
		}
		admin.SetPassword("password")
		db.Create(admin)
	}

	// Create default settings
	settings := map[string]string{
		"site_name":    "{{.AppName}}",
		"site_tagline": "Built with GoIgniter",
		"site_logo":    "",
	}

	for key, value := range settings {
		var setting models.Setting
		result := db.Where("key = ?", key).First(&setting)
		if result.Error != nil {
			db.Create(&models.Setting{Key: key, Value: value})
		}
	}

	return nil
}
```

- [ ] **Step 6: Commit controller templates**

```bash
git add setup/templates/auth.go.tmpl setup/templates/dashboard.go.tmpl setup/templates/users.go.tmpl setup/templates/settings.go.tmpl setup/templates/seed.go.tmpl
git commit -m "feat(setup): add controller and seed templates"
```

---

## Task 7: View Templates - Layout and Auth

**Files:**
- Create: `setup/templates/main.html.tmpl`
- Create: `setup/templates/login.html.tmpl`

- [ ] **Step 1: Create main layout template**

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} | {{.AppName}}</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    {{template "content" .}}
</body>
</html>
```

- [ ] **Step 2: Create login page template**

```html
{{define "content"}}
<div class="login-container">
    <div class="login-card">
        <h1>{{.AppName}}</h1>
        <h2>Login</h2>
        
        {{if .Error}}
        <div class="alert alert-error">
            {{.Error}}
        </div>
        {{end}}

        <form action="/auth/dologin" method="POST">
            <div class="form-group">
                <label for="email">Email</label>
                <input type="email" id="email" name="email" required autofocus>
            </div>
            
            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" required>
            </div>
            
            <button type="submit" class="btn btn-primary btn-block">Login</button>
        </form>
    </div>
</div>
{{end}}
```

- [ ] **Step 3: Commit layout and auth views**

```bash
git add setup/templates/main.html.tmpl setup/templates/login.html.tmpl
git commit -m "feat(setup): add layout and login view templates"
```

---

## Task 8: View Templates - Admin Layout Partials

**Files:**
- Create: `setup/templates/header.html.tmpl`
- Create: `setup/templates/sidebar.html.tmpl`
- Create: `setup/templates/footer.html.tmpl`

- [ ] **Step 1: Create header partial template**

```html
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} | Admin</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
```

- [ ] **Step 2: Create sidebar partial template**

```html
<aside class="sidebar">
    <div class="sidebar-header">
        <h2>{{.AppName}}</h2>
    </div>
    <nav class="sidebar-nav">
        <a href="/admin/dashboard" class="nav-link {{if eq .Title "Dashboard"}}active{{end}}">
            Dashboard
        </a>
        <a href="/admin/users/index" class="nav-link {{if eq .Title "Users"}}active{{end}}">
            Users
        </a>
        <a href="/admin/settings/index" class="nav-link {{if eq .Title "Settings"}}active{{end}}">
            Settings
        </a>
    </nav>
    <div class="sidebar-footer">
        <a href="/auth/logout" class="nav-link">Logout</a>
    </div>
</aside>
```

- [ ] **Step 3: Create footer partial template**

```html
<footer class="footer">
    <p>&copy; 2024 {{.AppName}}. Built with GoIgniter.</p>
</footer>
```

- [ ] **Step 4: Commit admin partials**

```bash
git add setup/templates/header.html.tmpl setup/templates/sidebar.html.tmpl setup/templates/footer.html.tmpl
git commit -m "feat(setup): add admin layout partial templates"
```

---

## Task 9: View Templates - Dashboard and Users

**Files:**
- Create: `setup/templates/dashboard.html.tmpl`
- Create: `setup/templates/users_index.html.tmpl`
- Create: `setup/templates/users_form.html.tmpl`
- Create: `setup/templates/users_row.html.tmpl`

- [ ] **Step 1: Create dashboard template**

```html
<!DOCTYPE html>
<html lang="en">
{{template "header" .}}
<body class="admin-body">
    <div class="admin-layout">
        {{template "sidebar" .}}
        
        <main class="main-content">
            {{template "messages" .}}
            
            <header class="page-header">
                <h1>Dashboard</h1>
                <p>Welcome back, {{.User.FullName}}!</p>
            </header>

            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-value">{{.UserCount}}</div>
                    <div class="stat-label">Total Users</div>
                </div>
            </div>
        </main>
    </div>
</body>
</html>
```

- [ ] **Step 2: Create users index template**

```html
<!DOCTYPE html>
<html lang="en">
{{template "header" .}}
<body class="admin-body">
    <div class="admin-layout">
        {{template "sidebar" .}}
        
        <main class="main-content">
            {{template "messages" .}}
            
            <header class="page-header">
                <h1>Users</h1>
                <a href="/admin/users/create" class="btn btn-primary">Add User</a>
            </header>

            <table class="table">
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Email</th>
                        <th>Name</th>
                        <th>Role</th>
                        <th>Status</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Users}}
                    <tr>
                        <td>{{.ID}}</td>
                        <td>{{.Email}}</td>
                        <td>{{.FullName}}</td>
                        <td>{{.Role}}</td>
                        <td>
                            {{if .IsActive}}
                            <span class="badge badge-success">Active</span>
                            {{else}}
                            <span class="badge badge-danger">Inactive</span>
                            {{end}}
                        </td>
                        <td>
                            <a href="/admin/users/edit/{{.ID}}" class="btn btn-sm btn-secondary">Edit</a>
                            <button hx-delete="/admin/users/delete/{{.ID}}"
                                    hx-confirm="Are you sure?"
                                    hx-target="closest tr"
                                    hx-swap="outerHTML"
                                    class="btn btn-sm btn-danger">Delete</button>
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>

            {{if .HasPages}}
            {{template "pagination" .}}
            {{end}}
        </main>
    </div>
</body>
</html>
```

- [ ] **Step 3: Create users form template**

```html
<!DOCTYPE html>
<html lang="en">
{{template "header" .}}
<body class="admin-body">
    <div class="admin-layout">
        {{template "sidebar" .}}
        
        <main class="main-content">
            {{template "messages" .}}
            
            <header class="page-header">
                <h1>{{.Title}}</h1>
                <a href="/admin/users/index" class="btn btn-secondary">Back</a>
            </header>

            <div class="card">
                <form action="/admin/users/{{if .EditUser.ID}}update/{{.EditUser.ID}}{{else}}store{{end}}" 
                      method="POST" 
                      hx-boost="true">
                    
                    {{if .Error}}
                    <div class="alert alert-error">{{.Error}}</div>
                    {{end}}

                    <div class="form-group">
                        <label for="email">Email</label>
                        <input type="email" id="email" name="email" value="{{.EditUser.Email}}" required>
                    </div>

                    <div class="form-group">
                        <label for="password">Password {{if .EditUser.ID}}(leave blank to keep){{end}}</label>
                        <input type="password" id="password" name="password" {{if not .EditUser.ID}}required{{end}}>
                    </div>

                    <div class="form-row">
                        <div class="form-group">
                            <label for="first_name">First Name</label>
                            <input type="text" id="first_name" name="first_name" value="{{.EditUser.FirstName}}">
                        </div>

                        <div class="form-group">
                            <label for="last_name">Last Name</label>
                            <input type="text" id="last_name" name="last_name" value="{{.EditUser.LastName}}">
                        </div>
                    </div>

                    <div class="form-group">
                        <label for="role">Role</label>
                        <select id="role" name="role">
                            <option value="user" {{if eq .EditUser.Role "user"}}selected{{end}}>User</option>
                            <option value="admin" {{if eq .EditUser.Role "admin"}}selected{{end}}>Admin</option>
                        </select>
                    </div>

                    {{if .EditUser.ID}}
                    <div class="form-group">
                        <label class="checkbox">
                            <input type="checkbox" name="is_active" {{if .EditUser.IsActive}}checked{{end}}>
                            Active
                        </label>
                    </div>
                    {{end}}

                    <button type="submit" class="btn btn-primary">Save</button>
                </form>
            </div>
        </main>
    </div>
</body>
</html>
```

- [ ] **Step 4: Create users row template (for HTMX)****

```html
<tr>
    <td>{{.ID}}</td>
    <td>{{.Email}}</td>
    <td>{{.FullName}}</td>
    <td>{{.Role}}</td>
    <td>
        {{if .IsActive}}
        <span class="badge badge-success">Active</span>
        {{else}}
        <span class="badge badge-danger">Inactive</span>
        {{end}}
    </td>
    <td>
        <a href="/admin/users/edit/{{.ID}}" class="btn btn-sm btn-secondary">Edit</a>
        <button hx-delete="/admin/users/delete/{{.ID}}"
                hx-confirm="Are you sure?"
                hx-target="closest tr"
                hx-swap="outerHTML"
                class="btn btn-sm btn-danger">Delete</button>
    </td>
</tr>
```

- [ ] **Step 5: Commit dashboard and users views**

```bash
git add setup/templates/dashboard.html.tmpl setup/templates/users_index.html.tmpl setup/templates/users_form.html.tmpl setup/templates/users_row.html.tmpl
git commit -m "feat(setup): add dashboard and users view templates"
```

---

## Task 10: View Templates - Settings and Partials

**Files:**
- Create: `setup/templates/settings_index.html.tmpl`
- Create: `setup/templates/messages.html.tmpl`
- Create: `setup/templates/pagination.html.tmpl`

- [ ] **Step 1: Create settings index template**

```html
<!DOCTYPE html>
<html lang="en">
{{template "header" .}}
<body class="admin-body">
    <div class="admin-layout">
        {{template "sidebar" .}}
        
        <main class="main-content">
            {{template "messages" .}}
            
            <header class="page-header">
                <h1>Settings</h1>
            </header>

            <div class="card">
                <form action="/admin/settings/update" method="POST">
                    
                    <div class="form-group">
                        <label for="site_name">Site Name</label>
                        <input type="text" id="site_name" name="site_name" value="{{.SiteName}}">
                    </div>

                    <div class="form-group">
                        <label for="site_tagline">Tagline</label>
                        <input type="text" id="site_tagline" name="site_tagline" value="{{.SiteTagline}}">
                    </div>

                    <div class="form-group">
                        <label for="site_logo">Logo URL</label>
                        <input type="text" id="site_logo" name="site_logo" value="{{.SiteLogo}}" placeholder="https://example.com/logo.png">
                    </div>

                    <button type="submit" class="btn btn-primary">Save Settings</button>
                </form>
            </div>
        </main>
    </div>
</body>
</html>
```

- [ ] **Step 2: Create messages partial template**

```html
{{define "messages"}}
{{if .Success}}
<div class="alert alert-success" hx-target="this" hx-swap="outerHTML">
    {{.Success}}
    <button class="close" onclick="this.parentElement.remove()">&times;</button>
</div>
{{end}}
{{if .Error}}
<div class="alert alert-error">
    {{.Error}}
    <button class="close" onclick="this.parentElement.remove()">&times;</button>
</div>
{{end}}
{{end}}
```

- [ ] **Step 3: Create pagination partial template**

```html
{{define "pagination"}}
<div class="pagination">
    {{if gt .Page 1}}
    <a href="?page={{sub .Page 1}}" class="btn btn-secondary">Previous</a>
    {{end}}
    
    <span class="page-info">Page {{.Page}} of {{divCeil .Total .PerPage}}</span>
    
    {{if gt (mul .Page .PerPage) .Total}}
    {{else}}
    <a href="?page={{add .Page 1}}" class="btn btn-secondary">Next</a>
    {{end}}
</div>
{{end}}
```

- [ ] **Step 4: Commit settings and partial views**

```bash
git add setup/templates/settings_index.html.tmpl setup/templates/messages.html.tmpl setup/templates/pagination.html.tmpl
git commit -m "feat(setup): add settings and partial view templates"
```

---

## Task 11: CSS Stylesheet Template

**Files:**
- Create: `setup/templates/style.css.tmpl`

- [ ] **Step 1: Create minimal CSS template**

```css
/* Reset & Base */
*, *::before, *::after {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}

:root {
    --primary: #3b82f6;
    --primary-dark: #2563eb;
    --secondary: #6b7280;
    --success: #10b981;
    --danger: #ef4444;
    --warning: #f59e0b;
    --text: #1f2937;
    --text-light: #6b7280;
    --bg: #f9fafb;
    --white: #ffffff;
    --border: #e5e7eb;
    --sidebar-bg: #1f2937;
    --sidebar-text: #f9fafb;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
    font-size: 16px;
    line-height: 1.5;
    color: var(--text);
    background: var(--bg);
}

/* Login Page */
.login-container {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
}

.login-card {
    background: var(--white);
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    width: 100%;
    max-width: 400px;
}

.login-card h1 {
    font-size: 1.5rem;
    text-align: center;
    margin-bottom: 0.5rem;
    color: var(--primary);
}

.login-card h2 {
    font-size: 1.25rem;
    text-align: center;
    margin-bottom: 1.5rem;
    color: var(--text-light);
}

/* Admin Layout */
.admin-body {
    min-height: 100vh;
}

.admin-layout {
    display: flex;
    min-height: 100vh;
}

/* Sidebar */
.sidebar {
    width: 250px;
    background: var(--sidebar-bg);
    color: var(--sidebar-text);
    display: flex;
    flex-direction: column;
    position: fixed;
    height: 100vh;
}

.sidebar-header {
    padding: 1.5rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.sidebar-header h2 {
    font-size: 1.25rem;
    color: var(--white);
}

.sidebar-nav {
    flex: 1;
    padding: 1rem 0;
}

.nav-link {
    display: block;
    padding: 0.75rem 1.5rem;
    color: var(--sidebar-text);
    text-decoration: none;
    transition: background 0.2s;
}

.nav-link:hover {
    background: rgba(255, 255, 255, 0.1);
}

.nav-link.active {
    background: var(--primary);
    color: var(--white);
}

.sidebar-footer {
    padding: 1rem 1.5rem;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
}

/* Main Content */
.main-content {
    flex: 1;
    margin-left: 250px;
    padding: 2rem;
}

/* Page Header */
.page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
}

.page-header h1 {
    font-size: 1.75rem;
    font-weight: 600;
}

/* Cards */
.card {
    background: var(--white);
    border-radius: 8px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    padding: 1.5rem;
}

/* Stats Grid */
.stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;
}

.stat-card {
    background: var(--white);
    border-radius: 8px;
    padding: 1.5rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.stat-value {
    font-size: 2rem;
    font-weight: 700;
    color: var(--primary);
}

.stat-label {
    color: var(--text-light);
    margin-top: 0.25rem;
}

/* Forms */
.form-group {
    margin-bottom: 1rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
}

.form-group input,
.form-group select,
.form-group textarea {
    width: 100%;
    padding: 0.625rem 0.875rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    font-size: 1rem;
    transition: border-color 0.2s, box-shadow 0.2s;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
    outline: none;
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
}

.checkbox {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
}

.checkbox input[type="checkbox"] {
    width: auto;
}

/* Buttons */
.btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0.625rem 1.25rem;
    font-size: 1rem;
    font-weight: 500;
    border-radius: 6px;
    border: none;
    cursor: pointer;
    text-decoration: none;
    transition: background 0.2s, transform 0.1s;
}

.btn:hover {
    transform: translateY(-1px);
}

.btn-primary {
    background: var(--primary);
    color: var(--white);
}

.btn-primary:hover {
    background: var(--primary-dark);
}

.btn-secondary {
    background: var(--secondary);
    color: var(--white);
}

.btn-secondary:hover {
    background: #4b5563;
}

.btn-danger {
    background: var(--danger);
    color: var(--white);
}

.btn-danger:hover {
    background: #dc2626;
}

.btn-block {
    width: 100%;
}

.btn-sm {
    padding: 0.375rem 0.75rem;
    font-size: 0.875rem;
}

/* Tables */
.table {
    width: 100%;
    border-collapse: collapse;
    background: var(--white);
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.table th,
.table td {
    padding: 0.875rem 1rem;
    text-align: left;
    border-bottom: 1px solid var(--border);
}

.table th {
    background: var(--bg);
    font-weight: 600;
}

.table tr:hover {
    background: var(--bg);
}

/* Badges */
.badge {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    font-size: 0.75rem;
    font-weight: 500;
    border-radius: 9999px;
}

.badge-success {
    background: rgba(16, 185, 129, 0.1);
    color: var(--success);
}

.badge-danger {
    background: rgba(239, 68, 68, 0.1);
    color: var(--danger);
}

/* Alerts */
.alert {
    padding: 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.alert-success {
    background: rgba(16, 185, 129, 0.1);
    color: var(--success);
    border: 1px solid var(--success);
}

.alert-error {
    background: rgba(239, 68, 68, 0.1);
    color: var(--danger);
    border: 1px solid var(--danger);
}

.alert .close {
    background: none;
    border: none;
    font-size: 1.25rem;
    cursor: pointer;
    opacity: 0.5;
}

.alert .close:hover {
    opacity: 1;
}

/* Pagination */
.pagination {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    margin-top: 1.5rem;
}

.page-info {
    color: var(--text-light);
}

/* HTMX Loading */
.htmx-indicator {
    opacity: 0;
    transition: opacity 200ms ease-in;
}

.htmx-request .htmx-indicator,
.htmx-request.htmx-indicator {
    opacity: 1;
}

/* Responsive */
@media (max-width: 768px) {
    .sidebar {
        width: 60px;
    }
    
    .sidebar-header h2,
    .nav-link span {
        display: none;
    }
    
    .main-content {
        margin-left: 60px;
    }
    
    .form-row {
        grid-template-columns: 1fr;
    }
}
```

- [ ] **Step 2: Commit CSS template**

```bash
git add setup/templates/style.css.tmpl
git commit -m "feat(setup): add minimal CSS template with admin layout"
```

---

## Task 12: Test Setup Script

**Files:**
- Test the generated project

- [ ] **Step 1: Create a test project using the setup script**

```bash
cd setup
./setup.sh --name=testapp
```

- [ ] **Step 2: Verify generated project structure**

```bash
ls -la testapp/
ls -la testapp/application/
ls -la testapp/application/controllers/
```

- [ ] **Step 3: Fix go.mod to use local goigniter (for testing)**

In generated `testapp/go.mod`, add replace directive:
```
replace github.com/semutdev/goigniter => ../
```

- [ ] **Step 4: Run the generated project**

```bash
cd testapp
go mod tidy
# Set DB_SEED=true for first run to create tables
DB_SEED=true go run main.go
```

- [ ] **Step 5: Test login functionality**

1. Open browser to `http://localhost:8080`
2. Should redirect to login page
3. Login with `admin@admin.com` / `password`
4. Should redirect to dashboard

- [ ] **Step 6: Test admin functionality**

1. Navigate to Users page
2. Create a new user
3. Edit the user
4. Delete the user
5. Navigate to Settings page
6. Update site name

- [ ] **Step 7: Clean up test project**

```bash
cd ..
rm -rf testapp
```

- [ ] **Step 8: Final commit with any fixes**

```bash
git add -A
git commit -m "feat(setup): complete setup tool with all templates"
```

---

## Summary

This plan creates a complete setup tool for GoIgniter that:

1. **Interactive Wizard**: Prompts for project name, database type, and MySQL config if needed
2. **Template Generation**: Uses bash string replacement to generate all files
3. **Login System**: Simple authentication with session management
4. **Admin Dashboard**: User management CRUD and settings
5. **HTMX + CSS**: Modern interactivity without page reloads
6. **One Command Setup**: `curl -sSL ... | bash`

Total: 12 tasks, ~50 steps