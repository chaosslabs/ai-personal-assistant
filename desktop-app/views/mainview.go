package views

import (
	"fmt"

	"github.com/platformlabs-co/personal-assist/logger"
	"github.com/platformlabs-co/personal-assist/models"
	"github.com/platformlabs-co/personal-assist/services"
)

// MainView handles main application view operations
type MainView struct {
	activityService *services.ActivityService
	audioService    *services.AudioService
}

// NewMainView creates a new main view
func NewMainView(activityService *services.ActivityService, audioService *services.AudioService) *MainView {
	return &MainView{
		activityService: activityService,
		audioService:    audioService,
	}
}

// StartRecordingButtonAction handles the "Start Recording" button action
// Creates a new "ManualRecording" activity and starts recording
func (v *MainView) StartRecordingButtonAction(userID string) (*RecordingSession, error) {
	logger.WithField("user_id", userID).Info("MainView StartRecordingButtonAction called")

	// Create a new ManualRecording activity
	logger.WithField("user_id", userID).Info("Creating manual recording activity")
	activity, err := v.activityService.CreateActivity(userID, models.ActivityTypeOther, "Manual Recording")
	if err != nil {
		logger.WithError(err).WithField("user_id", userID).Error("Failed to create manual recording activity")
		return nil, fmt.Errorf("failed to create manual recording activity: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"user_id":     userID,
		"activity_id": activity.ID,
		"title":       activity.Title,
	}).Info("Manual recording activity created successfully")

	// Set the activity metadata to indicate it's a manual recording
	logger.WithField("activity_id", activity.ID).Info("Setting activity metadata")
	activity.SetMetadata("recording_type", "manual")
	activity.SetMetadata("auto_created", true)
	if err := v.activityService.UpdateActivity(activity); err != nil {
		logger.WithError(err).WithField("activity_id", activity.ID).Error("Failed to update activity metadata")
		return nil, fmt.Errorf("failed to update activity metadata: %w", err)
	}

	// Start the activity (sets status to recording)
	logger.WithField("activity_id", activity.ID).Info("Starting activity (setting status to recording)")
	activity, err = v.activityService.StartActivity(userID, activity.ID)
	if err != nil {
		logger.WithError(err).WithField("activity_id", activity.ID).Error("Failed to start activity")
		return nil, fmt.Errorf("failed to start activity: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"activity_id": activity.ID,
		"status":      activity.Status,
	}).Info("Activity started successfully")

	// Create audio recording with built-in microphone if available
	logger.WithField("activity_id", activity.ID).Info("Getting built-in microphone for recording")
	builtInMic, err := v.audioService.AudioRecorder.GetBuiltInMicrophone()
	if err != nil {
		logger.WithError(err).WithField("activity_id", activity.ID).Warn("Failed to get built-in microphone, using default")
		builtInMic = &models.AudioDeviceInfo{
			Name:       "Default Audio Device",
			DeviceID:   "0",
			SampleRate: 44100,
			Channels:   2,
			DeviceType: "microphone",
		}
	}
	
	deviceInfo := *builtInMic
	
	logger.WithFields(map[string]interface{}{
		"device_name":   deviceInfo.Name,
		"device_id":     deviceInfo.DeviceID,
		"sample_rate":   deviceInfo.SampleRate,
		"channels":      deviceInfo.Channels,
		"device_type":   deviceInfo.DeviceType,
		"activity_id":   activity.ID,
	}).Info("Getting default recording config and creating audio recording")
	
	config := v.audioService.GetDefaultRecordingConfig()
	
	audioRecording, err := v.audioService.CreateAudioRecording(userID, activity.ID, deviceInfo, config)
	if err != nil {
		logger.WithError(err).WithFields(map[string]interface{}{
			"user_id":     userID,
			"activity_id": activity.ID,
		}).Error("Failed to create audio recording, cleaning up activity")
		
		// If audio recording creation fails, we should clean up the activity
		_, failErr := v.activityService.FailActivity(userID, activity.ID)
		if failErr != nil {
			logger.WithError(failErr).WithField("activity_id", activity.ID).Error("Failed to mark activity as failed during cleanup")
		}
		return nil, fmt.Errorf("failed to create audio recording: %w", err)
	}

	filePath := v.audioService.GetAudioFilePath(audioRecording)
	
	logger.WithFields(map[string]interface{}{
		"activity_id":   activity.ID,
		"recording_id":  audioRecording.ID,
		"file_path":     filePath,
	}).Info("Recording session created successfully")

	// Return recording session info
	return &RecordingSession{
		Activity:       activity,
		AudioRecording: audioRecording,
		FilePath:       filePath,
	}, nil
}

