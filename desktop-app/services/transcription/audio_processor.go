package transcription

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/sirupsen/logrus"
)

// AudioChunk represents a processed audio chunk ready for transcription
type AudioChunk struct {
	Samples         []float32 `json:"samples"`
	StartTime       float64   `json:"start_time"`       // Seconds from beginning of recording
	EndTime         float64   `json:"end_time"`         // Seconds from beginning of recording
	SampleRate      int       `json:"sample_rate"`
	ChunkIndex      int       `json:"chunk_index"`
	OriginalPath    string    `json:"original_path"`
	ActivityStartTime time.Time `json:"activity_start_time"` // For timeline correlation
}

// AudioProcessor handles audio preprocessing for optimal Whisper transcription
type AudioProcessor struct {
	sampleRate      int
	channels        int
	logger          *logrus.Logger
}

// NewAudioProcessor creates a new audio processor
func NewAudioProcessor(logger *logrus.Logger) *AudioProcessor {
	return &AudioProcessor{
		sampleRate: 16000, // Whisper optimal sample rate
		channels:   1,     // Mono for Whisper
		logger:     logger,
	}
}

// PrepareForWhisper converts audio to optimal format for Whisper processing
func (p *AudioProcessor) PrepareForWhisper(inputPath string) ([]float32, error) {
	p.logger.WithField("input_path", inputPath).Info("Preparing audio for Whisper")

	// Open the audio file
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Decode WAV file
	decoder := wav.NewDecoder(file)
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("invalid WAV file: %s", inputPath)
	}

	// Get audio format
	format := decoder.Format()
	p.logger.WithFields(logrus.Fields{
		"sample_rate": format.SampleRate,
		"channels":    format.NumChannels,
		"bit_depth":   decoder.BitDepth,
	}).Debug("Loaded audio file metadata")

	// Read all audio data
	buf := &audio.IntBuffer{
		Format: format,
		Data:   make([]int, 0),
	}

	// Read the entire file
	for {
		// Read in chunks to avoid memory issues with large files
		chunk := &audio.IntBuffer{
			Format: format,
			Data:   make([]int, 4096),
		}

		n, err := decoder.PCMBuffer(chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to read audio data: %w", err)
		}
		if n == 0 {
			break
		}

		buf.Data = append(buf.Data, chunk.Data[:n]...)
	}

	p.logger.WithField("samples", len(buf.Data)).Debug("Read audio samples")

	if len(buf.Data) == 0 {
		return nil, fmt.Errorf("no audio data found in file")
	}

	// Convert to float32 samples
	samples := p.convertToFloat32(buf.Data, int(decoder.BitDepth))

	// Convert to mono if stereo
	if format.NumChannels > 1 {
		samples = p.convertToMono(samples, format.NumChannels)
		p.logger.Debug("Converted stereo to mono")
	}

	// Resample if not 16kHz
	if format.SampleRate != p.sampleRate {
		samples = p.resample(samples, format.SampleRate, p.sampleRate)
		p.logger.WithFields(logrus.Fields{
			"from": format.SampleRate,
			"to":   p.sampleRate,
		}).Debug("Resampled audio")
	}

	// Normalize audio levels
	samples = p.NormalizeAudio(samples)
	p.logger.Debug("Normalized audio")

	// Remove silence from beginning and end
	samples = p.RemoveSilence(samples, 0.01)
	p.logger.Debug("Removed silence")

	p.logger.WithFields(logrus.Fields{
		"final_samples":    len(samples),
		"duration_seconds": float64(len(samples)) / float64(p.sampleRate),
	}).Info("Audio preprocessing completed")

	return samples, nil
}

