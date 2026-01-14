package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/platformlabs-co/personal-assist/models"
	"github.com/platformlabs-co/personal-assist/services/transcription"
	"github.com/platformlabs-co/personal-assist/storage"
	"github.com/sirupsen/logrus"
)

// TranscriptionService handles all transcription operations with activity integration
type TranscriptionService struct {
	dataDir        string
	modelsPath     string
	storage        *storage.SQLiteStorage
	logger         *logrus.Logger
	modelManager   *transcription.ModelManager
	processingJobs map[string]*TranscriptionJob
	jobMutex       sync.RWMutex
}

// TranscriptionJob represents an ongoing transcription operation
type TranscriptionJob struct {
	ActivityID    string
	Status        string
	Progress      float64
	CurrentFile   string
	StartTime     time.Time
	Error         error
}

// SearchFilter for transcript search (placeholder for future expansion)
type SearchFilter struct {
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
}

// NewTranscriptionService creates a new transcription service
func NewTranscriptionService(storage *storage.SQLiteStorage, dataDir, modelsPath string, logger *logrus.Logger) *TranscriptionService {
	// Initialize model manager
	modelManager := transcription.NewModelManager(modelsPath, logger)

	return &TranscriptionService{
		dataDir:        dataDir,
		modelsPath:     modelsPath,
		storage:        storage,
		logger:         logger,
		modelManager:   modelManager,
		processingJobs: make(map[string]*TranscriptionJob),
	}
}

// ProcessActivity transcribes all audio recordings in an activity
func (ts *TranscriptionService) ProcessActivity(userID, activityID string, config models.TranscriptionConfig) error {
	ts.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"activity_id": activityID,
	}).Info("Starting activity transcription")

	// Get recordings for the activity
	recordings, err := ts.storage.GetActivityRecordings(userID, activityID)
	if err != nil {
		ts.logger.WithError(err).Error("Failed to get activity recordings")
		return fmt.Errorf("failed to get activity recordings: %w", err)
	}

	if len(recordings) == 0 {
		ts.logger.WithField("activity_id", activityID).Info("No recordings found for activity")
		return fmt.Errorf("no recordings found for activity %s", activityID)
	}

	// Start transcription job
	job := &TranscriptionJob{
		ActivityID:  activityID,
		Status:      "processing",
		Progress:    0.0,
		CurrentFile: "",
		StartTime:   time.Now(),
	}

	ts.jobMutex.Lock()
	ts.processingJobs[activityID] = job
	ts.jobMutex.Unlock()

	// Process each recording (simplified implementation)
	go ts.processActivityAsync(userID, activityID, recordings, job)

	return nil
}

// ProcessRecording transcribes a specific audio recording
func (ts *TranscriptionService) ProcessRecording(userID, activityID, recordingID string, config models.TranscriptionConfig) error {
	ts.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"activity_id":  activityID,
		"recording_id": recordingID,
	}).Info("Starting recording transcription")

	// Get the specific recording
	recording, err := ts.storage.GetAudioRecording(userID, recordingID)
	if err != nil {
		ts.logger.WithError(err).Error("Failed to get audio recording")
		return fmt.Errorf("failed to get audio recording: %w", err)
	}

	// Start transcription job
	job := &TranscriptionJob{
		ActivityID:  activityID,
		Status:      "processing",
		Progress:    0.0,
		CurrentFile: recording.FilePath,
		StartTime:   time.Now(),
	}

	ts.jobMutex.Lock()
	ts.processingJobs[activityID] = job
	ts.jobMutex.Unlock()

	// Process the recording (simplified implementation)
	go ts.processRecordingAsync(userID, activityID, recording, job)

	return nil
}

// GetTranscript retrieves transcript chunks for an activity
func (ts *TranscriptionService) GetTranscript(userID, activityID string) ([]*models.TranscriptChunk, error) {
	return ts.storage.GetActivityTranscripts(userID, activityID)
}

// GetRecordingTranscript retrieves transcript chunks for a specific recording
func (ts *TranscriptionService) GetRecordingTranscript(userID, recordingID string) ([]*models.TranscriptChunk, error) {
	return ts.storage.GetRecordingTranscripts(userID, recordingID)
}

// SearchTranscripts searches through transcript content
func (ts *TranscriptionService) SearchTranscripts(userID string, query string, filter SearchFilter) ([]*models.TranscriptChunk, error) {
	// For now, implement a simple search
	// In a production system, you'd use FTS5 or similar
	return ts.storage.SearchTranscripts(userID, query)
}

