// go:build !darwin || !cgo
// +build !darwin !cgo

package coreaudio

import "fmt"

// SystemAudioTap stub for non-macOS platforms
type SystemAudioTap struct{}

// AudioCallback is called when audio data is available
type AudioCallback func(audioData []byte, channels int, sampleRate float64)

// IsAvailable always returns false on non-macOS platforms
func IsAvailable() bool {
	return false
}

// NewSystemAudioTap returns an error on non-macOS platforms
func NewSystemAudioTap(callback AudioCallback) (*SystemAudioTap, error) {
	return nil, fmt.Errorf("Core Audio Taps only available on macOS")
}

// Start returns an error on non-macOS platforms
func (t *SystemAudioTap) Start() error {
	return fmt.Errorf("Core Audio Taps only available on macOS")
}

// Stop does nothing on non-macOS platforms
func (t *SystemAudioTap) Stop() {}

// IsActive always returns false on non-macOS platforms
func (t *SystemAudioTap) IsActive() bool {
	return false
}

// Close does nothing on non-macOS platforms
func (t *SystemAudioTap) Close() error {
	return nil
}
