package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

type DB struct {
	*sql.DB
}

// NewDatabase creates a new database connection and initializes the schema
func NewDatabase(dbPath string) (*DB, error) {
	dbExists := true
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		dbExists = false
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}
	if dbExists {
		fmt.Printf("Using existing database at: %s\n", dbPath)
	} else {
		fmt.Printf("Creating new database at: %s\n", dbPath)
	}

	// Open database connection
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	db := &DB{sqlDB}

	// Initialize the database schema, unless it already exists
	if err := db.initializeSchema(); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return db, nil
		}
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}
	return db, nil
}

// initializeSchema reads and executes the schema.sql file
func (db *DB) initializeSchema() error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	return nil
}
