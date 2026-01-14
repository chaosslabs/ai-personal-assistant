.PHONY: dev build deps clean version-bump
.PHONY: whisper-deps-macos whisper-verify
.PHONY: frontend-install frontend-build
.PHONY: build-macos package-macos
.PHONY: ci-build-macos

# Detect OS
UNAME_S := $(shell uname -s)
WORKSPACE := $(PWD)

# =============================================================================
# Development Commands
# =============================================================================

# Run development server (macOS)
dev:
	CGO_CFLAGS="-I$(PWD)/whisper" CGO_LDFLAGS="-L$(PWD)/whisper -Wl,-rpath,$(PWD)/whisper -lwhisper -lggml" wails dev

# Install Go dependencies
deps:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -rf build/
	rm -rf whisper/
	go clean

# =============================================================================
# Whisper.cpp Dependencies
# =============================================================================

# Build whisper.cpp dependencies for macOS (arm64/Apple Silicon)
whisper-deps-macos:
	@echo "Checking for cmake..."
	@which cmake > /dev/null || (echo "cmake not found. Installing via Homebrew..." && brew install cmake)
	@echo "Building whisper.cpp dependencies for macOS arm64..."
	@rm -rf /tmp/whisper-build
	@mkdir -p /tmp/whisper-build
	cd /tmp/whisper-build && \
		git clone https://github.com/ggerganov/whisper.cpp . && \
		mkdir -p build && cd build && \
		cmake .. -DCMAKE_BUILD_TYPE=Release \
			-DBUILD_SHARED_LIBS=OFF \
			-DCMAKE_OSX_ARCHITECTURES="arm64" \
			-DGGML_NATIVE=OFF \
			-DGGML_METAL=ON \
			-DGGML_BLAS=ON \
			-DGGML_CPU=ON && \
		make -j8
	@echo "Creating whisper directory and copying libraries..."
	@mkdir -p whisper
	cp /tmp/whisper-build/build/ggml/src/libggml.a ./whisper/ 2>/dev/null || true
	cp /tmp/whisper-build/build/ggml/src/libggml-*.a ./whisper/ 2>/dev/null || true
	cp /tmp/whisper-build/build/ggml/src/ggml-blas/libggml-blas.a ./whisper/ 2>/dev/null || true
	cp /tmp/whisper-build/build/ggml/src/ggml-metal/libggml-metal.a ./whisper/ 2>/dev/null || true
	cp /tmp/whisper-build/build/ggml/src/ggml-cpu/libggml-cpu.a ./whisper/ 2>/dev/null || true
	cp /tmp/whisper-build/build/ggml/src/ggml-base/libggml-base.a ./whisper/ 2>/dev/null || true
	cp /tmp/whisper-build/build/ggml/src/libggml-base.a ./whisper/ 2>/dev/null || true
	cp /tmp/whisper-build/build/src/libwhisper.a ./whisper/
	cp /tmp/whisper-build/ggml/include/*.h ./whisper/
	cp /tmp/whisper-build/include/whisper.h ./whisper/
	@echo "Cleaning up temporary build..."
	rm -rf /tmp/whisper-build
	@echo "✅ Whisper dependencies installed for macOS arm64"

# Verify whisper setup
whisper-verify:
	@echo "Checking whisper directory..."
	@ls -la whisper/ || echo "whisper directory not found!"
	@echo ""
	@echo "Checking for header files..."
	@ls -la whisper/*.h || echo "No header files found!"
	@echo ""
	@echo "Checking for library files..."
	@ls -la whisper/*.a || echo "No library files found!"
ifeq ($(UNAME_S),Darwin)
	@echo ""
	@echo "Checking library architectures:"
	@for lib in whisper/*.a; do \
		echo "Library: $$lib"; \
		lipo -info "$$lib" 2>/dev/null || file "$$lib"; \
	done
endif

# =============================================================================
# Frontend Commands
# =============================================================================

# Install frontend dependencies
frontend-install:
	cd frontend && npm ci

# Build frontend
frontend-build:
	cd frontend && npm run build

# =============================================================================
# Platform-Specific Build Commands
# =============================================================================

# Build Wails app for macOS (arm64/Apple Silicon)
build-macos:
	@echo "Building Wails app for macOS arm64..."
	CGO_CFLAGS="-I$(WORKSPACE)/whisper" \
	CGO_LDFLAGS="-L$(WORKSPACE)/whisper -lwhisper -lggml-base -lggml-cpu -lggml-metal -lggml-blas -framework Accelerate -framework Foundation -framework Metal -framework MetalKit" \
	wails build -platform darwin/arm64 -clean
	@echo "✅ macOS build complete"

# Package macOS app
package-macos:
	@echo "Packaging macOS app..."
	cd build/bin && tar -czf memoria-darwin-arm64.tar.gz Memoria.app
	@echo "✅ macOS package created: build/bin/memoria-darwin-arm64.tar.gz"

# =============================================================================
# CI Build Commands (Full Build Pipeline)
# =============================================================================

# Complete CI build for macOS
ci-build-macos: deps whisper-deps-macos whisper-verify frontend-install frontend-build build-macos package-macos
	@echo "✅ Complete macOS CI build finished"

# =============================================================================
# Version Management
# =============================================================================

# Bump version in version.txt and wails.json
version-bump:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make version-bump VERSION=v0.2.0"; \
		exit 1; \
	fi
	@echo "Updating version to $(VERSION)..."
	@echo "$(VERSION)" > version.txt
	@sed -i.bak 's/"productVersion": ".*"/"productVersion": "$(VERSION)"/' wails.json && rm wails.json.bak
	@echo "✅ Version updated to $(VERSION) in version.txt and wails.json"

# =============================================================================
# Help
# =============================================================================

help:
	@echo "Memoria Build System"
	@echo ""
	@echo "Development Commands:"
	@echo "  make dev                 - Run development server (macOS)"
	@echo "  make deps                - Install Go dependencies"
	@echo "  make clean               - Clean build artifacts"
	@echo ""
	@echo "Whisper.cpp Dependencies:"
	@echo "  make whisper-deps-macos  - Build whisper.cpp for macOS (arm64)"
	@echo "  make whisper-verify      - Verify whisper.cpp installation"
	@echo ""
	@echo "Frontend Commands:"
	@echo "  make frontend-install    - Install frontend dependencies"
	@echo "  make frontend-build      - Build frontend"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build-macos         - Build app for macOS (arm64)"
	@echo "  make package-macos       - Package macOS app"
	@echo ""
	@echo "CI Commands (Full Pipeline):"
	@echo "  make ci-build-macos      - Complete macOS CI build"
	@echo ""
	@echo "Version Management:"
	@echo "  make version-bump VERSION=v0.2.0  - Bump version"
