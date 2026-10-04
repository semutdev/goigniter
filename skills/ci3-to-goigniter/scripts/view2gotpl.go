package scripts

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	reVarExpr  = regexp.MustCompile(`^\$[a-zA-Z_][a-zA-Z0-9_]*(?:->[a-zA-Z_][a-zA-Z0-9_]*|\[['"][a-zA-Z0-9_]+['"]\])*$`)
	reIdent    = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`)
	reForeach  = regexp.MustCompile(`^foreach\s*\(\s*(\$?[a-zA-Z0-9_>\[\]'"]+)\s+as\s+(?:(\$[a-zA-Z0-9_]+)\s*=>\s*)?(\$[a-zA-Z0-9_]+)\s*\)\s*[:{]?$`)
	reIf       = regexp.MustCompile(`^if\s*\((.*)\)\s*[:{]$`)
	reElseIf   = regexp.MustCompile(`^(?:\}\s*)?else\s*if\s*\((.*)\)\s*[:{]?$|^(?:\}\s*)?elseif\s*\((.*)\)\s*[:{]?$`)
	reViewLoad = regexp.MustCompile(`^\$this->load->view\(\s*['"]([^'"]+)['"]`)
	reBaseURL  = regexp.MustCompile(`^base_url\((.*)\)$`)
	reSiteURL  = regexp.MustCompile(`^site_url\((.*)\)$`)
	reEscaping = regexp.MustCompile(`^(?:htmlspecialchars|htmlentities)\((.*)\)$`)
)

type blockType int

const (
	blockIf blockType = iota
	blockForeach
)

type blockState struct {
	kind    blockType
	itemVar string
	keyVar  string
}

type loopScope struct {
	itemVar string
	keyVar  string
}

type viewTranspiler struct {
	loopStack  []loopScope
	blockStack []blockState
}

func (vt *viewTranspiler) pushLoop(itemVar, keyVar string) {
	vt.loopStack = append(vt.loopStack, loopScope{itemVar: itemVar, keyVar: keyVar})
	vt.blockStack = append(vt.blockStack, blockState{kind: blockForeach, itemVar: itemVar, keyVar: keyVar})
}

func (vt *viewTranspiler) pushIf() {
	vt.blockStack = append(vt.blockStack, blockState{kind: blockIf})
}

func (vt *viewTranspiler) popLoop() {
	if len(vt.loopStack) > 0 {
		vt.loopStack = vt.loopStack[:len(vt.loopStack)-1]
	}
	if len(vt.blockStack) > 0 && vt.blockStack[len(vt.blockStack)-1].kind == blockForeach {
		vt.blockStack = vt.blockStack[:len(vt.blockStack)-1]
	}
}

func (vt *viewTranspiler) popIf() {
	if len(vt.blockStack) > 0 && vt.blockStack[len(vt.blockStack)-1].kind == blockIf {
		vt.blockStack = vt.blockStack[:len(vt.blockStack)-1]
	}
}

func (vt *viewTranspiler) popAnyBlock() {
	if len(vt.blockStack) > 0 {
		last := vt.blockStack[len(vt.blockStack)-1]
		vt.blockStack = vt.blockStack[:len(vt.blockStack)-1]
		if last.kind == blockForeach && len(vt.loopStack) > 0 {
			vt.loopStack = vt.loopStack[:len(vt.loopStack)-1]
		}
	} else if len(vt.loopStack) > 0 {
		vt.loopStack = vt.loopStack[:len(vt.loopStack)-1]
	}
}

func (vt *viewTranspiler) currentLoopVar() string {
	if len(vt.loopStack) == 0 {
		return ""
	}
	return vt.loopStack[len(vt.loopStack)-1].itemVar
}

// TranspilePHPViewToTemplate converts a CI3 PHP view string into a Go html/template string.
func TranspilePHPViewToTemplate(phpContent string) string {
	vt := &viewTranspiler{}
	var sb strings.Builder
	pos := 0

	for pos < len(phpContent) {
		openStart, codeStart, codeEnd, nextIdx, isShortEcho, found := findNextPHPTag(phpContent, pos)
		if !found {
			sb.WriteString(phpContent[pos:])
			break
		}

		if openStart > pos {
			sb.WriteString(phpContent[pos:openStart])
		}

		rawCode := phpContent[codeStart:codeEnd]
		transpiled := vt.transpilePHPBlock(rawCode, isShortEcho)

		// If transpiled is empty (e.g. boilerplate like defined('BASEPATH')),
		// consume immediately following newline to avoid leaving empty lines.
		if transpiled == "" {
			if nextIdx < len(phpContent) {
				if phpContent[nextIdx] == '\r' && nextIdx+1 < len(phpContent) && phpContent[nextIdx+1] == '\n' {
					nextIdx += 2
				} else if phpContent[nextIdx] == '\n' {
					nextIdx++
				}
			}
		} else {
			sb.WriteString(transpiled)
		}

		pos = nextIdx
	}

	return sb.String()
}

func findNextPHPTag(s string, startIdx int) (openStart, codeStart, codeEnd, nextIdx int, isShortEcho bool, found bool) {
	for i := startIdx; i < len(s); {
		idx := strings.Index(s[i:], "<?")
		if idx == -1 {
			return 0, 0, 0, 0, false, false
		}
		tagStart := i + idx

		if strings.HasPrefix(s[tagStart:], "<?=") {
			openStart = tagStart
			codeStart = tagStart + 3
			isShortEcho = true
		} else if len(s[tagStart:]) >= 5 && strings.EqualFold(s[tagStart:tagStart+5], "<?php") {
			if len(s[tagStart:]) == 5 || unicode.IsSpace(rune(s[tagStart+5])) || strings.HasPrefix(s[tagStart+5:], "?>") {
				openStart = tagStart
				codeStart = tagStart + 5
				isShortEcho = false
			} else {
				i = tagStart + 2
				continue
			}
		} else if strings.HasPrefix(s[tagStart:], "<?xml") {
			i = tagStart + 5
			continue
		} else if len(s[tagStart:]) >= 2 {
			if len(s[tagStart:]) > 2 && (unicode.IsSpace(rune(s[tagStart+2])) || s[tagStart+2] == '$') {
				openStart = tagStart
				codeStart = tagStart + 2
				isShortEcho = false
			} else {
				i = tagStart + 2
				continue
			}
		} else {
			i = tagStart + 2
			continue
		}

		inQuote := byte(0)
		escaped := false
		foundClose := false

		for j := codeStart; j < len(s); j++ {
			c := s[j]
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' && inQuote != 0 {
				escaped = true
				continue
			}
			if inQuote == 0 {
				if c == '\'' || c == '"' {
					inQuote = c
				} else if c == '?' && j+1 < len(s) && s[j+1] == '>' {
					codeEnd = j
					nextIdx = j + 2
					foundClose = true
					return openStart, codeStart, codeEnd, nextIdx, isShortEcho, true
				}
			} else {
				if c == inQuote {
					inQuote = 0
				}
			}
		}

		if !foundClose {
			codeEnd = len(s)
			nextIdx = len(s)
			return openStart, codeStart, codeEnd, nextIdx, isShortEcho, true
		}
	}
	return 0, 0, 0, 0, false, false
}

func (vt *viewTranspiler) transpilePHPBlock(rawCode string, isShortEcho bool) string {
	trimmed := strings.TrimSpace(rawCode)
	if trimmed == "" {
		return ""
	}

	// 1. CI3 BASEPATH check
	if strings.Contains(trimmed, "defined('BASEPATH')") || strings.Contains(trimmed, "defined(\"BASEPATH\")") {
		return ""
	}

	// 2. View load: $this->load->view('partials/header'[, $data])
	if viewMatch := reViewLoad.FindStringSubmatch(trimmed); viewMatch != nil {
		viewName := viewMatch[1]
		if !strings.HasSuffix(viewName, ".html") {
			viewName += ".html"
		}
		return fmt.Sprintf(`{{ template "%s" . }}`, viewName)
	}

	// 3. Form open / Form close
	if strings.Contains(trimmed, "form_close") {
		return "</form>"
	}
	if strings.Contains(trimmed, "form_open") {
		isMultipart := strings.Contains(trimmed, "form_open_multipart")
		startIdx := strings.Index(trimmed, "form_open")
		openParen := strings.Index(trimmed[startIdx:], "(")
		actionURL := "/"
		if openParen != -1 {
			argStr := extractParenContent(trimmed[startIdx+openParen:])
			firstArg := extractFirstArg(argStr)
			if firstArg != "" {
				actionURL = vt.transpileURLArg(firstArg)
			}
		}
		if isMultipart {
			return fmt.Sprintf(`<form action="%s" method="POST" enctype="multipart/form-data">`, actionURL)
		}
		return fmt.Sprintf(`<form action="%s" method="POST">`, actionURL)
	}

	// 4. CSRF helpers
	if strings.Contains(trimmed, "get_csrf_token_name()") || strings.Contains(trimmed, "get_csrf_hash()") {
		return "{{ .csrf_token }}"
	}

	// 5. Loops: foreach
	if strings.HasPrefix(trimmed, "foreach") {
		if m := reForeach.FindStringSubmatch(trimmed); m != nil {
			targetArray := strings.TrimSpace(m[1])
			keyVar := strings.TrimPrefix(strings.TrimSpace(m[2]), "$")
			itemVar := strings.TrimPrefix(strings.TrimSpace(m[3]), "$")

			vt.pushLoop(itemVar, keyVar)
			targetTpl := vt.transpileVariableRef(targetArray)
			return fmt.Sprintf("{{ range %s }}", targetTpl)
		}
	}
	if strings.HasPrefix(trimmed, "endforeach") {
		vt.popLoop()
		return "{{ end }}"
	}

	// 6. Conditionals
	if strings.HasPrefix(trimmed, "if") && (strings.HasSuffix(trimmed, ":") || strings.HasSuffix(trimmed, "{")) {
		if m := reIf.FindStringSubmatch(trimmed); m != nil {
			vt.pushIf()
			cond := strings.TrimSpace(m[1])
			tplCond := vt.transpileCondition(cond)
			return fmt.Sprintf("{{ if %s }}", tplCond)
		}
	}
	if strings.HasPrefix(trimmed, "elseif") || strings.HasPrefix(trimmed, "else if") ||
		strings.HasPrefix(trimmed, "} elseif") || strings.HasPrefix(trimmed, "} else if") {
		if m := reElseIf.FindStringSubmatch(trimmed); m != nil {
			cond := m[1]
			if cond == "" {
				cond = m[2]
			}
			cond = strings.TrimSpace(cond)
			tplCond := vt.transpileCondition(cond)
			return fmt.Sprintf("{{ else if %s }}", tplCond)
		}
	}
	if trimmed == "else:" || trimmed == "else {" || trimmed == "} else {" || trimmed == "} else:" {
		return "{{ else }}"
	}
	if strings.HasPrefix(trimmed, "endif") {
		vt.popIf()
		return "{{ end }}"
	}
	if trimmed == "}" || strings.HasPrefix(trimmed, "} //") || strings.HasPrefix(trimmed, "}/*") {
		vt.popAnyBlock()
		return "{{ end }}"
	}

	// 7. Echo / Print / Short Echo expressions
	var expr string
	isEcho := false

	if isShortEcho {
		expr = strings.TrimSuffix(trimmed, ";")
		expr = strings.TrimSpace(expr)
		isEcho = true
	} else if strings.HasPrefix(trimmed, "echo ") || strings.HasPrefix(trimmed, "echo(") {
		expr = strings.TrimPrefix(trimmed, "echo")
		expr = strings.TrimSpace(expr)
		expr = strings.TrimSuffix(expr, ";")
		expr = strings.TrimSpace(expr)
		if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
			expr = strings.TrimSpace(expr[1 : len(expr)-1])
		}
		isEcho = true
	} else if strings.HasPrefix(trimmed, "print ") || strings.HasPrefix(trimmed, "print(") {
		expr = strings.TrimPrefix(trimmed, "print")
		expr = strings.TrimSpace(expr)
		expr = strings.TrimSuffix(expr, ";")
		expr = strings.TrimSpace(expr)
		if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
			expr = strings.TrimSpace(expr[1 : len(expr)-1])
		}
		isEcho = true
	}

	if isEcho {
		res, ok := vt.transpileEchoExpr(expr)
		if ok {
			return res
		}
	}

	// 8. Unconvertible / arbitrary PHP blocks
	safeCode := strings.ReplaceAll(trimmed, "*/", "* /")
	return fmt.Sprintf("{{/* TODO_MIGRATE: %s */}}", safeCode)
}

