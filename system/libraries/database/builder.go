package database

import (
	"fmt"
	"strings"
)

// Builder represents a query builder.
type Builder struct {
	db          *DB
	table       string
	columns     []string
	wheres      []whereClause
	orders      []orderClause
	joins       []joinClause
	limitVal    int
	offsetVal   int
	groupByVal  []string
	havingVal   string
	havingArgs  []any
}

type whereClause struct {
	column   string
	operator string
	value    any
	boolean  string // "AND" or "OR"
	isRaw    bool
	rawSQL   string
}

type orderClause struct {
	column    string
	direction string
}

type joinClause struct {
	joinType string // "INNER", "LEFT", "RIGHT"
	table    string
	col1     string
	operator string
	col2     string
}

// newBuilder creates a new query builder.
func newBuilder(db *DB, table string) *Builder {
	return &Builder{
		db:      db,
		table:   table,
		columns: []string{"*"},
	}
}

// Select specifies which columns to select.
func (b *Builder) Select(columns ...string) *Builder {
	b.columns = columns
	return b
}

// Where adds a WHERE clause.
func (b *Builder) Where(args ...any) *Builder {
	b.addWhere("AND", args...)
	return b
}

// OrWhere adds an OR WHERE clause.
func (b *Builder) OrWhere(args ...any) *Builder {
	b.addWhere("OR", args...)
	return b
}

func (b *Builder) addWhere(boolean string, args ...any) {
	var w whereClause
	w.boolean = boolean

	switch len(args) {
	case 2:
		// Where("column", value) -> column = value
		w.column = args[0].(string)
		w.operator = "="
		w.value = args[1]
	case 3:
		// Where("column", ">", value) -> column > value
		w.column = args[0].(string)
		w.operator = args[1].(string)
		w.value = args[2]
	default:
		return
	}

	b.wheres = append(b.wheres, w)
}

// WhereIn adds a WHERE IN clause.
func (b *Builder) WhereIn(column string, values any) *Builder {
	b.wheres = append(b.wheres, whereClause{
		column:   column,
		operator: "IN",
		value:    values,
		boolean:  "AND",
	})
	return b
}

// WhereNotIn adds a WHERE NOT IN clause.
func (b *Builder) WhereNotIn(column string, values any) *Builder {
	b.wheres = append(b.wheres, whereClause{
		column:   column,
		operator: "NOT IN",
		value:    values,
		boolean:  "AND",
	})
	return b
}

// WhereNull adds a WHERE IS NULL clause.
func (b *Builder) WhereNull(column string) *Builder {
	b.wheres = append(b.wheres, whereClause{
		column:   column,
		operator: "IS NULL",
		boolean:  "AND",
	})
	return b
}

// WhereNotNull adds a WHERE IS NOT NULL clause.
func (b *Builder) WhereNotNull(column string) *Builder {
	b.wheres = append(b.wheres, whereClause{
		column:   column,
		operator: "IS NOT NULL",
		boolean:  "AND",
	})
	return b
}

// WhereRaw adds a raw WHERE clause.
func (b *Builder) WhereRaw(sql string, args ...any) *Builder {
	b.wheres = append(b.wheres, whereClause{
		isRaw:   true,
		rawSQL:  sql,
		value:   args,
		boolean: "AND",
	})
	return b
}

// OrderBy adds an ORDER BY clause.
func (b *Builder) OrderBy(column, direction string) *Builder {
	b.orders = append(b.orders, orderClause{
		column:    column,
		direction: strings.ToUpper(direction),
	})
	return b
}

// Limit sets the LIMIT clause.
func (b *Builder) Limit(limit int) *Builder {
	b.limitVal = limit
	return b
}

// Offset sets the OFFSET clause.
func (b *Builder) Offset(offset int) *Builder {
	b.offsetVal = offset
	return b
}

// GroupBy sets the GROUP BY clause.
func (b *Builder) GroupBy(columns ...string) *Builder {
	b.groupByVal = columns
	return b
}

// Having sets the HAVING clause.
func (b *Builder) Having(sql string, args ...any) *Builder {
	b.havingVal = sql
	b.havingArgs = args
	return b
}

// Join adds an INNER JOIN clause.
func (b *Builder) Join(table, col1, operator, col2 string) *Builder {
	b.joins = append(b.joins, joinClause{
		joinType: "INNER",
		table:    table,
		col1:     col1,
		operator: operator,
		col2:     col2,
	})
	return b
}

