// go:build darwin && cgo
// +build darwin,cgo

package coreaudio

/*
#include "audio_tap_impl.h"
#include <stdlib.h>

// Wrapper is defined in callbacks_darwin.go to avoid duplicate symbols
*/
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/platformlabs-co/personal-assist/logger"
)

// SystemAudioTap represents a Core Audio tap for capturing system audio
type SystemAudioTap struct {
	handle   C.AudioTapHandle
	callback AudioCallback
	userData interface{}
	mutex    sync.Mutex
	isActive bool
}

// AudioCallback is called when audio data is available
type AudioCallback func(audioData []byte, channels int, sampleRate float64)

// Global registry to map C pointers to Go callbacks
var (
	callbackRegistry = make(map[uintptr]*SystemAudioTap)
	registryMutex    sync.RWMutex
	nextID           uintptr = 1
)

// IsAvailable checks if Core Audio Taps are available (macOS 14.2+)
func IsAvailable() bool {
	available := C.AudioTap_IsAvailable()
	return available == 1
}

// NewSystemAudioTap creates a new system audio tap
func NewSystemAudioTap(callback AudioCallback) (*SystemAudioTap, error) {
	if !IsAvailable() {
		return nil, fmt.Errorf("Core Audio Taps not available (requires macOS 14.2+)")
	}

	if callback == nil {
		return nil, fmt.Errorf("callback cannot be nil")
	}

	tap := &SystemAudioTap{
		callback: callback,
		isActive: false,
	}

	// Register the tap in the global registry
	registryMutex.Lock()
	id := nextID
	nextID++
	callbackRegistry[id] = tap
	tap.userData = id
	registryMutex.Unlock()

	logger.WithField("tap_id", id).Info("Creating Core Audio Tap")

	// Create the C audio tap with our C callback wrapper
	handle := C.AudioTap_Create(
		C.AudioTapCallback(C.coreaudio_callback_wrapper),
		unsafe.Pointer(id),
	)

	if handle == nil {
		registryMutex.Lock()
		delete(callbackRegistry, id)
		registryMutex.Unlock()

		errorMsg := C.GoString(C.AudioTap_GetLastError())
		logger.WithError(fmt.Errorf(errorMsg)).Error("Failed to create Core Audio Tap")
		return nil, fmt.Errorf("failed to create audio tap: %s", errorMsg)
	}

	tap.handle = handle

	// Keep the tap alive
	runtime.SetFinalizer(tap, (*SystemAudioTap).destroy)

	logger.WithField("tap_id", id).Info("Core Audio Tap created successfully")
	return tap, nil
}

// Start begins capturing audio
func (t *SystemAudioTap) Start() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.isActive {
		return fmt.Errorf("audio tap already active")
	}

	logger.Info("Starting Core Audio Tap")

	result := C.AudioTap_Start(t.handle)
	if result != 0 {
		errorMsg := C.GoString(C.AudioTap_GetLastError())
		logger.WithError(fmt.Errorf(errorMsg)).Error("Failed to start Core Audio Tap")
		return fmt.Errorf("failed to start audio tap: %s", errorMsg)
	}

	t.isActive = true
	logger.Info("Core Audio Tap started successfully")
	return nil
}

// Stop stops capturing audio
func (t *SystemAudioTap) Stop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.isActive {
		return
	}

	logger.Info("Stopping Core Audio Tap")
	C.AudioTap_Stop(t.handle)
	t.isActive = false
	logger.Info("Core Audio Tap stopped")
}

// IsActive returns whether the tap is currently capturing
func (t *SystemAudioTap) IsActive() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.isActive
}

// Close releases all resources
func (t *SystemAudioTap) Close() error {
	t.destroy()
	return nil
}

func (t *SystemAudioTap) destroy() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.handle == nil {
		return
	}

	logger.Info("Destroying Core Audio Tap")

	// Stop if active
	if t.isActive {
		C.AudioTap_Stop(t.handle)
		t.isActive = false
	}

	// Destroy the tap
	C.AudioTap_Destroy(t.handle)
	t.handle = nil

	// Remove from registry
	if id, ok := t.userData.(uintptr); ok {
		registryMutex.Lock()
		delete(callbackRegistry, id)
		registryMutex.Unlock()
	}

	logger.Info("Core Audio Tap destroyed")
}

//export audioTapGoCallback
func audioTapGoCallback(userData unsafe.Pointer, audioData unsafe.Pointer, dataSize C.size_t, channels C.int, sampleRate C.double) {
	// Get the tap from the registry
	id := uintptr(userData)

	registryMutex.RLock()
	tap, ok := callbackRegistry[id]
	registryMutex.RUnlock()

	if !ok || tap == nil || tap.callback == nil {
		return
	}

	// Convert C data to Go slice
	size := int(dataSize)
	if size <= 0 {
		return
	}

	// Create a copy of the audio data
	goData := C.GoBytes(audioData, C.int(dataSize))

	// Call the Go callback
	tap.callback(goData, int(channels), float64(sampleRate))
}