func (vt *viewTranspiler) transpileEchoExpr(expr string) (string, bool) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "", true
	}

	// Base URL
	if m := reBaseURL.FindStringSubmatch(expr); m != nil {
		return vt.transpileURLArg(m[1]), true
	}

	// Site URL
	if m := reSiteURL.FindStringSubmatch(expr); m != nil {
		return vt.transpileURLArg(m[1]), true
	}

	// Escaping functions: htmlspecialchars or htmlentities
	if m := reEscaping.FindStringSubmatch(expr); m != nil {
		firstArg := extractFirstArg(m[1])
		return vt.transpileEchoExpr(firstArg)
	}

	// Variable reference
	if reVarExpr.MatchString(expr) {
		ref := vt.transpileVariableRef(expr)
		return fmt.Sprintf("{{ %s }}", ref), true
	}

	return "", false
}

func (vt *viewTranspiler) transpileVariableRef(rawVar string) string {
	rawVar = strings.TrimSpace(rawVar)
	if rawVar == "" {
		return ""
	}

	parts := reIdent.FindAllString(rawVar, -1)
	if len(parts) == 0 {
		return ""
	}

	currentLoopVar := vt.currentLoopVar()
	if currentLoopVar != "" && strings.EqualFold(parts[0], currentLoopVar) {
		parts = parts[1:]
		if len(parts) == 0 {
			return "."
		}
	}

	var converted []string
	for _, p := range parts {
		converted = append(converted, ColumnNameToFieldName(p))
	}
	return "." + strings.Join(converted, ".")
}

