package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TranscriptChunk represents a segment of transcribed audio
type TranscriptChunk struct {
	ID                string    `json:"id" db:"id"`
	UserID            string    `json:"user_id" db:"user_id"`
	ActivityID        string    `json:"activity_id" db:"activity_id"`
	AudioRecordingID  string    `json:"audio_recording_id" db:"audio_recording_id"`
	Text              string    `json:"text" db:"text"`
	StartTime         float64   `json:"start_time" db:"start_time"`     // Seconds from activity start
	EndTime           float64   `json:"end_time" db:"end_time"`         // Seconds from activity start
	Speaker           *string   `json:"speaker,omitempty" db:"speaker"` // Optional speaker identification
	Confidence        *float64  `json:"confidence,omitempty" db:"confidence"` // 0-1 confidence score
	Language          *string   `json:"language,omitempty" db:"language"`     // Language code (e.g., "en", "es")
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// NewTranscriptChunk creates a new transcript chunk
func NewTranscriptChunk(userID, activityID, audioRecordingID, text string, startTime, endTime float64) *TranscriptChunk {
	return &TranscriptChunk{
		ID:               uuid.New().String(),
		UserID:           userID,
		ActivityID:       activityID,
		AudioRecordingID: audioRecordingID,
		Text:             text,
		StartTime:        startTime,
		EndTime:          endTime,
		CreatedAt:        time.Now(),
	}
}

// NewTranscriptChunkWithDetails creates a transcript chunk with optional details
func NewTranscriptChunkWithDetails(
	userID, activityID, audioRecordingID, text string,
	startTime, endTime float64,
	speaker *string,
	confidence *float64,
	language *string,
) *TranscriptChunk {
	return &TranscriptChunk{
		ID:               uuid.New().String(),
		UserID:           userID,
		ActivityID:       activityID,
		AudioRecordingID: audioRecordingID,
		Text:             text,
		StartTime:        startTime,
		EndTime:          endTime,
		Speaker:          speaker,
		Confidence:       confidence,
		Language:         language,
		CreatedAt:        time.Now(),
	}
}

// Duration returns the duration of this chunk in seconds
func (tc *TranscriptChunk) Duration() float64 {
	return tc.EndTime - tc.StartTime
}

// WordCount returns an approximate word count
func (tc *TranscriptChunk) WordCount() int {
	if tc.Text == "" {
		return 0
	}
	// Simple word count - split by whitespace
	words := 0
	inWord := false
	for _, char := range tc.Text {
		if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
			inWord = false
		} else if !inWord {
			words++
			inWord = true
		}
	}
	return words
}

// GetSpeaker returns the speaker name or "Unknown" if not set
func (tc *TranscriptChunk) GetSpeaker() string {
	if tc.Speaker != nil {
		return *tc.Speaker
	}
	return "Unknown"
}

// GetConfidence returns the confidence score or 0 if not set
func (tc *TranscriptChunk) GetConfidence() float64 {
	if tc.Confidence != nil {
		return *tc.Confidence
	}
	return 0.0
}

// GetLanguage returns the language code or "unknown" if not set
func (tc *TranscriptChunk) GetLanguage() string {
	if tc.Language != nil {
		return *tc.Language
	}
	return "unknown"
}

// SetSpeaker sets the speaker identification
func (tc *TranscriptChunk) SetSpeaker(speaker string) {
	tc.Speaker = &speaker
}

// SetConfidence sets the confidence score
func (tc *TranscriptChunk) SetConfidence(confidence float64) {
	tc.Confidence = &confidence
}

// SetLanguage sets the language code
func (tc *TranscriptChunk) SetLanguage(language string) {
	tc.Language = &language
}

// FormatTimestamp formats a timestamp for display (MM:SS or HH:MM:SS)
func (tc *TranscriptChunk) FormatTimestamp(timestamp float64) string {
	totalSeconds := int(timestamp)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// FormatStartTime returns formatted start time
func (tc *TranscriptChunk) FormatStartTime() string {
	return tc.FormatTimestamp(tc.StartTime)
}

// FormatEndTime returns formatted end time  
func (tc *TranscriptChunk) FormatEndTime() string {
	return tc.FormatTimestamp(tc.EndTime)
}

// FormatTimeRange returns formatted time range (start - end)
func (tc *TranscriptChunk) FormatTimeRange() string {
	return fmt.Sprintf("%s - %s", tc.FormatStartTime(), tc.FormatEndTime())
}