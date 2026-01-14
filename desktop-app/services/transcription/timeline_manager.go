package transcription

import (
	"fmt"
	"sort"
	"time"

	"github.com/platformlabs-co/personal-assist/models"
	"github.com/sirupsen/logrus"
)

// TimelineManager handles activity timeline correlation and multi-recording processing
type TimelineManager struct {
	activityStartTime time.Time
	logger           *logrus.Logger
}

// NewTimelineManager creates a new timeline manager
func NewTimelineManager(activityStartTime time.Time, logger *logrus.Logger) *TimelineManager {
	return &TimelineManager{
		activityStartTime: activityStartTime,
		logger:           logger,
	}
}

// ConvertToActivityTime converts absolute timestamp to activity-relative seconds
func (tm *TimelineManager) ConvertToActivityTime(absoluteTime time.Time) float64 {
	return absoluteTime.Sub(tm.activityStartTime).Seconds()
}

// ConvertToAbsoluteTime converts activity-relative seconds to absolute time
func (tm *TimelineManager) ConvertToAbsoluteTime(activityRelativeSeconds float64) time.Time {
	return tm.activityStartTime.Add(time.Duration(activityRelativeSeconds * float64(time.Second)))
}

// CorrelateChunks adjusts chunk timestamps relative to activity start
func (tm *TimelineManager) CorrelateChunks(chunks []*models.TranscriptChunk, recording *models.AudioRecording) {
	// Calculate recording start delay relative to activity
	recordingDelay := tm.ConvertToActivityTime(recording.CreatedAt)
	
	tm.logger.WithFields(logrus.Fields{
		"recording_id":    recording.ID,
		"recording_delay": recordingDelay,
		"chunk_count":     len(chunks),
	}).Debug("Correlating chunks to activity timeline")
	
	// Adjust all chunk timestamps
	for _, chunk := range chunks {
		chunk.StartTime += recordingDelay
		chunk.EndTime += recordingDelay
	}
}

// MergeOverlappingChunks merges overlapping transcript chunks from multiple recordings
func (tm *TimelineManager) MergeOverlappingChunks(allChunks []*models.TranscriptChunk) []*models.TranscriptChunk {
	if len(allChunks) == 0 {
		return allChunks
	}
	
	// Sort chunks by start time
	sort.Slice(allChunks, func(i, j int) bool {
		return allChunks[i].StartTime < allChunks[j].StartTime
	})
	
	tm.logger.WithField("chunk_count", len(allChunks)).Debug("Merging overlapping chunks")
	
	var merged []*models.TranscriptChunk
	current := allChunks[0]
	
	for i := 1; i < len(allChunks); i++ {
		next := allChunks[i]
		
		// Check if chunks overlap significantly
		overlap := tm.calculateOverlap(current, next)
		
		if overlap > 0.5 { // More than 50% overlap
			// Merge chunks
			current = tm.mergeChunks(current, next)
			tm.logger.WithFields(logrus.Fields{
				"overlap":    overlap,
				"merged_ids": []string{current.ID, next.ID},
			}).Debug("Merged overlapping chunks")
		} else {
			// No significant overlap, add current and move to next
			merged = append(merged, current)
			current = next
		}
	}
	
	// Add the last chunk
	merged = append(merged, current)
	
	tm.logger.WithFields(logrus.Fields{
		"original_count": len(allChunks),
		"merged_count":   len(merged),
	}).Info("Chunk merging completed")
	
	return merged
}

// calculateOverlap calculates the overlap ratio between two chunks
func (tm *TimelineManager) calculateOverlap(chunk1, chunk2 *models.TranscriptChunk) float64 {
	// Calculate overlap duration
	overlapStart := max(chunk1.StartTime, chunk2.StartTime)
	overlapEnd := min(chunk1.EndTime, chunk2.EndTime)
	
	if overlapEnd <= overlapStart {
		return 0 // No overlap
	}
	
	overlapDuration := overlapEnd - overlapStart
	
	// Calculate overlap relative to the shorter chunk
	duration1 := chunk1.EndTime - chunk1.StartTime
	duration2 := chunk2.EndTime - chunk2.StartTime
	shorterDuration := min(duration1, duration2)
	
	if shorterDuration <= 0 {
		return 0
	}
	
	return overlapDuration / shorterDuration
}

// mergeChunks combines two overlapping chunks into one
func (tm *TimelineManager) mergeChunks(chunk1, chunk2 *models.TranscriptChunk) *models.TranscriptChunk {
	// Use the chunk with higher confidence as the base
	var baseChunk, otherChunk *models.TranscriptChunk
	if chunk1.GetConfidence() >= chunk2.GetConfidence() {
		baseChunk, otherChunk = chunk1, chunk2
	} else {
		baseChunk, otherChunk = chunk2, chunk1
	}
	
	// Create merged chunk
	merged := &models.TranscriptChunk{
		ID:               baseChunk.ID, // Keep base chunk ID
		UserID:          baseChunk.UserID,
		ActivityID:      baseChunk.ActivityID,
		AudioRecordingID: baseChunk.AudioRecordingID,
		Text:            tm.mergeText(baseChunk.Text, otherChunk.Text),
		StartTime:       min(chunk1.StartTime, chunk2.StartTime),
		EndTime:         max(chunk1.EndTime, chunk2.EndTime),
		Speaker:         tm.mergeSpeaker(baseChunk.Speaker, otherChunk.Speaker),
		Confidence:      tm.mergeConfidence(baseChunk.Confidence, otherChunk.Confidence),
		Language:        baseChunk.Language,
		CreatedAt:       baseChunk.CreatedAt,
	}
	
	return merged
}