// LeftJoin adds a LEFT JOIN clause.
func (b *Builder) LeftJoin(table, col1, operator, col2 string) *Builder {
	b.joins = append(b.joins, joinClause{
		joinType: "LEFT",
		table:    table,
		col1:     col1,
		operator: operator,
		col2:     col2,
	})
	return b
}

// RightJoin adds a RIGHT JOIN clause.
func (b *Builder) RightJoin(table, col1, operator, col2 string) *Builder {
	b.joins = append(b.joins, joinClause{
		joinType: "RIGHT",
		table:    table,
		col1:     col1,
		operator: operator,
		col2:     col2,
	})
	return b
}

// buildSelect builds a SELECT query.
func (b *Builder) buildSelect() (string, []any) {
	var args []any
	var sql strings.Builder

	// SELECT columns
	sql.WriteString("SELECT ")
	sql.WriteString(strings.Join(b.columns, ", "))

	// FROM table
	sql.WriteString(" FROM ")
	sql.WriteString(b.table)

	// JOINs
	for _, j := range b.joins {
		sql.WriteString(fmt.Sprintf(" %s JOIN %s ON %s %s %s",
			j.joinType, j.table, j.col1, j.operator, j.col2))
	}

	// WHERE
	if len(b.wheres) > 0 {
		sql.WriteString(" WHERE ")
		whereSQL, whereArgs := b.buildWheres()
		sql.WriteString(whereSQL)
		args = append(args, whereArgs...)
	}

	// GROUP BY
	if len(b.groupByVal) > 0 {
		sql.WriteString(" GROUP BY ")
		sql.WriteString(strings.Join(b.groupByVal, ", "))
	}

	// HAVING
	if b.havingVal != "" {
		sql.WriteString(" HAVING ")
		sql.WriteString(b.havingVal)
		args = append(args, b.havingArgs...)
	}

	// ORDER BY
	if len(b.orders) > 0 {
		sql.WriteString(" ORDER BY ")
		var orderParts []string
		for _, o := range b.orders {
			orderParts = append(orderParts, fmt.Sprintf("%s %s", o.column, o.direction))
		}
		sql.WriteString(strings.Join(orderParts, ", "))
	}

	// LIMIT
	if b.limitVal > 0 {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", b.limitVal))
	}

	// OFFSET
	if b.offsetVal > 0 {
		sql.WriteString(fmt.Sprintf(" OFFSET %d", b.offsetVal))
	}

	return sql.String(), args
}

func (b *Builder) buildWheres() (string, []any) {
	var parts []string
	var args []any

	for i, w := range b.wheres {
		var part string

		if w.isRaw {
			part = w.rawSQL
			if rawArgs, ok := w.value.([]any); ok {
				args = append(args, rawArgs...)
			}
		} else if w.operator == "IS NULL" || w.operator == "IS NOT NULL" {
			part = fmt.Sprintf("%s %s", w.column, w.operator)
		} else if w.operator == "IN" || w.operator == "NOT IN" {
			placeholders, inArgs := b.buildInClause(w.value)
			part = fmt.Sprintf("%s %s (%s)", w.column, w.operator, placeholders)
			args = append(args, inArgs...)
		} else {
			part = fmt.Sprintf("%s %s ?", w.column, w.operator)
			args = append(args, w.value)
		}

		if i == 0 {
			parts = append(parts, part)
		} else {
			parts = append(parts, fmt.Sprintf("%s %s", w.boolean, part))
		}
	}

	return strings.Join(parts, " "), args
}

func (b *Builder) buildInClause(value any) (string, []any) {
	var placeholders []string
	var args []any

	switch v := value.(type) {
	case []int:
		for _, val := range v {
			placeholders = append(placeholders, "?")
			args = append(args, val)
		}
	case []int64:
		for _, val := range v {
			placeholders = append(placeholders, "?")
			args = append(args, val)
		}
	case []string:
		for _, val := range v {
			placeholders = append(placeholders, "?")
			args = append(args, val)
		}
	case []any:
		for _, val := range v {
			placeholders = append(placeholders, "?")
			args = append(args, val)
		}
	}

	return strings.Join(placeholders, ", "), args
}

// Get executes the query and scans results into dest.
func (b *Builder) Get(dest any) error {
	query, args := b.buildSelect()
	rows, err := b.db.execer().Query(query, args...)
	if err != nil {
		b.db.lastError = err
		return err
	}
	defer rows.Close()

	return scanRows(rows, dest)
}

