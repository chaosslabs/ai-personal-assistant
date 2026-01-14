# Settings Implementation Roadmap

This document tracks the implementation status of all settings in the Memoria (Personal Assist) desktop application.

## Implementation Status Legend
- ✅ **Fully Implemented** - Frontend UI + Backend persistence + Real functionality
- 🔄 **Partially Implemented** - Frontend UI exists, backend may be missing or incomplete
- ❌ **Not Implemented** - Only UI mockup, no backend functionality
- 🚫 **Disabled/Placeholder** - Intentionally disabled pending other features

---

## Appearance Settings

### Theme Selection ✅ **Fully Implemented**
- **Frontend**: Theme toggle UI with Light/Dark/System options
- **Backend**: Uses React context with localStorage persistence
- **Status**: Working - switches themes properly, persists across sessions

---

## Privacy & Security Settings

### Data Encryption 🔄 **Partially Implemented**
- **Frontend**: Toggle switch in UI
- **Backend**: ❌ No actual encryption implementation
- **Status**: UI only - needs backend encryption service

### Data Retention Period 🔄 **Partially Implemented**
- **Frontend**: Dropdown with 30/90/180/365 days and "never delete"
- **Backend**: ❌ No automatic cleanup implementation
- **Status**: UI only - needs cleanup service and scheduled tasks

### Auto-delete old activities 🔄 **Partially Implemented**
- **Frontend**: Toggle switch in UI
- **Backend**: ❌ No automatic deletion service
- **Status**: UI only - needs background cleanup service

---

## Recording Settings

### Audio Input Device ✅ **Fully Implemented**
- **Frontend**: Dropdown with real device list + refresh button
- **Backend**: `GetAudioDevices()` API, real device enumeration
- **Status**: Working - shows real audio devices, can refresh list

### Audio Quality ✅ **Fully Implemented**
- **Frontend**: Dropdown with Low/Medium/High/Lossless options
- **Backend**: `UserSettings.AudioQuality` field with persistence
- **Status**: Working - saves to backend, loads on startup

### Auto-transcribe recordings 🚫 **Disabled/Placeholder**
- **Frontend**: Toggle switch in UI
- **Backend**: 🚫 Transcription service disabled (Phase 5)
- **Status**: Placeholder - will be enabled in transcription phase

### Background recording 🔄 **Partially Implemented**
- **Frontend**: Toggle switch in UI
- **Backend**: ❌ No background recording implementation
- **Status**: UI only - needs system-level background recording service

---

## Notification Settings

### Recording notifications ❌ **Not Implemented**
- **Frontend**: Toggle switch in UI
- **Backend**: ❌ No notification system
- **Status**: UI only - needs macOS notification integration

### Transcript completion ❌ **Not Implemented**
- **Frontend**: Toggle switch in UI
- **Backend**: ❌ No notification system + transcription disabled
- **Status**: UI only - depends on transcription + notification system

### Daily summary ❌ **Not Implemented**
- **Frontend**: Toggle switch in UI
- **Backend**: ❌ No summary generation or notification system
- **Status**: UI only - needs AI summary service + notifications

---

## Storage Management

### Storage Usage Display ✅ **Fully Implemented**
- **Frontend**: Real-time storage stats with refresh button
- **Backend**: `GetSystemInfo()` with disk_usage_bytes/mb
- **Status**: Working - shows real storage usage, activity counts

### Storage Location 🔄 **Partially Implemented**
- **Frontend**: Display path + "Change" button
- **Backend**: ✅ Loads real path, ❌ Change functionality not implemented
- **Status**: Shows real path, but can't change location

### Maximum Storage Size 🔄 **Partially Implemented**
- **Frontend**: Dropdown with 5/10/25/50 GB and "Unlimited"
- **Backend**: ❌ No storage quota enforcement
- **Status**: UI only - needs storage quota monitoring service

### Auto-cleanup storage 🔄 **Partially Implemented**
- **Frontend**: Toggle switch + "Clean Old Data" button
- **Backend**: ❌ No automated cleanup, ✅ Analysis capability
- **Status**: Can analyze old data, but no actual cleanup implementation

### Export Data ❌ **Not Implemented**
- **Frontend**: Button in UI
- **Backend**: ❌ No export functionality
- **Status**: Button exists but has no click handler - needs data export service

### Clean Old Data 🔄 **Partially Implemented**
- **Frontend**: Button in UI with click handler
- **Backend**: ✅ Can analyze old activities, ❌ No actual deletion
- **Status**: Analyzes activities older than 30 days, logs count, but doesn't delete

### Clear All Data ❌ **Not Implemented**
- **Frontend**: Red destructive button in UI
- **Backend**: ❌ No bulk deletion functionality
- **Status**: Button exists but has no click handler - needs confirmation dialog + bulk delete API

---

## Advanced Settings

### Debug Mode ❌ **Not Implemented**
- **Frontend**: Toggle switch in UI
- **Backend**: ❌ No debug logging configuration
- **Status**: UI only - needs logging level configuration

### Hardware Acceleration 🚫 **Disabled/Placeholder**
- **Frontend**: Toggle switch (default checked)
- **Backend**: 🚫 Depends on transcription service (Phase 5)
- **Status**: Placeholder - will be relevant for GPU transcription

