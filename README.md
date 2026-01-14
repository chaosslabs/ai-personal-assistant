# Memoria (AI Personal Assistant)

A macOS desktop application for automatic meeting transcription and activity tracking. Built with Go, Wails, and SQLite for local-first privacy.

(Hopefully we can release to other distros in the future.)



## Project Structure

```
personal-assist/
├── CLAUDE.md           # Development plan and architecture docs
├── desktop-app/        # Main Wails desktop application
│   ├── models/         # Data models (User, Activity, AudioRecording, TranscriptChunk)
│   ├── database/       # SQLite schema and migrations
│   ├── storage/        # Data access layer
│   ├── services/       # Business logic layer
│   ├── frontend/       # Web UI (HTML/JS)
│   └── main.go         # Application entry point
└── README.md           # This file
```

## Prerequisites

- **Go 1.23+** - [Install Go](https://golang.org/doc/install)
- **Wails v2** - Install with: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Node.js** - For frontend dependencies

## Quick Start

### Development Mode (Hot Reload)
```bash
cd desktop-app
wails dev
```

### Production Build
```bash
cd desktop-app
wails build
```

### Run Built Application
```bash
cd desktop-app
"./build/bin/Personal Assist.app/Contents/MacOS/personal-assist"
```

## What It Does

### Current Features (MVP Foundation)
- ✅ **Local SQLite Database** - All data stored in `~/Library/Application Support/personal-assist/`
- ✅ **User Management** - Auto-created user on first launch
- ✅ **Activity Tracking** - Organize recordings by meeting/work session/call
- ✅ **Audio Recording Models** - Ready for audio capture integration
- ✅ **Transcript Storage** - Structured storage for processed audio
- ✅ **File Management** - Organized directory structure for recordings
- ✅ **Search & Export** - Text search and multiple export formats
- ✅ **Desktop UI** - Cross-platform Wails application with system tray

### Testing the Core System
The app includes built-in test methods:
- `GetCurrentUser()` - View auto-created user
- `CreateTestActivity()` - Create sample activity
- `ListActivities()` - View all activities
- `GetSystemInfo()` - Database paths and status

### Coming Next
- 🔄 **Audio Recording** - Capture system/microphone audio
- 🔄 **Whisper Integration** - OpenAI Whisper API for transcription
- 🔄 **Real-time Processing** - Background transcription pipeline
- 🔄 **Enhanced UI** - Rich interface for managing recordings
- 🔄 **Meeting Detection** - Auto-start based on meeting apps

## Architecture

**Tech Stack:**
- **Backend**: Go with Wails v2 framework
- **Database**: SQLite with WAL mode
- **Frontend**: Vanilla JS/HTML (upgradeable to React)
- **Storage**: Local-first in user's Application Support directory

**Data Models:**
- **User**: Settings, preferences, UUID-based
- **Activity**: Central concept - meetings, calls, work sessions
- **AudioRecording**: Linked to activities, file metadata
- **TranscriptChunk**: Processed audio segments with timestamps

**Storage Structure:**
```
~/Library/Application Support/memoria/
├── personal-assist.db              # Main SQLite database
├── activities/                     # Activity-specific files
│   └── {activity-id}/
│       └── audio/
│           ├── recording_001.m4a
│           └── recording_002.m4a
└── models/                         # Downloaded Whisper models
    ├── whisper-tiny.bin
    └── whisper-small.bin
```
## Privacy

- **Local-first**: All data stored locally, no cloud dependencies
- **User control**: Complete data ownership and deletion
- **Optional cloud**: Future OpenAI Whisper API integration with user consent
- **Secure storage**: Database encryption and keychain integration planned