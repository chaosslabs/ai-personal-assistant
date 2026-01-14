package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// DB holds the database connection
type DB struct {
	*sql.DB
	path string
}

// Config holds database configuration
type Config struct {
	DataDir string // App data directory
	DBName  string // Database filename
}

// NewDB creates a new database connection
func NewDB(config Config) (*DB, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Full path to database file
	dbPath := filepath.Join(config.DataDir, config.DBName)

	// Open database with SQLite-specific options
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_temp_store=memory", dbPath)
	
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(0) // No maximum lifetime

	db := &DB{
		DB:   sqlDB,
		path: dbPath,
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Path returns the database file path
func (db *DB) Path() string {
	return db.path
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.DB != nil {
		return db.DB.Close()
	}
	return nil
}

// DefaultConfig returns the default database configuration for personal-assist
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, "Library", "Application Support", "personal-assist")
	
	return Config{
		DataDir: dataDir,
		DBName:  "personal-assist.db",
	}
}

// ActivitiesDir returns the activities directory path
func ActivitiesDir(dataDir string) string {
	return filepath.Join(dataDir, "activities")
}

// ModelsDir returns the models directory path  
func ModelsDir(dataDir string) string {
	return filepath.Join(dataDir, "models")
}

// ActivityDir returns the specific activity directory path
func ActivityDir(dataDir, activityID string) string {
	return filepath.Join(ActivitiesDir(dataDir), activityID)
}

// ActivityAudioDir returns the activity audio directory path
func ActivityAudioDir(dataDir, activityID string) string {
	return filepath.Join(ActivityDir(dataDir, activityID), "audio")
}

// EnsureDirectories creates all necessary directories
func EnsureDirectories(dataDir string) error {
	dirs := []string{
		dataDir,
		ActivitiesDir(dataDir),
		ModelsDir(dataDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}