### Rebuild Database ❌ **Not Implemented**
- **Frontend**: Button with database icon in UI
- **Backend**: ❌ No database rebuild functionality
- **Status**: Button exists but has no click handler - needs database migration/rebuild service

### Reset to Defaults ❌ **Not Implemented**
- **Frontend**: Button with settings icon in UI
- **Backend**: ❌ No settings reset functionality
- **Status**: Button exists but has no click handler - needs default settings restoration

---

## Button Functionality Status Summary

### Storage Management Buttons (from screenshot)
| Button | Status | Click Handler | Backend API | Notes |
|--------|--------|---------------|-------------|-------|
| **Export Data** | ❌ Not Implemented | ❌ No | ❌ No | Button renders but does nothing |
| **Clean Old Data** | 🔄 Partial | ✅ Yes | 🔄 Analysis only | Finds old activities, logs count, no deletion |
| **Clear All Data** | ❌ Not Implemented | ❌ No | ❌ No | Destructive red button, no handler |

### Advanced Settings Buttons (from screenshot)
| Button | Status | Click Handler | Backend API | Notes |
|--------|--------|---------------|-------------|-------|
| **Rebuild Database** | ❌ Not Implemented | ❌ No | ❌ No | Database icon, no functionality |
| **Reset to Defaults** | ❌ Not Implemented | ❌ No | ❌ No | Settings icon, no functionality |

### Advanced Settings Toggles (from screenshot)
| Toggle | Status | State Persistence | Backend Effect | Notes |
|--------|--------|------------------|----------------|-------|
| **Debug Mode** | ❌ Not Implemented | ❌ No | ❌ No | Switch renders but no debug system |
| **Hardware Acceleration** | 🚫 Placeholder | ❌ No | 🚫 Disabled | Waiting on transcription (Phase 5) |

---

## Implementation Priority Recommendations

### High Priority (Core Functionality)
1. **Storage Location Change** - Users need to control where data is stored
2. **Data Encryption** - Privacy-first app needs real encryption
3. **Storage Quota Enforcement** - Prevent disk space issues
4. **Export Data** - Essential for data portability
5. **Clear All Data** - Users need data deletion capability

### Medium Priority (User Experience)
1. **Notification System** - Recording start/stop notifications
2. **Auto-cleanup Service** - Automated old data removal
3. **Background Recording** - Continue recording when minimized
4. **Debug Mode** - Essential for troubleshooting

### Low Priority (Nice to Have)
1. **Daily Summary** - Depends on AI summary service
2. **Reset to Defaults** - Admin convenience feature
3. **Rebuild Database** - Maintenance feature
4. **Hardware Acceleration** - Optimization for transcription

### Blocked (Waiting on Other Features)
1. **Auto-transcribe** - Waiting on Phase 5: Transcription
2. **Transcript notifications** - Waiting on transcription + notifications
3. **Hardware Acceleration** - Waiting on GPU transcription implementation

---

## Backend API Gaps

### Missing Backend Functions Needed:
```go
// Storage Management
func (a *App) ChangeStorageLocation(newPath string) error
func (a *App) ExportUserData() (string, error) // Returns export file path
func (a *App) ClearAllUserData() error
func (a *App) CleanupOldData(olderThanDays int) error

// Settings Management
func (a *App) ResetSettingsToDefaults() error
func (a *App) GetDefaultSettings() (*models.UserSettings, error)

// System Management
func (a *App) SetDebugMode(enabled bool) error
func (a *App) RebuildDatabase() error

// Notifications (macOS integration)
func (a *App) SendNotification(title, body string) error
func (a *App) ConfigureNotifications(settings NotificationSettings) error

// Storage Quotas
func (a *App) SetStorageQuota(maxBytes int64) error
func (a *App) GetStorageQuota() (int64, error)
func (a *App) IsStorageQuotaExceeded() (bool, error)
```

### Missing Services Needed:
- **EncryptionService** - AES encryption for sensitive data
- **NotificationService** - macOS notification center integration
- **CleanupService** - Background data cleanup scheduler
- **ExportService** - Data export in various formats (JSON, CSV)
- **StorageService** - Quota monitoring and enforcement

---

## Testing Status

### Tested and Working ✅
- Theme switching
- Audio device selection and refresh
- Real storage usage display
- Settings persistence (audioDevice, audioQuality)

### Needs Testing 🧪
- Settings loading on app restart
- Error handling for failed backend calls
- Storage refresh functionality
- Settings validation and limits

### Cannot Test (Not Implemented) ❌
- All notification features
- Data encryption/decryption
- Storage quota enforcement
- Data export/import
- Cleanup services

---

## Next Steps for Full Settings Implementation

1. **Implement Core Storage APIs** (ChangeStorageLocation, ExportData, ClearAllData)
2. **Add Encryption Service** for privacy compliance
3. **Create Notification Service** for macOS integration
4. **Build Cleanup Service** with scheduled tasks
5. **Add Storage Quota System** with monitoring
6. **Implement Settings Reset** functionality
7. **Add Debug/Logging Configuration**
8. **Create Data Export/Import** system

This roadmap will be updated as features are implemented and new requirements are discovered.