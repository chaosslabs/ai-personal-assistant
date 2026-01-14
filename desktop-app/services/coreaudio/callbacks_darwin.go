// go:build darwin && cgo
// +build darwin,cgo

package coreaudio

/*
#include <stdlib.h>

// Forward declaration of Go callback
void audioTapGoCallback(void* userData, void* audioData, size_t dataSize, int channels, double sampleRate);

// C callback wrapper - defined ONLY in this file to avoid duplicates
void coreaudio_callback_wrapper(void* userData, const void* audioData, size_t dataSize, int channels, double sampleRate) {
    audioTapGoCallback(userData, (void*)audioData, dataSize, channels, sampleRate);
}
*/
import "C"

// This file exists solely to define the C callback wrapper in a single compilation unit
// to avoid duplicate symbol errors during linking.
