// go:build darwin && cgo
// +build darwin,cgo

package coreaudio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/platformlabs-co/personal-assist/logger"
)

// MixedAudioRecorder records both system audio and microphone audio
type MixedAudioRecorder struct {
	systemTap        *SystemAudioTap
	micRecorder      *exec.Cmd
	outputFile       *os.File
	tempSystemFile   string
	tempMicFile      string
	finalFile        string
	isRecording      bool
	isMixedMode      bool // true if using Core Audio Taps, false if microphone-only
	mutex            sync.Mutex
	systemBuffer     *AudioBuffer
	startTime        time.Time
	callbackCount    int // Track number of audio callbacks received
	lastCallbackTime time.Time
}

// AudioBuffer is a thread-safe audio buffer
type AudioBuffer struct {
	data      []byte
	mutex     sync.Mutex
	sampleRate float64
	channels   int
}

// NewAudioBuffer creates a new audio buffer
func NewAudioBuffer() *AudioBuffer {
	return &AudioBuffer{
		data: make([]byte, 0, 1024*1024), // 1MB initial capacity
	}
}

// Append adds audio data to the buffer
func (b *AudioBuffer) Append(data []byte) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.data = append(b.data, data...)
}

// GetData returns a copy of the buffer data
func (b *AudioBuffer) GetData() []byte {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	result := make([]byte, len(b.data))
	copy(result, b.data)
	return result
}

// Clear clears the buffer
func (b *AudioBuffer) Clear() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.data = b.data[:0]
}

// SetFormat sets the audio format
func (b *AudioBuffer) SetFormat(sampleRate float64, channels int) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.sampleRate = sampleRate
	b.channels = channels
}

// NewMixedAudioRecorder creates a new mixed audio recorder
func NewMixedAudioRecorder(outputPath string) (*MixedAudioRecorder, error) {
	recorder := &MixedAudioRecorder{
		finalFile:    outputPath,
		systemBuffer: NewAudioBuffer(),
		isRecording:  false,
	}

	return recorder, nil
}

// Start begins recording both system audio and microphone
func (r *MixedAudioRecorder) Start(micDeviceIndex int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.isRecording {
		return fmt.Errorf("already recording")
	}

	logger.Info("Starting mixed audio recording")
	r.startTime = time.Now()

	// Check if Core Audio Taps are available
	version := GetMacOSVersion()
	logger.WithFields(map[string]interface{}{
		"macos_version": version.String(),
		"supports_taps": version.SupportsCoreAudioTaps(),
	}).Info("Checking macOS version for Core Audio Taps support")

	if !version.SupportsCoreAudioTaps() {
		logger.Warn("Core Audio Taps not available, falling back to microphone-only")
		return r.startMicrophoneOnly(micDeviceIndex)
	}

	// Create temporary files for system and mic audio
	r.tempSystemFile = fmt.Sprintf("%s.system.wav", r.finalFile)
	r.tempMicFile = fmt.Sprintf("%s.mic.wav", r.finalFile)

	// Start system audio capture
	if err := r.startSystemAudioCapture(); err != nil {
		logger.WithError(err).Error("Failed to start system audio capture, falling back to microphone-only")
		return r.startMicrophoneOnly(micDeviceIndex)
	}

	// Start microphone capture
	if err := r.startMicrophoneCapture(micDeviceIndex); err != nil {
		logger.WithError(err).Error("Failed to start microphone capture")
		r.stopSystemAudioCapture()
		return err
	}

	r.isRecording = true
	logger.Info("Mixed audio recording started successfully")
	return nil
}

// startSystemAudioCapture starts capturing system audio using Core Audio Taps
func (r *MixedAudioRecorder) startSystemAudioCapture() error {
	logger.Info("Starting system audio capture with Core Audio Taps")

	// Check and request Screen Recording permission
	if err := EnsureScreenRecordingPermission(); err != nil {
		logger.WithError(err).Error("Screen Recording permission check failed")
		return err
	}

	// Create the audio tap with callback
	tap, err := NewSystemAudioTap(func(audioData []byte, channels int, sampleRate float64) {
		// Track callback activity
		r.mutex.Lock()
		r.callbackCount++
		r.lastCallbackTime = time.Now()
		count := r.callbackCount
		r.mutex.Unlock()

		// Store audio format on first callback
		if r.systemBuffer.sampleRate == 0 {
			r.systemBuffer.SetFormat(sampleRate, channels)
			logger.WithFields(map[string]interface{}{
				"sample_rate": sampleRate,
				"channels":    channels,
			}).Info("System audio format detected")
		}

		// Log periodically (every 100 callbacks)
		if count%100 == 1 {
			logger.WithFields(map[string]interface{}{
				"callback_count": count,
				"data_size":      len(audioData),
				"buffer_size":    len(r.systemBuffer.data),
			}).Debug("System audio callback progress")
		}

		// Append audio data to buffer
		r.systemBuffer.Append(audioData)
	})

	if err != nil {
		return fmt.Errorf("failed to create system audio tap: %w", err)
	}

	// Start the tap
	if err := tap.Start(); err != nil {
		tap.Close()
		return fmt.Errorf("failed to start system audio tap: %w", err)
	}

	r.systemTap = tap
	r.isMixedMode = true // Mark that we're in mixed mode
	logger.Info("System audio capture started - ensure Screen Recording permission is granted in System Settings > Privacy & Security")
	return nil
}

