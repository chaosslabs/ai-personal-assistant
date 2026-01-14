package services

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/platformlabs-co/personal-assist/logger"
	"github.com/platformlabs-co/personal-assist/models"
	"github.com/platformlabs-co/personal-assist/services/coreaudio"
)

// AudioRecorder handles the actual audio recording using ffmpeg
type AudioRecorder struct {
	activeRecordings map[string]*RecordingProcess
	mutex           sync.RWMutex
}

// RecordingProcess represents an active recording process
type RecordingProcess struct {
	ID              string
	FilePath        string
	Config          models.RecordingConfig
	Device          models.AudioDeviceInfo
	Process         *exec.Cmd
	CoreAudioTapRec *coreaudio.MixedAudioRecorder // For Core Audio Taps recording
	StartTime       time.Time
	IsActive        bool
	UseCoreAudioTap bool // Whether using Core Audio Taps instead of ffmpeg
}

// NewAudioRecorder creates a new audio recorder
func NewAudioRecorder() *AudioRecorder {
	return &AudioRecorder{
		activeRecordings: make(map[string]*RecordingProcess),
	}
}

// ListAudioDevices lists available audio input devices using ffmpeg
func (r *AudioRecorder) ListAudioDevices() ([]models.AudioDeviceInfo, error) {
	logger.Info("Listing available audio input devices")
	
	// Run ffmpeg to list available AVFoundation devices
	cmd := exec.Command("ffmpeg", "-f", "avfoundation", "-list_devices", "true", "-i", "")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// This is expected - ffmpeg returns an error when listing devices, but still outputs the list
		logger.WithField("output", string(output)).Info("ffmpeg device list output (error expected)")
	}
	
	devices := r.parseAudioDevices(string(output))
	logger.WithField("device_count", len(devices)).Info("Found audio input devices")
	
	return devices, nil
}

// parseAudioDevices parses the ffmpeg device list output
func (r *AudioRecorder) parseAudioDevices(output string) []models.AudioDeviceInfo {
	var devices []models.AudioDeviceInfo

	// Look for both audio input and output devices in the ffmpeg output
	// Format: [AVFoundation indev @ 0x...] [0] Built-in Microphone
	// Format: [AVFoundation outdev @ 0x...] [0] Built-in Output
	lines := strings.Split(output, "\n")
	audioInputSection := false
	audioOutputSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check if we're in the audio input devices section
		if strings.Contains(line, "AVFoundation audio devices:") {
			audioInputSection = true
			audioOutputSection = false
			continue
		}

		// Check if we're in the audio output devices section
		if strings.Contains(line, "AVFoundation output devices:") {
			audioInputSection = false
			audioOutputSection = true
			continue
		}

		// Check if we've moved to video devices section
		if strings.Contains(line, "AVFoundation video devices:") {
			audioInputSection = false
			audioOutputSection = false
			continue
		}

		// Parse audio input device lines
		if audioInputSection && strings.Contains(line, "AVFoundation indev") {
			// Extract device index and name
			// Pattern: [AVFoundation indev @ 0x...] [0] Device Name
			re := regexp.MustCompile(`\[AVFoundation indev[^\]]+\]\s+\[(\d+)\]\s+(.+)`)
			matches := re.FindStringSubmatch(line)

			if len(matches) == 3 {
				deviceIndex := matches[1]
				deviceName := matches[2]

				// Convert index to integer
				if index, err := strconv.Atoi(deviceIndex); err == nil {
					device := models.AudioDeviceInfo{
						Name:       deviceName,
						DeviceID:   fmt.Sprintf("input_%d", index),
						SampleRate: 44100, // Default sample rate
						Channels:   2,     // Default stereo
						DeviceType: "microphone",
					}
					devices = append(devices, device)

					logger.WithFields(map[string]interface{}{
						"device_index": index,
						"device_name":  deviceName,
						"device_type":  "microphone",
					}).Info("Found audio input device")
				}
			}
		}

		// Parse audio output device lines for system audio capture
		if audioOutputSection && strings.Contains(line, "AVFoundation outdev") {
			// Extract device index and name
			// Pattern: [AVFoundation outdev @ 0x...] [0] Device Name
			re := regexp.MustCompile(`\[AVFoundation outdev[^\]]+\]\s+\[(\d+)\]\s+(.+)`)
			matches := re.FindStringSubmatch(line)

			if len(matches) == 3 {
				deviceIndex := matches[1]
				deviceName := matches[2]

				// Convert index to integer
				if index, err := strconv.Atoi(deviceIndex); err == nil {
					device := models.AudioDeviceInfo{
						Name:       fmt.Sprintf("System Audio (%s)", deviceName),
						DeviceID:   fmt.Sprintf("output_%d", index),
						SampleRate: 44100, // Default sample rate
						Channels:   2,     // Default stereo
						DeviceType: "system",
					}
					devices = append(devices, device)

					logger.WithFields(map[string]interface{}{
						"device_index": index,
						"device_name":  deviceName,
						"device_type":  "system",
					}).Info("Found audio output device for system capture")
				}
			}
		}
	}

	return devices
}

