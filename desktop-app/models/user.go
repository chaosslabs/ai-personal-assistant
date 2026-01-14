package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User represents a user of the personal assistant application
type User struct {
	ID        string          `json:"id" db:"id"`
	Username  string          `json:"username" db:"username"`
	Settings  UserSettings    `json:"settings" db:"settings"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

// UserSettings contains user preferences and configuration
type UserSettings struct {
	PreferredAudioDevice string  `json:"preferred_audio_device,omitempty"`
	WhisperModel        string  `json:"whisper_model,omitempty"`
	AutoStartRecording  bool    `json:"auto_start_recording"`
	TranscriptionLanguage string `json:"transcription_language,omitempty"`
	AudioQuality        string  `json:"audio_quality,omitempty"`
	StorageLocation     string  `json:"storage_location,omitempty"`
}

// NewUser creates a new user with default settings
func NewUser(username string) *User {
	now := time.Now()
	return &User{
		ID:       uuid.New().String(),
		Username: username,
		Settings: UserSettings{
			WhisperModel:       "whisper-small",
			AutoStartRecording: false,
			AudioQuality:      "high",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateSettings updates the user's settings
func (u *User) UpdateSettings(settings UserSettings) {
	u.Settings = settings
	u.UpdatedAt = time.Now()
}

// ToJSON converts the user to JSON string for database storage
func (u *User) ToJSON() (string, error) {
	data, err := json.Marshal(u)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SettingsToJSON converts settings to JSON string for database storage
func (u *User) SettingsToJSON() (string, error) {
	data, err := json.Marshal(u.Settings)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SettingsFromJSON parses settings from JSON string
func (u *User) SettingsFromJSON(jsonStr string) error {
	if jsonStr == "" {
		u.Settings = UserSettings{}
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), &u.Settings)
}