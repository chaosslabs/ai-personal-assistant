package services

import (
	"github.com/platformlabs-co/personal-assist/logger"
	"github.com/platformlabs-co/personal-assist/services/coreaudio"
)

// SystemCapabilities represents the recording capabilities of the system
type SystemCapabilities struct {
	MacOSVersion           string   `json:"macos_version"`
	SupportsCoreAudioTaps  bool     `json:"supports_core_audio_taps"`
	SupportsScreenCapture  bool     `json:"supports_screen_capture"`
	AvailableRecordingModes []string `json:"available_recording_modes"`
	RecommendedMode        string   `json:"recommended_mode"`
	RequiresPermissions    bool     `json:"requires_permissions"`
	PermissionMessage      string   `json:"permission_message,omitempty"`
}

// GetSystemCapabilities returns the audio recording capabilities of the system
func GetSystemCapabilities() SystemCapabilities {
	version := coreaudio.GetMacOSVersion()

	capabilities := SystemCapabilities{
		MacOSVersion:           version.String(),
		SupportsCoreAudioTaps:  version.SupportsCoreAudioTaps(),
		SupportsScreenCapture:  version.SupportsScreenCaptureKit(),
		AvailableRecordingModes: []string{"microphone"},
		RecommendedMode:        "microphone",
		RequiresPermissions:    false,
	}

	logger.WithFields(map[string]interface{}{
		"macos_version":          version.String(),
		"supports_core_audio":    capabilities.SupportsCoreAudioTaps,
		"supports_screen_capture": capabilities.SupportsScreenCapture,
	}).Info("Determining system capabilities")

	// Add mixed mode if Core Audio Taps are available (macOS 14.2+)
	if capabilities.SupportsCoreAudioTaps {
		capabilities.AvailableRecordingModes = append(capabilities.AvailableRecordingModes, "mixed")
		capabilities.RecommendedMode = "mixed"
		capabilities.RequiresPermissions = true
		capabilities.PermissionMessage = "Recording both your microphone and system audio requires Screen Recording permission. This allows capturing audio from meetings and calls."
		logger.Info("Core Audio Taps available - mixed mode enabled")
	} else if capabilities.SupportsScreenCapture {
		// ScreenCaptureKit is available but not Core Audio Taps (macOS 12.3 - 14.1)
		capabilities.PermissionMessage = "System audio capture requires macOS 14.2 or later. Please update macOS to record both sides of calls."
		logger.Info("ScreenCaptureKit available but Core Audio Taps not available - upgrade recommended")
	} else {
		// Older macOS version (< 12.3)
		capabilities.PermissionMessage = "System audio capture requires macOS 14.2 or later. Currently only microphone recording is available."
		logger.Warn("System audio capture not available - macOS too old")
	}

	return capabilities
}

// RecordingModeInfo provides information about a specific recording mode
type RecordingModeInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	Available    bool   `json:"available"`
	Recommended  bool   `json:"recommended"`
	Requirements string `json:"requirements,omitempty"`
}

// GetRecordingModes returns detailed information about all recording modes
func GetRecordingModes() []RecordingModeInfo {
	capabilities := GetSystemCapabilities()

	modes := []RecordingModeInfo{
		{
			ID:          "microphone",
			Name:        "Microphone Only",
			Description: "Record only your voice from the microphone",
			Icon:        "microphone",
			Available:   true,
			Recommended: capabilities.RecommendedMode == "microphone",
		},
		{
			ID:          "mixed",
			Name:        "Mixed Audio",
			Description: "Record both your microphone and system audio (recommended for calls)",
			Icon:        "microphone-speaker",
			Available:   capabilities.SupportsCoreAudioTaps,
			Recommended: capabilities.RecommendedMode == "mixed",
			Requirements: func() string {
				if !capabilities.SupportsCoreAudioTaps {
					return "Requires macOS 14.2 or later and Screen Recording permission"
				}
				return "Requires Screen Recording permission"
			}(),
		},
	}

	logger.WithField("mode_count", len(modes)).Info("Retrieved recording modes")
	return modes
}
