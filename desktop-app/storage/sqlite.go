package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/platformlabs-co/personal-assist/database"
	"github.com/platformlabs-co/personal-assist/models"
)

// SQLiteStorage provides SQLite-based storage operations
type SQLiteStorage struct {
	db *database.DB
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(db *database.DB) *SQLiteStorage {
	return &SQLiteStorage{db: db}
}

// User operations

// CreateUser creates a new user in the database
func (s *SQLiteStorage) CreateUser(user *models.User) error {
	settings, err := user.SettingsToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize user settings: %w", err)
	}

	query := `
		INSERT INTO users (id, username, settings, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`

	_, err = s.db.Exec(query,
		user.ID,
		user.Username,
		settings,
		user.CreatedAt.Unix(),
		user.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID
func (s *SQLiteStorage) GetUser(id string) (*models.User, error) {
	query := `SELECT id, username, settings, created_at, updated_at FROM users WHERE id = ?`

	var user models.User
	var settingsJSON string
	var createdAt, updatedAt int64

	err := s.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&settingsJSON,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Parse settings
	if err := user.SettingsFromJSON(settingsJSON); err != nil {
		return nil, fmt.Errorf("failed to parse user settings: %w", err)
	}

	user.CreatedAt = time.Unix(createdAt, 0)
	user.UpdatedAt = time.Unix(updatedAt, 0)

	return &user, nil
}

// UpdateUser updates an existing user
func (s *SQLiteStorage) UpdateUser(user *models.User) error {
	settings, err := user.SettingsToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize user settings: %w", err)
	}

	query := `
		UPDATE users 
		SET username = ?, settings = ?, updated_at = ?
		WHERE id = ?`

	result, err := s.db.Exec(query,
		user.Username,
		settings,
		time.Now().Unix(),
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	return nil
}

// GetFirstUser retrieves the first user (for single-user installations)
func (s *SQLiteStorage) GetFirstUser() (*models.User, error) {
	query := `SELECT id, username, settings, created_at, updated_at FROM users ORDER BY created_at ASC LIMIT 1`

	var user models.User
	var settingsJSON string
	var createdAt, updatedAt int64

	err := s.db.QueryRow(query).Scan(
		&user.ID,
		&user.Username,
		&settingsJSON,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no users found")
		}
		return nil, fmt.Errorf("failed to get first user: %w", err)
	}

	// Parse settings
	if err := user.SettingsFromJSON(settingsJSON); err != nil {
		return nil, fmt.Errorf("failed to parse user settings: %w", err)
	}

	user.CreatedAt = time.Unix(createdAt, 0)
	user.UpdatedAt = time.Unix(updatedAt, 0)

	return &user, nil
}

// Activity operations

// CreateActivity creates a new activity in the database
func (s *SQLiteStorage) CreateActivity(activity *models.Activity) error {
	tags, err := activity.TagsToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize activity tags: %w", err)
	}

	metadata, err := activity.MetadataToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize activity metadata: %w", err)
	}

	query := `
		INSERT INTO activities (id, user_id, type, title, start_time, end_time, status, tags, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var endTime *int64
	if activity.EndTime != nil {
		t := activity.EndTime.Unix()
		endTime = &t
	}

	_, err = s.db.Exec(query,
		activity.ID,
		activity.UserID,
		string(activity.Type),
		activity.Title,
		activity.StartTime.Unix(),
		endTime,
		string(activity.Status),
		tags,
		metadata,
		activity.CreatedAt.Unix(),
		activity.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}

	return nil
}

// GetActivity retrieves an activity by ID for any user
func (s *SQLiteStorage) GetActivity(userID, id string) (*models.Activity, error) {
	query := `
		SELECT id, user_id, type, title, start_time, end_time, status, tags, metadata, created_at, updated_at, deleted_at
		FROM activities WHERE user_id = ? AND id = ?`

	var activity models.Activity
	var tagsJSON, metadataJSON string
	var startTime, createdAt, updatedAt int64
	var endTime, deletedAt *int64

	err := s.db.QueryRow(query, userID, id).Scan(
		&activity.ID,
		&activity.UserID,
		&activity.Type,
		&activity.Title,
		&startTime,
		&endTime,
		&activity.Status,
		&tagsJSON,
		&metadataJSON,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("activity not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	// Parse JSON fields
	if err := activity.TagsFromJSON(tagsJSON); err != nil {
		return nil, fmt.Errorf("failed to parse activity tags: %w", err)
	}
	if err := activity.MetadataFromJSON(metadataJSON); err != nil {
		return nil, fmt.Errorf("failed to parse activity metadata: %w", err)
	}

	// Convert timestamps
	activity.StartTime = time.Unix(startTime, 0)
	if endTime != nil {
		t := time.Unix(*endTime, 0)
		activity.EndTime = &t
	}
	if deletedAt != nil {
		t := time.Unix(*deletedAt, 0)
		activity.DeletedAt = &t
	}
	activity.CreatedAt = time.Unix(createdAt, 0)
	activity.UpdatedAt = time.Unix(updatedAt, 0)

	return &activity, nil
}

// UpdateActivity updates an existing activity
func (s *SQLiteStorage) UpdateActivity(activity *models.Activity) error {
	tags, err := activity.TagsToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize activity tags: %w", err)
	}

	metadata, err := activity.MetadataToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize activity metadata: %w", err)
	}

	query := `
		UPDATE activities 
		SET type = ?, title = ?, start_time = ?, end_time = ?, status = ?, tags = ?, metadata = ?, updated_at = ?, deleted_at = ?
		WHERE id = ?`

	var endTime *int64
	if activity.EndTime != nil {
		t := activity.EndTime.Unix()
		endTime = &t
	}

	var deletedAt *int64
	if activity.DeletedAt != nil {
		t := activity.DeletedAt.Unix()
		deletedAt = &t
	}

	result, err := s.db.Exec(query,
		string(activity.Type),
		activity.Title,
		activity.StartTime.Unix(),
		endTime,
		string(activity.Status),
		tags,
		metadata,
		time.Now().Unix(),
		deletedAt,
		activity.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("activity not found: %s", activity.ID)
	}

	return nil
}

// GetActivitiesByUser retrieves activities for a specific user
func (s *SQLiteStorage) GetActivitiesByUser(userID string, limit, offset int) ([]*models.Activity, error) {
	query := `
		SELECT id, user_id, type, title, start_time, end_time, status, tags, metadata, created_at, updated_at, deleted_at
		FROM activities 
		WHERE user_id = ? AND deleted_at IS NULL
		ORDER BY start_time DESC
		LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities: %w", err)
	}
	defer rows.Close()

	var activities []*models.Activity
	for rows.Next() {
		activity := &models.Activity{}
		var tagsJSON, metadataJSON string
		var startTime, createdAt, updatedAt int64
		var endTime, deletedAt *int64

		err := rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.Type,
			&activity.Title,
			&startTime,
			&endTime,
			&activity.Status,
			&tagsJSON,
			&metadataJSON,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}

		// Parse JSON fields
		if err := activity.TagsFromJSON(tagsJSON); err != nil {
			return nil, fmt.Errorf("failed to parse activity tags: %w", err)
		}
		if err := activity.MetadataFromJSON(metadataJSON); err != nil {
			return nil, fmt.Errorf("failed to parse activity metadata: %w", err)
		}

		// Convert timestamps
		activity.StartTime = time.Unix(startTime, 0)
		if endTime != nil {
			t := time.Unix(*endTime, 0)
			activity.EndTime = &t
		}
		if deletedAt != nil {
			t := time.Unix(*deletedAt, 0)
			activity.DeletedAt = &t
		}
		activity.CreatedAt = time.Unix(createdAt, 0)
		activity.UpdatedAt = time.Unix(updatedAt, 0)

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activities: %w", err)
	}

	return activities, nil
}

