package main

/*
#cgo CFLAGS: -I${SRCDIR}/whisper
#cgo LDFLAGS: -L${SRCDIR}/whisper -lwhisper -lggml
*/
import "C"

// This file configures CGO to find the local whisper headers and libraries
// The whisper.cpp bindings will use these settings during compilation