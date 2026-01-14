package transcription

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/platformlabs-co/personal-assist/models"
	"github.com/sirupsen/logrus"
)

// TranscriptionCache handles caching of transcription results and models
type TranscriptionCache struct {
	audioHashes    map[string]*CachedTranscription
	modelCache     map[string]whisper.Model
	mutex          sync.RWMutex
	logger         *logrus.Logger
	maxCacheSize   int
	cleanupTicker  *time.Ticker
	cleanupStop    chan struct{}
}

// CachedTranscription represents a cached transcription result
type CachedTranscription struct {
	AudioHash    string                    `json:"audio_hash"`
	ModelID      string                    `json:"model_id"`
	Config       models.TranscriptionConfig `json:"config"`
	Chunks       []*models.TranscriptChunk `json:"chunks"`
	CachedAt     time.Time                 `json:"cached_at"`
	LastAccessed time.Time                 `json:"last_accessed"`
	AccessCount  int                       `json:"access_count"`
}

// NewTranscriptionCache creates a new transcription cache
func NewTranscriptionCache(maxCacheSize int, logger *logrus.Logger) *TranscriptionCache {
	cache := &TranscriptionCache{
		audioHashes:   make(map[string]*CachedTranscription),
		modelCache:    make(map[string]whisper.Model),
		logger:        logger,
		maxCacheSize:  maxCacheSize,
		cleanupStop:   make(chan struct{}),
	}

	// Start cleanup routine
	cache.cleanupTicker = time.NewTicker(30 * time.Minute)
	go cache.cleanupRoutine()

	return cache
}

// GetCachedTranscription checks if a transcription is cached
func (tc *TranscriptionCache) GetCachedTranscription(audioPath, modelID string, config models.TranscriptionConfig) ([]*models.TranscriptChunk, bool) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	// Generate cache key
	cacheKey, err := tc.generateCacheKey(audioPath, modelID, config)
	if err != nil {
		tc.logger.WithError(err).Warn("Failed to generate cache key")
		return nil, false
	}

	cached, exists := tc.audioHashes[cacheKey]
	if !exists {
		return nil, false
	}

	// Check if file hasn't changed
	if !tc.isFileUnchanged(audioPath, cached.AudioHash) {
		// File changed, remove from cache
		delete(tc.audioHashes, cacheKey)
		tc.logger.WithField("audio_path", audioPath).Debug("Cache invalidated due to file change")
		return nil, false
	}

	// Update access tracking
	cached.LastAccessed = time.Now()
	cached.AccessCount++

	tc.logger.WithFields(logrus.Fields{
		"audio_path":   audioPath,
		"model_id":     modelID,
		"chunks":       len(cached.Chunks),
		"access_count": cached.AccessCount,
	}).Debug("Cache hit for transcription")

	return cached.Chunks, true
}

// CacheTranscription stores a transcription result in the cache
func (tc *TranscriptionCache) CacheTranscription(audioPath, modelID string, config models.TranscriptionConfig, chunks []*models.TranscriptChunk) error {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	// Generate cache key
	cacheKey, err := tc.generateCacheKey(audioPath, modelID, config)
	if err != nil {
		return fmt.Errorf("failed to generate cache key: %w", err)
	}

	// Generate audio hash
	audioHash, err := tc.calculateFileHash(audioPath)
	if err != nil {
		return fmt.Errorf("failed to calculate audio hash: %w", err)
	}

	// Check cache size and cleanup if necessary
	if len(tc.audioHashes) >= tc.maxCacheSize {
		tc.cleanupOldestEntries()
	}

	cached := &CachedTranscription{
		AudioHash:    audioHash,
		ModelID:      modelID,
		Config:       config,
		Chunks:       chunks,
		CachedAt:     time.Now(),
		LastAccessed: time.Now(),
		AccessCount:  1,
	}

	tc.audioHashes[cacheKey] = cached

	tc.logger.WithFields(logrus.Fields{
		"audio_path": audioPath,
		"model_id":   modelID,
		"chunks":     len(chunks),
		"cache_size": len(tc.audioHashes),
	}).Debug("Transcription cached")

	return nil
}

// CacheModel stores a loaded Whisper model in memory
func (tc *TranscriptionCache) CacheModel(modelID string, model whisper.Model) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	// Close existing model if any
	if existingModel, exists := tc.modelCache[modelID]; exists {
		existingModel.Close()
	}

	tc.modelCache[modelID] = model

	tc.logger.WithField("model_id", modelID).Debug("Model cached")
}

// GetCachedModel retrieves a cached model
func (tc *TranscriptionCache) GetCachedModel(modelID string) (whisper.Model, bool) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	model, exists := tc.modelCache[modelID]
	return model, exists
}

// InvalidateAudioCache removes cached transcriptions for a specific audio file
func (tc *TranscriptionCache) InvalidateAudioCache(audioPath string) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	// Remove all cache entries for this audio file
	for key, cached := range tc.audioHashes {
		if cached.AudioHash != "" {
			// We'd need to store the original path to match properly
			// For now, this is a simplified implementation
			delete(tc.audioHashes, key)
		}
	}

	tc.logger.WithField("audio_path", audioPath).Debug("Audio cache invalidated")
}