// GetBuiltInMicrophone tries to find the built-in microphone device
func (r *AudioRecorder) GetBuiltInMicrophone() (*models.AudioDeviceInfo, error) {
	devices, err := r.ListAudioDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to list audio devices: %w", err)
	}
	
	// Look for built-in microphone patterns
	builtInPatterns := []string{
		"Built-in Microphone",
		"Internal Microphone", 
		"MacBook Pro Microphone",
		"MacBook Air Microphone",
		"iMac Microphone",
	}
	
	for _, device := range devices {
		for _, pattern := range builtInPatterns {
			if strings.Contains(device.Name, pattern) {
				logger.WithFields(map[string]interface{}{
					"device_name": device.Name,
					"device_id":   device.DeviceID,
				}).Info("Found built-in microphone")
				return &device, nil
			}
		}
	}
	
	// If no built-in microphone found, return the first device (usually index 0)
	if len(devices) > 0 {
		logger.WithField("device_name", devices[0].Name).Warn("Built-in microphone not found, using first available device")
		return &devices[0], nil
	}
	
	return nil, fmt.Errorf("no audio input devices found")
}

// StartRecording starts recording audio to the specified file
func (r *AudioRecorder) StartRecording(recordingID, filePath string, device models.AudioDeviceInfo, config models.RecordingConfig) error {
	logger.WithFields(map[string]interface{}{
		"recording_id": recordingID,
		"file_path":    filePath,
		"device_name":  device.Name,
		"format":       config.Format,
		"sample_rate":  config.SampleRate,
	}).Info("Starting audio recording")

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if recording is already active
	if _, exists := r.activeRecordings[recordingID]; exists {
		logger.WithField("recording_id", recordingID).Warn("Recording already active")
		return fmt.Errorf("recording %s is already active", recordingID)
	}

	// Create directory for the file if it doesn't exist
	dir := ""
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '/' {
			dir = filePath[:i]
			break
		}
	}
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.WithError(err).WithFields(map[string]interface{}{
			"recording_id": recordingID,
			"directory":    dir,
		}).Error("Failed to create recording directory")
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if we should use Core Audio Taps (macOS 14.2+ with mixed mode)
	useCoreAudioTap := false
	macVersion := coreaudio.GetMacOSVersion()

	if config.RecordingMode == "mixed" && macVersion.SupportsCoreAudioTaps() {
		logger.WithFields(map[string]interface{}{
			"recording_id":  recordingID,
			"macos_version": macVersion.String(),
		}).Info("Using Core Audio Taps for mixed recording")
		useCoreAudioTap = true
	}

	var cmd *exec.Cmd
	var coreAudioRec *coreaudio.MixedAudioRecorder

	if useCoreAudioTap {
		// Use Core Audio Taps for native system audio capture
		logger.WithField("recording_id", recordingID).Info("Creating Core Audio Taps recorder")

		recorder, err := coreaudio.NewMixedAudioRecorder(filePath)
		if err != nil {
			logger.WithError(err).Warn("Failed to create Core Audio Taps recorder, falling back to ffmpeg")
			useCoreAudioTap = false
			// Fall through to ffmpeg approach below
		} else {
			// Extract device index from device ID
			deviceIndex := 0
			if strings.HasPrefix(device.DeviceID, "input_") {
				fmt.Sscanf(device.DeviceID, "input_%d", &deviceIndex)
			}

			if err := recorder.Start(deviceIndex); err != nil {
				logger.WithError(err).Warn("Failed to start Core Audio Taps recorder, falling back to ffmpeg")
				useCoreAudioTap = false
				// Fall through to ffmpeg approach below
			} else {
				coreAudioRec = recorder
				logger.WithField("recording_id", recordingID).Info("Core Audio Taps recorder started successfully")
			}
		}
	}

	if !useCoreAudioTap {
		// Use traditional ffmpeg approach
		switch config.RecordingMode {
		case "microphone":
			cmd = r.createMicrophoneOnlyCommand(filePath, device, config)
		case "system":
			cmd = r.createSystemAudioCommand(filePath, device, config)
		case "mixed":
			cmd = r.createMixedAudioCommand(filePath, device, config)
		default:
			// Default to microphone only for backwards compatibility
			cmd = r.createMicrophoneOnlyCommand(filePath, device, config)
		}

		logger.WithFields(map[string]interface{}{
			"recording_id": recordingID,
			"command":      cmd.String(),
		}).Info("Starting ffmpeg recording process")

		// Start the recording process
		if err := cmd.Start(); err != nil {
			logger.WithError(err).WithField("recording_id", recordingID).Error("Failed to start ffmpeg recording process")
			return fmt.Errorf("failed to start recording process: %w", err)
		}
	}

	// Store the recording process
	r.activeRecordings[recordingID] = &RecordingProcess{
		ID:              recordingID,
		FilePath:        filePath,
		Config:          config,
		Device:          device,
		Process:         cmd,
		CoreAudioTapRec: coreAudioRec,
		StartTime:       time.Now(),
		IsActive:        true,
		UseCoreAudioTap: useCoreAudioTap,
	}

	logger.WithFields(map[string]interface{}{
		"recording_id":       recordingID,
		"file_path":          filePath,
		"use_core_audio_tap": useCoreAudioTap,
		"macos_version":      macVersion.String(),
	}).Info("Audio recording started successfully")

	return nil
}