// mergeText combines text from two chunks, preferring the more confident one
func (tm *TimelineManager) mergeText(text1, text2 string) string {
	// For now, just return the first text
	// In a more sophisticated implementation, you might:
	// - Compare confidence scores
	// - Use fuzzy string matching to find the best combination
	// - Apply NLP techniques to merge similar content
	
	if len(text1) >= len(text2) {
		return text1
	}
	return text2
}

// mergeSpeaker combines speaker information from two chunks
func (tm *TimelineManager) mergeSpeaker(speaker1, speaker2 *string) *string {
	if speaker1 != nil && *speaker1 != "Unknown" {
		return speaker1
	}
	if speaker2 != nil && *speaker2 != "Unknown" {
		return speaker2
	}
	return speaker1 // Return first one even if unknown
}

// mergeConfidence combines confidence scores from two chunks
func (tm *TimelineManager) mergeConfidence(confidence1, confidence2 *float64) *float64 {
	if confidence1 == nil && confidence2 == nil {
		return nil
	}
	if confidence1 == nil {
		return confidence2
	}
	if confidence2 == nil {
		return confidence1
	}
	
	// Take average of confidences
	avg := (*confidence1 + *confidence2) / 2.0
	return &avg
}

// ProcessMultipleRecordings processes all recordings for an activity in chronological order
func (tm *TimelineManager) ProcessMultipleRecordings(recordings []*models.AudioRecording, processor *WhisperProcessor, activity *models.Activity) ([]*models.TranscriptChunk, error) {
	// Sort recordings by creation time
	sort.Slice(recordings, func(i, j int) bool {
		return recordings[i].CreatedAt.Before(recordings[j].CreatedAt)
	})
	
	tm.logger.WithFields(logrus.Fields{
		"activity_id":     activity.ID,
		"recording_count": len(recordings),
	}).Info("Processing multiple recordings")
	
	var allChunks []*models.TranscriptChunk
	
	// Process each recording
	for i, recording := range recordings {
		tm.logger.WithFields(logrus.Fields{
			"recording_index": i,
			"recording_id":    recording.ID,
			"file_path":       recording.FilePath,
		}).Info("Processing recording")
		
		chunks, err := processor.ProcessRecording(recording, activity)
		if err != nil {
			tm.logger.WithError(err).WithField("recording_id", recording.ID).Error("Failed to process recording")
			continue
		}
		
		// Correlate chunks to activity timeline
		tm.CorrelateChunks(chunks, recording)
		
		allChunks = append(allChunks, chunks...)
	}
	
	// Merge overlapping chunks
	mergedChunks := tm.MergeOverlappingChunks(allChunks)
	
	tm.logger.WithFields(logrus.Fields{
		"activity_id":       activity.ID,
		"total_chunks":      len(allChunks),
		"merged_chunks":     len(mergedChunks),
		"recordings_processed": len(recordings),
	}).Info("Multiple recording processing completed")
	
	return mergedChunks, nil
}

// ValidateChunkTimeline ensures chunks are in chronological order and don't have major gaps
func (tm *TimelineManager) ValidateChunkTimeline(chunks []*models.TranscriptChunk) (bool, []string) {
	if len(chunks) == 0 {
		return true, nil
	}
	
	var issues []string
	
	// Check chronological order
	for i := 1; i < len(chunks); i++ {
		if chunks[i].StartTime < chunks[i-1].StartTime {
			issues = append(issues, fmt.Sprintf("Chunk %d starts before previous chunk", i))
		}
	}
	
	// Check for major gaps (more than 30 seconds)
	maxGap := 30.0
	for i := 1; i < len(chunks); i++ {
		gap := chunks[i].StartTime - chunks[i-1].EndTime
		if gap > maxGap {
			issues = append(issues, fmt.Sprintf("Large gap (%.1fs) between chunks %d and %d", gap, i-1, i))
		}
	}
	
	return len(issues) == 0, issues
}

// GetActivityDuration calculates the total duration covered by transcript chunks
func (tm *TimelineManager) GetActivityDuration(chunks []*models.TranscriptChunk) float64 {
	if len(chunks) == 0 {
		return 0
	}
	
	// Sort by start time
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].StartTime < chunks[j].StartTime
	})
	
	return chunks[len(chunks)-1].EndTime - chunks[0].StartTime
}

// Helper functions
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}