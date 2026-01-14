package services

import (
	"fmt"

	"github.com/platformlabs-co/personal-assist/logger"
	"github.com/platformlabs-co/personal-assist/models"
	"github.com/platformlabs-co/personal-assist/storage"
)

// AudioService handles audio recording operations
type AudioService struct {
	storage       *storage.SQLiteStorage
	fileManager   *storage.FileManager
	AudioRecorder *AudioRecorder // Made public for access from app.go
}

// NewAudioService creates a new audio service
func NewAudioService(storage *storage.SQLiteStorage, fileManager *storage.FileManager) *AudioService {
	return &AudioService{
		storage:       storage,
		fileManager:   fileManager,
		AudioRecorder: NewAudioRecorder(),
	}
}

// CreateAudioRecording creates a new audio recording record
func (s *AudioService) CreateAudioRecording(
	userID, activityID string,
	deviceInfo models.AudioDeviceInfo,
	config models.RecordingConfig,
) (*models.AudioRecording, error) {
	logger.WithFields(map[string]interface{}{
		"user_id":     userID,
		"activity_id": activityID,
		"device_name": deviceInfo.Name,
		"device_id":   deviceInfo.DeviceID,
		"format":      config.Format,
	}).Info("AudioService CreateAudioRecording called")

	// Generate unique filename
	fileName := s.fileManager.GenerateAudioFileName(activityID, config.Format)
	relativePath := s.fileManager.GetRelativeAudioFilePath(activityID, fileName)
	absolutePath := s.fileManager.GetAbsolutePathFromRelative(relativePath)

	logger.WithFields(map[string]interface{}{
		"activity_id":     activityID,
		"file_name":       fileName,
		"relative_path":   relativePath,
		"absolute_path":   absolutePath,
	}).Info("Generated audio recording file paths")

	// Create audio recording model
	recording := models.NewAudioRecording(userID, activityID, relativePath, deviceInfo, config)

	logger.WithFields(map[string]interface{}{
		"recording_id": recording.ID,
		"user_id":      userID,
		"activity_id":  activityID,
		"file_path":    relativePath,
		"status":       recording.Status,
	}).Info("Created audio recording model")

	// Create database record
	logger.WithField("recording_id", recording.ID).Info("Creating audio recording database record")
	if err := s.storage.CreateAudioRecording(recording); err != nil {
		logger.WithError(err).WithField("recording_id", recording.ID).Error("Failed to create audio recording in database")
		return nil, fmt.Errorf("failed to create audio recording in database: %w", err)
	}

	// Start actual audio recording
	logger.WithField("recording_id", recording.ID).Info("Starting actual audio recording")
	if err := s.AudioRecorder.StartRecording(recording.ID, absolutePath, deviceInfo, config); err != nil {
		logger.WithError(err).WithField("recording_id", recording.ID).Error("Failed to start audio recording")
		// Clean up database record if recording fails to start
		s.storage.DeleteAudioRecording(recording.ID)
		return nil, fmt.Errorf("failed to start audio recording: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"recording_id": recording.ID,
		"activity_id":  activityID,
		"file_path":    relativePath,
	}).Info("AudioService CreateAudioRecording completed successfully")

	return recording, nil
}

// GetAudioFilePath returns the absolute path for an audio recording
func (s *AudioService) GetAudioFilePath(recording *models.AudioRecording) string {
	return s.fileManager.GetAbsolutePathFromRelative(recording.FilePath)
}