// GetMap executes the query and returns results as []map[string]any.
func (b *Builder) GetMap() ([]map[string]any, error) {
	query, args := b.buildSelect()
	rows, err := b.db.execer().Query(query, args...)
	if err != nil {
		b.db.lastError = err
		return nil, err
	}
	defer rows.Close()

	return scanRowsToMap(rows)
}

// First gets the first result.
func (b *Builder) First(dest any) error {
	b.limitVal = 1
	return b.Get(dest)
}

// FirstMap gets the first result as map.
func (b *Builder) FirstMap() (map[string]any, error) {
	b.limitVal = 1
	results, err := b.GetMap()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Count returns the count of rows.
func (b *Builder) Count() (int64, error) {
	b.columns = []string{"COUNT(*) as count"}
	query, args := b.buildSelect()

	var count int64
	err := b.db.execer().QueryRow(query, args...).Scan(&count)
	if err != nil {
		b.db.lastError = err
		return 0, err
	}
	return count, nil
}

// Sum returns the sum of a column.
func (b *Builder) Sum(column string) (float64, error) {
	b.columns = []string{fmt.Sprintf("COALESCE(SUM(%s), 0) as sum", column)}
	query, args := b.buildSelect()

	var sum float64
	err := b.db.execer().QueryRow(query, args...).Scan(&sum)
	if err != nil {
		b.db.lastError = err
		return 0, err
	}
	return sum, nil
}

// Avg returns the average of a column.
func (b *Builder) Avg(column string) (float64, error) {
	b.columns = []string{fmt.Sprintf("COALESCE(AVG(%s), 0) as avg", column)}
	query, args := b.buildSelect()

	var avg float64
	err := b.db.execer().QueryRow(query, args...).Scan(&avg)
	if err != nil {
		b.db.lastError = err
		return 0, err
	}
	return avg, nil
}

// Min returns the minimum value of a column.
func (b *Builder) Min(column string) (float64, error) {
	b.columns = []string{fmt.Sprintf("MIN(%s) as min", column)}
	query, args := b.buildSelect()

	var min float64
	err := b.db.execer().QueryRow(query, args...).Scan(&min)
	if err != nil {
		b.db.lastError = err
		return 0, err
	}
	return min, nil
}

// Max returns the maximum value of a column.
func (b *Builder) Max(column string) (float64, error) {
	b.columns = []string{fmt.Sprintf("MAX(%s) as max", column)}
	query, args := b.buildSelect()

	var max float64
	err := b.db.execer().QueryRow(query, args...).Scan(&max)
	if err != nil {
		b.db.lastError = err
		return 0, err
	}
	return max, nil
}

// Insert inserts a new row.
func (b *Builder) Insert(data map[string]any) error {
	var columns []string
	var placeholders []string
	var args []any

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		b.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	_, err := b.db.execer().Exec(query, args...)
	if err != nil {
		b.db.lastError = err
	}
	return err
}

// InsertGetId inserts a new row and returns the last insert ID.
func (b *Builder) InsertGetId(data map[string]any) (int64, error) {
	var columns []string
	var placeholders []string
	var args []any

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		b.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	result, err := b.db.execer().Exec(query, args...)
	if err != nil {
		b.db.lastError = err
		return 0, err
	}

	return result.LastInsertId()
}

// Update updates rows.
func (b *Builder) Update(data map[string]any) error {
	var setParts []string
	var args []any

	for col, val := range data {
		setParts = append(setParts, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}

	query := fmt.Sprintf("UPDATE %s SET %s", b.table, strings.Join(setParts, ", "))

	if len(b.wheres) > 0 {
		whereSQL, whereArgs := b.buildWheres()
		query += " WHERE " + whereSQL
		args = append(args, whereArgs...)
	}

	_, err := b.db.execer().Exec(query, args...)
	if err != nil {
		b.db.lastError = err
	}
	return err
}

// Delete deletes rows.
func (b *Builder) Delete() error {
	query := fmt.Sprintf("DELETE FROM %s", b.table)
	var args []any

	if len(b.wheres) > 0 {
		whereSQL, whereArgs := b.buildWheres()
		query += " WHERE " + whereSQL
		args = whereArgs
	}

	_, err := b.db.execer().Exec(query, args...)
	if err != nil {
		b.db.lastError = err
	}
	return err
}

// ToSQL returns the SQL query string (for debugging).
func (b *Builder) ToSQL() string {
	query, _ := b.buildSelect()
	return query
}