// ClearModelCache removes all cached models
func (tc *TranscriptionCache) ClearModelCache() {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	// Close all models
	for _, model := range tc.modelCache {
		model.Close()
	}

	tc.modelCache = make(map[string]whisper.Model)
	tc.logger.Info("Model cache cleared")
}

// GetCacheStats returns cache statistics
func (tc *TranscriptionCache) GetCacheStats() CacheStats {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	stats := CacheStats{
		TranscriptionEntries: len(tc.audioHashes),
		ModelEntries:        len(tc.modelCache),
		MaxSize:             tc.maxCacheSize,
	}

	// Calculate total access count and average age
	var totalAccess int
	var totalAge time.Duration
	now := time.Now()

	for _, cached := range tc.audioHashes {
		totalAccess += cached.AccessCount
		totalAge += now.Sub(cached.CachedAt)
	}

	if len(tc.audioHashes) > 0 {
		stats.AverageAccessCount = float64(totalAccess) / float64(len(tc.audioHashes))
		stats.AverageAge = totalAge / time.Duration(len(tc.audioHashes))
	}

	return stats
}

// CacheStats represents cache performance statistics
type CacheStats struct {
	TranscriptionEntries int           `json:"transcription_entries"`
	ModelEntries        int           `json:"model_entries"`
	MaxSize             int           `json:"max_size"`
	AverageAccessCount  float64       `json:"average_access_count"`
	AverageAge          time.Duration `json:"average_age"`
}

// generateCacheKey creates a unique key for caching
func (tc *TranscriptionCache) generateCacheKey(audioPath, modelID string, config models.TranscriptionConfig) (string, error) {
	hasher := sha256.New()
	
	// Include audio file path, model ID, and relevant config parameters
	hasher.Write([]byte(audioPath))
	hasher.Write([]byte(modelID))
	hasher.Write([]byte(config.Language))
	hasher.Write([]byte(fmt.Sprintf("%.2f", config.Temperature)))
	hasher.Write([]byte(fmt.Sprintf("%t", config.EnableSpeakers)))
	hasher.Write([]byte(fmt.Sprintf("%t", config.EnableTimestamps)))
	
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// calculateFileHash calculates SHA256 hash of a file
func (tc *TranscriptionCache) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// isFileUnchanged checks if a file hasn't changed since caching
func (tc *TranscriptionCache) isFileUnchanged(filePath, cachedHash string) bool {
	currentHash, err := tc.calculateFileHash(filePath)
	if err != nil {
		return false
	}
	return currentHash == cachedHash
}

// cleanupRoutine periodically cleans up old cache entries
func (tc *TranscriptionCache) cleanupRoutine() {
	for {
		select {
		case <-tc.cleanupTicker.C:
			tc.cleanupOldEntries()
		case <-tc.cleanupStop:
			return
		}
	}
}

// cleanupOldEntries removes entries older than 24 hours with low access count
func (tc *TranscriptionCache) cleanupOldEntries() {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	now := time.Now()
	maxAge := 24 * time.Hour
	minAccessCount := 2

	var toDelete []string
	for key, cached := range tc.audioHashes {
		age := now.Sub(cached.CachedAt)
		if age > maxAge && cached.AccessCount < minAccessCount {
			toDelete = append(toDelete, key)
		}
	}

	for _, key := range toDelete {
		delete(tc.audioHashes, key)
	}

	if len(toDelete) > 0 {
		tc.logger.WithField("cleaned_entries", len(toDelete)).Info("Cache cleanup completed")
	}
}

// cleanupOldestEntries removes the oldest entries when cache is full
func (tc *TranscriptionCache) cleanupOldestEntries() {
	// Remove 25% of entries (oldest first)
	entriesToRemove := tc.maxCacheSize / 4
	if entriesToRemove == 0 {
		entriesToRemove = 1
	}

	// Sort by last accessed time
	type cacheEntry struct {
		key    string
		cached *CachedTranscription
	}

	var entries []cacheEntry
	for key, cached := range tc.audioHashes {
		entries = append(entries, cacheEntry{key, cached})
	}

	// Simple sort by last accessed time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].cached.LastAccessed.After(entries[j].cached.LastAccessed) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove oldest entries
	for i := 0; i < entriesToRemove && i < len(entries); i++ {
		delete(tc.audioHashes, entries[i].key)
	}

	tc.logger.WithField("removed_entries", entriesToRemove).Debug("Removed oldest cache entries")
}

// Close shuts down the cache and releases resources
func (tc *TranscriptionCache) Close() error {
	// Stop cleanup routine
	tc.cleanupTicker.Stop()
	close(tc.cleanupStop)

	// Clear all caches
	tc.ClearModelCache()
	
	tc.mutex.Lock()
	tc.audioHashes = make(map[string]*CachedTranscription)
	tc.mutex.Unlock()

	tc.logger.Info("Transcription cache closed")
	return nil
}