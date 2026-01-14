# Release Process

This document describes how to create a new release of Memoria (Personal AI Assistant).

## Version Management

The version number is stored in two places:
- `desktop-app/version.txt` - Source of truth for version number
- `desktop-app/wails.json` - Product version used by Wails build system

Use `make version-bump` to update both files simultaneously.

**Version Format:**
- Development: `v0.1.0-dev`
- Alpha: `v0.1.0-alpha.1`
- Beta: `v0.1.0-beta.1`
- Release: `v0.1.0`

## Creating a Release

### 1. Update Version

Use the Makefile command to update the version in both `version.txt` and `wails.json`:

```bash
cd desktop-app
make version-bump VERSION=v0.2.0
```

This will automatically update:
- `desktop-app/version.txt`
- `desktop-app/wails.json` (productVersion field)

### 2. Commit Version Change

```bash
git add desktop-app/version.txt desktop-app/wails.json
git commit -m "Bump version to v0.2.0"
git push origin main
```

### 3. Trigger Release Workflow

1. Go to GitHub Actions in your repository
2. Select the "Release Build" workflow
3. Click "Run workflow"
4. Optionally add release notes
5. Click "Run workflow"

The workflow will:
- Read the version from `version.txt`
- Create a Git tag and GitHub release
- Build binaries for all platforms:
  - macOS (Apple Silicon - arm64)
- Upload all binaries to the GitHub release

### 4. Verify Release

1. Check the GitHub Actions run completed successfully
2. Verify the release is created at `https://github.com/YOUR_USERNAME/ai-personal-assistant/releases`
3. Download and test binaries on target platforms

## Platform-Specific Builds

### macOS
- **Apple Silicon (M1/M2/M3)**: `memoria-darwin-arm64.tar.gz`

Users need to:
1. Extract the `.tar.gz` file
2. Move `Memoria.app` to Applications folder
3. Right-click and "Open" first time (for unsigned apps)

## Troubleshooting

### Build Fails
- Check GitHub Actions logs
- Ensure all dependencies are properly listed in `go.mod` and `package.json`
- Verify Wails configuration in `wails.json`

### Missing Dependencies
- **macOS**: Requires Xcode Command Line Tools

### Pre-release vs Release
- Versions containing `dev`, `alpha`, or `beta` are marked as pre-releases
- Pre-releases are shown differently on the GitHub releases page
- To create a stable release, use version format: `vX.Y.Z`

## Manual Release (Alternative)

If you need to build locally instead of using GitHub Actions:

```bash
# Build for current platform
cd desktop-app
wails build -clean

# Build for specific platform
wails build -platform darwin/arm64 -clean
```

Binaries will be in `desktop-app/build/bin/`

## Notes

- The workflow requires `GITHUB_TOKEN` which is automatically provided by GitHub Actions
- First-time releases require repository permissions to create releases
- Unsigned macOS builds require users to allow the app in System Preferences > Security & Privacy
- For signed releases, see macOS code signing documentation
