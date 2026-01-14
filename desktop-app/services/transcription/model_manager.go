package transcription

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/platformlabs-co/personal-assist/models"
	"github.com/sirupsen/logrus"
)

// ModelManager handles Whisper model downloading, storage, and loading
type ModelManager struct {
	modelsPath    string
	downloadQueue chan models.ModelDownloadRequest
	activeModel   *models.WhisperModel
	loadedModel   whisper.Model
	mutex         sync.RWMutex
	logger        *logrus.Logger
}

// NewModelManager creates a new model manager
func NewModelManager(modelsPath string, logger *logrus.Logger) *ModelManager {
	// Ensure models directory exists
	if err := os.MkdirAll(modelsPath, 0755); err != nil {
		logger.WithError(err).Error("Failed to create models directory")
	}

	mm := &ModelManager{
		modelsPath:    modelsPath,
		downloadQueue: make(chan models.ModelDownloadRequest, 10),
		logger:        logger,
	}

	// Start download worker
	go mm.downloadWorker()

	return mm
}

// GetAvailableModels returns all available models with download status
func (mm *ModelManager) GetAvailableModels() ([]models.WhisperModel, error) {
	models := models.AvailableWhisperModels()
	
	// Check which models are downloaded
	for i := range models {
		models[i].IsDownloaded = mm.isModelDownloaded(models[i].ID)
		models[i].IsActive = mm.isModelActive(models[i].ID)
	}
	
	return models, nil
}

// DownloadModel queues a model for download
func (mm *ModelManager) DownloadModel(modelID string) error {
	_, exists := models.GetModelByID(modelID)
	if !exists {
		return fmt.Errorf("unknown model ID: %s", modelID)
	}
	
	if mm.isModelDownloaded(modelID) {
		return fmt.Errorf("model %s is already downloaded", modelID)
	}
	
	mm.logger.WithField("model", modelID).Info("Queuing model for download")
	
	select {
	case mm.downloadQueue <- models.ModelDownloadRequest{ModelID: modelID, Priority: 1}:
		return nil
	default:
		return fmt.Errorf("download queue is full")
	}
}

// SetActiveModel sets the active model and loads it
func (mm *ModelManager) SetActiveModel(modelID string) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if !mm.isModelDownloaded(modelID) {
		return fmt.Errorf("model %s is not downloaded", modelID)
	}

	// Unload current model if any
	if mm.loadedModel != nil {
		mm.logger.Debug("Unloading existing model")
		mm.loadedModel.Close()
		mm.loadedModel = nil
	}

	// Load new model
	modelPath := mm.getModelPath(modelID)
	mm.logger.WithFields(logrus.Fields{
		"model_id":   modelID,
		"model_path": modelPath,
	}).Info("Loading Whisper model")

	// Check if file exists and is readable
	if stat, err := os.Stat(modelPath); err != nil {
		return fmt.Errorf("model file not accessible: %w", err)
	} else {
		mm.logger.WithField("file_size", stat.Size()).Debug("Model file found")
	}

	model, err := whisper.New(modelPath)
	if err != nil {
		mm.logger.WithError(err).Error("Failed to initialize Whisper model")
		return fmt.Errorf("failed to load model %s: %w", modelID, err)
	}

	if model == nil {
		mm.logger.Error("Whisper model is nil after loading")
		return fmt.Errorf("model loading returned nil")
	}

	mm.loadedModel = model

	// Update active model
	modelInfo, _ := models.GetModelByID(modelID)
	mm.activeModel = modelInfo

	mm.logger.WithField("model", modelID).Info("Successfully loaded model")

	return nil
}

// GetActiveModel returns the currently active model
func (mm *ModelManager) GetActiveModel() (*models.WhisperModel, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	if mm.activeModel == nil {
		return nil, fmt.Errorf("no active model")
	}
	
	return mm.activeModel, nil
}

// GetLoadedModel returns the loaded Whisper model for transcription
// Note: The returned model should not be closed by the caller - it's managed by ModelManager
func (mm *ModelManager) GetLoadedModel() whisper.Model {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	if mm.loadedModel == nil {
		mm.logger.Warn("GetLoadedModel called but no model is loaded")
	}

	return mm.loadedModel
}

// isModelDownloaded checks if a model file exists
func (mm *ModelManager) isModelDownloaded(modelID string) bool {
	modelPath := mm.getModelPath(modelID)
	_, err := os.Stat(modelPath)
	return err == nil
}

