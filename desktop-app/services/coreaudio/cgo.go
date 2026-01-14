// go:build darwin && cgo
// +build darwin,cgo

package coreaudio

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework CoreAudio -framework AudioToolbox -framework Foundation
import "C"

// This file sets up CGo compilation for the coreaudio package on macOS
