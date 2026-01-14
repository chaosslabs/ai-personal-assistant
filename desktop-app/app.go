package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/platformlabs-co/personal-assist/database"
	"github.com/platformlabs-co/personal-assist/logger"
	"github.com/platformlabs-co/personal-assist/models"
	"github.com/platformlabs-co/personal-assist/services"
	"github.com/platformlabs-co/personal-assist/storage"
	"github.com/platformlabs-co/personal-assist/views"
)

//go:embed version.txt
var versionData string

const APPNAME = "Personal Assist"

// App struct
type App struct {
	ctx                  context.Context
	db                   *database.DB
	activityService      *services.ActivityService
	audioService         *services.AudioService
	transcriptionService *services.TranscriptionService
	fileManager          *storage.FileManager
	currentUser          *models.User
	mainView             *views.MainView
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize logging first
	if err := a.initializeLogging(); err != nil {
		runtime.LogErrorf(ctx, "Failed to initialize logging: %v", err)
	}

	logger.Info("Application starting up")

	// Initialize database and services
	if err := a.initializeApp(); err != nil {
		logger.WithError(err).Error("Failed to initialize app")
		runtime.LogErrorf(ctx, "Failed to initialize app: %v", err)
		// Show error dialog but don't exit
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Initialization Error",
			Message: fmt.Sprintf("Failed to initialize Personal Assist:\n%v", err),
		})
	} else {
		logger.Info("Application initialized successfully")
	}
}

// initializeLogging initializes the logging system
func (a *App) initializeLogging() error {
	config := logger.DefaultLogConfig()
	return logger.InitLogger(config)
}