// startMicrophoneCapture starts capturing microphone audio using ffmpeg
func (r *MixedAudioRecorder) startMicrophoneCapture(deviceIndex int) error {
	logger.WithField("device_index", deviceIndex).Info("Starting microphone capture")

	// Use ffmpeg to capture microphone audio
	cmd := exec.Command("ffmpeg",
		"-f", "avfoundation",
		"-i", fmt.Sprintf(":%d", deviceIndex),
		"-ar", "44100",
		"-ac", "2",
		"-y",
		r.tempMicFile,
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	r.micRecorder = cmd
	logger.Info("Microphone capture started")
	return nil
}

// startMicrophoneOnly starts recording microphone only (fallback)
func (r *MixedAudioRecorder) startMicrophoneOnly(deviceIndex int) error {
	logger.Info("Starting microphone-only recording (fallback mode)")

	cmd := exec.Command("ffmpeg",
		"-f", "avfoundation",
		"-i", fmt.Sprintf(":%d", deviceIndex),
		"-ar", "44100",
		"-ac", "2",
		"-y",
		r.finalFile,
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start microphone recording: %w", err)
	}

	r.micRecorder = cmd
	r.isRecording = true
	r.isMixedMode = false // Mark that we're in microphone-only mode
	logger.Info("Microphone-only recording started")
	return nil
}

// Stop stops recording and mixes the audio streams
func (r *MixedAudioRecorder) Stop() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.isRecording {
		return fmt.Errorf("not currently recording")
	}

	logger.Info("Stopping mixed audio recording")

	// Remember if we're in mixed mode before stopping (systemTap will be set to nil)
	wasMixedMode := r.isMixedMode

	// Stop system audio capture
	r.stopSystemAudioCapture()

	// Stop microphone capture
	r.stopMicrophoneCapture()

	r.isRecording = false

	// If we were in microphone-only mode, we're done
	if !wasMixedMode {
		logger.Info("Microphone-only recording stopped")
		return nil
	}

	logger.Info("Processing mixed mode recording - saving and mixing audio streams")

	// Save system audio to WAV file
	if err := r.saveSystemAudioToWAV(); err != nil {
		logger.WithError(err).Error("Failed to save system audio, falling back to microphone-only")
		// If system audio failed, use microphone-only as fallback
		// The mic file should already exist from ffmpeg at tempMicFile location
		return r.fallbackToMicrophoneFile()
	}

	// Mix the two audio streams
	if err := r.mixAudioStreams(); err != nil {
		logger.WithError(err).Error("Failed to mix audio streams")
		return err
	}

	// Clean up temporary files
	r.cleanupTempFiles()

	logger.Info("Mixed audio recording stopped and saved")
	return nil
}

// stopSystemAudioCapture stops the system audio tap
func (r *MixedAudioRecorder) stopSystemAudioCapture() {
	if r.systemTap != nil {
		logger.Info("Stopping system audio capture")
		r.systemTap.Stop()
		r.systemTap.Close()
		r.systemTap = nil
	}
}

// stopMicrophoneCapture stops the microphone recorder
func (r *MixedAudioRecorder) stopMicrophoneCapture() {
	if r.micRecorder != nil {
		logger.Info("Stopping microphone capture")

		// Send interrupt signal for graceful shutdown
		r.micRecorder.Process.Signal(os.Interrupt)

		// Wait for process to finish with timeout
		done := make(chan error, 1)
		go func() {
			done <- r.micRecorder.Wait()
		}()

		select {
		case <-done:
			logger.Info("Microphone capture stopped gracefully")
		case <-time.After(5 * time.Second):
			logger.Warn("Microphone capture timeout, force killing")
			r.micRecorder.Process.Kill()
		}

		r.micRecorder = nil
	}
}