// StopRecording stops the recording process
func (r *AudioRecorder) StopRecording(recordingID string) error {
	logger.WithField("recording_id", recordingID).Info("Stopping audio recording")

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Get the recording process
	recording, exists := r.activeRecordings[recordingID]
	if !exists {
		logger.WithField("recording_id", recordingID).Warn("Recording not found or already stopped")
		return fmt.Errorf("recording %s not found", recordingID)
	}

	// Stop Core Audio Taps recorder if used
	if recording.UseCoreAudioTap && recording.CoreAudioTapRec != nil {
		logger.WithField("recording_id", recordingID).Info("Stopping Core Audio Taps recorder")

		if err := recording.CoreAudioTapRec.Stop(); err != nil {
			logger.WithError(err).WithField("recording_id", recordingID).Error("Failed to stop Core Audio Taps recorder")
			// Continue anyway to clean up
		}

		recording.IsActive = false
	}

	// Stop the ffmpeg process gracefully
	if recording.Process != nil && recording.IsActive {
		logger.WithFields(map[string]interface{}{
			"recording_id": recordingID,
			"pid":          recording.Process.Process.Pid,
		}).Info("Terminating ffmpeg process")

		// Send interrupt signal to ffmpeg for graceful shutdown
		if err := recording.Process.Process.Signal(os.Interrupt); err != nil {
			logger.WithError(err).WithField("recording_id", recordingID).Warn("Failed to send interrupt signal, forcing kill")
			recording.Process.Process.Kill()
		}

		// Wait for process to finish (with timeout)
		done := make(chan error, 1)
		go func() {
			done <- recording.Process.Wait()
		}()

		select {
		case err := <-done:
			if err != nil {
				logger.WithError(err).WithField("recording_id", recordingID).Info("ffmpeg process finished with error (expected for interrupt)")
			} else {
				logger.WithField("recording_id", recordingID).Info("ffmpeg process finished successfully")
			}
		case <-time.After(5 * time.Second):
			logger.WithField("recording_id", recordingID).Warn("ffmpeg process did not finish within timeout, force killing")
			recording.Process.Process.Kill()
		}

		recording.IsActive = false
	}

	// Remove from active recordings
	delete(r.activeRecordings, recordingID)

	// Check if file was created and has content
	if fileInfo, err := os.Stat(recording.FilePath); err == nil {
		logger.WithFields(map[string]interface{}{
			"recording_id": recordingID,
			"file_path":    recording.FilePath,
			"file_size":    fileInfo.Size(),
			"duration":     time.Since(recording.StartTime).Seconds(),
		}).Info("Audio recording stopped successfully")
	} else {
		logger.WithError(err).WithFields(map[string]interface{}{
			"recording_id": recordingID,
			"file_path":    recording.FilePath,
		}).Error("Audio file was not created or is not accessible")
		return fmt.Errorf("audio file not found after recording: %w", err)
	}

	return nil
}