func (vt *viewTranspiler) transpileCondition(cond string) string {
	cond = strings.TrimSpace(cond)
	for strings.HasPrefix(cond, "(") && strings.HasSuffix(cond, ")") && isValidParenthesized(cond) {
		cond = strings.TrimSpace(cond[1 : len(cond)-1])
	}

	// Negation check
	if strings.HasPrefix(cond, "!") {
		inner := strings.TrimSpace(cond[1:])
		if strings.HasPrefix(inner, "empty(") && strings.HasSuffix(inner, ")") {
			arg := strings.TrimSpace(inner[6 : len(inner)-1])
			return vt.transpileVariableRef(arg)
		}
		if strings.HasPrefix(inner, "isset(") && strings.HasSuffix(inner, ")") {
			arg := strings.TrimSpace(inner[6 : len(inner)-1])
			return "not " + vt.transpileVariableRef(arg)
		}
		return "not " + vt.transpileCondition(inner)
	}

	// empty(...) check
	if strings.HasPrefix(cond, "empty(") && strings.HasSuffix(cond, ")") {
		arg := strings.TrimSpace(cond[6 : len(cond)-1])
		return "not " + vt.transpileVariableRef(arg)
	}

	// isset(...) check
	if strings.HasPrefix(cond, "isset(") && strings.HasSuffix(cond, ")") {
		arg := strings.TrimSpace(cond[6 : len(cond)-1])
		return vt.transpileVariableRef(arg)
	}

	// Plain variable
	if reVarExpr.MatchString(cond) {
		return vt.transpileVariableRef(cond)
	}

	// Binary comparison operators
	ops := []struct {
		phpOp string
		goOp  string
	}{
		{"===", "eq"},
		{"!==", "ne"},
		{"==", "eq"},
		{"!=", "ne"},
		{">=", "ge"},
		{"<=", "le"},
		{">", "gt"},
		{"<", "lt"},
	}

	for _, entry := range ops {
		idx := findOpOutsideQuotes(cond, entry.phpOp)
		if idx != -1 {
			left := strings.TrimSpace(cond[:idx])
			right := strings.TrimSpace(cond[idx+len(entry.phpOp):])

			var leftTpl string
			if strings.HasPrefix(left, "$") {
				leftTpl = vt.transpileVariableRef(left)
			} else {
				leftTpl = left
			}

			var rightTpl string
			if strings.HasPrefix(right, "$") {
				rightTpl = vt.transpileVariableRef(right)
			} else if (strings.HasPrefix(right, "'") && strings.HasSuffix(right, "'")) ||
				(strings.HasPrefix(right, "\"") && strings.HasSuffix(right, "\"")) {
				lit := right[1 : len(right)-1]
				rightTpl = fmt.Sprintf("%q", lit)
			} else {
				rightTpl = right
			}

			return fmt.Sprintf("%s %s %s", entry.goOp, leftTpl, rightTpl)
		}
	}

	// Plain variable
	if strings.HasPrefix(cond, "$") {
		return vt.transpileVariableRef(cond)
	}

	return cond
}

