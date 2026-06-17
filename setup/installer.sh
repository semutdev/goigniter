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
    
    if [ ! -f "$template" ]; then
        echo -e "${RED}Error: Template not found: $template${NC}"
        return 1
    fi
    
    content=$(cat "$template")
    content=$(replace_vars "$content")
    
    mkdir -p "$(dirname "$output")"
    echo "$content" > "$output"
}

echo -e "${YELLOW}Creating project structure...${NC}"

# Create directories
mkdir -p "$PROJECT_DIR/application/config"
mkdir -p "$PROJECT_DIR/application/controllers/admin"
mkdir -p "$PROJECT_DIR/application/models"
mkdir -p "$PROJECT_DIR/application/views/layouts"
mkdir -p "$PROJECT_DIR/application/views/auth"
mkdir -p "$PROJECT_DIR/application/views/admin/users"
mkdir -p "$PROJECT_DIR/application/views/admin/settings"
mkdir -p "$PROJECT_DIR/application/views/admin/partials"
mkdir -p "$PROJECT_DIR/application/views/partials"
mkdir -p "$PROJECT_DIR/application/libs"
mkdir -p "$PROJECT_DIR/public/static/css"
mkdir -p "$PROJECT_DIR/database"

echo -e "${YELLOW}Generating configuration files...${NC}"

# Generate core files
generate_file "$TEMPLATES_DIR/gomod.tmpl" "$PROJECT_DIR/go.mod"
generate_file "$TEMPLATES_DIR/env.tmpl" "$PROJECT_DIR/.env"
generate_file "$TEMPLATES_DIR/env.example.tmpl" "$PROJECT_DIR/.env.example"
generate_file "$TEMPLATES_DIR/main.go.tmpl" "$PROJECT_DIR/main.go"

echo -e "${YELLOW}Generating application files...${NC}"

# Config
generate_file "$TEMPLATES_DIR/database.go.tmpl" "$PROJECT_DIR/application/config/database.go"

# Models
generate_file "$TEMPLATES_DIR/user.go.tmpl" "$PROJECT_DIR/application/models/user.go"
generate_file "$TEMPLATES_DIR/setting.go.tmpl" "$PROJECT_DIR/application/models/setting.go"

# Controllers
generate_file "$TEMPLATES_DIR/auth.go.tmpl" "$PROJECT_DIR/application/controllers/auth.go"
generate_file "$TEMPLATES_DIR/dashboard.go.tmpl" "$PROJECT_DIR/application/controllers/admin/dashboard.go"
generate_file "$TEMPLATES_DIR/users.go.tmpl" "$PROJECT_DIR/application/controllers/admin/users.go"
generate_file "$TEMPLATES_DIR/settings.go.tmpl" "$PROJECT_DIR/application/controllers/admin/settings.go"

# Libs
generate_file "$TEMPLATES_DIR/auth_lib.go.tmpl" "$PROJECT_DIR/application/libs/auth.go"

# Database
generate_file "$TEMPLATES_DIR/seed.go.tmpl" "$PROJECT_DIR/database/seed.go"

echo -e "${YELLOW}Generating views...${NC}"

# Layouts
generate_file "$TEMPLATES_DIR/main.html.tmpl" "$PROJECT_DIR/application/views/layouts/main.html"

# Auth views
generate_file "$TEMPLATES_DIR/login.html.tmpl" "$PROJECT_DIR/application/views/auth/login.html"

# Admin views
generate_file "$TEMPLATES_DIR/dashboard.html.tmpl" "$PROJECT_DIR/application/views/admin/dashboard.html"
generate_file "$TEMPLATES_DIR/users_index.html.tmpl" "$PROJECT_DIR/application/views/admin/users/index.html"
generate_file "$TEMPLATES_DIR/users_form.html.tmpl" "$PROJECT_DIR/application/views/admin/users/_form.html"
generate_file "$TEMPLATES_DIR/users_row.html.tmpl" "$PROJECT_DIR/application/views/admin/users/_row.html"
generate_file "$TEMPLATES_DIR/settings_index.html.tmpl" "$PROJECT_DIR/application/views/admin/settings/index.html"

# Partials
generate_file "$TEMPLATES_DIR/header.html.tmpl" "$PROJECT_DIR/application/views/admin/partials/_header.html"
generate_file "$TEMPLATES_DIR/sidebar.html.tmpl" "$PROJECT_DIR/application/views/admin/partials/_sidebar.html"
generate_file "$TEMPLATES_DIR/footer.html.tmpl" "$PROJECT_DIR/application/views/admin/partials/_footer.html"
generate_file "$TEMPLATES_DIR/messages.html.tmpl" "$PROJECT_DIR/application/views/partials/_messages.html"
generate_file "$TEMPLATES_DIR/pagination.html.tmpl" "$PROJECT_DIR/application/views/partials/_pagination.html"

echo -e "${YELLOW}Generating static files...${NC}"

# CSS
generate_file "$TEMPLATES_DIR/style.css.tmpl" "$PROJECT_DIR/public/static/css/style.css"

echo -e "${GREEN}✓ All files generated${NC}"