// GetTranscriptionStatus returns the status of an ongoing transcription
func (ts *TranscriptionService) GetTranscriptionStatus(userID, activityID string) (*models.TranscriptionStatus, error) {
	ts.jobMutex.RLock()
	defer ts.jobMutex.RUnlock()

	job, exists := ts.processingJobs[activityID]
	if !exists {
		// Check if we have completed transcripts
		chunks, err := ts.storage.GetActivityTranscripts(userID, activityID)
		if err == nil && len(chunks) > 0 {
			return &models.TranscriptionStatus{
				Stage:        "completed",
				Progress:     1.0,
				CurrentFile:  "",
				LastError:    "",
				StartedAt:    time.Now().Add(-5 * time.Minute), // Placeholder
			}, nil
		}

		return &models.TranscriptionStatus{
			Stage:      "not_started",
			Progress:   0.0,
		}, nil
	}

	status := &models.TranscriptionStatus{
		Stage:       job.Status,
		Progress:    job.Progress,
		CurrentFile: job.CurrentFile,
		StartedAt:   job.StartTime,
	}

	if job.Error != nil {
		status.LastError = job.Error.Error()
	}

	return status, nil
}

// GetAvailableModels returns available Whisper models (placeholder)
func (ts *TranscriptionService) GetAvailableModels() ([]models.WhisperModel, error) {
	// Placeholder implementation
	models := []models.WhisperModel{
		{
			ID:           "whisper-tiny",
			Name:         "Whisper Tiny",
			Size:         39 * 1024 * 1024, // 39 MB in bytes
			IsDownloaded: true,
			IsActive:     true,
			Languages:    []string{"en"},
			Accuracy:     "good",
			Speed:        "fast",
		},
		{
			ID:           "whisper-small",
			Name:         "Whisper Small",
			Size:         244 * 1024 * 1024, // 244 MB in bytes
			IsDownloaded: false,
			IsActive:     false,
			Languages:    []string{"en"},
			Accuracy:     "better",
			Speed:        "medium",
		},
	}
	return models, nil
}

// DownloadModel downloads a Whisper model (placeholder)
func (ts *TranscriptionService) DownloadModel(modelID string) error {
	ts.logger.WithField("model_id", modelID).Info("Model download requested (placeholder implementation)")
	return nil
}

// SetActiveModel sets the active Whisper model (placeholder)
func (ts *TranscriptionService) SetActiveModel(modelID string) error {
	ts.logger.WithField("model_id", modelID).Info("Setting active model (placeholder implementation)")
	return nil
}

// GetActiveModel returns the currently active model (placeholder)
func (ts *TranscriptionService) GetActiveModel() (*models.WhisperModel, error) {
	return &models.WhisperModel{
		ID:           "whisper-tiny",
		Name:         "Whisper Tiny",
		Size:         39 * 1024 * 1024, // 39 MB in bytes
		IsDownloaded: true,
		IsActive:     true,
		Languages:    []string{"en"},
		Accuracy:     "good",
		Speed:        "fast",
	}, nil
}

// processActivityAsync processes all recordings in an activity asynchronously using Whisper
func (ts *TranscriptionService) processActivityAsync(userID, activityID string, recordings []*models.AudioRecording, job *TranscriptionJob) {
	defer func() {
		ts.jobMutex.Lock()
		if job.Error == nil {
			job.Status = "completed"
			job.Progress = 1.0
		} else {
			job.Status = "failed"
		}
		ts.jobMutex.Unlock()
	}()

	// Get activity details
	activity, err := ts.storage.GetActivity(userID, activityID)
	if err != nil {
		ts.logger.WithError(err).Error("Failed to get activity for transcription")
		ts.jobMutex.Lock()
		job.Error = fmt.Errorf("failed to get activity: %w", err)
		ts.jobMutex.Unlock()
		return
	}

	// Get transcription config
	config := models.DefaultTranscriptionConfig()

	// Ensure we have a model available (downloads small if needed)
	ts.logger.Info("Ensuring Whisper model is available")
	err = ts.modelManager.EnsureDefaultModel()
	if err != nil {
		ts.logger.WithError(err).Error("Failed to ensure model availability")
		ts.jobMutex.Lock()
		job.Error = fmt.Errorf("model download in progress, please try again in a moment")
		ts.jobMutex.Unlock()
		return
	}

	// Get the loaded model from the manager
	loadedModel := ts.modelManager.GetLoadedModel()
	if loadedModel == nil {
		ts.logger.Error("No model loaded after ensuring default")
		ts.jobMutex.Lock()
		job.Error = fmt.Errorf("failed to load whisper model")
		ts.jobMutex.Unlock()
		return
	}

	ts.logger.Info("Creating Whisper processor with loaded model")

	// Create Whisper processor using the loaded model
	processor, err := transcription.NewWhisperProcessorFromModel(loadedModel, config, ts.logger)
	if err != nil {
		ts.logger.WithError(err).Error("Failed to create Whisper processor")
		ts.jobMutex.Lock()
		job.Error = fmt.Errorf("failed to load whisper model: %w", err)
		ts.jobMutex.Unlock()
		return
	}
	defer processor.Close()

	// Process each recording
	for i, recording := range recordings {
		// Construct full file path
		fullPath := ts.dataDir + "/" + recording.FilePath

		ts.logger.WithFields(logrus.Fields{
			"recording_id": recording.ID,
			"file_path":    fullPath,
		}).Info("Processing recording with Whisper")

		// Update progress
		progress := float64(i) / float64(len(recordings))
		ts.jobMutex.Lock()
		job.Progress = progress
		job.CurrentFile = fullPath
		ts.jobMutex.Unlock()

		// Update recording with full path for processing
		recordingWithFullPath := *recording
		recordingWithFullPath.FilePath = fullPath

		// Process the recording with Whisper
		chunks, err := processor.ProcessRecording(&recordingWithFullPath, activity)
		if err != nil {
			ts.logger.WithError(err).Error("Failed to process recording with Whisper")
			ts.jobMutex.Lock()
			job.Error = fmt.Errorf("failed to process recording: %w", err)
			ts.jobMutex.Unlock()
			return
		}

		// Store transcript chunks
		for _, chunk := range chunks {
			err := ts.storage.SaveTranscriptChunk(chunk)
			if err != nil {
				ts.logger.WithError(err).Error("Failed to save transcript chunk")
				ts.jobMutex.Lock()
				job.Error = fmt.Errorf("failed to save transcript chunk: %w", err)
				ts.jobMutex.Unlock()
				return
			}
		}

		ts.logger.WithFields(logrus.Fields{
			"recording_id": recording.ID,
			"chunk_count":  len(chunks),
		}).Info("Recording transcribed successfully")
	}
}

