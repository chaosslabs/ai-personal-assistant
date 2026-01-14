package models

// WhisperModel represents a Whisper model available for transcription
type WhisperModel struct {
	ID           string   `json:"id"`            // "tiny", "small", "medium", "large"
	Name         string   `json:"name"`          // Display name
	Size         int64    `json:"size"`          // File size in bytes
	IsDownloaded bool     `json:"is_downloaded"`
	IsActive     bool     `json:"is_active"`
	Languages    []string `json:"languages"`     // Supported languages
	Accuracy     string   `json:"accuracy"`      // "good", "better", "best"
	Speed        string   `json:"speed"`         // "fast", "medium", "slow"
}

// AvailableWhisperModels returns the list of available Whisper models
func AvailableWhisperModels() []WhisperModel {
	return []WhisperModel{
		{
			ID:           "tiny",
			Name:         "Tiny (39 MB)",
			Size:         39*1024*1024,
			IsDownloaded: false,
			IsActive:     false,
			Languages:    []string{"multilingual"},
			Accuracy:     "good",
			Speed:        "fast",
		},
		{
			ID:           "tiny.en",
			Name:         "Tiny English (39 MB)",
			Size:         39*1024*1024,
			IsDownloaded: false,
			IsActive:     false,
			Languages:    []string{"en"},
			Accuracy:     "good",
			Speed:        "fast",
		},
		{
			ID:           "small",
			Name:         "Small (465 MB)",
			Size:         487601967, // Actual size from HuggingFace
			IsDownloaded: false,
			IsActive:     false,
			Languages:    []string{"multilingual"},
			Accuracy:     "better",
			Speed:        "medium",
		},
		{
			ID:           "small.en",
			Name:         "Small English (465 MB)",
			Size:         487601967,
			IsDownloaded: false,
			IsActive:     false,
			Languages:    []string{"en"},
			Accuracy:     "better",
			Speed:        "medium",
		},
		{
			ID:           "medium",
			Name:         "Medium (769 MB)",
			Size:         769*1024*1024,
			IsDownloaded: false,
			IsActive:     false,
			Languages:    []string{"multilingual"},
			Accuracy:     "better",
			Speed:        "medium",
		},
		{
			ID:           "medium.en",
			Name:         "Medium English (769 MB)",
			Size:         769*1024*1024,
			IsDownloaded: false,
			IsActive:     false,
			Languages:    []string{"en"},
			Accuracy:     "better",
			Speed:        "medium",
		},
		{
			ID:           "large",
			Name:         "Large (1550 MB)",
			Size:         1550*1024*1024,
			IsDownloaded: false,
			IsActive:     false,
			Languages:    []string{"multilingual"},
			Accuracy:     "best",
			Speed:        "slow",
		},
		{
			ID:           "large-v1",
			Name:         "Large v1 (1550 MB)",
			Size:         1550*1024*1024,
			IsDownloaded: false,
			IsActive:     false,
			Languages:    []string{"multilingual"},
			Accuracy:     "best",
			Speed:        "slow",
		},
		{
			ID:           "large-v2",
			Name:         "Large v2 (1550 MB)",
			Size:         1550*1024*1024,
			IsDownloaded: false,
			IsActive:     false,
			Languages:    []string{"multilingual"},
			Accuracy:     "best",
			Speed:        "slow",
		},
		{
			ID:           "large-v3",
			Name:         "Large v3 (1550 MB)",
			Size:         1550*1024*1024,
			IsDownloaded: false,
			IsActive:     false,
			Languages:    []string{"multilingual"},
			Accuracy:     "best",
			Speed:        "slow",
		},
	}
}

// GetModelByID returns a model by its ID
func GetModelByID(id string) (*WhisperModel, bool) {
	models := AvailableWhisperModels()
	for i, model := range models {
		if model.ID == id {
			return &models[i], true
		}
	}
	return nil, false
}

// GetSizeMB returns the model size in megabytes
func (wm *WhisperModel) GetSizeMB() float64 {
	return float64(wm.Size) / (1024 * 1024)
}

// IsMultilingual returns true if the model supports multiple languages
func (wm *WhisperModel) IsMultilingual() bool {
	for _, lang := range wm.Languages {
		if lang == "multilingual" {
			return true
		}
	}
	return false
}

// SupportsLanguage returns true if the model supports the given language
func (wm *WhisperModel) SupportsLanguage(language string) bool {
	// Multilingual models support all languages
	if wm.IsMultilingual() {
		return true
	}
	
	// Check specific language support
	for _, lang := range wm.Languages {
		if lang == language {
			return true
		}
	}
	
	return false
}

// GetModelFilename returns the expected filename for this model
func (wm *WhisperModel) GetModelFilename() string {
	return "ggml-" + wm.ID + ".bin"
}

// ModelDownloadRequest represents a request to download a model
type ModelDownloadRequest struct {
	ModelID  string `json:"model_id"`
	Priority int    `json:"priority"`
}