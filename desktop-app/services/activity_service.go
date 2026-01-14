package services

import (
	"fmt"
	"os"

	"github.com/platformlabs-co/personal-assist/logger"
	"github.com/platformlabs-co/personal-assist/models"
	"github.com/platformlabs-co/personal-assist/storage"
)

// ActivityService handles activity-related operations
type ActivityService struct {
	storage     *storage.SQLiteStorage
	fileManager *storage.FileManager
}

// NewActivityService creates a new activity service
func NewActivityService(storage *storage.SQLiteStorage, fileManager *storage.FileManager) *ActivityService {
	return &ActivityService{
		storage:     storage,
		fileManager: fileManager,
	}
}

// CreateActivity creates a new activity and sets up its directories
func (s *ActivityService) CreateActivity(userID string, activityType models.ActivityType, title string) (*models.Activity, error) {
	activity := models.NewActivity(userID, activityType, title)

	// Create database record
	if err := s.storage.CreateActivity(activity); err != nil {
		return nil, fmt.Errorf("failed to create activity in database: %w", err)
	}

	// Ensure activity directories exist
	if err := s.fileManager.EnsureActivityDirectories(activity.ID); err != nil {
		return nil, fmt.Errorf("failed to create activity directories: %w", err)
	}

	return activity, nil
}

// GetActivity retrieves an activity by ID
func (s *ActivityService) GetActivity(userID, id string) (*models.Activity, error) {
	return s.storage.GetActivity(userID, id)
}

// UpdateActivity updates an existing activity
func (s *ActivityService) UpdateActivity(activity *models.Activity) error {
	return s.storage.UpdateActivity(activity)
}

