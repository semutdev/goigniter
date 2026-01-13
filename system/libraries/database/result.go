package database

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// RawResult represents a raw query result.
type RawResult struct {
	db    *DB
	query string
	args  []any
}

// Get executes the raw query and scans results into dest.
func (r *RawResult) Get(dest any) error {
	rows, err := r.db.execer().Query(r.query, r.args...)
	if err != nil {
		r.db.lastError = err
		return err
	}
	defer rows.Close()

	return scanRows(rows, dest)
}

// GetMap executes the raw query and returns results as []map[string]any.
func (r *RawResult) GetMap() ([]map[string]any, error) {
	rows, err := r.db.execer().Query(r.query, r.args...)
	if err != nil {
		r.db.lastError = err
		return nil, err
	}
	defer rows.Close()

	return scanRowsToMap(rows)
}

// First gets the first result.
func (r *RawResult) First(dest any) error {
	rows, err := r.db.execer().Query(r.query, r.args...)
	if err != nil {
		r.db.lastError = err
		return err
	}
	defer rows.Close()

	return scanRow(rows, dest)
}

// FirstMap gets the first result as map.
func (r *RawResult) FirstMap() (map[string]any, error) {
	results, err := r.GetMap()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// scanRows scans all rows into a slice of structs.
func scanRows(rows *sql.Rows, dest any) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("database: dest must be a pointer")
	}

	sliceValue := destValue.Elem()
	if sliceValue.Kind() != reflect.Slice {
		return fmt.Errorf("database: dest must be a pointer to a slice")
	}

	elemType := sliceValue.Type().Elem()
	isPtr := elemType.Kind() == reflect.Ptr
	if isPtr {
		elemType = elemType.Elem()
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		elem := reflect.New(elemType).Elem()
		scanDest := makeScanDest(elem, columns)

		if err := rows.Scan(scanDest...); err != nil {
			return err
		}

		if isPtr {
			sliceValue.Set(reflect.Append(sliceValue, elem.Addr()))
		} else {
			sliceValue.Set(reflect.Append(sliceValue, elem))
		}
	}

	return rows.Err()
}

// scanRow scans a single row into a struct.
func scanRow(rows *sql.Rows, dest any) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("database: dest must be a pointer")
	}

	elemValue := destValue.Elem()
	elemType := elemValue.Type()

	// Handle pointer to pointer
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
		newElem := reflect.New(elemType)
		elemValue.Set(newElem)
		elemValue = newElem.Elem()
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	if !rows.Next() {
		return sql.ErrNoRows
	}

	scanDest := makeScanDest(elemValue, columns)
	return rows.Scan(scanDest...)
}

// scanRowsToMap scans all rows into []map[string]any.
func scanRowsToMap(rows *sql.Rows) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]any

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return results, rows.Err()
}

// makeScanDest creates scan destinations for a struct.
func makeScanDest(elem reflect.Value, columns []string) []any {
	scanDest := make([]any, len(columns))
	fieldMap := makeFieldMap(elem.Type())

	for i, col := range columns {
		colLower := strings.ToLower(col)
		if fieldIdx, ok := fieldMap[colLower]; ok {
			scanDest[i] = elem.Field(fieldIdx).Addr().Interface()
		} else {
			// Column doesn't match any field, use a dummy variable
			var dummy any
			scanDest[i] = &dummy
		}
	}

	return scanDest
}

// makeFieldMap creates a map of lowercase column names to field indices.
func makeFieldMap(t reflect.Type) map[string]int {
	fieldMap := make(map[string]int)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Check for db tag
		if tag := field.Tag.Get("db"); tag != "" {
			if tag != "-" {
				fieldMap[strings.ToLower(tag)] = i
			}
			continue
		}

		// Check for json tag
		if tag := field.Tag.Get("json"); tag != "" {
			tagName := strings.Split(tag, ",")[0]
			if tagName != "-" {
				fieldMap[strings.ToLower(tagName)] = i
			}
			continue
		}

		// Use field name
		fieldMap[strings.ToLower(field.Name)] = i
	}

	return fieldMap
}

// InsertStruct inserts a struct into the table.
func (b *Builder) InsertStruct(src any) error {
	data := structToMap(src)
	return b.Insert(data)
}

// InsertStructGetId inserts a struct and returns the last insert ID.
func (b *Builder) InsertStructGetId(src any) (int64, error) {
	data := structToMap(src)
	return b.InsertGetId(data)
}

// UpdateStruct updates with struct data.
func (b *Builder) UpdateStruct(src any) error {
	data := structToMap(src)
	return b.Update(data)
}

// structToMap converts a struct to map[string]any.
func structToMap(src any) map[string]any {
	data := make(map[string]any)

	val := reflect.ValueOf(src)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if !fieldVal.CanInterface() {
			continue
		}

		// Get column name from tag or field name
		colName := ""

		if tag := field.Tag.Get("db"); tag != "" {
			if tag == "-" {
				continue
			}
			colName = tag
		} else if tag := field.Tag.Get("json"); tag != "" {
			tagName := strings.Split(tag, ",")[0]
			if tagName == "-" {
				continue
			}
			colName = tagName
		} else {
			colName = toSnakeCase(field.Name)
		}

		data[colName] = fieldVal.Interface()
	}

	return data
}

// toSnakeCase converts CamelCase to snake_case.
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
