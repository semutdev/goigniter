package scripts

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSQL2Struct_ParseMySQL(t *testing.T) {
	sql := `
-- Sample MySQL Schema
CREATE TABLE IF NOT EXISTS ` + "`users`" + ` (
  ` + "`id`" + ` INT AUTO_INCREMENT PRIMARY KEY,
  ` + "`username`" + ` VARCHAR(50) NOT NULL,
  ` + "`email`" + ` VARCHAR(100) NOT NULL UNIQUE,
  ` + "`bio`" + ` TEXT NULL,
  ` + "`is_active`" + ` TINYINT(1) DEFAULT 1,
  ` + "`balance`" + ` DECIMAL(10,2) DEFAULT 0.00,
  ` + "`created_at`" + ` DATETIME NOT NULL,
  ` + "`updated_at`" + ` TIMESTAMP NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`

	schemas, err := ParseSQLSchema(sql)
	if err != nil {
		t.Fatalf("ParseSQLSchema failed: %v", err)
	}

	if len(schemas) != 1 {
		t.Fatalf("expected 1 table schema, got %d", len(schemas))
	}

	table := schemas[0]
	if table.TableName != "users" {
		t.Errorf("expected TableName 'users', got '%s'", table.TableName)
	}
	if table.StructName != "User" {
		t.Errorf("expected StructName 'User', got '%s'", table.StructName)
	}
	if len(table.PrimaryKeys) != 1 || table.PrimaryKeys[0] != "id" {
		t.Errorf("expected PrimaryKeys ['id'], got %v", table.PrimaryKeys)
	}

	expectedCols := []struct {
		name      string
		goName    string
		sqlType   string
		goType    string
		nullable  bool
		isPrimary bool
	}{
		{"id", "ID", "INT", "int", false, true},
		{"username", "Username", "VARCHAR(50)", "string", false, false},
		{"email", "Email", "VARCHAR(100)", "string", false, false},
		{"bio", "Bio", "TEXT", "*string", true, false},
		{"is_active", "IsActive", "TINYINT(1)", "bool", false, false},
		{"balance", "Balance", "DECIMAL(10,2)", "float64", false, false},
		{"created_at", "CreatedAt", "DATETIME", "time.Time", false, false},
		{"updated_at", "UpdatedAt", "TIMESTAMP", "*time.Time", true, false},
	}

	if len(table.Columns) != len(expectedCols) {
		t.Fatalf("expected %d columns, got %d", len(expectedCols), len(table.Columns))
	}

	for i, exp := range expectedCols {
		col := table.Columns[i]
		if col.Name != exp.name {
			t.Errorf("col[%d] name: expected '%s', got '%s'", i, exp.name, col.Name)
		}
		if col.GoName != exp.goName {
			t.Errorf("col[%d] goName: expected '%s', got '%s'", i, exp.goName, col.GoName)
		}
		if !strings.EqualFold(col.SQLType, exp.sqlType) {
			t.Errorf("col[%d] sqlType: expected '%s', got '%s'", i, exp.sqlType, col.SQLType)
		}
		if col.GoType != exp.goType {
			t.Errorf("col[%d] goType: expected '%s', got '%s'", i, exp.goType, col.GoType)
		}
		if col.Nullable != exp.nullable {
			t.Errorf("col[%d] nullable: expected %v, got %v", i, exp.nullable, col.Nullable)
		}
		if col.IsPrimary != exp.isPrimary {
			t.Errorf("col[%d] isPrimary: expected %v, got %v", i, exp.isPrimary, col.IsPrimary)
		}
	}
}

func TestSQL2Struct_ParseSQLiteAndTableLevelPK(t *testing.T) {
	sql := `
CREATE TABLE "order_items" (
  "order_id" BIGINT NOT NULL,
  "item_id" INT NOT NULL,
  "quantity" INT NOT NULL DEFAULT 1,
  "unit_price" FLOAT NOT NULL,
  "notes" VARCHAR(255) NULL,
  PRIMARY KEY ("order_id", "item_id")
);
`

	schemas, err := ParseSQLSchema(sql)
	if err != nil {
		t.Fatalf("ParseSQLSchema failed: %v", err)
	}

	if len(schemas) != 1 {
		t.Fatalf("expected 1 table schema, got %d", len(schemas))
	}

	table := schemas[0]
	if table.TableName != "order_items" {
		t.Errorf("expected TableName 'order_items', got '%s'", table.TableName)
	}
	if table.StructName != "OrderItem" {
		t.Errorf("expected StructName 'OrderItem', got '%s'", table.StructName)
	}
	if len(table.PrimaryKeys) != 2 || table.PrimaryKeys[0] != "order_id" || table.PrimaryKeys[1] != "item_id" {
		t.Errorf("expected PrimaryKeys ['order_id', 'item_id'], got %v", table.PrimaryKeys)
	}

	colMap := make(map[string]ColumnSchema)
	for _, c := range table.Columns {
		colMap[c.Name] = c
	}

	if col, ok := colMap["order_id"]; !ok || col.GoType != "int64" || !col.IsPrimary {
		t.Errorf("unexpected order_id column: %+v", col)
	}
	if col, ok := colMap["item_id"]; !ok || col.GoType != "int" || !col.IsPrimary {
		t.Errorf("unexpected item_id column: %+v", col)
	}
	if col, ok := colMap["unit_price"]; !ok || col.GoType != "float64" {
		t.Errorf("unexpected unit_price column: %+v", col)
	}
	if col, ok := colMap["notes"]; !ok || col.GoType != "*string" || !col.Nullable {
		t.Errorf("unexpected notes column: %+v", col)
	}
}