// StopRecordingButtonAction handles stopping the current recording
func (v *MainView) StopRecordingButtonAction(userID, recordingID string) error {
	logger.WithField("recording_id", recordingID).Info("MainView StopRecordingButtonAction called")

	// Complete the audio recording
	logger.WithField("recording_id", recordingID).Info("Attempting to complete audio recording")
	audioRecording, err := v.audioService.CompleteAudioRecording(userID, recordingID)
	if err != nil {
		logger.WithError(err).WithField("recording_id", recordingID).Error("Failed to complete audio recording")
		return fmt.Errorf("failed to complete audio recording: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"recording_id": recordingID,
		"activity_id":  audioRecording.ActivityID,
		"user_id":      audioRecording.UserID,
		"status":       audioRecording.Status,
	}).Info("Audio recording completed successfully")

	// Complete the associated activity
	logger.WithField("activity_id", audioRecording.ActivityID).Info("Attempting to complete associated activity")
	completedActivity, err := v.activityService.CompleteActivity(audioRecording.UserID, audioRecording.ActivityID)
	if err != nil {
		logger.WithError(err).WithFields(map[string]interface{}{
			"recording_id": recordingID,
			"activity_id":  audioRecording.ActivityID,
		}).Error("Failed to complete activity")
		return fmt.Errorf("failed to complete activity: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"recording_id": recordingID,
		"activity_id":  audioRecording.ActivityID,
		"title":        completedActivity.Title,
		"status":       completedActivity.Status,
	}).Info("MainView StopRecordingButtonAction completed successfully")

	return nil
}

// DeleteActivityAction handles activity deletion with proper validation
func (v *MainView) DeleteActivityAction(userID, activityID string) error {
	logger.WithField("activity_id", activityID).Info("MainView delete activity action called")

	// Validate that the activity exists and get its current state
	activity, err := v.activityService.GetActivity(userID, activityID)
	if err != nil {
		logger.WithError(err).WithField("activity_id", activityID).Error("Failed to get activity for deletion in MainView")
		return fmt.Errorf("failed to get activity for deletion: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"activity_id": activityID,
		"title":       activity.Title,
		"status":      activity.Status,
	}).Info("Activity found for deletion")

	// Log if we're force-deleting an active/recording activity
	if activity.Status == models.ActivityStatusActive || activity.Status == models.ActivityStatusRecording {
		logger.WithFields(map[string]interface{}{
			"activity_id": activityID,
			"title":       activity.Title,
			"status":      activity.Status,
		}).Warn("Force deleting active or recording activity (user requested)")
	}

	// Perform the deletion using the service layer
	logger.WithField("activity_id", activityID).Info("Calling activity service DeleteActivity")
	if err := v.activityService.DeleteActivity(userID, activityID); err != nil {
		logger.WithError(err).WithField("activity_id", activityID).Error("Activity service DeleteActivity failed")
		return fmt.Errorf("failed to delete activity: %w", err)
	}

	logger.WithField("activity_id", activityID).Info("MainView delete activity action completed successfully")
	return nil
}

// RecordingSession represents an active recording session
type RecordingSession struct {
	Activity       *models.Activity      `json:"activity"`
	AudioRecording *models.AudioRecording `json:"audio_recording"`
	FilePath       string                `json:"file_path"`
}