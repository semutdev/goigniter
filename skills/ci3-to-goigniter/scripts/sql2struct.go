package scripts

import (
	"errors"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// ColumnSchema represents metadata for a table column.
type ColumnSchema struct {
	Name      string `json:"name"`
	GoName    string `json:"go_name"`
	SQLType   string `json:"sql_type"`
	GoType    string `json:"go_type"`
	Nullable  bool   `json:"nullable"`
	IsPrimary bool   `json:"is_primary"`
}

// TableSchema represents metadata for a database table.
type TableSchema struct {
	TableName   string         `json:"table_name"`
	StructName  string         `json:"struct_name"`
	Columns     []ColumnSchema `json:"columns"`
	PrimaryKeys []string       `json:"primary_keys"`
}

var initialisms = map[string]string{
	"id":    "ID",
	"ip":    "IP",
	"url":   "URL",
	"api":   "API",
	"http":  "HTTP",
	"https": "HTTPS",
	"html":  "HTML",
	"uuid":  "UUID",
	"sql":   "SQL",
	"sku":   "SKU",
	"json":  "JSON",
	"xml":   "XML",
	"db":    "DB",
	"ci":    "CI",
}

// ColumnNameToFieldName converts a SQL column name to an idiomatic Go struct field name.
func ColumnNameToFieldName(col string) string {
	cleaned := cleanIdentifier(col)
	if cleaned == "" {
		return ""
	}

	words := splitIntoWords(cleaned)
	var sb strings.Builder
	for _, word := range words {
		lower := strings.ToLower(word)
		if init, ok := initialisms[lower]; ok {
			sb.WriteString(init)
		} else {
			sb.WriteString(strings.ToUpper(word[:1]) + strings.ToLower(word[1:]))
		}
	}
	return sb.String()
}

// TableNameToStructName converts a SQL table name to a singularized CamelCase Go struct name.
func TableNameToStructName(table string) string {
	cleaned := cleanIdentifier(table)
	if cleaned == "" {
		return ""
	}

	words := splitIntoWords(cleaned)
	if len(words) == 0 {
		return ""
	}

	// Singularize the last word
	lastIdx := len(words) - 1
	words[lastIdx] = Singularize(words[lastIdx])

	var sb strings.Builder
	for _, word := range words {
		lower := strings.ToLower(word)
		if init, ok := initialisms[lower]; ok {
			sb.WriteString(init)
		} else {
			sb.WriteString(strings.ToUpper(word[:1]) + strings.ToLower(word[1:]))
		}
	}
	return sb.String()
}

// Singularize converts an English plural word to its singular form.
func Singularize(word string) string {
	lower := strings.ToLower(word)
	irregulars := map[string]string{
		"people":   "person",
		"children": "child",
		"men":      "man",
		"women":    "woman",
		"data":     "datum",
		"criteria": "criterion",
	}
	if sing, ok := irregulars[lower]; ok {
		return sing
	}
	if strings.HasSuffix(lower, "ies") && len(lower) > 3 {
		return word[:len(word)-3] + "y"
	}
	if strings.HasSuffix(lower, "sses") {
		return word[:len(word)-2]
	}
	if strings.HasSuffix(lower, "ses") {
		return word[:len(word)-2]
	}
	if strings.HasSuffix(lower, "xes") || strings.HasSuffix(lower, "ches") || strings.HasSuffix(lower, "shes") || strings.HasSuffix(lower, "zes") {
		return word[:len(word)-2]
	}
	if strings.HasSuffix(lower, "s") && !strings.HasSuffix(lower, "ss") {
		return word[:len(word)-1]
	}
	return word
}

// ToSnakeCase converts a CamelCase string to snake_case.
func ToSnakeCase(s string) string {
	var res strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			if unicode.IsLower(prev) || (i+1 < len(runes) && unicode.IsLower(runes[i+1])) {
				res.WriteRune('_')
			}
		}
		res.WriteRune(unicode.ToLower(r))
	}
	return res.String()
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func cleanIdentifier(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.LastIndex(s, "."); idx != -1 {
		s = s[idx+1:]
	}
	s = strings.Trim(s, "`'\"[]")
	return strings.TrimSpace(s)
}