func TestSQL2Struct_GenerateModelGoCode(t *testing.T) {
	sql := `
CREATE TABLE users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(50) NOT NULL,
  email VARCHAR(100) NOT NULL UNIQUE,
  bio TEXT NULL,
  is_active TINYINT(1) DEFAULT 1,
  balance DECIMAL(10,2) DEFAULT 0.00,
  created_at DATETIME NOT NULL,
  updated_at TIMESTAMP NULL
);
`

	schemas, err := ParseSQLSchema(sql)
	if err != nil {
		t.Fatalf("ParseSQLSchema failed: %v", err)
	}

	code, err := GenerateModelGoCode(schemas[0], "models")
	if err != nil {
		t.Fatalf("GenerateModelGoCode failed: %v", err)
	}

	// Verify generated code parses as valid Go
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "user.go", code, parser.AllErrors); err != nil {
		t.Fatalf("Generated code is not valid Go: %v\nCode:\n%s", err, code)
	}

	// Check package and imports
	if !strings.Contains(code, "package models") {
		t.Errorf("missing package declaration in generated code")
	}
	if !strings.Contains(code, `"github.com/semutdev/goigniter/system/libraries/database"`) {
		t.Errorf("missing database package import in generated code")
	}
	if !strings.Contains(code, `"time"`) {
		t.Errorf("missing time package import in generated code")
	}

	// Check struct definition and tags
	structChecks := []string{
		"type User struct {",
		"`json:\"id\" db:\"id\"`",
		"`json:\"username\" db:\"username\"`",
		"`json:\"email\" db:\"email\"`",
		"`json:\"bio\" db:\"bio\"`",
		"`json:\"is_active\" db:\"is_active\"`",
		"`json:\"balance\" db:\"balance\"`",
		"`json:\"created_at\" db:\"created_at\"`",
		"`json:\"updated_at\" db:\"updated_at\"`",
	}
	for _, check := range structChecks {
		if !strings.Contains(code, check) {
			t.Errorf("generated code missing struct element: %s", check)
		}
	}

	// Check TableName method
	if !strings.Contains(code, `func (User) TableName() string`) || !strings.Contains(code, `return "users"`) {
		t.Errorf("missing or invalid TableName method")
	}

	// Check Model CRUD boilerplate methods required by prompt
	crudChecks := []string{
		"type UserModel struct",
		"func NewUserModel() *UserModel",
		"func (m *UserModel) Find(id int) (*User, error)",
		"func (m *UserModel) FindAll() ([]User, error)",
		"func (m *UserModel) Insert(user *User) error",
		"func (m *UserModel) Update(user *User) error",
		"func (m *UserModel) Delete(id int) error",
	}
	for _, check := range crudChecks {
		if !strings.Contains(code, check) {
			t.Errorf("missing CRUD boilerplate method: %s\nGenerated code:\n%s", check, code)
		}
	}
}

func TestSQL2Struct_SingularizeAndCamelCase(t *testing.T) {
	testCases := []struct {
		table    string
		expected string
	}{
		{"users", "User"},
		{"order_items", "OrderItem"},
		{"categories", "Category"},
		{"companies", "Company"},
		{"addresses", "Address"},
		{"statuses", "Status"},
		{"user_profiles", "UserProfile"},
		{"posts", "Post"},
		{"people", "Person"},
		{"user", "User"},
	}

	for _, tc := range testCases {
		got := TableNameToStructName(tc.table)
		if got != tc.expected {
			t.Errorf("TableNameToStructName(%q): expected %q, got %q", tc.table, tc.expected, got)
		}
	}

	colCases := []struct {
		col      string
		expected string
	}{
		{"id", "ID"},
		{"user_id", "UserID"},
		{"ip_address", "IPAddress"},
		{"created_at", "CreatedAt"},
		{"is_active", "IsActive"},
		{"api_key", "APIKey"},
		{"url", "URL"},
	}

	for _, tc := range colCases {
		got := ColumnNameToFieldName(tc.col)
		if got != tc.expected {
			t.Errorf("ColumnNameToFieldName(%q): expected %q, got %q", tc.col, tc.expected, got)
		}
	}
}

func TestSQL2Struct_CLI(t *testing.T) {
	tempDir := t.TempDir()
	sqlFilePath := filepath.Join(tempDir, "schema.sql")
	sqlContent := `
CREATE TABLE products (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  price DECIMAL(10,2) NOT NULL DEFAULT 0.00
);
`
	if err := os.WriteFile(sqlFilePath, []byte(sqlContent), 0644); err != nil {
		t.Fatalf("failed to write test SQL file: %v", err)
	}

	outDir := filepath.Join(tempDir, "models")

	// Run CLI with -sql, -out, -pkg
	err := RunSQL2StructCLI([]string{
		"-sql", sqlFilePath,
		"-out", outDir,
		"-pkg", "custommodels",
	})
	if err != nil {
		t.Fatalf("RunSQL2StructCLI failed: %v", err)
	}

	outFile := filepath.Join(outDir, "product.go")
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("expected output file %s: %v", outFile, err)
	}

	code := string(content)
	if !strings.Contains(code, "package custommodels") {
		t.Errorf("expected package custommodels in output file")
	}
	if !strings.Contains(code, "type Product struct") {
		t.Errorf("expected Product struct in output file")
	}

	// Test missing -sql flag returns error
	if err := RunSQL2StructCLI([]string{}); err == nil {
		t.Errorf("expected error when -sql flag is missing, got nil")
	}
}
