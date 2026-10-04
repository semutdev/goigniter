package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ControllerInfo contains metadata about a discovered CI3 controller.
type ControllerInfo struct {
	Name    string   `json:"name"`
	File    string   `json:"file"`
	Methods []string `json:"methods"`
}

// ModelInfo contains metadata about a discovered CI3 model.
type ModelInfo struct {
	Name    string   `json:"name"`
	File    string   `json:"file"`
	Table   string   `json:"table,omitempty"`
	Methods []string `json:"methods,omitempty"`
}

// RouteInfo contains metadata about a discovered route in routes.php.
type RouteInfo struct {
	Pattern    string `json:"pattern"`
	Target     string `json:"target"`
	HTTPMethod string `json:"http_method,omitempty"`
}

// AutoloadInfo captures autoloaded components from config/autoload.php.
type AutoloadInfo struct {
	Libraries []string `json:"libraries"`
	Helpers   []string `json:"helpers"`
	Models    []string `json:"models"`
}

// InventoryStats provides summary counts of discovered CI3 artifacts.
type InventoryStats struct {
	TotalControllers int `json:"total_controllers"`
	TotalMethods     int `json:"total_methods"`
	TotalModels       int `json:"total_models"`
	TotalViews       int `json:"total_views"`
	TotalRoutes      int `json:"total_routes"`
}

// ProjectInventory represents the complete scanned inventory of a CodeIgniter 3 application.
type ProjectInventory struct {
	SourceDir   string           `json:"source_dir"`
	Controllers []ControllerInfo `json:"controllers"`
	Models      []ModelInfo      `json:"models"`
	Views       []string         `json:"views"`
	Routes      []RouteInfo      `json:"routes"`
	Autoload    AutoloadInfo     `json:"autoload"`
	Stats       InventoryStats   `json:"stats"`
}

var (
	ctrlClassRegex  = regexp.MustCompile(`(?i)class\s+(\w+)\s+extends\s+(?:CI_Controller|\w+)`)
	modelClassRegex = regexp.MustCompile(`(?i)class\s+(\w+)\s+extends\s+(?:CI_Model|\w+)`)
	methodRegex     = regexp.MustCompile(`(?i)(?:^|[\s;{}])(?:(public|protected|private)\s+)?function\s+(\w+)\s*\(`)
	modelTableRegex = regexp.MustCompile(`(?i)(?:protected|public|private)?\s*\$table\s*=\s*['"]([^'"]+)['"]`)

	// Routes regexes
	simpleRouteRegex = regexp.MustCompile(`\$route\['([^']+)'\]\s*=\s*['"]([^'"]*)['"];`)
	methodRouteRegex = regexp.MustCompile(`(?i)\$route\['([^']+)'\]\['(get|post|put|delete|patch|options|head)'\]\s*=\s*['"]([^'"]*)['"];`)

	// Autoload regex
	autoloadRegex = regexp.MustCompile(`(?i)\$autoload\['(libraries|helper|model)'\]\s*=\s*(?:array\(|\[)([^);\]]*)(?:\)|\]);`)
)

