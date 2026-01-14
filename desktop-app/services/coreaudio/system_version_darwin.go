// go:build darwin && cgo
// +build darwin,cgo

package coreaudio

/*
#import <Foundation/Foundation.h>

// Get macOS version as major.minor.patch
static void getMacOSVersion(int* major, int* minor, int* patch) {
    NSOperatingSystemVersion version = [[NSProcessInfo processInfo] operatingSystemVersion];
    *major = (int)version.majorVersion;
    *minor = (int)version.minorVersion;
    *patch = (int)version.patchVersion;
}
*/
import "C"
import (
	"fmt"
)

// MacOSVersion represents a macOS version
type MacOSVersion struct {
	Major int
	Minor int
	Patch int
}

// GetMacOSVersion returns the current macOS version
func GetMacOSVersion() MacOSVersion {
	var major, minor, patch C.int
	C.getMacOSVersion(&major, &minor, &patch)

	return MacOSVersion{
		Major: int(major),
		Minor: int(minor),
		Patch: int(patch),
	}
}

// String returns the version as a string
func (v MacOSVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// IsAtLeast checks if the current version is at least the specified version
func (v MacOSVersion) IsAtLeast(major, minor int) bool {
	if v.Major > major {
		return true
	}
	if v.Major == major && v.Minor >= minor {
		return true
	}
	return false
}

// SupportsCoreAudioTaps returns true if the macOS version supports Core Audio Taps (14.2+)
func (v MacOSVersion) SupportsCoreAudioTaps() bool {
	return v.IsAtLeast(14, 2)
}

// SupportsScreenCaptureKit returns true if the macOS version supports ScreenCaptureKit (12.3+)
func (v MacOSVersion) SupportsScreenCaptureKit() bool {
	return v.IsAtLeast(12, 3)
}