// initializeApp initializes the database and services
func (a *App) initializeApp() error {
	logger.Info("Starting application initialization")

	// Get default database configuration
	config := database.DefaultConfig()
	logger.WithField("data_dir", config.DataDir).Info("Using database configuration")

	// Ensure directories exist
	if err := database.EnsureDirectories(config.DataDir); err != nil {
		logger.WithError(err).Error("Failed to create directories")
		return fmt.Errorf("failed to create directories: %w", err)
	}
	logger.Info("Directories created successfully")

	// Initialize database connection
	dbPath := config.DataDir + "/" + config.DBName
	logger.WithField("db_path", dbPath).Info("Connecting to database")
	db, err := database.NewDB(config)
	if err != nil {
		logger.WithError(err).Error("Failed to connect to database")
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	a.db = db
	logger.Info("Database connected successfully")

	// Initialize database schema
	logger.Info("Initializing database schema")
	migrator := database.NewMigrator(db)
	if err := migrator.InitializeSchema(); err != nil {
		logger.WithError(err).Error("Failed to initialize schema")
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Validate schema
	logger.Info("Validating database schema")
	if err := migrator.Validate(); err != nil {
		logger.WithError(err).Error("Schema validation failed")
		return fmt.Errorf("schema validation failed: %w", err)
	}
	logger.Info("Database schema validated successfully")

	// Initialize file manager
	a.fileManager = storage.NewFileManager(config.DataDir)

	// Initialize storage layer
	sqliteStorage := storage.NewSQLiteStorage(db)

	// Initialize services
	logger.Info("Initializing services")
	a.activityService = services.NewActivityService(sqliteStorage, a.fileManager)
	a.audioService = services.NewAudioService(sqliteStorage, a.fileManager)

	// Initialize transcription service with Whisper integration
	modelsPath := a.fileManager.GetModelsDir()
	a.transcriptionService = services.NewTranscriptionService(sqliteStorage, config.DataDir, modelsPath, logger.GetLogger())

	// Initialize views
	logger.Info("Initializing views")
	a.mainView = views.NewMainView(a.activityService, a.audioService)

	// Initialize or get user
	logger.Info("Initializing user")
	if err := a.initializeUser(sqliteStorage); err != nil {
		logger.WithError(err).Error("Failed to initialize user")
		return fmt.Errorf("failed to initialize user: %w", err)
	}

	logger.Info("Application initialization completed successfully")
	return nil
}

// initializeUser creates or retrieves the current user
func (a *App) initializeUser(storage *storage.SQLiteStorage) error {
	// Try to get existing user
	user, err := storage.GetFirstUser()
	if err != nil && err.Error() != "no users found" {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Create new user if none exists
	if user == nil {
		// Get system username
		username := "User"
		if systemUser, err := os.UserHomeDir(); err == nil {
			// Extract username from home directory path
			parts := strings.Split(systemUser, "/")
			if len(parts) > 0 {
				username = parts[len(parts)-1]
			}
		}

		user = models.NewUser(username)
		if err := storage.CreateUser(user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	a.currentUser = user
	return nil
}

// ShowWindow shows the main window
func (a *App) ShowWindow() {
	runtime.Show(a.ctx)
}

// HideWindow hides the main window
func (a *App) HideWindow() {
	runtime.Hide(a.ctx)
}

// QuitApp quits the application
func (a *App) QuitApp() {
	// Cleanup services
	// if a.transcriptionService != nil {
	//	a.transcriptionService.Close()
	// }
	runtime.Quit(a.ctx)
}

// GetVersion returns the embedded version from version.txt
func (a *App) GetVersion() string {
	return strings.TrimSpace(versionData)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetCurrentUser returns the current user information
func (a *App) GetCurrentUser() map[string]interface{} {
	if a.currentUser == nil {
		return map[string]interface{}{
			"error": "No user initialized",
		}
	}

	return map[string]interface{}{
		"id":         a.currentUser.ID,
		"username":   a.currentUser.Username,
		"settings":   a.currentUser.Settings,
		"created_at": a.currentUser.CreatedAt.Format(time.RFC3339),
	}
}

// GetDatabasePath returns the path to the database file
func (a *App) GetDatabasePath() string {
	if a.db == nil {
		return "Database not initialized"
	}
	return a.db.Path()
}

// GetAppStatus returns the current application status
func (a *App) GetAppStatus() string {
	status := fmt.Sprintf("%s is running\n", APPNAME)

	if a.db != nil {
		status += "Database: Connected\n"
		status += fmt.Sprintf("Database Path: %s\n", a.db.Path())
	} else {
		status += "Database: Not connected\n"
	}

	if a.currentUser != nil {
		status += fmt.Sprintf("User: %s (%s)\n", a.currentUser.Username, a.currentUser.ID[:8])
	}

	return status
}

// CreateTestActivity creates a test activity for demonstration
func (a *App) CreateTestActivity() map[string]interface{} {
	if a.currentUser == nil || a.activityService == nil {
		return map[string]interface{}{
			"error": "Services not initialized",
		}
	}

	activity, err := a.activityService.CreateActivity(
		a.currentUser.ID,
		models.ActivityTypeMeeting,
		"Test Meeting",
	)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to create activity: %v", err),
		}
	}

	return map[string]interface{}{
		"id":         activity.ID,
		"title":      activity.Title,
		"type":       activity.Type,
		"status":     activity.Status,
		"start_time": activity.StartTime.Format(time.RFC3339),
		"created_at": activity.CreatedAt.Format(time.RFC3339),
	}
}

// ListActivities returns a list of activities for the current user
func (a *App) ListActivities() map[string]interface{} {
	if a.currentUser == nil || a.activityService == nil {
		return map[string]interface{}{
			"error": "Services not initialized",
		}
	}

	activities, err := a.activityService.ListActivities(a.currentUser.ID, 10, 0)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to list activities: %v", err),
		}
	}

	result := make([]map[string]interface{}, len(activities))
	for i, activity := range activities {
		result[i] = map[string]interface{}{
			"id":         activity.ID,
			"title":      activity.Title,
			"type":       activity.Type,
			"status":     activity.Status,
			"start_time": activity.StartTime.Format(time.RFC3339),
			"duration":   activity.Duration().String(),
		}
	}

	return map[string]interface{}{
		"activities": result,
		"count":      len(activities),
	}
}

// GetSystemInfo returns system information for debugging
func (a *App) GetSystemInfo() map[string]interface{} {
	config := database.DefaultConfig()

	info := map[string]interface{}{
		"data_directory": config.DataDir,
		"database_file":  config.DBName,
	}

	if a.db != nil {
		info["database_path"] = a.db.Path()
		info["database_connected"] = true
	} else {
		info["database_connected"] = false
	}

	if a.fileManager != nil {
		info["activities_dir"] = a.fileManager.GetActivitiesDir()
		info["models_dir"] = a.fileManager.GetModelsDir()

		// Get disk usage
		if usage, err := a.fileManager.GetDiskUsage(); err == nil {
			info["disk_usage_bytes"] = usage
			info["disk_usage_mb"] = float64(usage) / (1024 * 1024)
		}
	}

	return info
}

// ==================== ACTIVITY MANAGEMENT ====================

// CreateActivity creates a new activity
func (a *App) CreateActivity(actType, title string) (*models.Activity, error) {
	if a.currentUser == nil || a.activityService == nil {
		return nil, fmt.Errorf("services not initialized")
	}

	activity, err := a.activityService.CreateActivity(a.currentUser.ID, models.ActivityType(actType), title)
	if err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	return activity, nil
}

// GetActivities returns a list of activities with optional filtering
func (a *App) GetActivities() ([]*models.Activity, error) {
	logger.Info("GetActivities called")

	if a.currentUser == nil || a.activityService == nil {
		logger.Error("Services not initialized for GetActivities")
		return nil, fmt.Errorf("services not initialized")
	}

	logger.WithField("user_id", a.currentUser.ID).Info("Getting activities for user")
	activities, err := a.activityService.ListActivities(a.currentUser.ID, 50, 0)
	if err != nil {
		logger.WithError(err).WithField("user_id", a.currentUser.ID).Error("Failed to get activities from service")
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"user_id": a.currentUser.ID,
		"count":   len(activities),
	}).Info("Activities retrieved successfully")

	return activities, nil
}

// GetActivity returns a specific activity by ID
func (a *App) GetActivity(activityID string) (*models.Activity, error) {
	if a.currentUser == nil || a.activityService == nil {
		return nil, fmt.Errorf("services not initialized")
	}

	activity, err := a.activityService.GetActivity(a.currentUser.ID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return activity, nil
}

// StartActivity starts an existing activity
func (a *App) StartActivity(activityID string) error {
	if a.currentUser == nil || a.activityService == nil {
		return fmt.Errorf("services not initialized")
	}

	_, err := a.activityService.StartActivity(a.currentUser.ID, activityID)
	if err != nil {
		return fmt.Errorf("failed to start activity: %w", err)
	}

	// Emit event for frontend
	runtime.EventsEmit(a.ctx, "activity:started", activityID)
	return nil
}

// StopActivity stops an existing activity
func (a *App) StopActivity(activityID string) error {
	if a.currentUser == nil || a.activityService == nil {
		return fmt.Errorf("services not initialized")
	}

	_, err := a.activityService.CompleteActivity(a.currentUser.ID, activityID)
	if err != nil {
		return fmt.Errorf("failed to stop activity: %w", err)
	}

	// Emit event for frontend
	runtime.EventsEmit(a.ctx, "activity:stopped", activityID)
	return nil
}

// UpdateActivityTitle updates the title of an activity
func (a *App) UpdateActivityTitle(activityID, newTitle string) error {
	logger.WithFields(map[string]interface{}{
		"activity_id": activityID,
		"new_title":   newTitle,
	}).Info("UpdateActivityTitle called")

	if a.currentUser == nil || a.activityService == nil {
		return fmt.Errorf("services not initialized")
	}

	// Get the activity
	activity, err := a.activityService.GetActivity(a.currentUser.ID, activityID)
	if err != nil {
		logger.WithError(err).WithField("activity_id", activityID).Error("Failed to get activity for title update")
		return fmt.Errorf("failed to get activity: %w", err)
	}

	// Update the title
	activity.Title = newTitle

	// Save the updated activity
	if err := a.activityService.UpdateActivity(activity); err != nil {
		logger.WithError(err).WithField("activity_id", activityID).Error("Failed to update activity title")
		return fmt.Errorf("failed to update activity title: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"activity_id": activityID,
		"new_title":   newTitle,
	}).Info("Activity title updated successfully")

	// Emit event for frontend
	runtime.EventsEmit(a.ctx, "activity:updated", activityID)
	return nil
}

// UpdateActivityType updates the type of an activity
func (a *App) UpdateActivityType(activityID, newType string) error {
	logger.WithFields(map[string]interface{}{
		"activity_id": activityID,
		"new_type":    newType,
	}).Info("UpdateActivityType called")

	if a.currentUser == nil || a.activityService == nil {
		return fmt.Errorf("services not initialized")
	}

	// Validate the activity type
	validTypes := map[string]bool{
		"meeting":      true,
		"work_session": true,
		"call":         true,
		"other":        true,
	}
	if !validTypes[newType] {
		return fmt.Errorf("invalid activity type: %s", newType)
	}

	// Get the activity
	activity, err := a.activityService.GetActivity(a.currentUser.ID, activityID)
	if err != nil {
		logger.WithError(err).WithField("activity_id", activityID).Error("Failed to get activity for type update")
		return fmt.Errorf("failed to get activity: %w", err)
	}

	// Update the type
	activity.Type = models.ActivityType(newType)

	// Save the updated activity
	if err := a.activityService.UpdateActivity(activity); err != nil {
		logger.WithError(err).WithField("activity_id", activityID).Error("Failed to update activity type")
		return fmt.Errorf("failed to update activity type: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"activity_id": activityID,
		"new_type":    newType,
	}).Info("Activity type updated successfully")

	// Emit event for frontend
	runtime.EventsEmit(a.ctx, "activity:updated", activityID)
	return nil
}

// DeleteActivity deletes an activity with proper validation and cleanup
func (a *App) DeleteActivity(activityID string) error {
	logger.WithField("activity_id", activityID).Info("Delete activity requested")

	if a.mainView == nil {
		logger.Error("Main view not initialized for delete activity")
		return fmt.Errorf("main view not initialized")
	}

	// Get activity details for logging before deletion
	activity, err := a.GetActivity(activityID)
	if err != nil {
		logger.WithError(err).WithField("activity_id", activityID).Error("Failed to get activity for deletion")
		return fmt.Errorf("failed to get activity for deletion: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"activity_id": activityID,
		"title":       activity.Title,
		"type":        activity.Type,
		"status":      activity.Status,
	}).Info("Deleting activity")

	err = a.mainView.DeleteActivityAction(a.currentUser.ID, activityID)
	if err != nil {
		logger.WithError(err).WithFields(map[string]interface{}{
			"activity_id": activityID,
			"title":       activity.Title,
		}).Error("Failed to delete activity")
		return fmt.Errorf("failed to delete activity: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"activity_id": activityID,
		"title":       activity.Title,
	}).Info("Activity successfully deleted")

	// Emit event for frontend to refresh UI
	runtime.EventsEmit(a.ctx, "activity:deleted", activityID)
	return nil
}

// ==================== RECORDING MANAGEMENT ====================

// StartRecordingButtonAction handles the simple "Start Recording" button action
func (a *App) StartRecordingButtonAction() (*views.RecordingSession, error) {
	logger.Info("StartRecordingButtonAction called")

	if a.currentUser == nil || a.mainView == nil {
		logger.Error("Services not initialized for start recording")
		return nil, fmt.Errorf("services not initialized")
	}

	logger.WithField("user_id", a.currentUser.ID).Info("Starting recording button action")
	session, err := a.mainView.StartRecordingButtonAction(a.currentUser.ID)
	if err != nil {
		logger.WithError(err).WithField("user_id", a.currentUser.ID).Error("Failed to start recording in main view")
		return nil, fmt.Errorf("failed to start recording: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"activity_id":  session.Activity.ID,
		"recording_id": session.AudioRecording.ID,
		"file_path":    session.FilePath,
	}).Info("Recording started successfully")

	// Emit event for frontend
	runtime.EventsEmit(a.ctx, "recording:started", session.AudioRecording)
	return session, nil
}

// StopRecordingButtonAction handles stopping the current recording
func (a *App) StopRecordingButtonAction(recordingID string) error {
	logger.WithField("recording_id", recordingID).Info("StopRecordingButtonAction called")

	if a.mainView == nil {
		logger.Error("Main view not initialized for stop recording")
		return fmt.Errorf("main view not initialized")
	}

	logger.WithField("recording_id", recordingID).Info("Stopping recording via main view")
	err := a.mainView.StopRecordingButtonAction(a.currentUser.ID, recordingID)
	if err != nil {
		logger.WithError(err).WithField("recording_id", recordingID).Error("Failed to stop recording in main view")
		return fmt.Errorf("failed to stop recording: %w", err)
	}

	logger.WithField("recording_id", recordingID).Info("Recording stopped successfully")

	// Emit event for frontend
	runtime.EventsEmit(a.ctx, "recording:stopped", recordingID)
	return nil
}

// ==================== TRANSCRIPT MANAGEMENT ====================

// ProcessActivityTranscription starts transcription for an activity
func (a *App) ProcessActivityTranscription(activityID string) error {
	logger.WithField("activity_id", activityID).Info("ProcessActivityTranscription called")

	if a.transcriptionService == nil {
		return fmt.Errorf("transcription service not initialized")
	}
	if a.currentUser == nil {
		return fmt.Errorf("no user logged in")
	}

	// Use default transcription config
	config := models.DefaultTranscriptionConfig()
	err := a.transcriptionService.ProcessActivity(a.currentUser.ID, activityID, config)

	if err != nil {
		logger.WithError(err).Error("Failed to start transcription")
		runtime.EventsEmit(a.ctx, "transcription:error", map[string]interface{}{
			"activity_id": activityID,
			"error":       err.Error(),
		})
		return err
	}

	// Emit event that transcription started
	runtime.EventsEmit(a.ctx, "transcription:started", activityID)

	// Start a goroutine to monitor transcription progress
	go a.monitorTranscription(activityID)

	return nil
}

// ProcessRecordingTranscription starts transcription for a specific recording
func (a *App) ProcessRecordingTranscription(activityID, recordingID string) error {
	if a.transcriptionService == nil {
		return fmt.Errorf("transcription service not initialized")
	}
	if a.currentUser == nil {
		return fmt.Errorf("no user logged in")
	}

	// Use default transcription config
	config := models.DefaultTranscriptionConfig()
	return a.transcriptionService.ProcessRecording(a.currentUser.ID, activityID, recordingID, config)
}

// GetTranscriptionStatus returns the status of an ongoing transcription
func (a *App) GetTranscriptionStatus(activityID string) (*models.TranscriptionStatus, error) {
	if a.transcriptionService == nil {
		return nil, fmt.Errorf("transcription service not initialized")
	}
	if a.currentUser == nil {
		return nil, fmt.Errorf("no user logged in")
	}
	return a.transcriptionService.GetTranscriptionStatus(a.currentUser.ID, activityID)
}

// GetActivityTranscript returns transcript chunks for an activity
func (a *App) GetActivityTranscript(activityID string) ([]*models.TranscriptChunk, error) {
	if a.transcriptionService == nil {
		return nil, fmt.Errorf("transcription service not initialized")
	}
	if a.currentUser == nil {
		return nil, fmt.Errorf("no user logged in")
	}
	return a.transcriptionService.GetTranscript(a.currentUser.ID, activityID)
}

// GetRecordingTranscript returns transcript chunks for a specific recording
func (a *App) GetRecordingTranscript(recordingID string) ([]*models.TranscriptChunk, error) {
	if a.transcriptionService == nil {
		return nil, fmt.Errorf("transcription service not initialized")
	}
	if a.currentUser == nil {
		return nil, fmt.Errorf("no user logged in")
	}
	return a.transcriptionService.GetRecordingTranscript(a.currentUser.ID, recordingID)
}

// SearchActivityTranscripts searches across all transcripts
func (a *App) SearchActivityTranscripts(query string) ([]*models.TranscriptChunk, error) {
	if a.transcriptionService == nil {
		return nil, fmt.Errorf("transcription service not initialized")
	}
	if a.currentUser == nil {
		return nil, fmt.Errorf("no user logged in")
	}
	filter := services.SearchFilter{} // Empty filter for now
	return a.transcriptionService.SearchTranscripts(a.currentUser.ID, query, filter)
}

// ==================== WHISPER MODEL MANAGEMENT ====================

// GetAvailableModels returns all available Whisper models
func (a *App) GetAvailableModels() ([]models.WhisperModel, error) {
	if a.transcriptionService == nil {
		return nil, fmt.Errorf("transcription service not initialized")
	}
	return a.transcriptionService.GetAvailableModels()
}

// DownloadModel downloads a Whisper model
func (a *App) DownloadModel(modelID string) error {
	if a.transcriptionService == nil {
		return fmt.Errorf("transcription service not initialized")
	}
	return a.transcriptionService.DownloadModel(modelID)
}

// SetActiveModel sets the active Whisper model
func (a *App) SetActiveModel(modelID string) error {
	if a.transcriptionService == nil {
		return fmt.Errorf("transcription service not initialized")
	}
	return a.transcriptionService.SetActiveModel(modelID)
}

// GetActiveModel returns the currently active model
func (a *App) GetActiveModel() (*models.WhisperModel, error) {
	if a.transcriptionService == nil {
		return nil, fmt.Errorf("transcription service not initialized")
	}
	return a.transcriptionService.GetActiveModel()
}

// ==================== LEGACY TRANSCRIPT METHODS ====================

// GetTranscript returns transcript chunks for an activity
func (a *App) GetTranscript(activityID string) ([]*models.TranscriptChunk, error) {
	// Legacy method - delegate to GetActivityTranscript
	return a.GetActivityTranscript(activityID)
}

// SearchTranscripts searches across all transcripts
func (a *App) SearchTranscripts(query string) ([]*models.TranscriptChunk, error) {
	// Legacy method - delegate to SearchActivityTranscripts
	return a.SearchActivityTranscripts(query)
}

// ==================== SYSTEM MANAGEMENT ====================

// GetAudioDevices returns available audio input devices
func (a *App) GetAudioDevices() ([]map[string]interface{}, error) {
	logger.Info("GetAudioDevices called - getting real audio devices")

	if a.audioService == nil {
		logger.Error("Audio service not initialized")
		return nil, fmt.Errorf("audio service not initialized")
	}

	// Get real audio devices using the audio recorder
	devices, err := a.audioService.AudioRecorder.ListAudioDevices()
	if err != nil {
		logger.WithError(err).Error("Failed to get audio devices")
		// Return a fallback default device
		fallbackDevices := []map[string]interface{}{
			{
				"id":          "0",
				"name":        "Default Microphone",
				"type":        "input",
				"sample_rate": 44100,
				"channels":    2,
			},
		}
		return fallbackDevices, nil
	}

	// Convert to the expected format
	var deviceList []map[string]interface{}
	for _, device := range devices {
		deviceMap := map[string]interface{}{
			"id":          device.DeviceID,
			"name":        device.Name,
			"type":        "input",
			"sample_rate": device.SampleRate,
			"channels":    device.Channels,
			"device_type": device.DeviceType,
		}
		deviceList = append(deviceList, deviceMap)
	}

	logger.WithField("device_count", len(deviceList)).Info("Retrieved audio devices successfully")
	return deviceList, nil
}

// GetAppStatusDetailed returns detailed application status
func (a *App) GetAppStatusDetailed() map[string]interface{} {
	status := map[string]interface{}{
		"app_name": APPNAME,
		"version":  a.GetVersion(),
	}

	if a.db != nil {
		status["database"] = map[string]interface{}{
			"connected": true,
			"path":      a.db.Path(),
		}
	} else {
		status["database"] = map[string]interface{}{
			"connected": false,
		}
	}

	if a.currentUser != nil {
		status["user"] = map[string]interface{}{
			"id":       a.currentUser.ID,
			"username": a.currentUser.Username,
		}
	}

	return status
}

// ==================== RECORDING MODE CONFIGURATION ====================

// GetRecordingModes returns available recording modes with system capability awareness
func (a *App) GetRecordingModes() []map[string]interface{} {
	modes := services.GetRecordingModes()

	result := make([]map[string]interface{}, len(modes))
	for i, mode := range modes {
		result[i] = map[string]interface{}{
			"id":           mode.ID,
			"name":         mode.Name,
			"description":  mode.Description,
			"icon":         mode.Icon,
			"available":    mode.Available,
			"recommended":  mode.Recommended,
			"requirements": mode.Requirements,
		}
	}

	return result
}

// GetSystemCapabilities returns the audio recording capabilities of the system
func (a *App) GetSystemCapabilities() map[string]interface{} {
	capabilities := services.GetSystemCapabilities()

	return map[string]interface{}{
		"macos_version":            capabilities.MacOSVersion,
		"supports_core_audio_taps": capabilities.SupportsCoreAudioTaps,
		"supports_screen_capture":  capabilities.SupportsScreenCapture,
		"available_recording_modes": capabilities.AvailableRecordingModes,
		"recommended_mode":         capabilities.RecommendedMode,
		"requires_permissions":     capabilities.RequiresPermissions,
		"permission_message":       capabilities.PermissionMessage,
	}
}

// CreateRecordingWithMode creates a recording with a specific mode
func (a *App) CreateRecordingWithMode(activityID, recordingMode string) (*views.RecordingSession, error) {
	if a.currentUser == nil || a.mainView == nil {
		return nil, fmt.Errorf("service not initialized")
	}

	// Get default config and override the recording mode
	config := a.audioService.GetDefaultRecordingConfig()
	config.RecordingMode = recordingMode

	// Get built-in microphone for device info
	builtInMic, err := a.audioService.AudioRecorder.GetBuiltInMicrophone()
	if err != nil {
		logger.WithError(err).Warn("Failed to get built-in microphone, using default")
		builtInMic = &models.AudioDeviceInfo{
			Name:       "Default Audio Device",
			DeviceID:   "input_0",
			SampleRate: 44100,
			Channels:   2,
			DeviceType: "microphone",
		}
	}

	// Create audio recording with the specified mode
	audioRecording, err := a.audioService.CreateAudioRecording(a.currentUser.ID, activityID, *builtInMic, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio recording: %w", err)
	}

	// Get activity info
	activity, err := a.activityService.GetActivity(a.currentUser.ID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	filePath := a.audioService.GetAudioFilePath(audioRecording)

	// Return recording session info
	return &views.RecordingSession{
		Activity:       activity,
		AudioRecording: audioRecording,
		FilePath:       filePath,
	}, nil
}

// GetUserSettings returns the current user's settings
func (a *App) GetUserSettings() (map[string]interface{}, error) {
	if a.currentUser == nil {
		return nil, fmt.Errorf("no user logged in")
	}

	settings := map[string]interface{}{
		"preferred_audio_device": a.currentUser.Settings.PreferredAudioDevice,
		"whisper_model":          a.currentUser.Settings.WhisperModel,
		"auto_start_recording":   a.currentUser.Settings.AutoStartRecording,
		"transcription_language": a.currentUser.Settings.TranscriptionLanguage,
		"audio_quality":          a.currentUser.Settings.AudioQuality,
		"storage_location":       a.currentUser.Settings.StorageLocation,
	}

	return settings, nil
}

// UpdateUserSettings updates the current user's settings
func (a *App) UpdateUserSettings(settingsJSON map[string]interface{}) error {
	if a.currentUser == nil {
		return fmt.Errorf("no user logged in")
	}

	// Parse the settings from the map
	newSettings := models.UserSettings{}

	if val, ok := settingsJSON["preferred_audio_device"].(string); ok {
		newSettings.PreferredAudioDevice = val
	}
	if val, ok := settingsJSON["whisper_model"].(string); ok {
		newSettings.WhisperModel = val
	}
	if val, ok := settingsJSON["auto_start_recording"].(bool); ok {
		newSettings.AutoStartRecording = val
	}
	if val, ok := settingsJSON["transcription_language"].(string); ok {
		newSettings.TranscriptionLanguage = val
	}
	if val, ok := settingsJSON["audio_quality"].(string); ok {
		newSettings.AudioQuality = val
	}
	if val, ok := settingsJSON["storage_location"].(string); ok {
		newSettings.StorageLocation = val
	}

	// Update the user's settings
	a.currentUser.UpdateSettings(newSettings)

	// Save to database
	if a.db != nil {
		// Convert settings to JSON and save to database
		settingsJSON, err := a.currentUser.SettingsToJSON()
		if err != nil {
			return fmt.Errorf("failed to serialize settings: %w", err)
		}

		// Update user in database with new settings
		query := `UPDATE users SET settings = ?, updated_at = ? WHERE id = ?`
		_, err = a.db.Exec(query, settingsJSON, time.Now().Unix(), a.currentUser.ID)
		if err != nil {
			return fmt.Errorf("failed to save settings to database: %w", err)
		}

		logger.WithField("user_id", a.currentUser.ID).Info("User settings updated successfully")
	}

	return nil
}

// monitorTranscription monitors transcription progress and emits events
func (a *App) monitorTranscription(activityID string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		status, err := a.transcriptionService.GetTranscriptionStatus(a.currentUser.ID, activityID)
		if err != nil {
			logger.WithError(err).Error("Failed to get transcription status")
			return
		}

		// Emit progress event
		runtime.EventsEmit(a.ctx, "transcription:progress", map[string]interface{}{
			"activity_id": activityID,
			"stage":       status.Stage,
			"progress":    status.Progress,
		})

		// Check if completed or failed
		if status.Stage == "completed" {
			runtime.EventsEmit(a.ctx, "transcription:completed", activityID)
			logger.WithField("activity_id", activityID).Info("Transcription completed")
			return
		} else if status.Stage == "failed" {
			runtime.EventsEmit(a.ctx, "transcription:error", map[string]interface{}{
				"activity_id": activityID,
				"error":       status.LastError,
			})
			logger.WithField("activity_id", activityID).Error("Transcription failed")
			return
		}
	}
}