// InspectProject scans a legacy CodeIgniter 3 project directory and returns a structured inventory.
func InspectProject(srcDir string) (*ProjectInventory, error) {
	fi, err := os.Stat(srcDir)
	if err != nil {
		return nil, fmt.Errorf("invalid source directory: %w", err)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("source path is not a directory: %s", srcDir)
	}

	appDir := filepath.Join(srcDir, "application")
	if fi, err := os.Stat(appDir); err != nil || !fi.IsDir() {
		// Fallback: check if srcDir itself is the application directory
		ctrlCheck := filepath.Join(srcDir, "controllers")
		if cfi, err := os.Stat(ctrlCheck); err == nil && cfi.IsDir() {
			appDir = srcDir
		}
	}

	inv := &ProjectInventory{
		SourceDir:   srcDir,
		Controllers: make([]ControllerInfo, 0),
		Models:      make([]ModelInfo, 0),
		Views:       make([]string, 0),
		Routes:      make([]RouteInfo, 0),
		Autoload: AutoloadInfo{
			Libraries: make([]string, 0),
			Helpers:   make([]string, 0),
			Models:    make([]string, 0),
		},
	}

	// 1. Scan Controllers
	ctrlDir := filepath.Join(appDir, "controllers")
	if _, err := os.Stat(ctrlDir); err == nil {
		_ = filepath.WalkDir(ctrlDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".php") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			ctrl := parseController(path, srcDir, string(content))
			inv.Controllers = append(inv.Controllers, ctrl)
			return nil
		})
	}

	// Sort controllers by name
	sort.Slice(inv.Controllers, func(i, j int) bool {
		return inv.Controllers[i].Name < inv.Controllers[j].Name
	})

	// 2. Scan Models
	modelDir := filepath.Join(appDir, "models")
	if _, err := os.Stat(modelDir); err == nil {
		_ = filepath.WalkDir(modelDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".php") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			model := parseModel(path, srcDir, string(content))
			inv.Models = append(inv.Models, model)
			return nil
		})
	}

	// Sort models by name
	sort.Slice(inv.Models, func(i, j int) bool {
		return inv.Models[i].Name < inv.Models[j].Name
	})

	// 3. Scan Views
	viewDir := filepath.Join(appDir, "views")
	if _, err := os.Stat(viewDir); err == nil {
		_ = filepath.WalkDir(viewDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(d.Name()))
			if ext == ".php" || ext == ".html" {
				rel, err := filepath.Rel(viewDir, path)
				if err == nil {
					inv.Views = append(inv.Views, filepath.ToSlash(rel))
				}
			}
			return nil
		})
	}
	sort.Strings(inv.Views)

	// 4. Scan Routes
	routesFile := filepath.Join(appDir, "config", "routes.php")
	if content, err := os.ReadFile(routesFile); err == nil {
		inv.Routes = parseRoutes(string(content))
	}

	// 5. Scan Autoload
	autoloadFile := filepath.Join(appDir, "config", "autoload.php")
	if content, err := os.ReadFile(autoloadFile); err == nil {
		inv.Autoload = parseAutoload(string(content))
	}

	// Calculate Stats
	totalMethods := 0
	for _, c := range inv.Controllers {
		totalMethods += len(c.Methods)
	}
	inv.Stats = InventoryStats{
		TotalControllers: len(inv.Controllers),
		TotalMethods:     totalMethods,
		TotalModels:      len(inv.Models),
		TotalViews:       len(inv.Views),
		TotalRoutes:      len(inv.Routes),
	}

	return inv, nil
}

func parseController(fullPath, srcDir, content string) ControllerInfo {
	relPath, err := filepath.Rel(srcDir, fullPath)
	if err != nil {
		relPath = fullPath
	}
	relPath = filepath.ToSlash(relPath)

	name := strings.TrimSuffix(filepath.Base(fullPath), ".php")
	if match := ctrlClassRegex.FindStringSubmatch(content); len(match) > 1 {
		name = match[1]
	}

	var methods []string
	matches := methodRegex.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		visibility := strings.ToLower(m[1])
		methodName := m[2]

		// Skip private or protected methods
		if visibility == "private" || visibility == "protected" {
			continue
		}
		// In CI3, methods starting with '_' are not public action endpoints
		if strings.HasPrefix(methodName, "_") {
			continue
		}
		methods = append(methods, methodName)
	}

	if methods == nil {
		methods = make([]string, 0)
	}

	return ControllerInfo{
		Name:    name,
		File:    relPath,
		Methods: methods,
	}
}