func (vt *viewTranspiler) transpileURLArg(rawArg string) string {
	rawArg = strings.TrimSpace(rawArg)
	if rawArg == "" || rawArg == "''" || rawArg == "\"\"" || rawArg == "'/'" || rawArg == "\"/\"" {
		return "/"
	}

	parts := splitPHPConcat(rawArg)
	if len(parts) == 0 {
		return "/"
	}

	var sb strings.Builder
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if (strings.HasPrefix(part, "'") && strings.HasSuffix(part, "'")) ||
			(strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"")) {
			literal := part[1 : len(part)-1]
			sb.WriteString(literal)
		} else if strings.HasPrefix(part, "$") {
			varRef := vt.transpileVariableRef(part)
			sb.WriteString(fmt.Sprintf("{{ %s }}", varRef))
		} else {
			sb.WriteString(part)
		}
	}

	result := sb.String()
	if !strings.HasPrefix(result, "/") {
		result = "/" + result
	}
	return result
}

func splitPHPConcat(s string) []string {
	var parts []string
	var cur strings.Builder
	inQuote := byte(0)
	escaped := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			escaped = false
			cur.WriteByte(c)
			continue
		}
		if c == '\\' && inQuote != 0 {
			escaped = true
			cur.WriteByte(c)
			continue
		}
		if inQuote == 0 {
			if c == '\'' || c == '"' {
				inQuote = c
				cur.WriteByte(c)
			} else if c == '.' {
				parts = append(parts, strings.TrimSpace(cur.String()))
				cur.Reset()
			} else {
				cur.WriteByte(c)
			}
		} else {
			if c == inQuote {
				inQuote = 0
			}
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		parts = append(parts, strings.TrimSpace(cur.String()))
	}
	return parts
}

