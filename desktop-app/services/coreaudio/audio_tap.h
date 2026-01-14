#ifndef AUDIO_TAP_H
#define AUDIO_TAP_H

// This header provides the Core Audio Taps interface for Go CGo
// The implementation is in audio_tap_impl.h (header-only for CGo compatibility)

#ifdef __APPLE__

#include "audio_tap_impl.h"

#else

// Stub declarations for non-Apple platforms
typedef void* AudioTapHandle;
typedef void (*AudioTapCallback)(void* userData, const void* audioData, size_t dataSize, int channels, double sampleRate);

static inline AudioTapHandle AudioTap_Create(AudioTapCallback callback, void* userData) { return NULL; }
static inline int AudioTap_Start(AudioTapHandle handle) { return -1; }
static inline void AudioTap_Stop(AudioTapHandle handle) {}
static inline void AudioTap_Destroy(AudioTapHandle handle) {}
static inline int AudioTap_IsAvailable() { return 0; }
static inline const char* AudioTap_GetLastError() { return "Not available on this platform"; }

#endif // __APPLE__

#endif // AUDIO_TAP_H
