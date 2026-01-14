package models

import "time"

// TranscriptionConfig contains configuration for Whisper transcription
type TranscriptionConfig struct {
	Model              string   `json:"model"`                          // "tiny", "small", "medium", "large"
	Language           string   `json:"language"`                       // "auto", "en", "es", etc.
	Temperature        float64  `json:"temperature"`                    // 0.0-1.0
	EnableTimestamps   bool     `json:"enable_timestamps"`
	EnableSpeakers     bool     `json:"enable_speakers"`
	ChunkDuration      int      `json:"chunk_duration"`                 // seconds
	OverlapDuration    int      `json:"overlap_duration"`               // seconds
	NoSpeechThreshold  float64  `json:"no_speech_threshold"`
	CustomVocabulary   []string `json:"custom_vocabulary,omitempty"`
}

// DefaultTranscriptionConfig returns a sensible default configuration
func DefaultTranscriptionConfig() TranscriptionConfig {
	return TranscriptionConfig{
		Model:             "small",
		Language:          "auto",
		Temperature:       0.0,
		EnableTimestamps:  true,
		EnableSpeakers:    false,
		ChunkDuration:     30,
		OverlapDuration:   5,
		NoSpeechThreshold: 0.6,
		CustomVocabulary:  make([]string, 0),
	}
}

// TranscriptionStatus represents the current status of transcription processing
type TranscriptionStatus struct {
	Stage           string        `json:"stage"`              // "queued", "processing", "completed", "failed"
	Progress        float64       `json:"progress"`           // 0.0-1.0
	ProcessedChunks int           `json:"processed_chunks"`
	TotalChunks     int           `json:"total_chunks"`
	CurrentFile     string        `json:"current_file"`
	EstimatedTime   time.Duration `json:"estimated_time"`
	StartedAt       time.Time     `json:"started_at"`
	LastError       string        `json:"last_error,omitempty"`
}

// NewTranscriptionStatus creates a new transcription status
func NewTranscriptionStatus(totalChunks int, currentFile string) *TranscriptionStatus {
	return &TranscriptionStatus{
		Stage:           "queued",
		Progress:        0.0,
		ProcessedChunks: 0,
		TotalChunks:     totalChunks,
		CurrentFile:     currentFile,
		EstimatedTime:   0,
		StartedAt:       time.Now(),
		LastError:       "",
	}
}

// Start marks the transcription as started
func (ts *TranscriptionStatus) Start() {
	ts.Stage = "processing"
	ts.StartedAt = time.Now()
}

// UpdateProgress updates the transcription progress
func (ts *TranscriptionStatus) UpdateProgress(processedChunks int) {
	ts.ProcessedChunks = processedChunks
	if ts.TotalChunks > 0 {
		ts.Progress = float64(processedChunks) / float64(ts.TotalChunks)
	}
}

// Complete marks the transcription as completed
func (ts *TranscriptionStatus) Complete() {
	ts.Stage = "completed"
	ts.Progress = 1.0
	ts.ProcessedChunks = ts.TotalChunks
}

// Fail marks the transcription as failed with an error message
func (ts *TranscriptionStatus) Fail(err string) {
	ts.Stage = "failed"
	ts.LastError = err
}

// EstimateTimeRemaining calculates estimated time remaining based on current progress
func (ts *TranscriptionStatus) EstimateTimeRemaining() time.Duration {
	if ts.Progress <= 0 {
		return 0
	}
	
	elapsed := time.Since(ts.StartedAt)
	totalEstimated := time.Duration(float64(elapsed) / ts.Progress)
	remaining := totalEstimated - elapsed
	
	if remaining < 0 {
		remaining = 0
	}
	
	return remaining
}