//go:build !darwin || !cgo
// +build !darwin !cgo

package coreaudio

// HasScreenRecordingPermission always returns false on non-macOS platforms
func HasScreenRecordingPermission() bool {
	return false
}

// RequestScreenRecordingPermission always returns false on non-macOS platforms
func RequestScreenRecordingPermission() bool {
	return false
}

// EnsureScreenRecordingPermission always returns error on non-macOS platforms
func EnsureScreenRecordingPermission() error {
	return nil // Not applicable on non-macOS
}

// GetPermissionStatusMessage returns a message for non-macOS platforms
func GetPermissionStatusMessage() string {
	return "Screen Recording permission is only required on macOS"
}
