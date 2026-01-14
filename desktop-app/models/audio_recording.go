package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AudioRecordingStatus represents the status of an audio recording
type AudioRecordingStatus string

const (
	AudioRecordingStatusRecording AudioRecordingStatus = "recording"
	AudioRecordingStatusCompleted AudioRecordingStatus = "completed"
	AudioRecordingStatusFailed    AudioRecordingStatus = "failed"
)

// AudioRecording represents an audio recording linked to an activity
type AudioRecording struct {
	ID             string               `json:"id" db:"id"`
	UserID         string               `json:"user_id" db:"user_id"`
	ActivityID     string               `json:"activity_id" db:"activity_id"`
	FilePath       string               `json:"file_path" db:"file_path"`
	DeviceInfo     AudioDeviceInfo      `json:"device_info" db:"device_info"`
	Status         AudioRecordingStatus `json:"status" db:"status"`
	Duration       *float64             `json:"duration,omitempty" db:"duration"`
	FileSize       *int64               `json:"file_size,omitempty" db:"file_size"`
	Config         RecordingConfig      `json:"config" db:"config"`
	CreatedAt      time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at" db:"updated_at"`
}

// AudioDeviceInfo contains information about the recording device
type AudioDeviceInfo struct {
	Name        string  `json:"name"`
	DeviceID    string  `json:"device_id,omitempty"`
	SampleRate  int     `json:"sample_rate"`
	Channels    int     `json:"channels"`
	BitDepth    int     `json:"bit_depth,omitempty"`
	DeviceType  string  `json:"device_type,omitempty"` // "microphone", "system", "line_in"
}

// RecordingConfig contains the configuration used for recording
type RecordingConfig struct {
	Format         string  `json:"format"`         // "wav", "mp3", "m4a"
	Quality        string  `json:"quality"`        // "low", "medium", "high"
	SampleRate     int     `json:"sample_rate"`
	Bitrate        int     `json:"bitrate,omitempty"`
	AutoGainControl bool   `json:"auto_gain_control"`
	NoiseReduction bool   `json:"noise_reduction"`
	ChunkSize      int     `json:"chunk_size,omitempty"` // For streaming/processing
	RecordingMode  string  `json:"recording_mode"`      // "microphone", "system", "mixed"
}

// NewAudioRecording creates a new audio recording
func NewAudioRecording(userID, activityID, filePath string, deviceInfo AudioDeviceInfo, config RecordingConfig) *AudioRecording {
	now := time.Now()
	return &AudioRecording{
		ID:         uuid.New().String(),
		UserID:     userID,
		ActivityID: activityID,
		FilePath:   filePath,
		DeviceInfo: deviceInfo,
		Status:     AudioRecordingStatusRecording,
		Config:     config,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// StartRecording marks the recording as started
func (ar *AudioRecording) StartRecording() {
	ar.Status = AudioRecordingStatusRecording
	ar.UpdatedAt = time.Now()
}

// Complete marks the recording as completed with file information
func (ar *AudioRecording) Complete(duration float64, fileSize int64) {
	ar.Status = AudioRecordingStatusCompleted
	ar.Duration = &duration
	ar.FileSize = &fileSize
	ar.UpdatedAt = time.Now()
}

// Fail marks the recording as failed
func (ar *AudioRecording) Fail() {
	ar.Status = AudioRecordingStatusFailed
	ar.UpdatedAt = time.Now()
}

// GetDurationMinutes returns the duration in minutes
func (ar *AudioRecording) GetDurationMinutes() float64 {
	if ar.Duration == nil {
		return 0
	}
	return *ar.Duration / 60.0
}

// GetFileSizeMB returns the file size in megabytes
func (ar *AudioRecording) GetFileSizeMB() float64 {
	if ar.FileSize == nil {
		return 0
	}
	return float64(*ar.FileSize) / (1024 * 1024)
}

// IsCompleted returns true if the recording is completed
func (ar *AudioRecording) IsCompleted() bool {
	return ar.Status == AudioRecordingStatusCompleted
}

// DeviceInfoToJSON converts device info to JSON string for database storage
func (ar *AudioRecording) DeviceInfoToJSON() (string, error) {
	data, err := json.Marshal(ar.DeviceInfo)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DeviceInfoFromJSON parses device info from JSON string
func (ar *AudioRecording) DeviceInfoFromJSON(jsonStr string) error {
	if jsonStr == "" {
		ar.DeviceInfo = AudioDeviceInfo{}
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), &ar.DeviceInfo)
}

// ConfigToJSON converts config to JSON string for database storage
func (ar *AudioRecording) ConfigToJSON() (string, error) {
	data, err := json.Marshal(ar.Config)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ConfigFromJSON parses config from JSON string
func (ar *AudioRecording) ConfigFromJSON(jsonStr string) error {
	if jsonStr == "" {
		ar.Config = RecordingConfig{}
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), &ar.Config)
}

// DefaultRecordingConfig returns a default recording configuration
func DefaultRecordingConfig() RecordingConfig {
	return RecordingConfig{
		Format:         "wav",
		Quality:        "high",
		SampleRate:     44100,
		Bitrate:        128000,
		AutoGainControl: true,
		NoiseReduction: true,
		ChunkSize:      4096,
		RecordingMode:  "mixed", // Default to capturing both mic and system audio
	}
}