// SearchTranscripts performs text search on transcript content using LIKE
func (s *SQLiteStorage) SearchTranscripts(userID, query string) ([]*models.TranscriptChunk, error) {
	return s.SearchTranscriptsWithLimit(userID, query, 100, 0)
}

// SearchTranscriptsWithLimit performs text search with pagination
func (s *SQLiteStorage) SearchTranscriptsWithLimit(userID, query string, limit, offset int) ([]*models.TranscriptChunk, error) {
	// Simple LIKE search since FTS5 is not available
	searchQuery := `
		SELECT id, user_id, activity_id, audio_recording_id, text, 
		       start_time, end_time, speaker, confidence, language, created_at
		FROM transcript_chunks 
		WHERE user_id = ? AND text LIKE ?
		ORDER BY start_time
		LIMIT ? OFFSET ?`

	// Add wildcards for LIKE search
	likeQuery := "%" + query + "%"

	rows, err := s.db.Query(searchQuery, userID, likeQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search transcripts: %w", err)
	}
	defer rows.Close()

	return s.scanTranscriptChunks(rows)
}

// scanTranscriptChunks scans transcript chunks from SQL rows
func (s *SQLiteStorage) scanTranscriptChunks(rows *sql.Rows) ([]*models.TranscriptChunk, error) {
	var chunks []*models.TranscriptChunk
	for rows.Next() {
		chunk := &models.TranscriptChunk{}
		var createdAt int64
		var speaker, language sql.NullString
		var confidence sql.NullFloat64

		err := rows.Scan(
			&chunk.ID,
			&chunk.UserID,
			&chunk.ActivityID,
			&chunk.AudioRecordingID,
			&chunk.Text,
			&chunk.StartTime,
			&chunk.EndTime,
			&speaker,
			&confidence,
			&language,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transcript chunk: %w", err)
		}

		// Handle nullable fields
		if speaker.Valid {
			chunk.Speaker = &speaker.String
		}
		if confidence.Valid {
			chunk.Confidence = &confidence.Float64
		}
		if language.Valid {
			chunk.Language = &language.String
		}

		chunk.CreatedAt = time.Unix(createdAt, 0)
		chunks = append(chunks, chunk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transcript chunks: %w", err)
	}

	return chunks, nil
}

// Audio Recording operations

// CreateAudioRecording creates a new audio recording in the database
func (s *SQLiteStorage) CreateAudioRecording(recording *models.AudioRecording) error {
	deviceInfo, err := recording.DeviceInfoToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize device info: %w", err)
	}

	config, err := recording.ConfigToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	query := `
		INSERT INTO audio_recordings (id, user_id, activity_id, file_path, device_info, status, duration, file_size, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = s.db.Exec(query,
		recording.ID,
		recording.UserID,
		recording.ActivityID,
		recording.FilePath,
		deviceInfo,
		string(recording.Status),
		recording.Duration,
		recording.FileSize,
		config,
		recording.CreatedAt.Unix(),
		recording.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to create audio recording: %w", err)
	}

	return nil
}

// GetAudioRecording retrieves an audio recording by ID
func (s *SQLiteStorage) GetAudioRecording(userID, id string) (*models.AudioRecording, error) {
	query := `
		SELECT id, user_id, activity_id, file_path, device_info, status, duration, file_size, config, created_at, updated_at
		FROM audio_recordings WHERE user_id = ? AND id = ?`

	var recording models.AudioRecording
	var deviceInfoJSON, configJSON string
	var createdAt, updatedAt int64
	var duration sql.NullFloat64
	var fileSize sql.NullInt64

	err := s.db.QueryRow(query, userID, id).Scan(
		&recording.ID,
		&recording.UserID,
		&recording.ActivityID,
		&recording.FilePath,
		&deviceInfoJSON,
		&recording.Status,
		&duration,
		&fileSize,
		&configJSON,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("audio recording not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get audio recording: %w", err)
	}

	// Parse JSON fields
	if err := recording.DeviceInfoFromJSON(deviceInfoJSON); err != nil {
		return nil, fmt.Errorf("failed to parse device info: %w", err)
	}
	if err := recording.ConfigFromJSON(configJSON); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Handle nullable fields
	if duration.Valid {
		recording.Duration = &duration.Float64
	}
	if fileSize.Valid {
		recording.FileSize = &fileSize.Int64
	}

	recording.CreatedAt = time.Unix(createdAt, 0)
	recording.UpdatedAt = time.Unix(updatedAt, 0)

	return &recording, nil
}

// UpdateAudioRecording updates an existing audio recording
func (s *SQLiteStorage) UpdateAudioRecording(recording *models.AudioRecording) error {
	deviceInfo, err := recording.DeviceInfoToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize device info: %w", err)
	}

	config, err := recording.ConfigToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	query := `
		UPDATE audio_recordings 
		SET status = ?, duration = ?, file_size = ?, device_info = ?, config = ?, updated_at = ?
		WHERE id = ?`

	result, err := s.db.Exec(query,
		string(recording.Status),
		recording.Duration,
		recording.FileSize,
		deviceInfo,
		config,
		time.Now().Unix(),
		recording.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update audio recording: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("audio recording not found: %s", recording.ID)
	}

	return nil
}

// GetActivityRecordings retrieves all audio recordings for an activity
func (s *SQLiteStorage) GetActivityRecordings(userID, activityID string) ([]*models.AudioRecording, error) {
	query := `
		SELECT id, user_id, activity_id, file_path, device_info, status, duration, file_size, config, created_at, updated_at
		FROM audio_recordings 
		WHERE user_id = ? AND activity_id = ?
		ORDER BY created_at ASC`

	rows, err := s.db.Query(query, userID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audio recordings: %w", err)
	}
	defer rows.Close()

	var recordings []*models.AudioRecording
	for rows.Next() {
		recording := &models.AudioRecording{}
		var deviceInfoJSON, configJSON string
		var createdAt, updatedAt int64
		var duration sql.NullFloat64
		var fileSize sql.NullInt64

		err := rows.Scan(
			&recording.ID,
			&recording.UserID,
			&recording.ActivityID,
			&recording.FilePath,
			&deviceInfoJSON,
			&recording.Status,
			&duration,
			&fileSize,
			&configJSON,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audio recording: %w", err)
		}

		// Parse JSON fields
		if err := recording.DeviceInfoFromJSON(deviceInfoJSON); err != nil {
			return nil, fmt.Errorf("failed to parse device info: %w", err)
		}
		if err := recording.ConfigFromJSON(configJSON); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}

		// Handle nullable fields
		if duration.Valid {
			recording.Duration = &duration.Float64
		}
		if fileSize.Valid {
			recording.FileSize = &fileSize.Int64
		}

		recording.CreatedAt = time.Unix(createdAt, 0)
		recording.UpdatedAt = time.Unix(updatedAt, 0)

		recordings = append(recordings, recording)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audio recordings: %w", err)
	}

	return recordings, nil
}

// Transcript Chunk operations

// SaveTranscriptChunk saves a transcript chunk to the database (alias for CreateTranscriptChunk)
func (s *SQLiteStorage) SaveTranscriptChunk(chunk *models.TranscriptChunk) error {
	return s.CreateTranscriptChunk(chunk)
}

// CreateTranscriptChunk creates a new transcript chunk in the database
func (s *SQLiteStorage) CreateTranscriptChunk(chunk *models.TranscriptChunk) error {
	query := `
		INSERT INTO transcript_chunks (id, user_id, activity_id, audio_recording_id, text, start_time, end_time, speaker, confidence, language, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		chunk.ID,
		chunk.UserID,
		chunk.ActivityID,
		chunk.AudioRecordingID,
		chunk.Text,
		chunk.StartTime,
		chunk.EndTime,
		chunk.Speaker,
		chunk.Confidence,
		chunk.Language,
		chunk.CreatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to create transcript chunk: %w", err)
	}

	return nil
}

