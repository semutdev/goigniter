package scripts

import (
	"encoding/json"
	"errors"
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
	IsSpecial  bool   `json:"is_special,omitempty"`
	Notes      string `json:"notes,omitempty"`
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
	TotalModels      int `json:"total_models"`
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
	methodRegex     = regexp.MustCompile(`(?i)(?:^|[\s;{}])(?:static\s+|final\s+)*(?:(public|protected|private)\s+)?(?:static\s+|final\s+)*function\s+(\w+)\s*\(`)
	modelTableRegex = regexp.MustCompile(`(?i)(?:protected|public|private)?\s*\$table\s*=\s*['"]([^'"]+)['"]`)

	// Routes regexes with bracket spacing and single/double quotes
	simpleRouteRegex = regexp.MustCompile(`\$route\[\s*['"]([^'"]+)['"]\s*\]\s*=\s*['"]([^'"]*)['"]\s*;`)
	methodRouteRegex = regexp.MustCompile(`(?i)\$route\[\s*['"]([^'"]+)['"]\s*\]\s*\[\s*['"](get|post|put|delete|patch|options|head)['"]\s*\]\s*=\s*['"]([^'"]*)['"]\s*;`)

	// Autoload regex
	autoloadRegex = regexp.MustCompile(`(?i)\$autoload\['(libraries|helper|model)'\]\s*=\s*(?:array\(|\[)([^);\]]*)(?:\)|\]);`)

	// Array items regex compiled at package level
	itemRegex = regexp.MustCompile(`['"]([^'"]+)['"]`)
)

