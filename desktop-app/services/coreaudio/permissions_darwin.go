//go:build darwin && cgo
// +build darwin,cgo

package coreaudio

/*
#cgo LDFLAGS: -framework ApplicationServices
#include <ApplicationServices/ApplicationServices.h>

// Check if screen recording permission is granted
static int hasScreenRecordingPermission() {
	// CGPreflightScreenCaptureAccess checks if permission is granted
	// Returns 1 if granted, 0 if not
	return CGPreflightScreenCaptureAccess() ? 1 : 0;
}

// Request screen recording permission
// This will show the system permission dialog if not already granted
static int requestScreenRecordingPermission() {
	// CGRequestScreenCaptureAccess triggers the permission prompt
	// Returns 1 if granted (or already granted), 0 if denied
	return CGRequestScreenCaptureAccess() ? 1 : 0;
}
*/
import "C"
import (
	"fmt"

	"github.com/platformlabs-co/personal-assist/logger"
)

// HasScreenRecordingPermission checks if Screen Recording permission is granted
func HasScreenRecordingPermission() bool {
	result := C.hasScreenRecordingPermission()
	hasPermission := result == 1

	logger.WithField("has_permission", hasPermission).Info("Checked Screen Recording permission status")
	return hasPermission
}

// RequestScreenRecordingPermission requests Screen Recording permission
// This will show a system dialog if permission is not already granted
func RequestScreenRecordingPermission() bool {
	logger.Info("Requesting Screen Recording permission - dialog may appear")

	result := C.requestScreenRecordingPermission()
	granted := result == 1

	if granted {
		logger.Info("Screen Recording permission granted")
	} else {
		logger.Warn("Screen Recording permission denied or requires app restart")
	}

	return granted
}

// EnsureScreenRecordingPermission checks and requests permission if needed
func EnsureScreenRecordingPermission() error {
	// First check if already granted
	if HasScreenRecordingPermission() {
		return nil
	}

	// Not granted, request it
	logger.Warn("Screen Recording permission not granted - requesting now")

	if !RequestScreenRecordingPermission() {
		return fmt.Errorf("Screen Recording permission required. Please grant permission in System Settings > Privacy & Security > Screen Recording, then restart the app")
	}

	return nil
}

// GetPermissionStatusMessage returns a user-friendly message about permission status
func GetPermissionStatusMessage() string {
	if HasScreenRecordingPermission() {
		return "Screen Recording permission is granted ✓"
	}

	return "Screen Recording permission is required to capture system audio. " +
		"Please grant permission in System Settings > Privacy & Security > Screen Recording, " +
		"then restart Memoria."
}
