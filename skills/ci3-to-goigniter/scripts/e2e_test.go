package scripts

import (
	"bytes"
	"encoding/json"
	"go/parser"
	"go/token"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// findFixtureDir locates the ci3_fixture directory regardless of current working directory.
func findFixtureDir(t *testing.T) string {
	t.Helper()
	candidates := []string{
		filepath.Join("..", "testdata", "ci3_fixture"),
		filepath.Join("skills", "ci3-to-goigniter", "testdata", "ci3_fixture"),
		filepath.Join(".", "testdata", "ci3_fixture"),
	}

	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			abs, err := filepath.Abs(c)
			if err == nil {
				return abs
			}
			return c
		}
	}

	t.Fatalf("could not locate ci3_fixture directory from candidates: %v", candidates)
	return ""
}

func TestE2E(t *testing.T) {
	fixtureDir := findFixtureDir(t)

	// Subtest 1: InspectProject
	t.Run("InspectProject", func(t *testing.T) {
		inv, err := InspectProject(fixtureDir)
		if err != nil {
			t.Fatalf("InspectProject failed: %v", err)
		}

		if inv == nil {
			t.Fatal("expected non-nil ProjectInventory")
		}

		// Verify Controllers
		if len(inv.Controllers) != 1 {
			t.Fatalf("expected 1 controller, got %d", len(inv.Controllers))
		}
		ctrl := inv.Controllers[0]
		if ctrl.Name != "Users" {
			t.Errorf("expected controller name 'Users', got %q", ctrl.Name)
		}
		if !strings.HasSuffix(filepath.ToSlash(ctrl.File), "application/controllers/Users.php") {
			t.Errorf("unexpected controller file path: %q", ctrl.File)
		}
		expectedCtrlMethods := []string{"create", "detail", "index"}
		for _, m := range expectedCtrlMethods {
			found := false
			for _, act := range ctrl.Methods {
				if act == m {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("controller missing expected action %q, found: %v", m, ctrl.Methods)
			}
		}

		// Verify Models
		if len(inv.Models) != 1 {
			t.Fatalf("expected 1 model, got %d", len(inv.Models))
		}
		model := inv.Models[0]
		if model.Name != "User_model" {
			t.Errorf("expected model name 'User_model', got %q", model.Name)
		}
		if model.Table != "users" {
			t.Errorf("expected model table 'users', got %q", model.Table)
		}
		if !strings.HasSuffix(filepath.ToSlash(model.File), "application/models/User_model.php") {
			t.Errorf("unexpected model file path: %q", model.File)
		}
		expectedModelMethods := []string{"delete_user", "get_all", "get_by_id", "insert_user", "update_user"}
		for _, m := range expectedModelMethods {
			found := false
			for _, act := range model.Methods {
				if act == m {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("model missing expected method %q, found: %v", m, model.Methods)
			}
		}

		// Verify Views
		if len(inv.Views) != 2 {
			t.Fatalf("expected 2 views, got %d (%v)", len(inv.Views), inv.Views)
		}
		if inv.Views[0] != "users/detail.php" || inv.Views[1] != "users/index.php" {
			t.Errorf("unexpected views list: %v", inv.Views)
		}

		// Verify Routes
		if len(inv.Routes) < 6 {
			t.Fatalf("expected at least 6 routes, got %d", len(inv.Routes))
		}
		routeMap := make(map[string]RouteInfo)
		for _, r := range inv.Routes {
			routeMap[r.Pattern] = r
		}

		if r, ok := routeMap["default_controller"]; !ok || !r.IsSpecial || r.Target != "users" {
			t.Errorf("default_controller route missing or incorrect: %+v", r)
		}
		if r, ok := routeMap["404_override"]; !ok || !r.IsSpecial {
			t.Errorf("404_override route missing or incorrect: %+v", r)
		}
		if r, ok := routeMap["users"]; !ok || r.Target != "users/index" {
			t.Errorf("users route missing or incorrect: %+v", r)
		}
		if r, ok := routeMap["users/create"]; !ok || r.Target != "users/create" {
			t.Errorf("users/create route missing or incorrect: %+v", r)
		}
		if r, ok := routeMap["users/(:num)"]; !ok || r.Target != "users/detail/$1" {
			t.Errorf("users/(:num) route missing or incorrect: %+v", r)
		}
		if r, ok := routeMap["api/users"]; !ok || r.Target != "users/api_index" || r.HTTPMethod != "GET" {
			t.Errorf("api/users route missing or incorrect: %+v", r)
		}

		// Verify Autoload
		if len(inv.Autoload.Libraries) != 2 || inv.Autoload.Libraries[0] != "database" || inv.Autoload.Libraries[1] != "session" {
			t.Errorf("unexpected autoload libraries: %v", inv.Autoload.Libraries)
		}
		if len(inv.Autoload.Helpers) != 2 || inv.Autoload.Helpers[0] != "url" || inv.Autoload.Helpers[1] != "form" {
			t.Errorf("unexpected autoload helpers: %v", inv.Autoload.Helpers)
		}
		if len(inv.Autoload.Models) != 1 || inv.Autoload.Models[0] != "user_model" {
			t.Errorf("unexpected autoload models: %v", inv.Autoload.Models)
		}

		// Verify Stats
		if inv.Stats.TotalControllers != 1 {
			t.Errorf("expected TotalControllers 1, got %d", inv.Stats.TotalControllers)
		}
		if inv.Stats.TotalMethods != 3 {
			t.Errorf("expected TotalMethods 3, got %d", inv.Stats.TotalMethods)
		}
		if inv.Stats.TotalModels != 1 {
			t.Errorf("expected TotalModels 1, got %d", inv.Stats.TotalModels)
		}
		if inv.Stats.TotalViews != 2 {
			t.Errorf("expected TotalViews 2, got %d", inv.Stats.TotalViews)
		}
		if inv.Stats.TotalRoutes != len(inv.Routes) {
			t.Errorf("expected TotalRoutes %d, got %d", len(inv.Routes), inv.Stats.TotalRoutes)
		}

		// Verify Markdown generation
		md := inv.GenerateMarkdown()
		if len(md) == 0 {
			t.Fatal("GenerateMarkdown returned empty string")
		}
		expectedMDSections := []string{
			"# CodeIgniter 3 Migration Inventory Report",
			"**Total Controllers:**",
			"**Total Action Methods:**",
			"**Total Models:**",
			"**Total Views:**",
			"**Total Routes:**",
			"## Controllers",
			"## Models",
			"## Routes",
			"## Views",
			"## Autoload Configuration",
			"Users",
			"User_model",
			"users/index.php",
			"users/detail.php",
		}
		for _, s := range expectedMDSections {
			if !strings.Contains(md, s) {
				t.Errorf("markdown missing expected content %q", s)
			}
		}

		// Verify JSON generation
		jsonData, err := inv.GenerateJSON()
		if err != nil {
			t.Fatalf("GenerateJSON failed: %v", err)
		}
		var parsedInv ProjectInventory
		if err := json.Unmarshal(jsonData, &parsedInv); err != nil {
			t.Fatalf("failed to unmarshal generated JSON: %v", err)
		}
		if parsedInv.Stats.TotalControllers != inv.Stats.TotalControllers {
			t.Errorf("unmarshaled stats do not match original: %v vs %v", parsedInv.Stats, inv.Stats)
		}
	})

	// Subtest 2: SQL2Struct
	t.Run("SQL2Struct", func(t *testing.T) {
		sqlPath := filepath.Join(fixtureDir, "schema.sql")
		sqlContent, err := os.ReadFile(sqlPath)
		if err != nil {
			t.Fatalf("failed reading schema.sql: %v", err)
		}

		tables, err := ParseSQLSchema(string(sqlContent))
		if err != nil {
			t.Fatalf("ParseSQLSchema failed: %v", err)
		}

		if len(tables) != 2 {
			t.Fatalf("expected 2 tables from schema.sql, got %d", len(tables))
		}

		tableMap := make(map[string]TableSchema)
		for _, tbl := range tables {
			tableMap[tbl.TableName] = tbl
		}

		// Verify users table schema
		usersTable, ok := tableMap["users"]
		if !ok {
			t.Fatal("users table not found in parsed schema")
		}
		if usersTable.StructName != "User" {
			t.Errorf("expected struct name 'User', got %q", usersTable.StructName)
		}
		if len(usersTable.PrimaryKeys) != 1 || usersTable.PrimaryKeys[0] != "id" {
			t.Errorf("expected primary key ['id'], got %v", usersTable.PrimaryKeys)
		}

		// Verify orders table schema
		ordersTable, ok := tableMap["orders"]
		if !ok {
			t.Fatal("orders table not found in parsed schema")
		}
		if ordersTable.StructName != "Order" {
			t.Errorf("expected struct name 'Order', got %q", ordersTable.StructName)
		}
		if len(ordersTable.PrimaryKeys) != 1 || ordersTable.PrimaryKeys[0] != "id" {
			t.Errorf("expected primary key ['id'], got %v", ordersTable.PrimaryKeys)
		}

		// Verify multi-model combined Go code generation and parsing
		fset := token.NewFileSet()
		combinedCode, err := GenerateMultiModelGoCode(tables, "models")
		if err != nil {
			t.Fatalf("GenerateMultiModelGoCode failed: %v", err)
		}
		if _, err := parser.ParseFile(fset, "schema_models.go", combinedCode, parser.AllErrors); err != nil {
			t.Fatalf("combined Go code failed to parse with go/parser: %v\nCode:\n%s", err, combinedCode)
		}

		// Verify single model Go code generation for each table
		userCode, err := GenerateModelGoCode(usersTable, "models")
		if err != nil {
			t.Fatalf("GenerateModelGoCode for users failed: %v", err)
		}
		if _, err := parser.ParseFile(fset, "user.go", userCode, parser.AllErrors); err != nil {
			t.Fatalf("User model Go code failed to parse with go/parser: %v\nCode:\n%s", err, userCode)
		}
		if !strings.Contains(userCode, "type User struct") {
			t.Errorf("user code missing User struct definition")
		}
		if !strings.Contains(userCode, "type UserModel struct") {
			t.Errorf("user code missing UserModel struct definition")
		}

		orderCode, err := GenerateModelGoCode(ordersTable, "models")
		if err != nil {
			t.Fatalf("GenerateModelGoCode for orders failed: %v", err)
		}
		if _, err := parser.ParseFile(fset, "order.go", orderCode, parser.AllErrors); err != nil {
			t.Fatalf("Order model Go code failed to parse with go/parser: %v\nCode:\n%s", err, orderCode)
		}
		if !strings.Contains(orderCode, "type Order struct") {
			t.Errorf("order code missing Order struct definition")
		}
		if !strings.Contains(orderCode, "type OrderModel struct") {
			t.Errorf("order code missing OrderModel struct definition")
		}
	})

	// Subtest 3: View2GoTpl
	t.Run("View2GoTpl", func(t *testing.T) {
		srcViewsDir := filepath.Join(fixtureDir, "application", "views")
		outViewsDir := t.TempDir()

		count, err := TranspileViewsDir(srcViewsDir, outViewsDir)
		if err != nil {
			t.Fatalf("TranspileViewsDir failed: %v", err)
		}
		if count != 2 {
			t.Fatalf("expected 2 transpiled views, got %d", count)
		}

		// Check converted files exist with .html extension
		indexHTMLPath := filepath.Join(outViewsDir, "users", "index.html")
		detailHTMLPath := filepath.Join(outViewsDir, "users", "detail.html")

		if _, err := os.Stat(indexHTMLPath); err != nil {
			t.Fatalf("expected index.html to exist at %s: %v", indexHTMLPath, err)
		}
		if _, err := os.Stat(detailHTMLPath); err != nil {
			t.Fatalf("expected detail.html to exist at %s: %v", detailHTMLPath, err)
		}

		// Ensure old .php files do not exist in the output directory
		if _, err := os.Stat(filepath.Join(outViewsDir, "users", "index.php")); !os.IsNotExist(err) {
			t.Errorf("unexpected index.php in output directory")
		}
		if _, err := os.Stat(filepath.Join(outViewsDir, "users", "detail.php")); !os.IsNotExist(err) {
			t.Errorf("unexpected detail.php in output directory")
		}

		// Verify no PHP tags remain in transpiled templates
		for _, f := range []string{indexHTMLPath, detailHTMLPath} {
			content, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("failed reading %s: %v", f, err)
			}
			str := string(content)
			if strings.Contains(str, "<?php") || strings.Contains(str, "<?=") || strings.Contains(str, "?>") {
				t.Errorf("file %s still contains PHP tags:\n%s", f, str)
			}
		}

		// Verify standard html/template.ParseFiles succeeds without error
		tplIndex, err := template.ParseFiles(indexHTMLPath)
		if err != nil {
			indexContent, _ := os.ReadFile(indexHTMLPath)
			t.Fatalf("html/template.ParseFiles failed on index.html: %v\nContent:\n%s", err, string(indexContent))
		}

		tplDetail, err := template.ParseFiles(detailHTMLPath)
		if err != nil {
			detailContent, _ := os.ReadFile(detailHTMLPath)
			t.Fatalf("html/template.ParseFiles failed on detail.html: %v\nContent:\n%s", err, string(detailContent))
		}

		// Test rendering index.html with representative data
		type ViewUser struct {
			ID       int
			Name     string
			Email    string
			IsActive bool
		}
		type IndexData struct {
			Title string
			Users []ViewUser
		}

		var bufIndex bytes.Buffer
		indexData := IndexData{
			Title: "Users List",
			Users: []ViewUser{
				{ID: 1, Name: "Alice Smith", Email: "alice@example.com", IsActive: true},
				{ID: 2, Name: "Bob Jones", Email: "bob@example.com", IsActive: false},
			},
		}
		if err := tplIndex.Execute(&bufIndex, indexData); err != nil {
			t.Fatalf("executing parsed index template failed: %v", err)
		}
		renderedIndex := bufIndex.String()
		if !strings.Contains(renderedIndex, "Alice Smith") || !strings.Contains(renderedIndex, "alice@example.com") {
			t.Errorf("rendered index missing user Alice: %s", renderedIndex)
		}
		if !strings.Contains(renderedIndex, "Active") || !strings.Contains(renderedIndex, "Inactive") {
			t.Errorf("rendered index missing status badges: %s", renderedIndex)
		}
		if !strings.Contains(renderedIndex, "/users/detail/1") {
			t.Errorf("rendered index missing detail link: %s", renderedIndex)
		}

		// Test rendering detail.html with representative data
		type DetailData struct {
			Title string
			User  ViewUser
		}
		var bufDetail bytes.Buffer
		detailData := DetailData{
			Title: "User Detail",
			User: ViewUser{
				ID:       1,
				Name:     "Alice Smith",
				Email:    "alice@example.com",
				IsActive: true,
			},
		}
		if err := tplDetail.Execute(&bufDetail, detailData); err != nil {
			t.Fatalf("executing parsed detail template failed: %v", err)
		}
		renderedDetail := bufDetail.String()
		if !strings.Contains(renderedDetail, "Alice Smith") || !strings.Contains(renderedDetail, "alice@example.com") {
			t.Errorf("rendered detail missing user info: %s", renderedDetail)
		}
		if !strings.Contains(renderedDetail, "/users") {
			t.Errorf("rendered detail missing back link: %s", renderedDetail)
		}
	})

	// Subtest 4: CLI Execution of all 3 subcommands
	t.Run("CLIExecution", func(t *testing.T) {
		tempDir := t.TempDir()

		// 1. CLI inspect subcommand
		outMD := filepath.Join(tempDir, "report.md")
		err := RunInspectCLI([]string{
			"-src", fixtureDir,
			"-out", outMD,
			"-format", "markdown",
		})
		if err != nil {
			t.Fatalf("RunInspectCLI markdown failed: %v", err)
		}
		mdBytes, err := os.ReadFile(outMD)
		if err != nil || !strings.Contains(string(mdBytes), "# CodeIgniter 3 Migration Inventory Report") {
			t.Fatalf("RunInspectCLI output invalid: %v, content:\n%s", err, string(mdBytes))
		}

		outJSON := filepath.Join(tempDir, "report.json")
		err = RunInspectCLI([]string{
			"-src", fixtureDir,
			"-out", outJSON,
			"-format", "json",
		})
		if err != nil {
			t.Fatalf("RunInspectCLI json failed: %v", err)
		}
		jsonBytes, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading inspect JSON output: %v", err)
		}
		var cliInv ProjectInventory
		if err := json.Unmarshal(jsonBytes, &cliInv); err != nil {
			t.Fatalf("failed unmarshaling inspect JSON output: %v", err)
		}
		if cliInv.Stats.TotalControllers != 1 || cliInv.Stats.TotalModels != 1 {
			t.Errorf("unexpected stats from CLI inspect: %+v", cliInv.Stats)
		}

		// 2. CLI sql2struct subcommand
		outModelsDir := filepath.Join(tempDir, "models")
		err = RunSQL2StructCLI([]string{
			"-sql", filepath.Join(fixtureDir, "schema.sql"),
			"-out", outModelsDir,
			"-pkg", "models",
		})
		if err != nil {
			t.Fatalf("RunSQL2StructCLI directory output failed: %v", err)
		}

		fset := token.NewFileSet()
		userModelFile := filepath.Join(outModelsDir, "user.go")
		orderModelFile := filepath.Join(outModelsDir, "order.go")
		for _, f := range []string{userModelFile, orderModelFile} {
			content, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("expected CLI generated model file %s: %v", f, err)
			}
			if _, err := parser.ParseFile(fset, filepath.Base(f), string(content), parser.AllErrors); err != nil {
				t.Fatalf("CLI generated model %s failed to parse: %v", f, err)
			}
		}

		// Combined single file CLI output
		outCombinedGo := filepath.Join(tempDir, "combined.go")
		err = RunSQL2StructCLI([]string{
			"-sql", filepath.Join(fixtureDir, "schema.sql"),
			"-out", outCombinedGo,
			"-pkg", "models",
		})
		if err != nil {
			t.Fatalf("RunSQL2StructCLI single file output failed: %v", err)
		}
		combinedBytes, err := os.ReadFile(outCombinedGo)
		if err != nil {
			t.Fatalf("failed reading CLI generated combined file: %v", err)
		}
		if _, err := parser.ParseFile(fset, "combined.go", string(combinedBytes), parser.AllErrors); err != nil {
			t.Fatalf("CLI generated combined file failed to parse: %v", err)
		}

		// 3. CLI view2gotpl subcommand
		outViewsDir := filepath.Join(tempDir, "views")
		err = RunView2GoTplCLI([]string{
			"-src", filepath.Join(fixtureDir, "application", "views"),
			"-out", outViewsDir,
		})
		if err != nil {
			t.Fatalf("RunView2GoTplCLI failed: %v", err)
		}

		cliIndexHTML := filepath.Join(outViewsDir, "users", "index.html")
		cliDetailHTML := filepath.Join(outViewsDir, "users", "detail.html")
		for _, tplFile := range []string{cliIndexHTML, cliDetailHTML} {
			if _, err := os.Stat(tplFile); err != nil {
				t.Fatalf("expected CLI converted template %s: %v", tplFile, err)
			}
			if _, err := template.ParseFiles(tplFile); err != nil {
				t.Fatalf("CLI converted template %s failed html/template.ParseFiles: %v", tplFile, err)
			}
		}
	})
}