// isModelActive checks if a model is currently active
func (mm *ModelManager) isModelActive(modelID string) bool {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	return mm.activeModel != nil && mm.activeModel.ID == modelID
}

// getModelPath returns the file path for a model
func (mm *ModelManager) getModelPath(modelID string) string {
	model, _ := models.GetModelByID(modelID)
	filename := model.GetModelFilename()
	return filepath.Join(mm.modelsPath, filename)
}

// downloadWorker processes model download requests
func (mm *ModelManager) downloadWorker() {
	for req := range mm.downloadQueue {
		if err := mm.downloadModelFile(req.ModelID); err != nil {
			mm.logger.WithError(err).WithField("model", req.ModelID).Error("Failed to download model")
		}
	}
}

// downloadModelFile downloads a model file from the official repository
func (mm *ModelManager) downloadModelFile(modelID string) error {
	model, exists := models.GetModelByID(modelID)
	if !exists {
		return fmt.Errorf("unknown model ID: %s", modelID)
	}
	
	url := mm.getModelDownloadURL(modelID)
	modelPath := mm.getModelPath(modelID)
	tempPath := modelPath + ".tmp"
	
	mm.logger.WithFields(logrus.Fields{
		"model": modelID,
		"url":   url,
		"path":  modelPath,
	}).Info("Starting model download")
	
	// Create HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}
	
	// Create temporary file
	out, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()
	
	// Download with progress tracking
	hasher := sha256.New()
	writer := io.MultiWriter(out, hasher)
	
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to download model data: %w", err)
	}
	
	// Verify file size
	stat, err := out.Stat()
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to stat downloaded file: %w", err)
	}
	
	expectedSize := model.Size
	if stat.Size() != expectedSize {
		os.Remove(tempPath)
		return fmt.Errorf("downloaded file size mismatch: got %d, expected %d", stat.Size(), expectedSize)
	}
	
	// Move temporary file to final location
	if err := os.Rename(tempPath, modelPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to move model file: %w", err)
	}
	
	mm.logger.WithField("model", modelID).Info("Model download completed successfully")
	
	return nil
}

// getModelDownloadURL returns the download URL for a model
func (mm *ModelManager) getModelDownloadURL(modelID string) string {
	baseURL := "https://huggingface.co/ggerganov/whisper.cpp/resolve/main"
	model, _ := models.GetModelByID(modelID)
	filename := model.GetModelFilename()
	return fmt.Sprintf("%s/%s", baseURL, filename)
}

// EnsureDefaultModel ensures a default model is available and loaded
// This will block until a model is available (either already downloaded or download completes)
func (mm *ModelManager) EnsureDefaultModel() error {
	// Check if we have any model loaded
	if mm.loadedModel != nil {
		return nil
	}

	// Try to load a downloaded model
	availableModels, _ := mm.GetAvailableModels()
	for _, model := range availableModels {
		if model.IsDownloaded {
			return mm.SetActiveModel(model.ID)
		}
	}

	// No model available, download the small model as default
	mm.logger.Info("No models available, downloading default small model")
	if err := mm.DownloadModel("small"); err != nil {
		return fmt.Errorf("failed to queue default model download: %w", err)
	}

	// Wait for download to complete
	mm.logger.Info("Waiting for model download to complete...")
	modelPath := mm.getModelPath("small")
	maxWait := 10 * time.Minute // 10 minutes for 465MB download
	checkInterval := 2 * time.Second
	elapsed := time.Duration(0)

	for elapsed < maxWait {
		time.Sleep(checkInterval)
		elapsed += checkInterval

		// Check if file exists and has the expected size
		if stat, err := os.Stat(modelPath); err == nil {
			modelInfo, _ := models.GetModelByID("small")
			if stat.Size() == modelInfo.Size {
				mm.logger.Info("Model download completed, loading model")
				return mm.SetActiveModel("small")
			}
		}
	}

	return fmt.Errorf("model download timed out after %v", maxWait)
}

// Close closes the model manager and cleans up resources
func (mm *ModelManager) Close() error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	// Close download queue
	close(mm.downloadQueue)
	
	// Unload model
	if mm.loadedModel != nil {
		mm.loadedModel.Close()
		mm.loadedModel = nil
	}
	
	return nil
}