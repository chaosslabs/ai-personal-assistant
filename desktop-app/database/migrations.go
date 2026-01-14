package database

import (
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"
)

//go:embed schema.sql
var schemaFS embed.FS

// Migration represents a database migration
type Migration struct {
	Version int
	SQL     string
}

// Migrator handles database migrations
type Migrator struct {
	db *DB
}

// NewMigrator creates a new migrator
func NewMigrator(db *DB) *Migrator {
	return &Migrator{db: db}
}

// InitializeSchema initializes the database schema
func (m *Migrator) InitializeSchema() error {
	// Check if database is already initialized
	initialized, err := m.isSchemaInitialized()
	if err != nil {
		return fmt.Errorf("failed to check schema initialization: %w", err)
	}

	if initialized {
		return nil
	}

	// Read and execute schema.sql
	schemaSQL, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute schema in a transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute entire schema as a single statement since SQLite can handle multiple statements
	_, err = tx.Exec(string(schemaSQL))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit schema transaction: %w", err)
	}

	return nil
}

// isSchemaInitialized checks if the schema has been initialized
func (m *Migrator) isSchemaInitialized() (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'`
	err := m.db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetCurrentVersion returns the current schema version
func (m *Migrator) GetCurrentVersion() (int, error) {
	var version int
	query := `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`
	err := m.db.QueryRow(query).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get current version: %w", err)
	}
	return version, nil
}

// AddMigration records a new migration in the database
func (m *Migrator) AddMigration(version int) error {
	query := `INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`
	_, err := m.db.Exec(query, version, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("failed to record migration %d: %w", version, err)
	}
	return nil
}

// GetAppliedMigrations returns all applied migration versions
func (m *Migrator) GetAppliedMigrations() ([]int, error) {
	query := `SELECT version FROM schema_migrations ORDER BY version`
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		versions = append(versions, version)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration versions: %w", err)
	}

	return versions, nil
}

// RunMigrations runs any pending migrations
func (m *Migrator) RunMigrations(migrations []Migration) error {
	if len(migrations) == 0 {
		return nil
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	currentVersion, err := m.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue // Skip already applied migrations
		}

		if err := m.runMigration(migration); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", migration.Version, err)
		}
	}

	return nil
}

// runMigration runs a single migration
func (m *Migrator) runMigration(migration Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin migration transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	statements := splitSQL(migration.SQL)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute migration statement: %w\nStatement: %s", err, stmt)
		}
	}

	// Record migration as applied
	_, err = tx.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`,
		migration.Version, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	return nil
}

// Validate checks the database schema integrity
func (m *Migrator) Validate() error {
	// Check that all required tables exist
	requiredTables := []string{
		"users",
		"activities", 
		"audio_recordings",
		"transcript_chunks",
		"schema_migrations",
	}

	for _, table := range requiredTables {
		exists, err := m.tableExists(table)
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}
		if !exists {
			return fmt.Errorf("required table %s does not exist", table)
		}
	}

	// Check foreign key constraints are enabled
	var fkEnabled int
	err := m.db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	if err != nil {
		return fmt.Errorf("failed to check foreign keys setting: %w", err)
	}
	if fkEnabled != 1 {
		return fmt.Errorf("foreign keys are not enabled")
	}

	return nil
}

// tableExists checks if a table exists
func (m *Migrator) tableExists(tableName string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`
	err := m.db.QueryRow(query, tableName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// splitSQL splits SQL text into individual statements
func splitSQL(sql string) []string {
	// Simple SQL statement splitter - splits on semicolons not in quotes
	var statements []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(sql); i++ {
		char := sql[i]
		
		if !inQuotes {
			if char == '\'' || char == '"' {
				inQuotes = true
				quoteChar = char
			} else if char == ';' {
				stmt := strings.TrimSpace(current.String())
				if stmt != "" {
					statements = append(statements, stmt)
				}
				current.Reset()
				continue
			}
		} else if char == quoteChar {
			// Check for escaped quotes
			if i+1 < len(sql) && sql[i+1] == quoteChar {
				current.WriteByte(char)
				i++ // Skip next character
			} else {
				inQuotes = false
				quoteChar = 0
			}
		}
		
		current.WriteByte(char)
	}
	
	// Add final statement if any
	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}
	
	return statements
}

// CreateMigration creates a new migration with the given SQL
func CreateMigration(version int, sql string) Migration {
	return Migration{
		Version: version,
		SQL:     sql,
	}
}