// processRecordingAsync processes a single recording asynchronously using Whisper
func (ts *TranscriptionService) processRecordingAsync(userID, activityID string, recording *models.AudioRecording, job *TranscriptionJob) {
	defer func() {
		ts.jobMutex.Lock()
		if job.Error == nil {
			job.Status = "completed"
			job.Progress = 1.0
		} else {
			job.Status = "failed"
		}
		ts.jobMutex.Unlock()
	}()

	// Construct full file path
	fullPath := ts.dataDir + "/" + recording.FilePath

	ts.logger.WithFields(logrus.Fields{
		"recording_id": recording.ID,
		"file_path":    fullPath,
	}).Info("Processing single recording with Whisper")

	// Update recording with full path for processing
	recordingWithFullPath := *recording
	recordingWithFullPath.FilePath = fullPath

	// Get activity details
	activity, err := ts.storage.GetActivity(userID, activityID)
	if err != nil {
		ts.logger.WithError(err).Error("Failed to get activity for transcription")
		ts.jobMutex.Lock()
		job.Error = fmt.Errorf("failed to get activity: %w", err)
		ts.jobMutex.Unlock()
		return
	}

	// Get transcription config
	config := models.DefaultTranscriptionConfig()

	// Ensure we have a model available (downloads small if needed)
	ts.logger.Info("Ensuring Whisper model is available")
	err = ts.modelManager.EnsureDefaultModel()
	if err != nil {
		ts.logger.WithError(err).Error("Failed to ensure model availability")
		ts.jobMutex.Lock()
		job.Error = fmt.Errorf("model download in progress, please try again in a moment")
		ts.jobMutex.Unlock()
		return
	}

	// Get the loaded model from the manager
	loadedModel := ts.modelManager.GetLoadedModel()
	if loadedModel == nil {
		ts.logger.Error("No model loaded after ensuring default")
		ts.jobMutex.Lock()
		job.Error = fmt.Errorf("failed to load whisper model")
		ts.jobMutex.Unlock()
		return
	}

	ts.logger.Info("Creating Whisper processor with loaded model")

	// Create Whisper processor using the loaded model
	processor, err := transcription.NewWhisperProcessorFromModel(loadedModel, config, ts.logger)
	if err != nil {
		ts.logger.WithError(err).Error("Failed to create Whisper processor")
		ts.jobMutex.Lock()
		job.Error = fmt.Errorf("failed to load whisper model: %w", err)
		ts.jobMutex.Unlock()
		return
	}
	defer processor.Close()

	// Process the recording
	chunks, err := processor.ProcessRecording(&recordingWithFullPath, activity)
	if err != nil {
		ts.logger.WithError(err).Error("Failed to process recording with Whisper")
		ts.jobMutex.Lock()
		job.Error = fmt.Errorf("failed to process recording: %w", err)
		ts.jobMutex.Unlock()
		return
	}

	ts.logger.WithField("chunk_count", len(chunks)).Info("Transcription completed, saving chunks")

	// Store transcript chunks
	for _, chunk := range chunks {
		err := ts.storage.SaveTranscriptChunk(chunk)
		if err != nil {
			ts.logger.WithError(err).Error("Failed to save transcript chunk")
			ts.jobMutex.Lock()
			job.Error = fmt.Errorf("failed to save transcript chunk: %w", err)
			ts.jobMutex.Unlock()
			return
		}
	}

	ts.logger.WithFields(logrus.Fields{
		"activity_id":  activityID,
		"recording_id": recording.ID,
		"chunk_count":  len(chunks),
	}).Info("Transcription saved successfully")
}

// Close cleans up the transcription service
func (ts *TranscriptionService) Close() error {
	ts.logger.Info("Closing transcription service")
	if ts.modelManager != nil {
		return ts.modelManager.Close()
	}
	return nil
}