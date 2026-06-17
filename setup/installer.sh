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