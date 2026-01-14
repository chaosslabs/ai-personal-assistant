#ifndef AUDIO_TAP_IMPL_H
#define AUDIO_TAP_IMPL_H

#ifdef __APPLE__

#import <Foundation/Foundation.h>
#import <CoreAudio/CoreAudio.h>
#import <AudioToolbox/AudioToolbox.h>
#include <os/log.h>

// Callback function type for audio data
typedef void (*AudioTapCallback)(void* userData, const void* audioData, size_t dataSize, int channels, double sampleRate);

// Forward declaration of Go callback - implemented in Go via //export
extern void audioTapGoCallback(void* userData, void* audioData, size_t dataSize, int channels, double sampleRate);

// Forward declaration of wrapper - defined in callbacks_darwin.go CGo preamble
extern void coreaudio_callback_wrapper(void* userData, const void* audioData, size_t dataSize, int channels, double sampleRate);

// Audio tap handle
typedef void* AudioTapHandle;

// Audio tap context structure
typedef struct {
    AudioTapCallback callback;
    void* userData;
    AudioDeviceID deviceID;
    AudioDeviceIOProcID procID;
    CFTypeRef tap;
    int isRunning;
    int channels;
    double sampleRate;
} AudioTapContext;

// Storage for the last error message
static char audioTapLastError[256] = "";

// Set the last error message
static inline void AudioTap_SetLastError(const char* error) {
    strncpy(audioTapLastError, error, sizeof(audioTapLastError) - 1);
    audioTapLastError[sizeof(audioTapLastError) - 1] = '\0';
}

// Check if Core Audio Taps API is available (macOS 14.2+)
static inline int AudioTap_IsAvailable() {
    if (@available(macOS 14.2, *)) {
        return 1;
    }
    return 0;
}