func extractFirstArg(s string) string {
	inQuote := byte(0)
	parenDepth := 0
	bracketDepth := 0
	escaped := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' && inQuote != 0 {
			escaped = true
			continue
		}
		if inQuote == 0 {
			switch c {
			case '\'', '"':
				inQuote = c
			case '(':
				parenDepth++
			case ')':
				parenDepth--
			case '[', '{':
				bracketDepth++
			case ']', '}':
				bracketDepth--
			case ',':
				if parenDepth == 0 && bracketDepth == 0 {
					return strings.TrimSpace(s[:i])
				}
			}
		} else {
			if c == inQuote {
				inQuote = 0
			}
		}
	}
	return strings.TrimSpace(s)
}

func extractParenContent(s string) string {
	if !strings.HasPrefix(s, "(") {
		return ""
	}
	depth := 0
	inQuote := byte(0)
	escaped := false
	start := 1

	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' && inQuote != 0 {
			escaped = true
			continue
		}
		if inQuote == 0 {
			if c == '\'' || c == '"' {
				inQuote = c
			} else if c == '(' {
				depth++
			} else if c == ')' {
				depth--
				if depth == 0 {
					return strings.TrimSpace(s[start:i])
				}
			}
		} else {
			if c == inQuote {
				inQuote = 0
			}
		}
	}
	return strings.TrimSpace(s[start:])
}

func findOpOutsideQuotes(s, op string) int {
	inQuote := byte(0)
	escaped := false

	for i := 0; i <= len(s)-len(op); i++ {
		c := s[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' && inQuote != 0 {
			escaped = true
			continue
		}
		if inQuote == 0 {
			if c == '\'' || c == '"' {
				inQuote = c
			} else if s[i:i+len(op)] == op {
				// Prevent matching '>' inside '->'
				if op == ">" && i > 0 && s[i-1] == '-' {
					continue
				}
				// Prevent matching '>' inside '>=' or '>>'
				if op == ">" && i+1 < len(s) && (s[i+1] == '=' || s[i+1] == '>') {
					continue
				}
				// Prevent matching '<' inside '<=' or '<<'
				if op == "<" && i+1 < len(s) && (s[i+1] == '=' || s[i+1] == '<') {
					continue
				}
				return i
			}
		} else {
			if c == inQuote {
				inQuote = 0
			}
		}
	}
	return -1
}