func parseModel(fullPath, srcDir, content string) ModelInfo {
	relPath, err := filepath.Rel(srcDir, fullPath)
	if err != nil {
		relPath = fullPath
	}
	relPath = filepath.ToSlash(relPath)

	name := strings.TrimSuffix(filepath.Base(fullPath), ".php")
	if match := modelClassRegex.FindStringSubmatch(content); len(match) > 1 {
		name = match[1]
	}

	table := ""
	if match := modelTableRegex.FindStringSubmatch(content); len(match) > 1 {
		table = match[1]
	}

	var methods []string
	matches := methodRegex.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		visibility := strings.ToLower(m[1])
		methodName := m[2]
		if visibility == "private" || visibility == "protected" || strings.HasPrefix(methodName, "_") {
			continue
		}
		methods = append(methods, methodName)
	}
	if methods == nil {
		methods = make([]string, 0)
	}

	return ModelInfo{
		Name:    name,
		File:    relPath,
		Table:   table,
		Methods: methods,
	}
}

func parseRoutes(content string) []RouteInfo {
	routes := make([]RouteInfo, 0)

	// First match method-specific routes like $route['api/users']['post'] = 'users/create';
	methodMatches := methodRouteRegex.FindAllStringSubmatch(content, -1)
	for _, m := range methodMatches {
		pattern := m[1]
		httpMethod := strings.ToUpper(m[2])
		target := m[3]
		routes = append(routes, RouteInfo{
			Pattern:    pattern,
			Target:     target,
			HTTPMethod: httpMethod,
		})
	}

	// Next match standard routes like $route['users'] = 'users/index';
	simpleMatches := simpleRouteRegex.FindAllStringSubmatch(content, -1)
	for _, m := range simpleMatches {
		pattern := m[1]
		target := m[2]
		// Skip CI3 system config routes unless desired, but keep them for full accuracy
		routes = append(routes, RouteInfo{
			Pattern:    pattern,
			Target:     target,
			HTTPMethod: "ANY",
		})
	}

	return routes
}

func parseAutoload(content string) AutoloadInfo {
	info := AutoloadInfo{
		Libraries: make([]string, 0),
		Helpers:   make([]string, 0),
		Models:    make([]string, 0),
	}

	matches := autoloadRegex.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		category := strings.ToLower(m[1])
		rawItems := m[2]
		items := parsePHPArrayItems(rawItems)

		switch category {
		case "libraries":
			info.Libraries = append(info.Libraries, items...)
		case "helper":
			info.Helpers = append(info.Helpers, items...)
		case "model":
			info.Models = append(info.Models, items...)
		}
	}

	return info
}

func parsePHPArrayItems(raw string) []string {
	var items []string
	parts := strings.Split(raw, ",")
	itemRegex := regexp.MustCompile(`['"]([^'"]+)['"]`)
	for _, p := range parts {
		if m := itemRegex.FindStringSubmatch(p); len(m) > 1 {
			items = append(items, m[1])
		}
	}
	return items
}