func splitIntoWords(s string) []string {
	var words []string
	var current strings.Builder
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '_' || r == '-' || unicode.IsSpace(r) {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			continue
		}

		if unicode.IsUpper(r) && current.Len() > 0 {
			prev := runes[i-1]
			if unicode.IsLower(prev) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]) && !unicode.IsUpper(prev)) {
				words = append(words, current.String())
				current.Reset()
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

func sqlTypeToGoType(sqlType string, nullable bool) (goType string, baseType string) {
	upper := strings.ToUpper(strings.TrimSpace(sqlType))

	switch {
	case strings.HasPrefix(upper, "TINYINT(1)") || upper == "BOOLEAN" || upper == "BOOL":
		baseType = "bool"
	case strings.HasPrefix(upper, "BIGINT") || upper == "BIGSERIAL":
		baseType = "int64"
	case strings.HasPrefix(upper, "INT") || strings.HasPrefix(upper, "INTEGER") || strings.HasPrefix(upper, "SMALLINT") ||
		strings.HasPrefix(upper, "TINYINT") || strings.HasPrefix(upper, "MEDIUMINT") || upper == "SERIAL":
		baseType = "int"
	case strings.HasPrefix(upper, "FLOAT") || strings.HasPrefix(upper, "DOUBLE") || strings.HasPrefix(upper, "DECIMAL") ||
		strings.HasPrefix(upper, "NUMERIC") || strings.HasPrefix(upper, "REAL"):
		baseType = "float64"
	case strings.HasPrefix(upper, "DATETIME") || strings.HasPrefix(upper, "TIMESTAMP") ||
		strings.HasPrefix(upper, "DATE") || strings.HasPrefix(upper, "TIME") || upper == "YEAR":
		baseType = "time.Time"
	case strings.HasPrefix(upper, "BLOB") || strings.HasPrefix(upper, "LONGBLOB") || strings.HasPrefix(upper, "MEDIUMBLOB") ||
		strings.HasPrefix(upper, "TINYBLOB") || strings.HasPrefix(upper, "BYTEA") || strings.HasPrefix(upper, "BINARY") ||
		strings.HasPrefix(upper, "VARBINARY"):
		baseType = "[]byte"
	default:
		baseType = "string"
	}

	if nullable {
		if baseType == "[]byte" {
			goType = "[]byte"
		} else {
			goType = "*" + baseType
		}
	} else {
		goType = baseType
	}

	return goType, baseType
}

func stripSQLComments(sql string) string {
	var sb strings.Builder
	runes := []rune(sql)
	n := len(runes)
	i := 0
	inSingleQuote := false
	inDoubleQuote := false
	inBacktick := false

	for i < n {
		r := runes[i]
		if r == '\'' && !inDoubleQuote && !inBacktick {
			inSingleQuote = !inSingleQuote
			sb.WriteRune(r)
			i++
			continue
		}
		if r == '"' && !inSingleQuote && !inBacktick {
			inDoubleQuote = !inDoubleQuote
			sb.WriteRune(r)
			i++
			continue
		}
		if r == '`' && !inSingleQuote && !inDoubleQuote {
			inBacktick = !inBacktick
			sb.WriteRune(r)
			i++
			continue
		}

		if !inSingleQuote && !inDoubleQuote && !inBacktick {
			if r == '-' && i+1 < n && runes[i+1] == '-' {
				i += 2
				for i < n && runes[i] != '\n' {
					i++
				}
				continue
			}
			if r == '#' {
				i++
				for i < n && runes[i] != '\n' {
					i++
				}
				continue
			}
			if r == '/' && i+1 < n && runes[i+1] == '*' {
				i += 2
				for i+1 < n && !(runes[i] == '*' && runes[i+1] == '/') {
					i++
				}
				if i+1 < n {
					i += 2
				} else {
					i = n
				}
				continue
			}
		}

		sb.WriteRune(r)
		i++
	}
	return sb.String()
}

func splitByTopLevelCommas(s string) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	inSingleQuote := false
	inDoubleQuote := false
	inBacktick := false

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\'' && !inDoubleQuote && !inBacktick {
			inSingleQuote = !inSingleQuote
			current.WriteRune(r)
			continue
		}
		if r == '"' && !inSingleQuote && !inBacktick {
			inDoubleQuote = !inDoubleQuote
			current.WriteRune(r)
			continue
		}
		if r == '`' && !inSingleQuote && !inDoubleQuote {
			inBacktick = !inBacktick
			current.WriteRune(r)
			continue
		}

		if !inSingleQuote && !inDoubleQuote && !inBacktick {
			if r == '(' {
				depth++
			} else if r == ')' {
				depth--
			} else if r == ',' && depth == 0 {
				trimmed := strings.TrimSpace(current.String())
				if trimmed != "" {
					parts = append(parts, trimmed)
				}
				current.Reset()
				continue
			}
		}
		current.WriteRune(r)
	}
	trimmed := strings.TrimSpace(current.String())
	if trimmed != "" {
		parts = append(parts, trimmed)
	}
	return parts
}