// convertToFloat32 converts integer PCM samples to float32 normalized to [-1.0, 1.0]
func (p *AudioProcessor) convertToFloat32(intSamples []int, bitDepth int) []float32 {
	samples := make([]float32, len(intSamples))

	// Calculate normalization factor based on bit depth
	var maxValue float64
	switch bitDepth {
	case 8:
		maxValue = 128.0
	case 16:
		maxValue = 32768.0
	case 24:
		maxValue = 8388608.0
	case 32:
		maxValue = 2147483648.0
	default:
		maxValue = 32768.0 // Assume 16-bit if unknown
	}

	for i, sample := range intSamples {
		samples[i] = float32(float64(sample) / maxValue)
	}

	return samples
}

// convertToMono converts multi-channel audio to mono by averaging channels
func (p *AudioProcessor) convertToMono(samples []float32, numChannels int) []float32 {
	if numChannels == 1 {
		return samples
	}

	monoSamples := make([]float32, len(samples)/numChannels)

	for i := 0; i < len(monoSamples); i++ {
		var sum float32
		for ch := 0; ch < numChannels; ch++ {
			sum += samples[i*numChannels+ch]
		}
		monoSamples[i] = sum / float32(numChannels)
	}

	return monoSamples
}

// resample converts audio from one sample rate to another using linear interpolation
func (p *AudioProcessor) resample(samples []float32, fromRate, toRate int) []float32 {
	if fromRate == toRate {
		return samples
	}

	// Calculate output length
	ratio := float64(toRate) / float64(fromRate)
	outputLen := int(float64(len(samples)) * ratio)
	resampled := make([]float32, outputLen)

	// Simple linear interpolation
	for i := 0; i < outputLen; i++ {
		// Map output index to input space
		srcPos := float64(i) / ratio
		srcIndex := int(srcPos)
		srcFrac := srcPos - float64(srcIndex)

		// Boundary check
		if srcIndex >= len(samples)-1 {
			resampled[i] = samples[len(samples)-1]
			continue
		}

		// Linear interpolation
		sample1 := samples[srcIndex]
		sample2 := samples[srcIndex+1]
		resampled[i] = sample1 + float32(srcFrac)*(sample2-sample1)
	}

	return resampled
}

// ChunkAudio splits audio into overlapping chunks for better transcription accuracy
func (p *AudioProcessor) ChunkAudio(samples []float32, chunkDuration time.Duration, overlapDuration time.Duration, activityStartTime time.Time, inputPath string) []AudioChunk {
	if len(samples) == 0 {
		return []AudioChunk{}
	}
	
	chunkSamples := int(float64(p.sampleRate) * chunkDuration.Seconds())
	overlapSamples := int(float64(p.sampleRate) * overlapDuration.Seconds())
	
	var chunks []AudioChunk
	chunkIndex := 0
	
	for start := 0; start < len(samples); start += (chunkSamples - overlapSamples) {
		end := start + chunkSamples
		if end > len(samples) {
			end = len(samples)
		}
		
		// Skip tiny chunks at the end
		if end-start < p.sampleRate { // Less than 1 second
			break
		}
		
		chunkSamples := make([]float32, end-start)
		copy(chunkSamples, samples[start:end])
		
		startTime := float64(start) / float64(p.sampleRate)
		endTime := float64(end) / float64(p.sampleRate)
		
		chunk := AudioChunk{
			Samples:           chunkSamples,
			StartTime:         startTime,
			EndTime:           endTime,
			SampleRate:        p.sampleRate,
			ChunkIndex:        chunkIndex,
			OriginalPath:      inputPath,
			ActivityStartTime: activityStartTime,
		}
		
		chunks = append(chunks, chunk)
		chunkIndex++
	}
	
	p.logger.WithFields(logrus.Fields{
		"input_path":   inputPath,
		"total_chunks": len(chunks),
		"chunk_duration": chunkDuration,
		"overlap_duration": overlapDuration,
	}).Info("Audio chunked successfully")
	
	return chunks
}