// GetActivityTranscripts retrieves all transcript chunks for an activity
func (s *SQLiteStorage) GetActivityTranscripts(userID, activityID string) ([]*models.TranscriptChunk, error) {
	query := `
		SELECT id, user_id, activity_id, audio_recording_id, text, start_time, end_time, speaker, confidence, language, created_at
		FROM transcript_chunks 
		WHERE user_id = ? AND activity_id = ?
		ORDER BY start_time ASC`

	rows, err := s.db.Query(query, userID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transcript chunks: %w", err)
	}
	defer rows.Close()

	return s.scanTranscriptChunks(rows)
}

// GetTranscriptChunksByActivity retrieves all transcript chunks for an activity
func (s *SQLiteStorage) GetTranscriptChunksByActivity(activityID string) ([]*models.TranscriptChunk, error) {
	query := `
		SELECT id, user_id, activity_id, audio_recording_id, text, start_time, end_time, speaker, confidence, language, created_at
		FROM transcript_chunks 
		WHERE activity_id = ?
		ORDER BY start_time ASC`

	rows, err := s.db.Query(query, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transcript chunks: %w", err)
	}
	defer rows.Close()

	return s.scanTranscriptChunks(rows)
}

// GetRecordingTranscripts retrieves transcript chunks for a specific recording
func (s *SQLiteStorage) GetRecordingTranscripts(userID, recordingID string) ([]*models.TranscriptChunk, error) {
	query := `
		SELECT id, user_id, activity_id, audio_recording_id, text, start_time, end_time, speaker, confidence, language, created_at
		FROM transcript_chunks 
		WHERE user_id = ? AND audio_recording_id = ?
		ORDER BY start_time ASC`

	rows, err := s.db.Query(query, userID, recordingID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transcript chunks: %w", err)
	}
	defer rows.Close()

	return s.scanTranscriptChunks(rows)
}