// CompleteAudioRecording marks an audio recording as completed
func (s *AudioService) CompleteAudioRecording(userID, recordingID string) (*models.AudioRecording, error) {
	logger.WithField("recording_id", recordingID).Info("AudioService CompleteAudioRecording called")

	// Stop the actual audio recording first
	logger.WithField("recording_id", recordingID).Info("Stopping actual audio recording")
	if err := s.AudioRecorder.StopRecording(recordingID); err != nil {
		logger.WithError(err).WithField("recording_id", recordingID).Error("Failed to stop audio recording")
		return nil, fmt.Errorf("failed to stop audio recording: %w", err)
	}

	// Get the recording from database
	logger.WithField("recording_id", recordingID).Info("Getting audio recording from database")
	recording, err := s.storage.GetAudioRecording(userID, recordingID)
	if err != nil {
		logger.WithError(err).WithField("recording_id", recordingID).Error("Failed to get audio recording from database")
		return nil, fmt.Errorf("failed to get audio recording: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"recording_id": recordingID,
		"activity_id":  recording.ActivityID,
		"user_id":      recording.UserID,
		"file_path":    recording.FilePath,
		"status":       recording.Status,
	}).Info("Retrieved audio recording from database")

	// Get file size and duration
	filePath := s.GetAudioFilePath(recording)
	logger.WithFields(map[string]interface{}{
		"recording_id":       recordingID,
		"absolute_file_path": filePath,
		"relative_file_path": recording.FilePath,
	}).Info("Getting audio file size")

	// Check if file exists before getting size
	if !s.fileManager.FileExists(filePath) {
		logger.WithFields(map[string]interface{}{
			"recording_id": recordingID,
			"file_path":    filePath,
		}).Warn("Audio file does not exist, creating placeholder file with zero size")
		
		// For now, handle missing file gracefully since we might not have actual recording yet
		fileSize := int64(0)
		duration := 0.0
		
		recording.Complete(duration, fileSize)
		logger.WithFields(map[string]interface{}{
			"recording_id": recordingID,
			"duration":     duration,
			"file_size":    fileSize,
		}).Info("Completed audio recording with placeholder values (file not found)")
	} else {
		fileSize, err := s.fileManager.GetFileSize(filePath)
		if err != nil {
			logger.WithError(err).WithFields(map[string]interface{}{
				"recording_id": recordingID,
				"file_path":    filePath,
			}).Error("Failed to get audio file size")
			return nil, fmt.Errorf("failed to get file size: %w", err)
		}

		logger.WithFields(map[string]interface{}{
			"recording_id": recordingID,
			"file_size":    fileSize,
		}).Info("Retrieved audio file size")

		// TODO: Calculate duration from audio file
		// For now, estimate based on file size (rough approximation)
		duration := s.estimateAudioDuration(fileSize, recording.Config)
		logger.WithFields(map[string]interface{}{
			"recording_id":     recordingID,
			"file_size":        fileSize,
			"estimated_duration": duration,
			"format":           recording.Config.Format,
		}).Info("Estimated audio duration from file size")

		recording.Complete(duration, fileSize)
	}

	// Update database record
	logger.WithField("recording_id", recordingID).Info("Updating audio recording in database")
	if err := s.storage.UpdateAudioRecording(recording); err != nil {
		logger.WithError(err).WithField("recording_id", recordingID).Error("Failed to update audio recording in database")
		return nil, fmt.Errorf("failed to update audio recording: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"recording_id": recordingID,
		"status":       recording.Status,
		"duration":     recording.Duration,
		"file_size":    recording.FileSize,
	}).Info("AudioService CompleteAudioRecording completed successfully")

	return recording, nil
}

// FailAudioRecording marks an audio recording as failed
func (s *AudioService) FailAudioRecording(userID, recordingID string) (*models.AudioRecording, error) {
	recording, err := s.storage.GetAudioRecording(userID, recordingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio recording: %w", err)
	}

	recording.Fail()

	if err := s.storage.UpdateAudioRecording(recording); err != nil {
		return nil, fmt.Errorf("failed to update audio recording: %w", err)
	}

	return recording, nil
}

// DeleteAudioRecording deletes an audio recording and its file
func (s *AudioService) DeleteAudioRecording(userID, recordingID string) error {
	// Get recording to get file path
	recording, err := s.storage.GetAudioRecording(userID, recordingID)
	if err != nil {
		return fmt.Errorf("failed to get audio recording: %w", err)
	}

	// Delete file
	filePath := s.GetAudioFilePath(recording)
	if err := s.fileManager.DeleteFile(filePath); err != nil {
		return fmt.Errorf("failed to delete audio file: %w", err)
	}

	// Delete from database
	if err := s.storage.DeleteAudioRecording(recordingID); err != nil {
		return fmt.Errorf("failed to delete audio recording from database: %w", err)
	}

	return nil
}

