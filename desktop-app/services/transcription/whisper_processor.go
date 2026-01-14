package transcription

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/platformlabs-co/personal-assist/models"
	"github.com/sirupsen/logrus"
)

// WhisperProcessor handles transcription using Whisper.cpp
type WhisperProcessor struct {
	model   whisper.Model
	context whisper.Context
	config  models.TranscriptionConfig
	logger  *logrus.Logger
}

// NewWhisperProcessor creates a new Whisper processor
func NewWhisperProcessor(modelPath string, config models.TranscriptionConfig, logger *logrus.Logger) (*WhisperProcessor, error) {
	// Initialize whisper model
	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load whisper model from %s: %w", modelPath, err)
	}

	// Create context from model
	context, err := model.NewContext()
	if err != nil {
		model.Close()
		return nil, fmt.Errorf("failed to create whisper context: %w", err)
	}

	wp := &WhisperProcessor{
		model:   model,
		context: context,
		config:  config,
		logger:  logger,
	}

	return wp, nil
}

// NewWhisperProcessorFromModel creates a new Whisper processor from an already loaded model
func NewWhisperProcessorFromModel(model whisper.Model, config models.TranscriptionConfig, logger *logrus.Logger) (*WhisperProcessor, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	logger.Debug("Creating Whisper context from loaded model")

	// Create context from model with error handling
	context, err := model.NewContext()
	if err != nil {
		logger.WithError(err).Error("Failed to create Whisper context - this may indicate model corruption or memory issues")
		return nil, fmt.Errorf("failed to create whisper context: %w", err)
	}

	if context == nil {
		logger.Error("Whisper context is nil after creation")
		return nil, fmt.Errorf("whisper context creation returned nil")
	}

	logger.Debug("Whisper context created successfully")

	wp := &WhisperProcessor{
		model:   model,
		context: context,
		config:  config,
		logger:  logger,
	}

	return wp, nil
}

// TranscribeChunk processes a single audio chunk through Whisper
func (wp *WhisperProcessor) TranscribeChunk(chunk AudioChunk, activityStartTime time.Time) (*models.TranscriptChunk, error) {
	if len(chunk.Samples) == 0 {
		return nil, fmt.Errorf("chunk has no audio data")
	}

	wp.logger.WithFields(logrus.Fields{
		"chunk_index": chunk.ChunkIndex,
		"start_time":  chunk.StartTime,
		"end_time":    chunk.EndTime,
		"duration":    chunk.EndTime - chunk.StartTime,
	}).Debug("Processing audio chunk")

	// Configure context based on our config
	if wp.config.Language != "auto" {
		if err := wp.context.SetLanguage(wp.config.Language); err != nil {
			wp.logger.WithError(err).Warn("Failed to set language, using auto-detect")
		}
	}
	wp.context.SetTemperature(float32(wp.config.Temperature))
	wp.context.SetTokenTimestamps(wp.config.EnableTimestamps)
	
	// Process the audio with whisper
	err := wp.context.Process(chunk.Samples, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("whisper processing failed: %w", err)
	}

	// Extract results
	segments := wp.extractSegments(chunk, activityStartTime)
	
	if len(segments) == 0 {
		// Return empty chunk if no speech detected
		return models.NewTranscriptChunk(
			"", // UserID will be set by caller
			"", // ActivityID will be set by caller  
			"", // AudioRecordingID will be set by caller
			"",
			wp.convertToActivityTime(chunk.StartTime, activityStartTime),
			wp.convertToActivityTime(chunk.EndTime, activityStartTime),
		), nil
	}

	// Combine all segments into one chunk
	var combinedText strings.Builder
	var confidence float64
	var language string
	var speaker string

	for i, segment := range segments {
		if i > 0 {
			combinedText.WriteString(" ")
		}
		combinedText.WriteString(strings.TrimSpace(segment.Text))
		
		// Use confidence from first segment
		if i == 0 {
			confidence = segment.Confidence
			language = segment.Language
			speaker = segment.Speaker
		}
	}

	text := combinedText.String()
	if text == "" {
		return nil, fmt.Errorf("no transcribed text in chunk")
	}

	// Create transcript chunk with activity-relative timestamps
	transcriptChunk := models.NewTranscriptChunkWithDetails(
		"", // UserID will be set by caller
		"", // ActivityID will be set by caller
		"", // AudioRecordingID will be set by caller
		text,
		wp.convertToActivityTime(chunk.StartTime, activityStartTime),
		wp.convertToActivityTime(chunk.EndTime, activityStartTime),
		wp.stringPtr(speaker),
		&confidence,
		wp.stringPtr(language),
	)

	wp.logger.WithFields(logrus.Fields{
		"chunk_index":       chunk.ChunkIndex,
		"transcribed_text":  wp.truncateString(text, 50),
		"confidence":        confidence,
		"language":          language,
		"speaker":           speaker,
	}).Debug("Chunk transcribed successfully")

	return transcriptChunk, nil
}