// GetTranscriptChunksByAudioRecording retrieves transcript chunks for a specific recording
func (s *SQLiteStorage) GetTranscriptChunksByAudioRecording(audioRecordingID string) ([]*models.TranscriptChunk, error) {
	query := `
		SELECT id, user_id, activity_id, audio_recording_id, text, start_time, end_time, speaker, confidence, language, created_at
		FROM transcript_chunks 
		WHERE audio_recording_id = ?
		ORDER BY start_time ASC`

	rows, err := s.db.Query(query, audioRecordingID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transcript chunks: %w", err)
	}
	defer rows.Close()

	return s.scanTranscriptChunks(rows)
}

// DeleteActivity deletes an activity and all associated data
func (s *SQLiteStorage) DeleteActivity(activityID string) error {
	// SQLite will cascade delete due to foreign key constraints
	query := `DELETE FROM activities WHERE id = ?`

	result, err := s.db.Exec(query, activityID)
	if err != nil {
		return fmt.Errorf("failed to delete activity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("activity not found: %s", activityID)
	}

	return nil
}

// DeleteAudioRecording deletes an audio recording
func (s *SQLiteStorage) DeleteAudioRecording(recordingID string) error {
	query := `DELETE FROM audio_recordings WHERE id = ?`

	result, err := s.db.Exec(query, recordingID)
	if err != nil {
		return fmt.Errorf("failed to delete audio recording: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("audio recording not found: %s", recordingID)
	}

	return nil
}
