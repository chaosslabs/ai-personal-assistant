-- Personal Assistant Database Schema
-- SQLite database for activity-based audio transcription
-- All timestamps stored as INTEGER (Unix seconds) unless specified

PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

-- Users table - single user per installation
CREATE TABLE users (
    id TEXT PRIMARY KEY,                -- UUID generated on first launch
    username TEXT NOT NULL,             -- System username or user-provided
    settings TEXT NOT NULL DEFAULT '{}', -- JSON: audio device preferences, whisper model, etc.
    created_at INTEGER NOT NULL,        -- Unix timestamp
    updated_at INTEGER NOT NULL         -- Unix timestamp
);

-- Activities table - central concept for organizing recordings
CREATE TABLE activities (
    id TEXT PRIMARY KEY,                -- UUID
    user_id TEXT NOT NULL,              -- Foreign key to users.id
    type TEXT NOT NULL CHECK (type IN ('meeting', 'work_session', 'call', 'other')),
    title TEXT NOT NULL,                -- User provided or auto-generated
    start_time INTEGER NOT NULL,        -- Unix timestamp
    end_time INTEGER,                   -- Unix timestamp, NULL for active activities
    status TEXT NOT NULL CHECK (status IN ('active', 'recording', 'processing', 'completed', 'failed')),
    tags TEXT NOT NULL DEFAULT '[]',    -- JSON array of tags
    metadata TEXT NOT NULL DEFAULT '{}', -- JSON for extensibility
    created_at INTEGER NOT NULL,        -- Unix timestamp
    updated_at INTEGER NOT NULL,        -- Unix timestamp
    deleted_at INTEGER,                 -- Unix timestamp for soft delete, NULL if not deleted
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Audio recordings table - links to activities
CREATE TABLE audio_recordings (
    id TEXT PRIMARY KEY,                -- UUID
    user_id TEXT NOT NULL,              -- Foreign key to users.id
    activity_id TEXT NOT NULL,          -- Foreign key to activities.id
    file_path TEXT NOT NULL,            -- Relative path to app data directory
    device_info TEXT NOT NULL DEFAULT '{}', -- JSON: device name, sample rate, channels, etc.
    status TEXT NOT NULL CHECK (status IN ('recording', 'completed', 'failed')),
    duration REAL,                      -- Duration in seconds, calculated after recording
    file_size INTEGER,                  -- File size in bytes
    config TEXT NOT NULL DEFAULT '{}', -- JSON: recording settings used
    created_at INTEGER NOT NULL,        -- Unix timestamp
    updated_at INTEGER NOT NULL,        -- Unix timestamp
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE
);

-- Transcript chunks table - processed audio segments
CREATE TABLE transcript_chunks (
    id TEXT PRIMARY KEY,                -- UUID
    user_id TEXT NOT NULL,              -- Foreign key to users.id  
    activity_id TEXT NOT NULL,          -- Foreign key to activities.id
    audio_recording_id TEXT NOT NULL,   -- Foreign key to audio_recordings.id
    text TEXT NOT NULL,                 -- The transcribed text
    start_time REAL NOT NULL,           -- Start time in seconds from activity start
    end_time REAL NOT NULL,             -- End time in seconds from activity start
    speaker TEXT,                       -- Speaker identification (optional)
    confidence REAL,                    -- Confidence score 0-1 from Whisper
    language TEXT,                      -- Detected language code (e.g., 'en', 'es')
    created_at INTEGER NOT NULL,        -- Unix timestamp
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
    FOREIGN KEY (audio_recording_id) REFERENCES audio_recordings(id) ON DELETE CASCADE
);

-- Indexes for common query patterns
CREATE INDEX idx_activities_user_id ON activities(user_id);
CREATE INDEX idx_activities_status ON activities(status);
CREATE INDEX idx_activities_type ON activities(type);
CREATE INDEX idx_activities_start_time ON activities(start_time);
CREATE INDEX idx_activities_user_start ON activities(user_id, start_time);

CREATE INDEX idx_audio_recordings_user_id ON audio_recordings(user_id);
CREATE INDEX idx_audio_recordings_activity_id ON audio_recordings(activity_id);
CREATE INDEX idx_audio_recordings_status ON audio_recordings(status);

CREATE INDEX idx_transcript_chunks_user_id ON transcript_chunks(user_id);
CREATE INDEX idx_transcript_chunks_activity_id ON transcript_chunks(activity_id);
CREATE INDEX idx_transcript_chunks_audio_recording_id ON transcript_chunks(audio_recording_id);
CREATE INDEX idx_transcript_chunks_start_time ON transcript_chunks(start_time);

-- Database metadata for migrations
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL        -- Unix timestamp
);

-- Note: FTS5 not available in this SQLite build, using LIKE for text search instead
-- Full-text search can be implemented at application level

-- Insert initial schema version
INSERT INTO schema_migrations (version, applied_at) VALUES (1, strftime('%s', 'now'));