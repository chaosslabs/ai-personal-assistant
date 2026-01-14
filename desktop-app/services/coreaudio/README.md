# Core Audio Taps - Native macOS System Audio Capture

This package implements native macOS system audio capture using the Core Audio Taps API introduced in macOS 14.2 (Sonoma).

## Overview

Core Audio Taps allows capturing system audio output without requiring third-party virtual audio drivers like BlackHole. This is the official Apple API for system audio capture.

## Requirements

- **macOS 14.2 or later** (Sonoma, released November 2023)
- **Screen Recording permission** (required by macOS for audio capture)
- **ffmpeg** (for audio mixing and format conversion)

## Architecture

### Components

1. **audio_tap.m/h** - Objective-C implementation of Core Audio Taps
   - Creates audio taps for system audio output
   - Handles real-time audio callbacks
   - Manages audio device lifecycle

2. **audio_tap_darwin.go** - Go CGo bridge
   - Wraps Objective-C implementation for Go
   - Manages callback registration
   - Provides Go-friendly API

3. **system_version_darwin.go** - macOS version detection
   - Detects current macOS version
   - Checks Core Audio Taps availability
   - Determines feature support

4. **mixed_audio_recorder.go** - Integrated recorder
   - Captures system audio via Core Audio Taps
   - Captures microphone via ffmpeg
   - Mixes both streams into final output
   - Gracefully falls back to microphone-only if system audio fails

## Usage

### Basic System Audio Capture

```go
import "github.com/platformlabs-co/personal-assist/services/coreaudio"

// Check if available
if !coreaudio.IsAvailable() {
    log.Println("Core Audio Taps not available (requires macOS 14.2+)")
    return
}

// Create audio tap with callback
tap, err := coreaudio.NewSystemAudioTap(func(audioData []byte, channels int, sampleRate float64) {
    // Process audio data
    log.Printf("Received %d bytes, %d channels, %.0f Hz", len(audioData), channels, sampleRate)
})
if err != nil {
    log.Fatal(err)
}
defer tap.Close()

// Start capturing
if err := tap.Start(); err != nil {
    log.Fatal(err)
}

// ... do work ...

// Stop capturing
tap.Stop()
```

### Mixed Audio Recording (Microphone + System)

```go
import "github.com/platformlabs-co/personal-assist/services/coreaudio"

// Create mixed recorder
recorder, err := coreaudio.NewMixedAudioRecorder("/path/to/output.wav")
if err != nil {
    log.Fatal(err)
}

// Start recording (device index 0 = built-in mic)
if err := recorder.Start(0); err != nil {
    log.Fatal(err)
}

// ... record meeting ...

// Stop and save
if err := recorder.Stop(); err != nil {
    log.Fatal(err)
}
```

### Check macOS Version

```go
import "github.com/platformlabs-co/personal-assist/services/coreaudio"

version := coreaudio.GetMacOSVersion()
log.Printf("macOS %s", version.String())

if version.SupportsCoreAudioTaps() {
    log.Println("Core Audio Taps available")
} else {
    log.Println("Core Audio Taps not available - upgrade to macOS 14.2+")
}
```

## Integration with Audio Recorder

The `AudioRecorder` service automatically uses Core Audio Taps when:

1. macOS version is 14.2 or later
2. Recording mode is set to `"mixed"`
3. Core Audio Taps initialization succeeds

If any condition fails, it gracefully falls back to ffmpeg-based recording.

```go
// In services/audio_recorder.go
config := models.RecordingConfig{
    RecordingMode: "mixed",  // Triggers Core Audio Taps on macOS 14.2+
    Format:        "wav",
    SampleRate:    44100,
}

// Start recording - will use Core Audio Taps if available
err := audioRecorder.StartRecording(recordingID, filePath, device, config)
```

## Permissions

### Screen Recording Permission

Core Audio Taps requires **Screen Recording permission**, even though it only captures audio. This is a macOS security requirement.

#### Requesting Permission

The app must include this key in `Info.plist`:

```xml
<key>NSMicrophoneUsageDescription</key>
<string>Memoria needs microphone access to record your voice during meetings and work sessions.</string>

<key>NSAudioCaptureUsageDescription</key>
<string>Memoria needs audio capture permission to record system audio from meetings and calls.</string>
```

#### User Permission Flow

1. First time system audio capture is attempted
2. macOS shows permission prompt: "Memoria wants to record your screen"
3. User grants Screen Recording permission in System Settings
4. App needs to restart for permission to take effect

#### Checking Permission Status

The app now automatically checks and requests Screen Recording permission when you start a mixed mode recording:

```go
import "github.com/platformlabs-co/personal-assist/services/coreaudio"

// Check if permission is granted
if coreaudio.HasScreenRecordingPermission() {
    log.Println("Permission granted!")
} else {
    log.Println("Permission not granted")
}

// Request permission (shows system dialog)
if coreaudio.RequestScreenRecordingPermission() {
    log.Println("Permission granted or already granted")
} else {
    log.Println("Permission denied - user needs to restart app after granting")
}

// Check and request in one call (used internally)
if err := coreaudio.EnsureScreenRecordingPermission(); err != nil {
    log.Printf("Permission error: %v", err)
}

// Get user-friendly status message
message := coreaudio.GetPermissionStatusMessage()
log.Println(message)
```

**Note**: After granting Screen Recording permission, the app must be restarted for the permission to take effect.

## Troubleshooting

### "Core Audio Taps not available"

**Cause:** macOS version < 14.2

**Solution:**
- Upgrade to macOS 14.2 (Sonoma) or later
- App will fall back to microphone-only recording automatically

### "Failed to create audio tap"

**Causes:**
- Screen Recording permission not granted
- Audio device not available
- System audio not playing

**Solutions:**
1. Check System Settings → Privacy & Security → Screen Recording
2. Ensure app is listed and enabled
3. Restart app after granting permission
4. Ensure audio is playing from system

### "No system audio captured"

**Causes:**
- No audio playing during recording
- Audio output muted
- Incorrect audio device selected

**Solutions:**
1. Ensure system audio is playing
2. Check system volume is not muted
3. Verify correct output device is selected in macOS Sound settings

### Recording only captures microphone

**Cause:** Core Audio Taps failed to initialize, fell back to microphone-only

**Solutions:**
1. Check logs for specific error
2. Verify macOS version: `sw_vers`
3. Grant Screen Recording permission
4. Restart application

## Platform Support

### macOS (Darwin)

Full support on macOS 14.2+:
- Native Core Audio Taps implementation
- System audio + microphone mixing
- Graceful fallback to microphone-only

### Linux / Other Platforms

Stub implementation that returns errors:
- `IsAvailable()` returns `false`
- `NewSystemAudioTap()` returns error
- Build succeeds but functionality unavailable

This allows the codebase to compile on all platforms while only providing functionality on macOS.

## Build Flags

The implementation uses Go build tags for platform-specific code:

```go
// go:build darwin
// +build darwin
```

- `*_darwin.go` files compile only on macOS
- `*_stub.go` files compile on all other platforms

## CGo Configuration

The package uses CGo to interface with Objective-C:

```go
/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreAudio -framework AudioToolbox -framework Foundation
*/
```

Required frameworks:
- **CoreAudio** - Core audio APIs
- **AudioToolbox** - Audio processing tools
- **Foundation** - macOS foundation classes

## Performance Considerations

### Memory Usage

- Audio buffers: ~1MB initial capacity
- Scales with recording duration
- Buffers are released when recording stops

### CPU Usage

- Minimal overhead from Core Audio Taps
- Audio mixing (ffmpeg) adds ~5-10% CPU
- Comparable to traditional recording methods

### Latency

- Near-zero latency for audio capture
- Real-time callbacks from Core Audio
- Post-processing mixing adds 1-2 seconds after stop

## Testing

### Manual Testing Checklist

1. **Version Detection**
   - [ ] Detects correct macOS version
   - [ ] Returns true for SupportsCoreAudioTaps on 14.2+
   - [ ] Returns false on macOS < 14.2

2. **Permission Handling**
   - [ ] Shows permission prompt on first use
   - [ ] Fails gracefully if permission denied
   - [ ] Falls back to microphone-only

3. **Audio Capture**
   - [ ] Captures system audio (play music during recording)
   - [ ] Captures microphone audio
   - [ ] Mixes both streams correctly
   - [ ] Output file contains both audio sources

4. **Fallback Behavior**
   - [ ] Falls back to ffmpeg if Core Audio Taps fails
   - [ ] Falls back to microphone-only if system audio unavailable
   - [ ] Logs appropriate warnings

### Unit Testing

Currently manual. Automated tests TODO:
- Mock Core Audio callbacks
- Test version detection
- Test graceful degradation

## Future Enhancements

1. **Permission Detection**
   - Pre-check Screen Recording permission
   - Guide users to grant permission
   - Better error messages

2. **Audio Format Options**
   - Support more output formats
   - Configurable sample rates
   - Bitrate control

3. **Per-Application Capture**
   - Capture specific app audio (e.g., only Zoom)
   - Requires ScreenCaptureKit integration

4. **Audio Processing**
   - Noise reduction
   - Echo cancellation
   - Volume normalization

## References

- [Core Audio Taps Documentation](https://developer.apple.com/documentation/coreaudio/capturing-system-audio-with-core-audio-taps)
- [AudioCap Sample Code](https://github.com/insidegui/AudioCap)
- [ScreenCaptureKit Overview](https://developer.apple.com/documentation/screencapturekit)

## License

Part of Memoria (Personal Assist) project.