// ProcessRecording processes an entire audio recording through Whisper
func (wp *WhisperProcessor) ProcessRecording(recording *models.AudioRecording, activity *models.Activity) ([]*models.TranscriptChunk, error) {
	wp.logger.WithFields(logrus.Fields{
		"recording_id": recording.ID,
		"activity_id":  activity.ID,
		"file_path":    recording.FilePath,
	}).Info("Starting recording transcription")

	// Load and preprocess audio file
	audioProcessor := NewAudioProcessor(wp.logger)
	samples, err := audioProcessor.PrepareForWhisper(recording.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare audio: %w", err)
	}

	// Validate audio quality
	if isValid, message := audioProcessor.ValidateAudioQuality(samples); !isValid {
		return nil, fmt.Errorf("audio quality validation failed: %s", message)
	}

	// Split into chunks
	chunkDuration := time.Duration(wp.config.ChunkDuration) * time.Second
	overlapDuration := time.Duration(wp.config.OverlapDuration) * time.Second
	chunks := audioProcessor.ChunkAudio(samples, chunkDuration, overlapDuration, activity.StartTime, recording.FilePath)

	wp.logger.WithField("chunk_count", len(chunks)).Info("Audio split into chunks")

	// Process each chunk
	var allChunks []*models.TranscriptChunk
	for i, chunk := range chunks {
		transcriptChunk, err := wp.TranscribeChunk(chunk, activity.StartTime)
		if err != nil {
			wp.logger.WithError(err).WithField("chunk_index", i).Warn("Failed to transcribe chunk, skipping")
			continue
		}

		// Set required IDs
		transcriptChunk.ActivityID = activity.ID
		transcriptChunk.AudioRecordingID = recording.ID
		transcriptChunk.UserID = activity.UserID

		allChunks = append(allChunks, transcriptChunk)
	}

	wp.logger.WithFields(logrus.Fields{
		"recording_id":      recording.ID,
		"chunks_processed":  len(chunks),
		"chunks_transcribed": len(allChunks),
	}).Info("Recording transcription completed")

	return allChunks, nil
}

// extractSegments extracts transcript segments from the Whisper context using NextSegment API
func (wp *WhisperProcessor) extractSegments(chunk AudioChunk, activityStartTime time.Time) []TranscriptSegment {
	var segments []TranscriptSegment

	// Use NextSegment to retrieve all segments
	for {
		segment, err := wp.context.NextSegment()
		if err != nil {
			if err == io.EOF {
				break // End of segments
			}
			wp.logger.WithError(err).Warn("Error reading segment")
			continue
		}

		// Skip empty segments
		text := strings.TrimSpace(segment.Text)
		if text == "" {
			continue
		}

		// Convert segment times to our format
		startTime := chunk.StartTime + segment.Start.Seconds()
		endTime := chunk.StartTime + segment.End.Seconds()

		// Create transcript segment
		transcriptSegment := TranscriptSegment{
			Text:       text,
			StartTime:  startTime,
			EndTime:    endTime,
			Confidence: 0.8, // Default confidence
			Language:   wp.detectLanguage(text),
			Speaker:    wp.detectSpeaker(segment.Num, text),
		}

		segments = append(segments, transcriptSegment)
	}

	return segments
}

// TranscriptSegment represents a segment of transcribed text
type TranscriptSegment struct {
	Text       string  `json:"text"`
	StartTime  float64 `json:"start_time"`
	EndTime    float64 `json:"end_time"`
	Confidence float64 `json:"confidence"`
	Language   string  `json:"language"`
	Speaker    string  `json:"speaker"`
}

// convertToActivityTime converts absolute timestamp to activity-relative time
func (wp *WhisperProcessor) convertToActivityTime(absoluteTime float64, activityStartTime time.Time) float64 {
	// For now, assume absoluteTime is already relative to the recording start
	// In a full implementation, this would account for the recording start time relative to activity start
	return absoluteTime
}

// detectLanguage attempts to detect the language of transcribed text
func (wp *WhisperProcessor) detectLanguage(text string) string {
	if wp.config.Language != "auto" {
		return wp.config.Language
	}
	
	// Get the detected language from context
	detectedLang := wp.context.DetectedLanguage()
	if detectedLang != "" {
		return detectedLang
	}
	
	return "en" // Default fallback
}

// detectSpeaker attempts to identify the speaker for a segment
func (wp *WhisperProcessor) detectSpeaker(segmentIndex int, text string) string {
	if !wp.config.EnableSpeakers {
		return "Unknown"
	}
	
	// This is a simplified speaker detection
	// In practice, you'd use more sophisticated speaker diarization
	// The whisper.cpp bindings may provide speaker information in newer versions
	
	return fmt.Sprintf("Speaker_%d", segmentIndex%2+1) // Simple alternating speakers
}

// stringPtr returns a pointer to a string, handling empty strings
func (wp *WhisperProcessor) stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// truncateString truncates a string to a maximum length
func (wp *WhisperProcessor) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Close cleans up the Whisper processor
func (wp *WhisperProcessor) Close() error {
	// Note: We don't close the model here because it's owned by the ModelManager
	// We only close the context if it exists
	wp.logger.Debug("Closing Whisper processor")
	return nil
}