// GenerateMarkdown formats the project inventory as a human- and agent-readable Markdown report.
func (inv *ProjectInventory) GenerateMarkdown() string {
	var sb strings.Builder

	sb.WriteString("# CodeIgniter 3 Migration Inventory Report\n\n")
	sb.WriteString(fmt.Sprintf("**Source Directory:** `%s`\n", inv.SourceDir))
	sb.WriteString(fmt.Sprintf("- **Total Controllers:** %d\n", inv.Stats.TotalControllers))
	sb.WriteString(fmt.Sprintf("- **Total Action Methods:** %d\n", inv.Stats.TotalMethods))
	sb.WriteString(fmt.Sprintf("- **Total Models:** %d\n", inv.Stats.TotalModels))
	sb.WriteString(fmt.Sprintf("- **Total Views:** %d\n", inv.Stats.TotalViews))
	sb.WriteString(fmt.Sprintf("- **Total Routes:** %d\n\n", inv.Stats.TotalRoutes))

	// Controllers Table
	sb.WriteString("## Controllers\n\n")
	if len(inv.Controllers) == 0 {
		sb.WriteString("_No controllers found._\n\n")
	} else {
		sb.WriteString("| Controller | File | Action Methods |\n")
		sb.WriteString("|---|---|---|\n")
		for _, c := range inv.Controllers {
			var methodList string
			if len(c.Methods) > 0 {
				var formatted []string
				for _, m := range c.Methods {
					formatted = append(formatted, fmt.Sprintf("`%s()`", m))
				}
				methodList = strings.Join(formatted, ", ")
			} else {
				methodList = "*(none)*"
			}
			sb.WriteString(fmt.Sprintf("| **%s** | `%s` | %s |\n", c.Name, c.File, methodList))
		}
		sb.WriteString("\n")
	}

	// Models Table
	sb.WriteString("## Models\n\n")
	if len(inv.Models) == 0 {
		sb.WriteString("_No models found._\n\n")
	} else {
		sb.WriteString("| Model | File | Table | Methods |\n")
		sb.WriteString("|---|---|---|---|\n")
		for _, m := range inv.Models {
			tableStr := "-"
			if m.Table != "" {
				tableStr = fmt.Sprintf("`%s`", m.Table)
			}
			var methodList string
			if len(m.Methods) > 0 {
				var formatted []string
				for _, fn := range m.Methods {
					formatted = append(formatted, fmt.Sprintf("`%s()`", fn))
				}
				methodList = strings.Join(formatted, ", ")
			} else {
				methodList = "*(none)*"
			}
			sb.WriteString(fmt.Sprintf("| **%s** | `%s` | %s | %s |\n", m.Name, m.File, tableStr, methodList))
		}
		sb.WriteString("\n")
	}

	// Routes Table
	sb.WriteString("## Routes\n\n")
	if len(inv.Routes) == 0 {
		sb.WriteString("_No custom routes found._\n\n")
	} else {
		sb.WriteString("| Route Pattern | HTTP Method | Target Controller/Action |\n")
		sb.WriteString("|---|---|---|\n")
		for _, r := range inv.Routes {
			method := r.HTTPMethod
			if method == "" {
				method = "ANY"
			}
			sb.WriteString(fmt.Sprintf("| `%s` | **%s** | `%s` |\n", r.Pattern, method, r.Target))
		}
		sb.WriteString("\n")
	}

	// Views List
	sb.WriteString("## Views\n\n")
	if len(inv.Views) == 0 {
		sb.WriteString("_No views found._\n\n")
	} else {
		for _, v := range inv.Views {
			sb.WriteString(fmt.Sprintf("- `%s`\n", v))
		}
		sb.WriteString("\n")
	}

	// Autoload Configuration
	sb.WriteString("## Autoload Configuration\n\n")
	if len(inv.Autoload.Libraries) > 0 {
		sb.WriteString(fmt.Sprintf("- **Libraries:** `%s`\n", strings.Join(inv.Autoload.Libraries, "`, `")))
	}
	if len(inv.Autoload.Helpers) > 0 {
		sb.WriteString(fmt.Sprintf("- **Helpers:** `%s`\n", strings.Join(inv.Autoload.Helpers, "`, `")))
	}
	if len(inv.Autoload.Models) > 0 {
		sb.WriteString(fmt.Sprintf("- **Models:** `%s`\n", strings.Join(inv.Autoload.Models, "`, `")))
	}
	sb.WriteString("\n")

	return sb.String()
}

// GenerateJSON serializes the project inventory as indented JSON.
func (inv *ProjectInventory) GenerateJSON() ([]byte, error) {
	return json.MarshalIndent(inv, "", "  ")
}

func main() {
	src := flag.String("src", ".", "Path to CodeIgniter 3 project root")
	out := flag.String("out", "", "Output file path (default stdout)")
	format := flag.String("format", "markdown", "Output format: 'markdown' or 'json'")
	flag.Parse()

	inventory, err := InspectProject(*src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inspecting project: %v\n", err)
		os.Exit(1)
	}

	var output []byte
	switch strings.ToLower(*format) {
	case "json":
		data, err := inventory.GenerateJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
			os.Exit(1)
		}
		output = data
	default:
		output = []byte(inventory.GenerateMarkdown())
	}

	if *out == "" || *out == "-" {
		fmt.Println(string(output))
	} else {
		if err := os.WriteFile(*out, output, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Inventory successfully generated: %s\n", *out)
	}
}