// Get default output device
static inline AudioDeviceID AudioTap_GetDefaultOutputDevice() {
    AudioDeviceID deviceID = kAudioDeviceUnknown;
    UInt32 size = sizeof(deviceID);
    AudioObjectPropertyAddress propertyAddress = {
        kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    OSStatus status = AudioObjectGetPropertyData(
        kAudioObjectSystemObject,
        &propertyAddress,
        0,
        NULL,
        &size,
        &deviceID
    );

    if (status != noErr) {
        AudioTap_SetLastError("Failed to get default output device");
        return kAudioDeviceUnknown;
    }

    return deviceID;
}

// Get device stream format
static inline OSStatus AudioTap_GetDeviceStreamFormat(AudioDeviceID deviceID, AudioStreamBasicDescription* format) {
    UInt32 size = sizeof(AudioStreamBasicDescription);
    AudioObjectPropertyAddress propertyAddress = {
        kAudioDevicePropertyStreamFormat,
        kAudioDevicePropertyScopeOutput,
        kAudioObjectPropertyElementMain
    };

    return AudioObjectGetPropertyData(
        deviceID,
        &propertyAddress,
        0,
        NULL,
        &size,
        format
    );
}

// Audio tap callback - called when audio data is available
static OSStatus AudioTap_IOProc(
    AudioObjectID inObjectID,
    const AudioTimeStamp* inNow,
    const AudioBufferList* inInputData,
    const AudioTimeStamp* inInputTime,
    AudioBufferList* outOutputData,
    const AudioTimeStamp* inOutputTime,
    void* inClientData)
{
    AudioTapContext* context = (AudioTapContext*)inClientData;

    if (context == NULL || context->callback == NULL || !context->isRunning) {
        return noErr;
    }

    // Process each buffer in the buffer list
    for (UInt32 i = 0; i < inInputData->mNumberBuffers; i++) {
        AudioBuffer buffer = inInputData->mBuffers[i];

        // Call the Go callback with the audio data
        context->callback(
            context->userData,
            buffer.mData,
            buffer.mDataByteSize,
            buffer.mNumberChannels,
            context->sampleRate
        );
    }

    return noErr;
}

// Create an audio tap for system audio capture
static inline AudioTapHandle AudioTap_Create(AudioTapCallback callback, void* userData) {
    if (!AudioTap_IsAvailable()) {
        AudioTap_SetLastError("Core Audio Taps API not available (requires macOS 14.2+)");
        return NULL;
    }

    if (callback == NULL) {
        AudioTap_SetLastError("Callback cannot be NULL");
        return NULL;
    }

    @autoreleasepool {
        AudioTapContext* context = (AudioTapContext*)calloc(1, sizeof(AudioTapContext));
        if (context == NULL) {
            AudioTap_SetLastError("Failed to allocate context");
            return NULL;
        }

        context->callback = callback;
        context->userData = userData;
        context->isRunning = 0;

        // Get the default output device
        context->deviceID = AudioTap_GetDefaultOutputDevice();
        if (context->deviceID == kAudioDeviceUnknown) {
            free(context);
            return NULL;
        }

        // Get the device stream format
        AudioStreamBasicDescription format;
        OSStatus status = AudioTap_GetDeviceStreamFormat(context->deviceID, &format);
        if (status != noErr) {
            AudioTap_SetLastError("Failed to get device stream format");
            free(context);
            return NULL;
        }

        context->channels = format.mChannelsPerFrame;
        context->sampleRate = format.mSampleRate;

        // Create the IO proc
        status = AudioDeviceCreateIOProcID(
            context->deviceID,
            AudioTap_IOProc,
            context,
            &context->procID
        );

        if (status != noErr) {
            char errorMsg[256];
            snprintf(errorMsg, sizeof(errorMsg), "Failed to create audio IO proc: %d", status);
            AudioTap_SetLastError(errorMsg);
            free(context);
            return NULL;
        }

        context->tap = NULL;
        os_log(OS_LOG_DEFAULT, "Core Audio Tap created successfully");
        return (AudioTapHandle)context;
    }
}

// Start the audio tap
static inline int AudioTap_Start(AudioTapHandle handle) {
    if (handle == NULL) {
        AudioTap_SetLastError("Invalid handle");
        return -1;
    }

    AudioTapContext* context = (AudioTapContext*)handle;

    if (context->isRunning) {
        AudioTap_SetLastError("Audio tap already running");
        return -1;
    }

    OSStatus status = AudioDeviceStart(context->deviceID, context->procID);
    if (status != noErr) {
        char errorMsg[256];
        snprintf(errorMsg, sizeof(errorMsg), "Failed to start audio device: %d", status);
        AudioTap_SetLastError(errorMsg);
        return -1;
    }

    context->isRunning = 1;
    os_log(OS_LOG_DEFAULT, "Core Audio Tap started");
    return 0;
}

// Stop the audio tap
static inline void AudioTap_Stop(AudioTapHandle handle) {
    if (handle == NULL) {
        return;
    }

    AudioTapContext* context = (AudioTapContext*)handle;

    if (!context->isRunning) {
        return;
    }

    AudioDeviceStop(context->deviceID, context->procID);
    context->isRunning = 0;

    os_log(OS_LOG_DEFAULT, "Core Audio Tap stopped");
}

// Destroy the audio tap
static inline void AudioTap_Destroy(AudioTapHandle handle) {
    if (handle == NULL) {
        return;
    }

    AudioTapContext* context = (AudioTapContext*)handle;

    // Stop if running
    if (context->isRunning) {
        AudioTap_Stop(handle);
    }

    // Destroy the IO proc
    if (context->procID != NULL) {
        AudioDeviceDestroyIOProcID(context->deviceID, context->procID);
    }

    // Release the tap if it exists
    if (context->tap != NULL) {
        CFRelease(context->tap);
    }

    free(context);
    os_log(OS_LOG_DEFAULT, "Core Audio Tap destroyed");
}

// Get the last error message
static inline const char* AudioTap_GetLastError() {
    return audioTapLastError;
}

#endif // __APPLE__

#endif // AUDIO_TAP_IMPL_H