// GetAudioRecordingsByActivity returns all audio recordings for an activity
func (s *AudioService) GetAudioRecordingsByActivity(userID, activityID string) ([]*models.AudioRecording, error) {
	return s.storage.GetActivityRecordings(userID, activityID)
}

// ValidateAudioDevice checks if an audio device is available
func (s *AudioService) ValidateAudioDevice(deviceInfo models.AudioDeviceInfo) error {
	// TODO: Implement device validation
	// This would typically involve:
	// 1. Enumerating available audio devices
	// 2. Checking if the specified device exists and is accessible
	// 3. Verifying sample rate and other capabilities
	return nil
}

// GetDefaultRecordingConfig returns a default recording configuration
func (s *AudioService) GetDefaultRecordingConfig() models.RecordingConfig {
	return models.DefaultRecordingConfig()
}

// GetSupportedFormats returns supported audio formats
func (s *AudioService) GetSupportedFormats() []string {
	return []string{"wav", "mp3", "m4a", "aac"}
}

// GetSupportedSampleRates returns supported sample rates
func (s *AudioService) GetSupportedSampleRates() []int {
	return []int{8000, 16000, 22050, 44100, 48000}
}

// EstimateFileSize estimates the file size for a recording duration
func (s *AudioService) EstimateFileSize(config models.RecordingConfig, durationMinutes float64) int64 {
	// Rough estimation based on format and quality
	var bytesPerMinute int64

	switch config.Format {
	case "wav":
		// Uncompressed: sample_rate * channels * bit_depth/8 * 60
		bytesPerMinute = int64(config.SampleRate * 2 * 2 * 60) // Assuming 16-bit stereo
	case "mp3":
		// Compressed: roughly based on bitrate
		if config.Bitrate > 0 {
			bytesPerMinute = int64(config.Bitrate/8) * 60
		} else {
			bytesPerMinute = 16000 * 60 // Default ~128kbps
		}
	case "m4a", "aac":
		// Similar to MP3 but slightly more efficient
		if config.Bitrate > 0 {
			bytesPerMinute = int64(config.Bitrate/8) * 60
		} else {
			bytesPerMinute = 12000 * 60 // Default ~96kbps
		}
	default:
		bytesPerMinute = 16000 * 60 // Conservative estimate
	}

	return int64(float64(bytesPerMinute) * durationMinutes)
}

// calculateAudioDuration calculates the duration of an audio file
// This would require an audio library to implement properly
func (s *AudioService) calculateAudioDuration(filePath string) float64 {
	// TODO: Implement using audio library like github.com/hajimehoshi/oto
	// or call external tool like ffprobe
	return 0.0
}

// estimateAudioDuration estimates duration based on file size and format
func (s *AudioService) estimateAudioDuration(fileSize int64, config models.RecordingConfig) float64 {
	// Rough estimation based on format
	var bytesPerSecond int64

	switch config.Format {
	case "wav":
		// Uncompressed: sample_rate * channels * bit_depth/8
		bytesPerSecond = int64(config.SampleRate * 2 * 2) // Assuming 16-bit stereo
	case "mp3":
		if config.Bitrate > 0 {
			bytesPerSecond = int64(config.Bitrate / 8)
		} else {
			bytesPerSecond = 16000 // Default ~128kbps
		}
	case "m4a", "aac":
		if config.Bitrate > 0 {
			bytesPerSecond = int64(config.Bitrate / 8)
		} else {
			bytesPerSecond = 12000 // Default ~96kbps
		}
	default:
		bytesPerSecond = 16000 // Conservative estimate
	}

	if bytesPerSecond == 0 {
		return 0.0
	}

	return float64(fileSize) / float64(bytesPerSecond)
}

// GetAudioFileInfo returns information about an audio file
func (s *AudioService) GetAudioFileInfo(filePath string) (map[string]interface{}, error) {
	// TODO: Implement using audio library
	// Should return info like:
	// - Duration
	// - Sample rate
	// - Channels
	// - Bit depth
	// - Format
	// - File size
	return nil, fmt.Errorf("not implemented yet")
}