func extractSQLType(s string) (sqlType string, remainder string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ""
	}
	depth := 0
	hasParen := false
	typeEnd := -1
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '(' {
			depth++
			hasParen = true
		} else if r == ')' {
			depth--
			if depth == 0 {
				typeEnd = i + 1
				break
			}
		} else if unicode.IsSpace(r) && !hasParen {
			typeEnd = i
			break
		}
	}

	if typeEnd == -1 {
		return s, ""
	}
	sqlType = strings.TrimSpace(string(runes[:typeEnd]))
	remainder = strings.TrimSpace(string(runes[typeEnd:]))
	return sqlType, remainder
}

var (
	createTableRegex = regexp.MustCompile(`(?i)\bCREATE\s+(?:TEMPORARY\s+)?TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([^\s(]+)`)
	primaryKeyRegex  = regexp.MustCompile(`(?i)^\s*(?:CONSTRAINT\s+\S+\s+)?PRIMARY\s+KEY\s*\((.*?)\)`)
	constraintRegex  = regexp.MustCompile(`(?i)^\s*(?:CONSTRAINT\s+\S+\s+)?(?:FOREIGN\s+KEY|UNIQUE\s+KEY|UNIQUE|KEY|INDEX|CHECK)\b`)
	nullKeywordRegex = regexp.MustCompile(`\bNULL\b`)
)