// saveSystemAudioToWAV saves the system audio buffer to a WAV file
func (r *MixedAudioRecorder) saveSystemAudioToWAV() error {
	logger.WithFields(map[string]interface{}{
		"callback_count": r.callbackCount,
		"buffer_size":    len(r.systemBuffer.data),
	}).Info("Saving system audio to WAV file")

	data := r.systemBuffer.GetData()
	if len(data) == 0 {
		logger.WithFields(map[string]interface{}{
			"callback_count":      r.callbackCount,
			"last_callback_time":  r.lastCallbackTime,
			"recording_duration":  time.Since(r.startTime),
		}).Warn("No system audio data captured - this usually means Screen Recording permission is not granted or no system audio was playing")

		if r.callbackCount == 0 {
			return fmt.Errorf("no system audio callbacks received - Screen Recording permission may not be granted")
		}
		return fmt.Errorf("no system audio data - ensure audio was playing during recording")
	}

	// Create WAV file
	file, err := os.Create(r.tempSystemFile)
	if err != nil {
		return fmt.Errorf("failed to create system audio file: %w", err)
	}
	defer file.Close()

	// Write WAV header
	sampleRate := int(r.systemBuffer.sampleRate)
	channels := r.systemBuffer.channels
	bitsPerSample := 16

	if err := writeWAVHeader(file, len(data), sampleRate, channels, bitsPerSample); err != nil {
		return fmt.Errorf("failed to write WAV header: %w", err)
	}

	// Write audio data
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write audio data: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"file_size":    len(data),
		"sample_rate":  sampleRate,
		"channels":     channels,
		"duration_sec": float64(len(data)) / float64(sampleRate*channels*bitsPerSample/8),
	}).Info("System audio saved to WAV")

	return nil
}

// mixAudioStreams mixes the system and microphone audio streams using ffmpeg
func (r *MixedAudioRecorder) mixAudioStreams() error {
	logger.WithFields(map[string]interface{}{
		"system_file": r.tempSystemFile,
		"mic_file":    r.tempMicFile,
		"output_file": r.finalFile,
	}).Info("Mixing audio streams")

	// Use ffmpeg to mix the two audio streams
	cmd := exec.Command("ffmpeg",
		"-i", r.tempSystemFile,
		"-i", r.tempMicFile,
		"-filter_complex", "[0:a][1:a]amix=inputs=2:duration=longest",
		"-ar", "44100",
		"-ac", "2",
		"-y",
		r.finalFile,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.WithError(err).WithField("output", string(output)).Error("Failed to mix audio streams")
		return fmt.Errorf("failed to mix audio: %w", err)
	}

	logger.Info("Audio streams mixed successfully")
	return nil
}

// fallbackToMicrophoneFile handles fallback when system audio capture fails
func (r *MixedAudioRecorder) fallbackToMicrophoneFile() error {
	logger.WithFields(map[string]interface{}{
		"temp_mic_file": r.tempMicFile,
		"final_file":    r.finalFile,
	}).Info("Using microphone-only file as fallback")

	// Check if microphone temp file exists
	if _, err := os.Stat(r.tempMicFile); err != nil {
		logger.WithError(err).Error("Microphone temp file not found")
		return fmt.Errorf("microphone file not found: %w", err)
	}

	// Rename/move the temp mic file to final location
	if err := os.Rename(r.tempMicFile, r.finalFile); err != nil {
		logger.WithError(err).Error("Failed to move microphone file to final location")
		return fmt.Errorf("failed to move microphone file: %w", err)
	}

	logger.Info("Successfully saved microphone-only recording")
	return nil
}

// cleanupTempFiles removes temporary audio files
func (r *MixedAudioRecorder) cleanupTempFiles() {
	if r.tempSystemFile != "" {
		os.Remove(r.tempSystemFile)
		logger.WithField("file", r.tempSystemFile).Debug("Removed temporary system audio file")
	}
	if r.tempMicFile != "" {
		os.Remove(r.tempMicFile)
		logger.WithField("file", r.tempMicFile).Debug("Removed temporary microphone file")
	}
}

// IsRecording returns whether recording is active
func (r *MixedAudioRecorder) IsRecording() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.isRecording
}

// writeWAVHeader writes a WAV file header
func writeWAVHeader(w *os.File, dataSize int, sampleRate int, channels int, bitsPerSample int) error {
	var buf bytes.Buffer

	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(36+dataSize))
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))              // fmt chunk size
	binary.Write(&buf, binary.LittleEndian, uint16(1))               // audio format (PCM)
	binary.Write(&buf, binary.LittleEndian, uint16(channels))        // number of channels
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))      // sample rate
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate*channels*bitsPerSample/8)) // byte rate
	binary.Write(&buf, binary.LittleEndian, uint16(channels*bitsPerSample/8))            // block align
	binary.Write(&buf, binary.LittleEndian, uint16(bitsPerSample))   // bits per sample

	// data chunk
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(dataSize))

	_, err := w.Write(buf.Bytes())
	return err
}
