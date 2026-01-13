package database

import (
	"os"
	"testing"
)

type User struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Email  string `db:"email"`
	Status string `db:"status"`
}

func setupTestDB(t *testing.T) *DB {
	// Use in-memory SQLite database
	db, err := Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			status TEXT DEFAULT 'active'
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func TestInsert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := db.Table("users").Insert(map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	count, _ := db.Table("users").Count()
	if count != 1 {
		t.Errorf("Expected 1 row, got %d", count)
	}
}

func TestInsertGetId(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id, err := db.Table("users").InsertGetId(map[string]any{
		"name":  "Jane Doe",
		"email": "jane@example.com",
	})
	if err != nil {
		t.Fatalf("InsertGetId failed: %v", err)
	}

	if id != 1 {
		t.Errorf("Expected id 1, got %d", id)
	}
}

func TestSelect(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data
	db.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com", "status": "active"})
	db.Table("users").Insert(map[string]any{"name": "Jane", "email": "jane@test.com", "status": "inactive"})

	var users []User
	err := db.Table("users").Get(&users)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestSelectColumns(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com"})

	var users []User
	err := db.Table("users").Select("name", "email").Get(&users)
	if err != nil {
		t.Fatalf("Select with columns failed: %v", err)
	}

	if users[0].Name != "John" {
		t.Errorf("Expected name 'John', got '%s'", users[0].Name)
	}
}

func TestWhere(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com", "status": "active"})
	db.Table("users").Insert(map[string]any{"name": "Jane", "email": "jane@test.com", "status": "inactive"})

	var users []User
	err := db.Table("users").Where("status", "active").Get(&users)
	if err != nil {
		t.Fatalf("Where failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	if users[0].Name != "John" {
		t.Errorf("Expected name 'John', got '%s'", users[0].Name)
	}
}

func TestWhereOperator(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com"})
	db.Table("users").Insert(map[string]any{"name": "Jane", "email": "jane@test.com"})
	db.Table("users").Insert(map[string]any{"name": "Bob", "email": "bob@test.com"})

	count, _ := db.Table("users").Where("id", ">", 1).Count()
	if count != 2 {
		t.Errorf("Expected 2 users with id > 1, got %d", count)
	}
}

func TestWhereIn(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com"})
	db.Table("users").Insert(map[string]any{"name": "Jane", "email": "jane@test.com"})
	db.Table("users").Insert(map[string]any{"name": "Bob", "email": "bob@test.com"})

	var users []User
	err := db.Table("users").WhereIn("id", []int{1, 3}).Get(&users)
	if err != nil {
		t.Fatalf("WhereIn failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestOrWhere(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com", "status": "active"})
	db.Table("users").Insert(map[string]any{"name": "Jane", "email": "jane@test.com", "status": "inactive"})
	db.Table("users").Insert(map[string]any{"name": "Bob", "email": "bob@test.com", "status": "pending"})

	var users []User
	err := db.Table("users").Where("status", "active").OrWhere("status", "inactive").Get(&users)
	if err != nil {
		t.Fatalf("OrWhere failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestOrderBy(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Table("users").Insert(map[string]any{"name": "Charlie", "email": "charlie@test.com"})
	db.Table("users").Insert(map[string]any{"name": "Alice", "email": "alice@test.com"})
	db.Table("users").Insert(map[string]any{"name": "Bob", "email": "bob@test.com"})

	var users []User
	err := db.Table("users").OrderBy("name", "ASC").Get(&users)
	if err != nil {
		t.Fatalf("OrderBy failed: %v", err)
	}

	if users[0].Name != "Alice" {
		t.Errorf("Expected first user 'Alice', got '%s'", users[0].Name)
	}
}

func TestLimitOffset(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Table("users").Insert(map[string]any{"name": "User1", "email": "user1@test.com"})
	db.Table("users").Insert(map[string]any{"name": "User2", "email": "user2@test.com"})
	db.Table("users").Insert(map[string]any{"name": "User3", "email": "user3@test.com"})

	var users []User
	err := db.Table("users").Limit(2).Offset(1).Get(&users)
	if err != nil {
		t.Fatalf("Limit/Offset failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestFirst(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com"})
	db.Table("users").Insert(map[string]any{"name": "Jane", "email": "jane@test.com"})

	var users []User
	err := db.Table("users").Where("id", 1).First(&users)
	if err != nil {
		t.Fatalf("First failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
}

func TestUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com", "status": "active"})

	err := db.Table("users").Where("id", 1).Update(map[string]any{
		"status": "inactive",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	var users []User
	db.Table("users").Where("id", 1).Get(&users)
	if users[0].Status != "inactive" {
		t.Errorf("Expected status 'inactive', got '%s'", users[0].Status)
	}
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com"})
	db.Table("users").Insert(map[string]any{"name": "Jane", "email": "jane@test.com"})

	err := db.Table("users").Where("id", 1).Delete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	count, _ := db.Table("users").Count()
	if count != 1 {
		t.Errorf("Expected 1 row after delete, got %d", count)
	}
}

func TestCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com", "status": "active"})
	db.Table("users").Insert(map[string]any{"name": "Jane", "email": "jane@test.com", "status": "active"})
	db.Table("users").Insert(map[string]any{"name": "Bob", "email": "bob@test.com", "status": "inactive"})

	count, err := db.Table("users").Where("status", "active").Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestGetMap(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com"})

	results, err := db.Table("users").GetMap()
	if err != nil {
		t.Fatalf("GetMap failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0]["name"] != "John" {
		t.Errorf("Expected name 'John', got '%v'", results[0]["name"])
	}
}

func TestTransaction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Test successful transaction
	err := db.Transaction(func(tx *DB) error {
		tx.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com"})
		tx.Table("users").Insert(map[string]any{"name": "Jane", "email": "jane@test.com"})
		return nil
	})
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	count, _ := db.Table("users").Count()
	if count != 2 {
		t.Errorf("Expected 2 rows after commit, got %d", count)
	}
}

func TestTransactionRollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert initial data
	db.Table("users").Insert(map[string]any{"name": "Initial", "email": "initial@test.com"})

	// Test rollback
	err := db.Transaction(func(tx *DB) error {
		tx.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com"})
		return os.ErrInvalid // Return error to trigger rollback
	})

	if err == nil {
		t.Error("Expected error from transaction")
	}

	count, _ := db.Table("users").Count()
	if count != 1 {
		t.Errorf("Expected 1 row after rollback, got %d", count)
	}
}

func TestRawQuery(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Table("users").Insert(map[string]any{"name": "John", "email": "john@test.com"})

	var users []User
	err := db.Query("SELECT * FROM users WHERE name = ?", "John").Get(&users)
	if err != nil {
		t.Fatalf("Raw query failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
}

func TestInsertStruct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := User{
		Name:   "John Struct",
		Email:  "john.struct@test.com",
		Status: "active",
	}

	err := db.Table("users").InsertStruct(&user)
	if err != nil {
		t.Fatalf("InsertStruct failed: %v", err)
	}

	count, _ := db.Table("users").Count()
	if count != 1 {
		t.Errorf("Expected 1 row, got %d", count)
	}
}

func TestToSQL(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	sql := db.Table("users").
		Select("id", "name").
		Where("status", "active").
		OrderBy("name", "ASC").
		Limit(10).
		ToSQL()

	expected := "SELECT id, name FROM users WHERE status = ? ORDER BY name ASC LIMIT 10"
	if sql != expected {
		t.Errorf("Expected SQL:\n%s\nGot:\n%s", expected, sql)
	}
}
