package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// FileManager handles file operations for the personal assistant
type FileManager struct {
	dataDir string
}

// NewFileManager creates a new file manager
func NewFileManager(dataDir string) *FileManager {
	return &FileManager{
		dataDir: dataDir,
	}
}

// GetDataDir returns the base data directory
func (fm *FileManager) GetDataDir() string {
	return fm.dataDir
}

// GetActivitiesDir returns the activities directory
func (fm *FileManager) GetActivitiesDir() string {
	return filepath.Join(fm.dataDir, "activities")
}

// GetModelsDir returns the models directory
func (fm *FileManager) GetModelsDir() string {
	return filepath.Join(fm.dataDir, "models")
}

// GetActivityDir returns the directory for a specific activity
func (fm *FileManager) GetActivityDir(activityID string) string {
	return filepath.Join(fm.GetActivitiesDir(), activityID)
}

// GetActivityAudioDir returns the audio directory for a specific activity
func (fm *FileManager) GetActivityAudioDir(activityID string) string {
	return filepath.Join(fm.GetActivityDir(activityID), "audio")
}

// EnsureDirectories creates all necessary directories
func (fm *FileManager) EnsureDirectories() error {
	dirs := []string{
		fm.dataDir,
		fm.GetActivitiesDir(),
		fm.GetModelsDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// EnsureActivityDirectories creates directories for a specific activity
func (fm *FileManager) EnsureActivityDirectories(activityID string) error {
	dirs := []string{
		fm.GetActivityDir(activityID),
		fm.GetActivityAudioDir(activityID),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create activity directory %s: %w", dir, err)
		}
	}

	return nil
}

// GenerateAudioFileName generates a unique audio file name for an activity
func (fm *FileManager) GenerateAudioFileName(activityID, extension string) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s.%s", activityID[:8], timestamp, extension)
}

// GetAudioFilePath returns the full path for an audio file
func (fm *FileManager) GetAudioFilePath(activityID, fileName string) string {
	return filepath.Join(fm.GetActivityAudioDir(activityID), fileName)
}

// GetRelativeAudioFilePath returns the relative path from the data directory
func (fm *FileManager) GetRelativeAudioFilePath(activityID, fileName string) string {
	return filepath.Join("activities", activityID, "audio", fileName)
}

// GetAbsolutePathFromRelative converts a relative path to absolute
func (fm *FileManager) GetAbsolutePathFromRelative(relativePath string) string {
	return filepath.Join(fm.dataDir, relativePath)
}

// FileExists checks if a file exists
func (fm *FileManager) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// GetFileSize returns the size of a file in bytes
func (fm *FileManager) GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	return info.Size(), nil
}

// CopyFile copies a file from source to destination
func (fm *FileManager) CopyFile(src, dst string) error {
	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Sync to ensure data is written
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

// MoveFile moves a file from source to destination
func (fm *FileManager) MoveFile(src, dst string) error {
	// Try to rename first (fastest if on same filesystem)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// If rename fails, copy and delete
	if err := fm.CopyFile(src, dst); err != nil {
		return fmt.Errorf("failed to copy file during move: %w", err)
	}

	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to remove source file during move: %w", err)
	}

	return nil
}

// DeleteFile deletes a file
func (fm *FileManager) DeleteFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, consider it deleted
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// DeleteActivityFiles deletes all files associated with an activity
func (fm *FileManager) DeleteActivityFiles(activityID string) error {
	activityDir := fm.GetActivityDir(activityID)
	if !fm.FileExists(activityDir) {
		return nil // Directory doesn't exist, nothing to delete
	}

	if err := os.RemoveAll(activityDir); err != nil {
		return fmt.Errorf("failed to delete activity directory: %w", err)
	}

	return nil
}

// ListAudioFiles lists all audio files for an activity
func (fm *FileManager) ListAudioFiles(activityID string) ([]string, error) {
	audioDir := fm.GetActivityAudioDir(activityID)
	
	// Check if directory exists
	if !fm.FileExists(audioDir) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(audioDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio directory: %w", err)
	}

	var audioFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		ext := filepath.Ext(name)
		
		// Check for common audio file extensions
		if isAudioExtension(ext) {
			audioFiles = append(audioFiles, name)
		}
	}

	return audioFiles, nil
}

// isAudioExtension checks if a file extension is for an audio file
func isAudioExtension(ext string) bool {
	audioExts := []string{".wav", ".mp3", ".m4a", ".aac", ".flac", ".ogg"}
	for _, audioExt := range audioExts {
		if ext == audioExt {
			return true
		}
	}
	return false
}

// GetDiskUsage returns disk usage information for the data directory
func (fm *FileManager) GetDiskUsage() (int64, error) {
	var totalSize int64

	err := filepath.Walk(fm.dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to calculate disk usage: %w", err)
	}

	return totalSize, nil
}

// CleanupOldFiles removes files older than the specified duration
func (fm *FileManager) CleanupOldFiles(maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)
	
	err := filepath.Walk(fm.GetActivitiesDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}
		
		if !info.IsDir() && info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				// Log error but continue with cleanup
				fmt.Printf("Warning: failed to remove old file %s: %v\n", path, err)
			}
		}
		
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to cleanup old files: %w", err)
	}

	return nil
}