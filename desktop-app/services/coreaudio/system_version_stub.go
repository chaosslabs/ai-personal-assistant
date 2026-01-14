// go:build !darwin || !cgo
// +build !darwin !cgo

package coreaudio

import "fmt"

// MacOSVersion represents a macOS version (stub for non-macOS)
type MacOSVersion struct {
	Major int
	Minor int
	Patch int
}

// GetMacOSVersion returns a zero version on non-macOS platforms
func GetMacOSVersion() MacOSVersion {
	return MacOSVersion{0, 0, 0}
}

// String returns the version as a string
func (v MacOSVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// IsAtLeast always returns false on non-macOS platforms
func (v MacOSVersion) IsAtLeast(major, minor int) bool {
	return false
}

// SupportsCoreAudioTaps always returns false on non-macOS platforms
func (v MacOSVersion) SupportsCoreAudioTaps() bool {
	return false
}

// SupportsScreenCaptureKit always returns false on non-macOS platforms
func (v MacOSVersion) SupportsScreenCaptureKit() bool {
	return false
}