// StartActivity starts an activity (marks it as recording)
func (s *ActivityService) StartActivity(userID, id string) (*models.Activity, error) {
	activity, err := s.storage.GetActivity(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	activity.Start()

	if err := s.storage.UpdateActivity(activity); err != nil {
		return nil, fmt.Errorf("failed to update activity status: %w", err)
	}

	return activity, nil
}

// CompleteActivity completes an activity
func (s *ActivityService) CompleteActivity(userID, id string) (*models.Activity, error) {
	activity, err := s.storage.GetActivity(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	activity.Complete()

	if err := s.storage.UpdateActivity(activity); err != nil {
		return nil, fmt.Errorf("failed to update activity status: %w", err)
	}

	return activity, nil
}

// FailActivity marks an activity as failed
func (s *ActivityService) FailActivity(userID, id string) (*models.Activity, error) {
	activity, err := s.storage.GetActivity(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	activity.Fail()

	if err := s.storage.UpdateActivity(activity); err != nil {
		return nil, fmt.Errorf("failed to update activity status: %w", err)
	}

	return activity, nil
}

// SetActivityProcessing marks an activity as processing
func (s *ActivityService) SetActivityProcessing(userID, id string) (*models.Activity, error) {
	activity, err := s.storage.GetActivity(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	activity.SetProcessing()

	if err := s.storage.UpdateActivity(activity); err != nil {
		return nil, fmt.Errorf("failed to update activity status: %w", err)
	}

	return activity, nil
}

// ListActivities retrieves activities for a user with pagination
func (s *ActivityService) ListActivities(userID string, limit, offset int) ([]*models.Activity, error) {
	return s.storage.GetActivitiesByUser(userID, limit, offset)
}

// AddActivityTag adds a tag to an activity
func (s *ActivityService) AddActivityTag(userID, id, tag string) (*models.Activity, error) {
	activity, err := s.storage.GetActivity(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	activity.AddTag(tag)

	if err := s.storage.UpdateActivity(activity); err != nil {
		return nil, fmt.Errorf("failed to update activity: %w", err)
	}

	return activity, nil
}

// RemoveActivityTag removes a tag from an activity
func (s *ActivityService) RemoveActivityTag(userID, id, tag string) (*models.Activity, error) {
	activity, err := s.storage.GetActivity(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	activity.RemoveTag(tag)

	if err := s.storage.UpdateActivity(activity); err != nil {
		return nil, fmt.Errorf("failed to update activity: %w", err)
	}

	return activity, nil
}

// SetActivityMetadata sets metadata for an activity
func (s *ActivityService) SetActivityMetadata(userID, id, key string, value interface{}) (*models.Activity, error) {
	activity, err := s.storage.GetActivity(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	activity.SetMetadata(key, value)

	if err := s.storage.UpdateActivity(activity); err != nil {
		return nil, fmt.Errorf("failed to update activity: %w", err)
	}

	return activity, nil
}

// DeleteActivity soft deletes an activity (keeps in database but marks as deleted)
func (s *ActivityService) DeleteActivity(userID, id string) error {
	logger.WithField("activity_id", id).Info("ActivityService DeleteActivity called (soft delete)")

	// Get activity to ensure it exists
	activity, err := s.storage.GetActivity(userID, id)
	if err != nil {
		logger.WithError(err).WithField("activity_id", id).Error("Failed to get activity from storage")
		return fmt.Errorf("failed to get activity: %w", err)
	}

	// Check if already deleted
	if activity.IsDeleted() {
		logger.WithField("activity_id", id).Warn("Activity is already soft deleted")
		return fmt.Errorf("activity is already deleted")
	}

	logger.WithFields(map[string]interface{}{
		"activity_id": id,
		"title":       activity.Title,
		"type":        activity.Type,
		"status":      activity.Status,
	}).Info("Found activity for soft deletion")

	// Soft delete the activity (sets deleted_at timestamp)
	activity.SoftDelete()

	// Update in database
	logger.WithField("activity_id", id).Info("Soft deleting activity in database")
	if err := s.storage.UpdateActivity(activity); err != nil {
		logger.WithError(err).WithField("activity_id", id).Error("Failed to soft delete activity in database")
		return fmt.Errorf("failed to soft delete activity: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"activity_id": id,
		"deleted_at":  activity.DeletedAt,
	}).Info("ActivityService soft delete completed successfully")
	
	return nil
}

// RestoreActivity restores a soft-deleted activity
func (s *ActivityService) RestoreActivity(userID, id string) error {
	logger.WithField("activity_id", id).Info("ActivityService RestoreActivity called")

	// Get activity to ensure it exists
	activity, err := s.storage.GetActivity(userID, id)
	if err != nil {
		logger.WithError(err).WithField("activity_id", id).Error("Failed to get activity from storage")
		return fmt.Errorf("failed to get activity: %w", err)
	}

	// Check if it's actually deleted
	if !activity.IsDeleted() {
		logger.WithField("activity_id", id).Warn("Activity is not deleted, cannot restore")
		return fmt.Errorf("activity is not deleted")
	}

	logger.WithFields(map[string]interface{}{
		"activity_id": id,
		"title":       activity.Title,
	}).Info("Restoring soft-deleted activity")

	// Restore the activity
	activity.Restore()

	// Update in database
	if err := s.storage.UpdateActivity(activity); err != nil {
		logger.WithError(err).WithField("activity_id", id).Error("Failed to restore activity in database")
		return fmt.Errorf("failed to restore activity: %w", err)
	}

	logger.WithField("activity_id", id).Info("ActivityService restore completed successfully")
	return nil
}

// GetActivityAudioFiles lists audio files associated with an activity
func (s *ActivityService) GetActivityAudioFiles(id string) ([]string, error) {
	return s.fileManager.ListAudioFiles(id)
}

// GetActivityDiskUsage returns disk usage for a specific activity
func (s *ActivityService) GetActivityDiskUsage(id string) (int64, error) {
	activityDir := s.fileManager.GetActivityDir(id)

	if !s.fileManager.FileExists(activityDir) {
		return 0, nil
	}

	var totalSize int64
	entries, err := os.ReadDir(activityDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read activity directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				totalSize += info.Size()
			}
		}
	}

	return totalSize, nil
}

// GenerateActivityTitle generates a default title for an activity
func (s *ActivityService) GenerateActivityTitle(activityType models.ActivityType) string {
	switch activityType {
	case models.ActivityTypeMeeting:
		return "Meeting"
	case models.ActivityTypeWorkSession:
		return "Work Session"
	case models.ActivityTypeCall:
		return "Call"
	default:
		return "Activity"
	}
}