// NormalizeAudio applies audio normalization to improve transcription quality
func (p *AudioProcessor) NormalizeAudio(samples []float32) []float32 {
	if len(samples) == 0 {
		return samples
	}
	
	// Find peak amplitude
	var maxAmp float32 = 0
	for _, sample := range samples {
		amp := float32(math.Abs(float64(sample)))
		if amp > maxAmp {
			maxAmp = amp
		}
	}
	
	// Avoid division by zero
	if maxAmp == 0 {
		return samples
	}
	
	// Normalize to 70% of maximum to avoid clipping
	targetLevel := float32(0.7)
	factor := targetLevel / maxAmp
	
	normalized := make([]float32, len(samples))
	for i, sample := range samples {
		normalized[i] = sample * factor
	}
	
	return normalized
}

// RemoveSilence removes silence from beginning and end of audio
func (p *AudioProcessor) RemoveSilence(samples []float32, threshold float32) []float32 {
	if len(samples) == 0 {
		return samples
	}
	
	// Find first non-silent sample
	start := 0
	for i, sample := range samples {
		if math.Abs(float64(sample)) > float64(threshold) {
			start = i
			break
		}
	}
	
	// Find last non-silent sample
	end := len(samples) - 1
	for i := len(samples) - 1; i >= 0; i-- {
		if math.Abs(float64(samples[i])) > float64(threshold) {
			end = i
			break
		}
	}
	
	// Ensure we have valid range
	if start >= end {
		return samples
	}
	
	// Add small padding to avoid cutting speech
	padding := p.sampleRate / 10 // 0.1 seconds
	start = int(math.Max(0, float64(start-padding)))
	end = int(math.Min(float64(len(samples)-1), float64(end+padding)))
	
	return samples[start:end+1]
}

// GetAudioDuration calculates the duration of audio samples
func (p *AudioProcessor) GetAudioDuration(samples []float32) time.Duration {
	if len(samples) == 0 || p.sampleRate == 0 {
		return 0
	}
	
	seconds := float64(len(samples)) / float64(p.sampleRate)
	return time.Duration(seconds * float64(time.Second))
}

// ValidateAudioQuality checks if audio quality is suitable for transcription
func (p *AudioProcessor) ValidateAudioQuality(samples []float32) (bool, string) {
	if len(samples) == 0 {
		return false, "no audio data"
	}
	
	duration := p.GetAudioDuration(samples)
	if duration < time.Second {
		return false, "audio too short (less than 1 second)"
	}
	
	// Check for clipping
	clippedSamples := 0
	for _, sample := range samples {
		if math.Abs(float64(sample)) >= 0.99 {
			clippedSamples++
		}
	}
	
	clippingRate := float64(clippedSamples) / float64(len(samples))
	if clippingRate > 0.1 {
		return false, fmt.Sprintf("audio severely clipped (%.1f%% of samples)", clippingRate*100)
	}
	
	// Check signal level
	var sum float64
	for _, sample := range samples {
		sum += math.Abs(float64(sample))
	}
	avgLevel := sum / float64(len(samples))
	
	if avgLevel < 0.01 {
		return false, "audio level too low (possible silence or very quiet recording)"
	}
	
	if clippingRate > 0.05 {
		return true, fmt.Sprintf("some clipping detected (%.1f%% of samples)", clippingRate*100)
	}
	
	return true, "audio quality is good"
}

// EstimateProcessingTime estimates how long transcription will take
func (p *AudioProcessor) EstimateProcessingTime(samples []float32, modelSize string) time.Duration {
	audioDuration := p.GetAudioDuration(samples)
	
	// Rough estimates based on model size (these would be calibrated in practice)
	var processingFactor float64
	switch modelSize {
	case "tiny":
		processingFactor = 0.1 // 10x faster than real-time
	case "small":
		processingFactor = 0.3 // 3x faster than real-time
	case "medium":
		processingFactor = 0.5 // 2x faster than real-time
	case "large":
		processingFactor = 1.0 // roughly real-time
	default:
		processingFactor = 0.5
	}
	
	estimated := time.Duration(float64(audioDuration) * processingFactor)
	
	// Add base overhead
	overhead := 10 * time.Second
	return estimated + overhead
}