func isValidParenthesized(s string) bool {
	if !strings.HasPrefix(s, "(") || !strings.HasSuffix(s, ")") {
		return false
	}
	depth := 0
	inQuote := byte(0)
	escaped := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' && inQuote != 0 {
			escaped = true
			continue
		}
		if inQuote == 0 {
			if c == '\'' || c == '"' {
				inQuote = c
			} else if c == '(' {
				depth++
			} else if c == ')' {
				depth--
				if depth == 0 && i < len(s)-1 {
					return false
				}
			}
		} else {
			if c == inQuote {
				inQuote = 0
			}
		}
	}
	return depth == 0
}

// TranspileViewsDir walks srcDir and converts all .php view files to .html in outDir.
// It ignores security stub files (such as CI3's index.html Directory access forbidden).
func TranspileViewsDir(srcDir, outDir string) (int, error) {
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return 0, fmt.Errorf("stat srcDir: %w", err)
	}
	if !srcInfo.IsDir() {
		return 0, fmt.Errorf("srcDir is not a directory: %s", srcDir)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return 0, fmt.Errorf("creating outDir: %w", err)
	}

	count := 0
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		lowerBase := strings.ToLower(filepath.Base(path))

		// Ignore CI3 directory-listing security stubs
		if lowerBase == "index.html" || lowerBase == "index.htm" {
			return nil
		}

		// Only process PHP files
		if !strings.HasSuffix(lowerBase, ".php") {
			return nil
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		transpiled := TranspilePHPViewToTemplate(string(contentBytes))

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		relDir := filepath.Dir(relPath)
		baseWithoutExt := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		outFileName := baseWithoutExt + ".html"
		targetDir := filepath.Join(outDir, relDir)

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", targetDir, err)
		}

		targetFile := filepath.Join(targetDir, outFileName)
		if err := os.WriteFile(targetFile, []byte(transpiled), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", targetFile, err)
		}

		count++
		return nil
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}

// RunView2GoTplCLI parses CLI arguments and runs the view transpiler.
func RunView2GoTplCLI(args []string) error {
	fs := flag.NewFlagSet("view2gotpl", flag.ContinueOnError)
	srcFlag := fs.String("src", "", "Input views directory (or single PHP view file)")
	outFlag := fs.String("out", "./app/views", "Output directory for Go templates (or '-' for stdout)")
	fileFlag := fs.String("file", "", "Single PHP view file to transpile")

	fs.Usage = func() {
		fmt.Println("Usage: go run ./skills/ci3-to-goigniter view2gotpl [flags]")
		fmt.Println()
		fmt.Println("Converts CI3 PHP views to Go html/template files.")
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

	src := *srcFlag
	if src == "" && *fileFlag != "" {
		src = *fileFlag
	}

	if src == "" {
		return errors.New("missing required flag: -src <path to views dir or file>")
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat src %s: %w", src, err)
	}

	if !info.IsDir() {
		content, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("reading file %s: %w", src, err)
		}
		transpiled := TranspilePHPViewToTemplate(string(content))

		if *outFlag == "-" || *outFlag == "" {
			fmt.Print(transpiled)
			return nil
		}

		targetFile := *outFlag
		if strings.HasSuffix(targetFile, string(filepath.Separator)) || strings.HasSuffix(targetFile, "/") || !strings.HasSuffix(strings.ToLower(targetFile), ".html") {
			base := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src)) + ".html"
			targetFile = filepath.Join(targetFile, base)
		}

		dir := filepath.Dir(targetFile)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", dir, err)
			}
		}

		if err := os.WriteFile(targetFile, []byte(transpiled), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", targetFile, err)
		}
		fmt.Printf("Transpiled: %s -> %s\n", src, targetFile)
		return nil
	}

	count, err := TranspileViewsDir(src, *outFlag)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully transpiled %d view(s) to %s\n", count, *outFlag)
	return nil
}
