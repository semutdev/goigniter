#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# GitHub repo URL
REPO_URL="https://raw.githubusercontent.com/semutdev/goigniter/main"

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

DB_TYPE_PROVIDED=false

# Parse flags
while [[ $# -gt 0 ]]; do
    case $1 in
        --name=*) APP_NAME="${1#*=}" ;;
        --db=*) DB_TYPE="${1#*=}"; DB_TYPE_PROVIDED=true ;;
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

# Ask for db type if not explicitly set via flag
if [ "$DB_TYPE_PROVIDED" == "false" ]; then
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

# Project directory
PROJECT_DIR="$PWD/$APP_NAME"

# Create project directory
if [ -d "$PROJECT_DIR" ]; then
    echo -e "${RED}Error: Directory '$APP_NAME' already exists${NC}"
    exit 1
fi

mkdir -p "$PROJECT_DIR"

# Check if running locally or via curl
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)" 2>/dev/null || SCRIPT_DIR=""
TEMPLATES_DIR="$SCRIPT_DIR/templates"

if [ -f "$SCRIPT_DIR/installer.sh" ] && [ -d "$TEMPLATES_DIR" ]; then
    # Running locally - source installer directly
    source "$SCRIPT_DIR/installer.sh"
else
    # Running via curl - download files from GitHub
    echo -e "${YELLOW}Downloading templates from GitHub...${NC}"
    
    # Create temp directory for templates
    TEMP_DIR=$(mktemp -d)
    TEMPLATES_DIR="$TEMP_DIR/templates"
    mkdir -p "$TEMPLATES_DIR"
    
    # Download installer.sh
    curl -sSL "$REPO_URL/setup/installer.sh" -o "$TEMP_DIR/installer.sh"
    
    # Download all templates
    TEMPLATE_FILES=(
        "gomod.tmpl"
        "env.tmpl"
        "env.example.tmpl"
        "main.go.tmpl"
        "database.go.tmpl"
        "user.go.tmpl"
        "setting.go.tmpl"
        "auth.go.tmpl"
        "auth_lib.go.tmpl"
        "dashboard.go.tmpl"
        "users.go.tmpl"
        "settings.go.tmpl"
        "seed.go.tmpl"
        "main.html.tmpl"
        "login.html.tmpl"
        "dashboard.html.tmpl"
        "users_index.html.tmpl"
        "users_form.html.tmpl"
        "users_row.html.tmpl"
        "settings_index.html.tmpl"
        "header.html.tmpl"
        "sidebar.html.tmpl"
        "footer.html.tmpl"
        "messages.html.tmpl"
        "pagination.html.tmpl"
        "style.css.tmpl"
    )
    
    for tmpl in "${TEMPLATE_FILES[@]}"; do
        curl -sSL "$REPO_URL/setup/templates/$tmpl" -o "$TEMPLATES_DIR/$tmpl"
    done
    
    # Source installer
    source "$TEMP_DIR/installer.sh"
    
    # Cleanup temp directory
    rm -rf "$TEMP_DIR"
fi

echo ""
echo -e "${GREEN}✓ Done!${NC}"
echo ""
echo "Run: cd $APP_NAME && go mod tidy && go run main.go"
echo ""