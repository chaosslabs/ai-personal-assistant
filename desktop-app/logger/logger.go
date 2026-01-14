package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	Logger *logrus.Logger
)

// LogConfig holds configuration for logging
type LogConfig struct {
	Level    string
	LogsDir  string
	MaxSize  int64  // Maximum size of a single log file in bytes
	MaxFiles int    // Maximum number of log files to keep
}

// DefaultLogConfig returns default logging configuration
func DefaultLogConfig() *LogConfig {
	homeDir, _ := os.UserHomeDir()
	logsDir := filepath.Join(homeDir, "Library", "Application Support", "personal-assist", "logs")
	
	return &LogConfig{
		Level:    "info",
		LogsDir:  logsDir,
		MaxSize:  10 * 1024 * 1024, // 10MB
		MaxFiles: 5,
	}
}

// InitLogger initializes the global logger with the provided configuration
func InitLogger(config *LogConfig) error {
	Logger = logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	Logger.SetLevel(level)

	// Create logs directory
	if err := os.MkdirAll(config.LogsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(config.LogsDir, fmt.Sprintf("personal-assist_%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	// Set up multi-writer to write to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, file)
	Logger.SetOutput(multiWriter)

	// Set custom formatter
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     false,
	})

	// Log startup message
	Logger.WithFields(logrus.Fields{
		"log_file": logFile,
		"level":    config.Level,
	}).Info("Logger initialized")

	// Clean up old log files
	if err := cleanupOldLogs(config); err != nil {
		Logger.WithError(err).Warn("Failed to cleanup old log files")
	}

	return nil
}

// cleanupOldLogs removes old log files based on configuration
func cleanupOldLogs(config *LogConfig) error {
	files, err := filepath.Glob(filepath.Join(config.LogsDir, "personal-assist_*.log"))
	if err != nil {
		return err
	}

	if len(files) <= config.MaxFiles {
		return nil
	}

	// Sort files by modification time (oldest first)
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var fileInfos []fileInfo
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{
			path:    file,
			modTime: info.ModTime(),
		})
	}

	// Sort by modification time
	for i := 0; i < len(fileInfos)-1; i++ {
		for j := i + 1; j < len(fileInfos); j++ {
			if fileInfos[i].modTime.After(fileInfos[j].modTime) {
				fileInfos[i], fileInfos[j] = fileInfos[j], fileInfos[i]
			}
		}
	}

	// Remove oldest files
	filesToRemove := len(fileInfos) - config.MaxFiles
	for i := 0; i < filesToRemove; i++ {
		if err := os.Remove(fileInfos[i].path); err != nil {
			Logger.WithError(err).WithField("file", fileInfos[i].path).Warn("Failed to remove old log file")
		} else {
			Logger.WithField("file", fileInfos[i].path).Info("Removed old log file")
		}
	}

	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *logrus.Logger {
	if Logger == nil {
		// Fallback to default logger if not initialized
		Logger = logrus.New()
		Logger.SetLevel(logrus.InfoLevel)
	}
	return Logger
}

// WithField creates a new logger entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

// WithFields creates a new logger entry with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

// WithError creates a new logger entry with an error field
func WithError(err error) *logrus.Entry {
	return GetLogger().WithError(err)
}

// Info logs an info message
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// Panic logs a panic message and panics
func Panic(args ...interface{}) {
	GetLogger().Panic(args...)
}

// Panicf logs a formatted panic message and panics
func Panicf(format string, args ...interface{}) {
	GetLogger().Panicf(format, args...)
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}