// ParseSQLSchema parses SQL DDL containing one or more CREATE TABLE statements into TableSchemas.
func ParseSQLSchema(sqlContent string) ([]TableSchema, error) {
	cleanSQL := stripSQLComments(sqlContent)

	matches := createTableRegex.FindAllStringSubmatchIndex(cleanSQL, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	var tables []TableSchema

	for _, loc := range matches {
		rawTableName := cleanSQL[loc[2]:loc[3]]
		tableName := cleanIdentifier(rawTableName)
		if tableName == "" {
			continue
		}

		startParen := strings.Index(cleanSQL[loc[1]:], "(")
		if startParen == -1 {
			continue
		}
		bodyStart := loc[1] + startParen + 1

		bodyRunes := []rune(cleanSQL[bodyStart:])
		depth := 1
		inSingleQuote := false
		inDoubleQuote := false
		inBacktick := false
		bodyEnd := -1

		for idx := 0; idx < len(bodyRunes); idx++ {
			r := bodyRunes[idx]
			if r == '\'' && !inDoubleQuote && !inBacktick {
				inSingleQuote = !inSingleQuote
				continue
			}
			if r == '"' && !inSingleQuote && !inBacktick {
				inDoubleQuote = !inDoubleQuote
				continue
			}
			if r == '`' && !inSingleQuote && !inDoubleQuote {
				inBacktick = !inBacktick
				continue
			}

			if !inSingleQuote && !inDoubleQuote && !inBacktick {
				if r == '(' {
					depth++
				} else if r == ')' {
					depth--
					if depth == 0 {
						bodyEnd = idx
						break
					}
				}
			}
		}

		if bodyEnd == -1 {
			continue
		}

		tableBody := string(bodyRunes[:bodyEnd])
		lines := splitByTopLevelCommas(tableBody)

		table := TableSchema{
			TableName:  tableName,
			StructName: TableNameToStructName(tableName),
		}

		for _, rawLine := range lines {
			line := strings.TrimSpace(rawLine)
			if line == "" {
				continue
			}

			// Check table-level primary key
			if pkMatch := primaryKeyRegex.FindStringSubmatch(line); len(pkMatch) > 1 {
				pkCols := strings.Split(pkMatch[1], ",")
				for _, pk := range pkCols {
					cleanPK := cleanIdentifier(pk)
					if cleanPK != "" {
						table.PrimaryKeys = append(table.PrimaryKeys, cleanPK)
					}
				}
				continue
			}

			// Check other constraints or indexes to skip
			if constraintRegex.MatchString(line) {
				continue
			}

			// Column definition
			var colName string
			var rest string

			if line[0] == '`' || line[0] == '"' || line[0] == '[' {
				closeChar := line[0]
				if closeChar == '[' {
					closeChar = ']'
				}
				endIdx := strings.IndexRune(line[1:], rune(closeChar))
				if endIdx == -1 {
					parts := strings.Fields(line)
					colName = cleanIdentifier(parts[0])
					rest = strings.TrimSpace(line[len(parts[0]):])
				} else {
					colName = line[1 : endIdx+1]
					rest = strings.TrimSpace(line[endIdx+2:])
				}
			} else {
				parts := strings.Fields(line)
				if len(parts) < 2 {
					continue
				}
				colName = cleanIdentifier(parts[0])
				rest = strings.TrimSpace(line[len(parts[0]):])
			}

			sqlType, remainder := extractSQLType(rest)
			if sqlType == "" {
				continue
			}

			upperRemainder := strings.ToUpper(remainder)
			isPrimary := strings.Contains(upperRemainder, "PRIMARY KEY")
			nullable := false

			if strings.Contains(upperRemainder, "NOT NULL") {
				nullable = false
			} else if nullKeywordRegex.MatchString(upperRemainder) {
				nullable = true
			} else if isPrimary {
				nullable = false
			} else if strings.Contains(upperRemainder, "DEFAULT") {
				nullable = false
			} else {
				nullable = true
			}

			goType, _ := sqlTypeToGoType(sqlType, nullable)

			col := ColumnSchema{
				Name:      colName,
				GoName:    ColumnNameToFieldName(colName),
				SQLType:   sqlType,
				GoType:    goType,
				Nullable:  nullable,
				IsPrimary: isPrimary,
			}

			if isPrimary {
				found := false
				for _, pk := range table.PrimaryKeys {
					if strings.EqualFold(pk, colName) {
						found = true
						break
					}
				}
				if !found {
					table.PrimaryKeys = append(table.PrimaryKeys, colName)
				}
			}

			table.Columns = append(table.Columns, col)
		}

		// Sync table-level primary keys back to columns
		for i := range table.Columns {
			for _, pk := range table.PrimaryKeys {
				if strings.EqualFold(table.Columns[i].Name, pk) {
					table.Columns[i].IsPrimary = true
					table.Columns[i].Nullable = false
					table.Columns[i].GoType = strings.TrimPrefix(table.Columns[i].GoType, "*")
				}
			}
		}

		// Ensure all IsPrimary columns are recorded in PrimaryKeys
		for _, col := range table.Columns {
			if col.IsPrimary {
				found := false
				for _, pk := range table.PrimaryKeys {
					if strings.EqualFold(pk, col.Name) {
						found = true
						break
					}
				}
				if !found {
					table.PrimaryKeys = append(table.PrimaryKeys, col.Name)
				}
			}
		}

		tables = append(tables, table)
	}

	return tables, nil
}

// GenerateModelGoCode generates idiomatic Go model code with CRUD boilerplate for GoIgniter.
func GenerateModelGoCode(schema TableSchema, pkgName string) (string, error) {
	if pkgName == "" {
		pkgName = "models"
	}

	var sb strings.Builder

	// Package
	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Imports
	needTime := false
	for _, col := range schema.Columns {
		if strings.Contains(col.GoType, "time.Time") {
			needTime = true
			break
		}
	}

	sb.WriteString("import (\n")
	if needTime {
		sb.WriteString("\t\"time\"\n\n")
	}
	sb.WriteString("\t\"github.com/semutdev/goigniter/system/libraries/database\"\n")
	sb.WriteString(")\n\n")

	// Struct
	sb.WriteString(fmt.Sprintf("// %s represents the %s table.\n", schema.StructName, schema.TableName))
	sb.WriteString(fmt.Sprintf("type %s struct {\n", schema.StructName))
	for _, col := range schema.Columns {
		sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\" db:\"%s\"`\n", col.GoName, col.GoType, col.Name, col.Name))
	}
	sb.WriteString("}\n\n")

	// TableName
	sb.WriteString(fmt.Sprintf("// TableName returns the database table name for %s.\n", schema.StructName))
	sb.WriteString(fmt.Sprintf("func (%s) TableName() string {\n", schema.StructName))
	sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", schema.TableName))
	sb.WriteString("}\n\n")

	// Model receiver struct
	modelName := schema.StructName + "Model"
	sb.WriteString(fmt.Sprintf("// %s provides data access methods for %s.\n", modelName, schema.TableName))
	sb.WriteString(fmt.Sprintf("type %s struct{}\n\n", modelName))

	// NewModel constructor
	sb.WriteString(fmt.Sprintf("// New%s creates a new %s instance.\n", modelName, modelName))
	sb.WriteString(fmt.Sprintf("func New%s() *%s {\n", modelName, modelName))
	sb.WriteString(fmt.Sprintf("\treturn &%s{}\n", modelName))
	sb.WriteString("}\n\n")

	// Primary key info
	pkCol := "id"
	pkGoName := "ID"
	pkType := "int"
	if len(schema.PrimaryKeys) > 0 {
		pkCol = schema.PrimaryKeys[0]
		for _, col := range schema.Columns {
			if strings.EqualFold(col.Name, pkCol) {
				pkGoName = col.GoName
				pkType = strings.TrimPrefix(col.GoType, "*")
				break
			}
		}
	}
	pkParam := "id"
	if !strings.EqualFold(pkCol, "id") {
		pkParam = lowerFirst(pkGoName)
	}

	entityVar := lowerFirst(schema.StructName)

	// Find(id <pkType>) (*<StructName>, error)
	sb.WriteString(fmt.Sprintf("// Find retrieves a single %s by primary key.\n", schema.StructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) Find(%s %s) (*%s, error) {\n", modelName, pkParam, pkType, schema.StructName))
	sb.WriteString(fmt.Sprintf("\tvar item %s\n", schema.StructName))
	sb.WriteString(fmt.Sprintf("\terr := database.Table(\"%s\").Where(\"%s\", %s).First(&item)\n", schema.TableName, pkCol, pkParam))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\treturn nil, err\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn &item, nil\n")
	sb.WriteString("}\n\n")

	// FindAll() ([]<StructName>, error)
	sb.WriteString(fmt.Sprintf("// FindAll retrieves all %s records.\n", schema.StructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) FindAll() ([]%s, error) {\n", modelName, schema.StructName))
	sb.WriteString(fmt.Sprintf("\tvar items []%s\n", schema.StructName))
	sb.WriteString(fmt.Sprintf("\terr := database.Table(\"%s\").Get(&items)\n", schema.TableName))
	sb.WriteString("\treturn items, err\n")
	sb.WriteString("}\n\n")

	// Insert(<entityVar> *<StructName>) error
	var insertCols []ColumnSchema
	for _, col := range schema.Columns {
		if len(schema.PrimaryKeys) == 1 && col.IsPrimary && len(schema.Columns) > 1 {
			continue
		}
		insertCols = append(insertCols, col)
	}

	sb.WriteString(fmt.Sprintf("// Insert inserts a new %s record into the database.\n", schema.StructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) Insert(%s *%s) error {\n", modelName, entityVar, schema.StructName))
	sb.WriteString(fmt.Sprintf("\treturn database.Table(\"%s\").Insert(map[string]any{\n", schema.TableName))
	for _, col := range insertCols {
		sb.WriteString(fmt.Sprintf("\t\t\"%s\": %s.%s,\n", col.Name, entityVar, col.GoName))
	}
	sb.WriteString("\t})\n")
	sb.WriteString("}\n\n")

	// Update(<entityVar> *<StructName>) error
	var updateCols []ColumnSchema
	for _, col := range schema.Columns {
		if col.IsPrimary && len(schema.Columns) > len(schema.PrimaryKeys) {
			continue
		}
		updateCols = append(updateCols, col)
	}

	sb.WriteString(fmt.Sprintf("// Update updates an existing %s record in the database.\n", schema.StructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) Update(%s *%s) error {\n", modelName, entityVar, schema.StructName))
	whereChain := fmt.Sprintf(".Where(\"%s\", %s.%s)", pkCol, entityVar, pkGoName)
	if len(schema.PrimaryKeys) > 1 {
		var whereParts []string
		for _, pk := range schema.PrimaryKeys {
			goName := ColumnNameToFieldName(pk)
			whereParts = append(whereParts, fmt.Sprintf(".Where(\"%s\", %s.%s)", pk, entityVar, goName))
		}
		whereChain = strings.Join(whereParts, "")
	}

	sb.WriteString(fmt.Sprintf("\treturn database.Table(\"%s\")%s.Update(map[string]any{\n", schema.TableName, whereChain))
	for _, col := range updateCols {
		sb.WriteString(fmt.Sprintf("\t\t\"%s\": %s.%s,\n", col.Name, entityVar, col.GoName))
	}
	sb.WriteString("\t})\n")
	sb.WriteString("}\n\n")

	// Delete(<pkParam> <pkType>) error
	sb.WriteString(fmt.Sprintf("// Delete deletes a %s record by primary key.\n", schema.StructName))
	sb.WriteString(fmt.Sprintf("func (m *%s) Delete(%s %s) error {\n", modelName, pkParam, pkType))
	sb.WriteString(fmt.Sprintf("\treturn database.Table(\"%s\").Where(\"%s\", %s).Delete()\n", schema.TableName, pkCol, pkParam))
	sb.WriteString("}\n")

	formatted, err := format.Source([]byte(sb.String()))
	if err != nil {
		return sb.String(), fmt.Errorf("formatting generated Go code: %w", err)
	}

	return string(formatted), nil
}

// RunSQL2StructCLI executes the sql2struct command line tool with given arguments.
func RunSQL2StructCLI(args []string) error {
	fs := flag.NewFlagSet("sql2struct", flag.ContinueOnError)
	sqlPath := fs.String("sql", "", "Path to SQL file")
	outDir := fs.String("out", "", "Output directory or file (default stdout)")
	pkgName := fs.String("pkg", "models", "Go package name (default 'models')")

	fs.Usage = func() {
		fmt.Println("Usage: go run ./skills/ci3-to-goigniter sql2struct [flags]")
		fmt.Println()
		fmt.Println("Converts SQL schema / table definitions to Go model structs and CRUD boilerplate.")
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

	if *sqlPath == "" {
		return errors.New("missing required flag: -sql")
	}

	content, err := os.ReadFile(*sqlPath)
	if err != nil {
		return fmt.Errorf("reading SQL file %s: %w", *sqlPath, err)
	}

	tables, err := ParseSQLSchema(string(content))
	if err != nil {
		return fmt.Errorf("parsing SQL schema: %w", err)
	}

	if len(tables) == 0 {
		return fmt.Errorf("no CREATE TABLE statements found in %s", *sqlPath)
	}

	// If -out is empty or "-", print to stdout
	if *outDir == "" || *outDir == "-" {
		for i, table := range tables {
			code, err := GenerateModelGoCode(table, *pkgName)
			if err != nil {
				return fmt.Errorf("generating code for %s: %w", table.TableName, err)
			}
			if i > 0 {
				fmt.Println("\n// " + strings.Repeat("-", 60) + "\n")
			}
			fmt.Print(code)
		}
		return nil
	}

	// If -out ends with .go, write single/combined file
	if strings.HasSuffix(strings.ToLower(*outDir), ".go") {
		dir := filepath.Dir(*outDir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
		var fullCode strings.Builder
		for i, table := range tables {
			code, err := GenerateModelGoCode(table, *pkgName)
			if err != nil {
				return fmt.Errorf("generating code for %s: %w", table.TableName, err)
			}
			if i > 0 {
				fullCode.WriteString("\n\n")
			}
			fullCode.WriteString(code)
		}
		if err := os.WriteFile(*outDir, []byte(fullCode.String()), 0644); err != nil {
			return fmt.Errorf("writing output file %s: %w", *outDir, err)
		}
		fmt.Printf("Model generated: %s\n", *outDir)
		return nil
	}

	// Otherwise, -out is a directory: generate one file per table
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", *outDir, err)
	}

	for _, table := range tables {
		code, err := GenerateModelGoCode(table, *pkgName)
		if err != nil {
			return fmt.Errorf("generating code for %s: %w", table.TableName, err)
		}
		fileName := ToSnakeCase(table.StructName) + ".go"
		targetPath := filepath.Join(*outDir, fileName)
		if err := os.WriteFile(targetPath, []byte(code), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", targetPath, err)
		}
		fmt.Printf("Model generated: %s\n", targetPath)
	}

	return nil
}
