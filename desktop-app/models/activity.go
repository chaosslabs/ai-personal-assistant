package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityTypeMeeting     ActivityType = "meeting"
	ActivityTypeWorkSession ActivityType = "work_session"
	ActivityTypeCall        ActivityType = "call"
	ActivityTypeOther       ActivityType = "other"
)

// ActivityStatus represents the current status of an activity
type ActivityStatus string

const (
	ActivityStatusActive     ActivityStatus = "active"
	ActivityStatusRecording  ActivityStatus = "recording"
	ActivityStatusProcessing ActivityStatus = "processing"
	ActivityStatusCompleted  ActivityStatus = "completed"
	ActivityStatusFailed     ActivityStatus = "failed"
)

// Activity represents a user activity session
type Activity struct {
	ID        string         `json:"id" db:"id"`
	UserID    string         `json:"user_id" db:"user_id"`
	Type      ActivityType   `json:"type" db:"type"`
	Title     string         `json:"title" db:"title"`
	StartTime time.Time      `json:"start_time" db:"start_time"`
	EndTime   *time.Time     `json:"end_time,omitempty" db:"end_time"`
	Status    ActivityStatus `json:"status" db:"status"`
	Tags      []string       `json:"tags" db:"tags"`
	Metadata  Metadata       `json:"metadata" db:"metadata"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Metadata contains extensible activity metadata
type Metadata map[string]interface{}

// NewActivity creates a new activity
func NewActivity(userID string, activityType ActivityType, title string) *Activity {
	now := time.Now()
	return &Activity{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      activityType,
		Title:     title,
		StartTime: now,
		Status:    ActivityStatusActive,
		Tags:      make([]string, 0),
		Metadata:  make(Metadata),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Start marks the activity as started and recording
func (a *Activity) Start() {
	a.Status = ActivityStatusRecording
	a.UpdatedAt = time.Now()
}

// Complete marks the activity as completed
func (a *Activity) Complete() {
	now := time.Now()
	a.Status = ActivityStatusCompleted
	a.EndTime = &now
	a.UpdatedAt = now
}

// Fail marks the activity as failed
func (a *Activity) Fail() {
	now := time.Now()
	a.Status = ActivityStatusFailed
	a.EndTime = &now
	a.UpdatedAt = now
}

// SetProcessing marks the activity as processing
func (a *Activity) SetProcessing() {
	a.Status = ActivityStatusProcessing
	a.UpdatedAt = time.Now()
}

// Duration returns the duration of the activity
func (a *Activity) Duration() time.Duration {
	if a.EndTime != nil {
		return a.EndTime.Sub(a.StartTime)
	}
	return time.Since(a.StartTime)
}

// AddTag adds a tag to the activity
func (a *Activity) AddTag(tag string) {
	// Check if tag already exists
	for _, existing := range a.Tags {
		if existing == tag {
			return
		}
	}
	a.Tags = append(a.Tags, tag)
	a.UpdatedAt = time.Now()
}

// RemoveTag removes a tag from the activity
func (a *Activity) RemoveTag(tag string) {
	for i, existing := range a.Tags {
		if existing == tag {
			a.Tags = append(a.Tags[:i], a.Tags[i+1:]...)
			a.UpdatedAt = time.Now()
			return
		}
	}
}

// SetMetadata sets a metadata key-value pair
func (a *Activity) SetMetadata(key string, value interface{}) {
	a.Metadata[key] = value
	a.UpdatedAt = time.Now()
}

// GetMetadata gets a metadata value by key
func (a *Activity) GetMetadata(key string) (interface{}, bool) {
	value, exists := a.Metadata[key]
	return value, exists
}

// TagsToJSON converts tags to JSON string for database storage
func (a *Activity) TagsToJSON() (string, error) {
	data, err := json.Marshal(a.Tags)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// TagsFromJSON parses tags from JSON string
func (a *Activity) TagsFromJSON(jsonStr string) error {
	if jsonStr == "" {
		a.Tags = make([]string, 0)
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), &a.Tags)
}

// MetadataToJSON converts metadata to JSON string for database storage
func (a *Activity) MetadataToJSON() (string, error) {
	data, err := json.Marshal(a.Metadata)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MetadataFromJSON parses metadata from JSON string
func (a *Activity) MetadataFromJSON(jsonStr string) error {
	if jsonStr == "" {
		a.Metadata = make(Metadata)
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), &a.Metadata)
}

// IsDeleted checks if the activity has been soft deleted
func (a *Activity) IsDeleted() bool {
	return a.DeletedAt != nil
}

// SoftDelete marks the activity as deleted with the current timestamp
func (a *Activity) SoftDelete() {
	now := time.Now()
	a.DeletedAt = &now
	a.UpdatedAt = now
}

// Restore removes the soft delete flag from the activity
func (a *Activity) Restore() {
	a.DeletedAt = nil
	a.UpdatedAt = time.Now()
}