// IsRecording checks if a recording is currently active
func (r *AudioRecorder) IsRecording(recordingID string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	recording, exists := r.activeRecordings[recordingID]
	return exists && recording.IsActive
}

// GetActiveRecordings returns a list of currently active recording IDs
func (r *AudioRecorder) GetActiveRecordings() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var activeIDs []string
	for id, recording := range r.activeRecordings {
		if recording.IsActive {
			activeIDs = append(activeIDs, id)
		}
	}
	return activeIDs
}

// GetRecordingDuration returns the current duration of an active recording
func (r *AudioRecorder) GetRecordingDuration(recordingID string) time.Duration {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	if recording, exists := r.activeRecordings[recordingID]; exists && recording.IsActive {
		return time.Since(recording.StartTime)
	}
	return 0
}

// StopAllRecordings stops all active recordings
func (r *AudioRecorder) StopAllRecordings() error {
	var recordingIDs []string
	
	// Get list of recording IDs first (avoid deadlock)
	r.mutex.RLock()
	for recordingID := range r.activeRecordings {
		recordingIDs = append(recordingIDs, recordingID)
	}
	r.mutex.RUnlock()
	
	// Stop each recording
	var errors []error
	for _, recordingID := range recordingIDs {
		if err := r.StopRecording(recordingID); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop recording %s: %w", recordingID, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to stop some recordings: %v", errors)
	}

	return nil
}

// createMicrophoneOnlyCommand creates an ffmpeg command for microphone-only recording
func (r *AudioRecorder) createMicrophoneOnlyCommand(filePath string, device models.AudioDeviceInfo, config models.RecordingConfig) *exec.Cmd {
	var audioInput string

	// Parse device ID to get the actual index
	if device.DeviceID == "default" || device.DeviceID == "" {
		audioInput = ":0"  // Default input device
	} else if strings.HasPrefix(device.DeviceID, "input_") {
		// Extract the index from "input_0" format
		index := strings.TrimPrefix(device.DeviceID, "input_")
		audioInput = fmt.Sprintf(":%s", index)
	} else {
		audioInput = fmt.Sprintf(":%s", device.DeviceID)
	}

	return exec.Command("ffmpeg",
		"-f", "avfoundation",                        // macOS audio framework
		"-i", audioInput,                            // Audio input device (microphone)
		"-ar", fmt.Sprintf("%d", config.SampleRate), // Sample rate
		"-ac", "2",                                  // Stereo output
		"-y",                                        // Overwrite output file
		filePath,
	)
}

// createSystemAudioCommand creates an ffmpeg command for system audio recording
func (r *AudioRecorder) createSystemAudioCommand(filePath string, device models.AudioDeviceInfo, config models.RecordingConfig) *exec.Cmd {
	// For system audio, we need to use the output device as input
	// This requires special macOS permissions and setup

	return exec.Command("ffmpeg",
		"-f", "avfoundation",                        // macOS audio framework
		"-i", ":1",                                  // System audio (typically index 1)
		"-ar", fmt.Sprintf("%d", config.SampleRate), // Sample rate
		"-ac", "2",                                  // Stereo output
		"-y",                                        // Overwrite output file
		filePath,
	)
}

// createMixedAudioCommand creates an ffmpeg command for mixed microphone + system audio recording
func (r *AudioRecorder) createMixedAudioCommand(filePath string, device models.AudioDeviceInfo, config models.RecordingConfig) *exec.Cmd {
	var micInput string

	// Parse device ID to get the actual microphone index
	if device.DeviceID == "default" || device.DeviceID == "" {
		micInput = ":0"  // Default input device
	} else if strings.HasPrefix(device.DeviceID, "input_") {
		// Extract the index from "input_0" format
		index := strings.TrimPrefix(device.DeviceID, "input_")
		micInput = fmt.Sprintf(":%s", index)
	} else {
		micInput = fmt.Sprintf(":%s", device.DeviceID)
	}

	// Create ffmpeg command to capture both microphone and system audio
	// and mix them together
	return exec.Command("ffmpeg",
		"-f", "avfoundation",                        // macOS audio framework
		"-i", micInput,                              // Microphone input
		"-f", "avfoundation",                        // macOS audio framework
		"-i", ":1",                                  // System audio input
		"-filter_complex", "[0:a][1:a]amix=inputs=2:duration=longest", // Mix the two audio streams
		"-ar", fmt.Sprintf("%d", config.SampleRate), // Sample rate
		"-ac", "2",                                  // Stereo output
		"-y",                                        // Overwrite output file
		filePath,
	)
}