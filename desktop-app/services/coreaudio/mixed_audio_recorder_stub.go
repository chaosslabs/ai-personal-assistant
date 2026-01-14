// go:build !darwin || !cgo
// +build !darwin !cgo

package coreaudio

import "fmt"

// MixedAudioRecorder stub for non-macOS platforms
type MixedAudioRecorder struct{}

// NewMixedAudioRecorder returns an error on non-macOS platforms
func NewMixedAudioRecorder(outputFile string) (*MixedAudioRecorder, error) {
	return nil, fmt.Errorf("Mixed audio recording only available on macOS 14.2+")
}

// Start returns an error on non-macOS platforms
func (r *MixedAudioRecorder) Start(micDeviceIndex int) error {
	return fmt.Errorf("Mixed audio recording only available on macOS 14.2+")
}

// Stop does nothing on non-macOS platforms
func (r *MixedAudioRecorder) Stop() error {
	return fmt.Errorf("Mixed audio recording only available on macOS 14.2+")
}

// IsRecording always returns false on non-macOS platforms
func (r *MixedAudioRecorder) IsRecording() bool {
	return false
}

// Close does nothing on non-macOS platforms
func (r *MixedAudioRecorder) Close() error {
	return nil
}