// stripPHPComments removes single-line (// and #) and multi-line (/* ... */) comments
// from PHP source code while preserving comments inside string literals.
func stripPHPComments(src string) string {
	var sb strings.Builder
	n := len(src)
	i := 0
	for i < n {
		// Multi-line comment /* ... */
		if i+1 < n && src[i] == '/' && src[i+1] == '*' {
			i += 2
			for i < n {
				if i+1 < n && src[i] == '*' && src[i+1] == '/' {
					i += 2
					break
				}
				if src[i] == '\n' {
					sb.WriteByte('\n')
				}
				i++
			}
			continue
		}

		// Single-line comment // or #
		if (i+1 < n && src[i] == '/' && src[i+1] == '/') || src[i] == '#' {
			for i < n && src[i] != '\n' {
				i++
			}
			continue
		}

		// String literals '...' or "..."
		if src[i] == '\'' || src[i] == '"' {
			quote := src[i]
			sb.WriteByte(quote)
			i++
			for i < n {
				c := src[i]
				sb.WriteByte(c)
				i++
				if c == '\\' && i < n {
					sb.WriteByte(src[i])
					i++
				} else if c == quote {
					break
				}
			}
			continue
		}

		sb.WriteByte(src[i])
		i++
	}
	return sb.String()
}

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

	// Sort controllers by name, then file for tie-breaker
	sort.Slice(inv.Controllers, func(i, j int) bool {
		if inv.Controllers[i].Name == inv.Controllers[j].Name {
			return inv.Controllers[i].File < inv.Controllers[j].File
		}
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

	// Sort models by name, then file for tie-breaker
	sort.Slice(inv.Models, func(i, j int) bool {
		if inv.Models[i].Name == inv.Models[j].Name {
			return inv.Models[i].File < inv.Models[j].File
		}
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
				// Filter CI3 index.html security stubs
				if ext == ".html" {
					content, err := os.ReadFile(path)
					if err == nil {
						str := string(content)
						if strings.Contains(str, "Directory access is forbidden") || strings.Contains(str, "403 Forbidden") {
							return nil
						}
					}
				}
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
	cleanContent := stripPHPComments(content)
	relPath, err := filepath.Rel(srcDir, fullPath)
	if err != nil {
		relPath = fullPath
	}
	relPath = filepath.ToSlash(relPath)

	name := strings.TrimSuffix(filepath.Base(fullPath), ".php")
	if match := ctrlClassRegex.FindStringSubmatch(cleanContent); len(match) > 1 {
		name = match[1]
	}

	var methods []string
	matches := methodRegex.FindAllStringSubmatch(cleanContent, -1)
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
	cleanContent := stripPHPComments(content)
	relPath, err := filepath.Rel(srcDir, fullPath)
	if err != nil {
		relPath = fullPath
	}
	relPath = filepath.ToSlash(relPath)

	name := strings.TrimSuffix(filepath.Base(fullPath), ".php")
	if match := modelClassRegex.FindStringSubmatch(cleanContent); len(match) > 1 {
		name = match[1]
	}

	table := ""
	if match := modelTableRegex.FindStringSubmatch(cleanContent); len(match) > 1 {
		table = match[1]
	}

	var methods []string
	matches := methodRegex.FindAllStringSubmatch(cleanContent, -1)
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

type routeMatch struct {
	pos   int
	route RouteInfo
}

func parseRoutes(content string) []RouteInfo {
	cleanContent := stripPHPComments(content)
	var matches []routeMatch

	// Match method-specific routes like $route['api/users']['post'] = 'users/create';
	methodIndices := methodRouteRegex.FindAllStringSubmatchIndex(cleanContent, -1)
	for _, idx := range methodIndices {
		pattern := cleanContent[idx[2]:idx[3]]
		httpMethod := strings.ToUpper(cleanContent[idx[4]:idx[5]])
		target := cleanContent[idx[6]:idx[7]]
		isSpecial, notes := isSpecialRoute(pattern)
		matches = append(matches, routeMatch{
			pos: idx[0],
			route: RouteInfo{
				Pattern:    pattern,
				Target:     target,
				HTTPMethod: httpMethod,
				IsSpecial:  isSpecial,
				Notes:      notes,
			},
		})
	}

	// Match standard routes like $route['users'] = 'users/index';
	simpleIndices := simpleRouteRegex.FindAllStringSubmatchIndex(cleanContent, -1)
	for _, idx := range simpleIndices {
		pattern := cleanContent[idx[2]:idx[3]]
		target := cleanContent[idx[4]:idx[5]]
		isSpecial, notes := isSpecialRoute(pattern)
		matches = append(matches, routeMatch{
			pos: idx[0],
			route: RouteInfo{
				Pattern:    pattern,
				Target:     target,
				HTTPMethod: "ANY",
				IsSpecial:  isSpecial,
				Notes:      notes,
			},
		})
	}

	// Sort routes in the order they appeared in the file
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].pos < matches[j].pos
	})

	routes := make([]RouteInfo, 0, len(matches))
	for _, m := range matches {
		routes = append(routes, m.route)
	}
	return routes
}

func isSpecialRoute(pattern string) (bool, string) {
	switch pattern {
	case "default_controller":
		return true, "CI3 Default Controller"
	case "404_override":
		return true, "CI3 Custom 404 Override"
	case "translate_uri_dashes":
		return true, "CI3 Dash Translation"
	default:
		return false, ""
	}
}

func parseAutoload(content string) AutoloadInfo {
	cleanContent := stripPHPComments(content)
	info := AutoloadInfo{
		Libraries: make([]string, 0),
		Helpers:   make([]string, 0),
		Models:    make([]string, 0),
	}

	matches := autoloadRegex.FindAllStringSubmatch(cleanContent, -1)
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
	for _, p := range parts {
		if m := itemRegex.FindStringSubmatch(p); len(m) > 1 {
			items = append(items, m[1])
		}
	}
	return items
}

func escapeMarkdownTable(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

// GenerateMarkdown formats the project inventory as a human- and agent-readable Markdown report.
func (inv *ProjectInventory) GenerateMarkdown() string {
	var sb strings.Builder

	sb.WriteString("# CodeIgniter 3 Migration Inventory Report\n\n")
	sb.WriteString(fmt.Sprintf("**Source Directory:** `%s`\n", escapeMarkdownTable(inv.SourceDir)))
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
					formatted = append(formatted, fmt.Sprintf("`%s()`", escapeMarkdownTable(m)))
				}
				methodList = strings.Join(formatted, ", ")
			} else {
				methodList = "*(none)*"
			}
			sb.WriteString(fmt.Sprintf("| **%s** | `%s` | %s |\n", escapeMarkdownTable(c.Name), escapeMarkdownTable(c.File), methodList))
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
				tableStr = fmt.Sprintf("`%s`", escapeMarkdownTable(m.Table))
			}
			var methodList string
			if len(m.Methods) > 0 {
				var formatted []string
				for _, fn := range m.Methods {
					formatted = append(formatted, fmt.Sprintf("`%s()`", escapeMarkdownTable(fn)))
				}
				methodList = strings.Join(formatted, ", ")
			} else {
				methodList = "*(none)*"
			}
			sb.WriteString(fmt.Sprintf("| **%s** | `%s` | %s | %s |\n", escapeMarkdownTable(m.Name), escapeMarkdownTable(m.File), tableStr, methodList))
		}
		sb.WriteString("\n")
	}

	// Routes Table
	sb.WriteString("## Routes\n\n")
	if len(inv.Routes) == 0 {
		sb.WriteString("_No custom routes found._\n\n")
	} else {
		sb.WriteString("| Route Pattern | HTTP Method | Target Controller/Action | Notes |\n")
		sb.WriteString("|---|---|---|---|\n")
		for _, r := range inv.Routes {
			method := r.HTTPMethod
			if method == "" {
				method = "ANY"
			}
			targetStr := "-"
			if r.Target != "" {
				targetStr = fmt.Sprintf("`%s`", escapeMarkdownTable(r.Target))
			} else if r.IsSpecial {
				targetStr = "*(none)*"
			}
			notes := "-"
			if r.Notes != "" {
				notes = escapeMarkdownTable(r.Notes)
			}
			sb.WriteString(fmt.Sprintf("| `%s` | **%s** | %s | %s |\n", escapeMarkdownTable(r.Pattern), escapeMarkdownTable(method), targetStr, notes))
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

// RunInspectCLI executes the inspect command line tool with given arguments.
func RunInspectCLI(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	src := fs.String("src", ".", "Path to CodeIgniter 3 project root")
	out := fs.String("out", "", "Output file path (default stdout)")
	format := fs.String("format", "markdown", "Output format: 'markdown' or 'json'")

	fs.Usage = func() {
		fmt.Println("Usage: go run ./skills/ci3-to-goigniter inspect [flags]")
		fmt.Println()
		fmt.Println("Scans a legacy CodeIgniter 3 project and produces an inventory report.")
		fmt.Println()
		fmt.Println("Flags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	inventory, err := InspectProject(*src)
	if err != nil {
		return fmt.Errorf("inspecting project: %w", err)
	}

	var output []byte
	switch strings.ToLower(*format) {
	case "json":
		data, err := inventory.GenerateJSON()
		if err != nil {
			return fmt.Errorf("generating JSON: %w", err)
		}
		output = data
	default:
		output = []byte(inventory.GenerateMarkdown())
	}

	if *out == "" || *out == "-" {
		fmt.Println(string(output))
	} else {
		if err := os.WriteFile(*out, output, 0644); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		fmt.Printf("Inventory successfully generated: %s\n", *out)
	}

	return nil
}
