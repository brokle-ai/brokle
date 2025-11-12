package openai

import (
	"brokle/internal/core/domain/gateway"
)

// OpenAI model registry with pricing, capabilities, and metadata
// This data is used to populate the database and make routing decisions

// ModelDefinition defines the structure for OpenAI model metadata
type ModelDefinition struct {
	Name                  string                 `json:"name"`
	DisplayName           string                 `json:"display_name"`
	Type                  gateway.ModelType      `json:"type"`
	InputCostPer1kTokens  float64                `json:"input_cost_per_1k_tokens"`
	OutputCostPer1kTokens float64                `json:"output_cost_per_1k_tokens"`
	MaxContextTokens      int                    `json:"max_context_tokens"`
	SupportsStreaming     bool                   `json:"supports_streaming"`
	SupportsFunctions     bool                   `json:"supports_functions"`
	SupportsVision        bool                   `json:"supports_vision"`
	QualityScore          *float64               `json:"quality_score,omitempty"`
	SpeedScore            *float64               `json:"speed_score,omitempty"`
	Metadata              map[string]interface{} `json:"metadata"`
	IsDeprecated          bool                   `json:"is_deprecated"`
	IsEnabled             bool                   `json:"is_enabled"`
}

// GetOpenAIModels returns the complete registry of OpenAI models
func GetOpenAIModels() map[string]ModelDefinition {
	return map[string]ModelDefinition{
		// GPT-4 Models
		"gpt-4": {
			Name:                  "gpt-4",
			DisplayName:           "GPT-4",
			Type:                  gateway.ModelTypeText,
			InputCostPer1kTokens:  0.03,
			OutputCostPer1kTokens: 0.06,
			MaxContextTokens:      8192,
			SupportsStreaming:     true,
			SupportsFunctions:     true,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.95),
			SpeedScore:            float64Ptr(0.60),
			Metadata: map[string]interface{}{
				"family":          "gpt-4",
				"training_cutoff": "2023-04",
				"capabilities":    []string{"text_generation", "function_calling", "json_mode"},
				"context_window":  8192,
				"description":     "Most capable GPT-4 model. Best for complex, multi-step tasks.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
		"gpt-4-0613": {
			Name:                  "gpt-4-0613",
			DisplayName:           "GPT-4 (June 2023)",
			Type:                  gateway.ModelTypeText,
			InputCostPer1kTokens:  0.03,
			OutputCostPer1kTokens: 0.06,
			MaxContextTokens:      8192,
			SupportsStreaming:     true,
			SupportsFunctions:     true,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.95),
			SpeedScore:            float64Ptr(0.60),
			Metadata: map[string]interface{}{
				"family":          "gpt-4",
				"training_cutoff": "2023-04",
				"snapshot_date":   "2023-06-13",
				"capabilities":    []string{"text_generation", "function_calling"},
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
		"gpt-4-turbo": {
			Name:                  "gpt-4-turbo",
			DisplayName:           "GPT-4 Turbo",
			Type:                  gateway.ModelTypeText,
			InputCostPer1kTokens:  0.01,
			OutputCostPer1kTokens: 0.03,
			MaxContextTokens:      128000,
			SupportsStreaming:     true,
			SupportsFunctions:     true,
			SupportsVision:        true,
			QualityScore:          float64Ptr(0.93),
			SpeedScore:            float64Ptr(0.75),
			Metadata: map[string]interface{}{
				"family":          "gpt-4",
				"training_cutoff": "2024-04",
				"capabilities":    []string{"text_generation", "function_calling", "vision", "json_mode"},
				"context_window":  128000,
				"description":     "GPT-4 Turbo with vision. More efficient than GPT-4.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
		"gpt-4-turbo-preview": {
			Name:                  "gpt-4-turbo-preview",
			DisplayName:           "GPT-4 Turbo Preview",
			Type:                  gateway.ModelTypeText,
			InputCostPer1kTokens:  0.01,
			OutputCostPer1kTokens: 0.03,
			MaxContextTokens:      128000,
			SupportsStreaming:     true,
			SupportsFunctions:     true,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.92),
			SpeedScore:            float64Ptr(0.75),
			Metadata: map[string]interface{}{
				"family":          "gpt-4",
				"training_cutoff": "2024-04",
				"capabilities":    []string{"text_generation", "function_calling", "json_mode"},
				"context_window":  128000,
				"preview":         true,
			},
			IsDeprecated: true,
			IsEnabled:    false,
		},
		"gpt-4o": {
			Name:                  "gpt-4o",
			DisplayName:           "GPT-4o",
			Type:                  gateway.ModelTypeMultimodal,
			InputCostPer1kTokens:  0.005,
			OutputCostPer1kTokens: 0.015,
			MaxContextTokens:      128000,
			SupportsStreaming:     true,
			SupportsFunctions:     true,
			SupportsVision:        true,
			QualityScore:          float64Ptr(0.92),
			SpeedScore:            float64Ptr(0.85),
			Metadata: map[string]interface{}{
				"family":          "gpt-4o",
				"training_cutoff": "2023-10",
				"capabilities":    []string{"text_generation", "function_calling", "vision", "json_mode", "multimodal"},
				"context_window":  128000,
				"description":     "GPT-4 Omni: multimodal flagship model, cheaper and faster than GPT-4 Turbo.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
		"gpt-4o-mini": {
			Name:                  "gpt-4o-mini",
			DisplayName:           "GPT-4o Mini",
			Type:                  gateway.ModelTypeMultimodal,
			InputCostPer1kTokens:  0.00015,
			OutputCostPer1kTokens: 0.0006,
			MaxContextTokens:      128000,
			SupportsStreaming:     true,
			SupportsFunctions:     true,
			SupportsVision:        true,
			QualityScore:          float64Ptr(0.85),
			SpeedScore:            float64Ptr(0.95),
			Metadata: map[string]interface{}{
				"family":          "gpt-4o",
				"training_cutoff": "2023-10",
				"capabilities":    []string{"text_generation", "function_calling", "vision", "json_mode", "multimodal"},
				"context_window":  128000,
				"description":     "Affordable and intelligent small model for fast, lightweight tasks.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
		"gpt-4-vision-preview": {
			Name:                  "gpt-4-vision-preview",
			DisplayName:           "GPT-4 Vision Preview",
			Type:                  gateway.ModelTypeMultimodal,
			InputCostPer1kTokens:  0.01,
			OutputCostPer1kTokens: 0.03,
			MaxContextTokens:      128000,
			SupportsStreaming:     false,
			SupportsFunctions:     false,
			SupportsVision:        true,
			QualityScore:          float64Ptr(0.90),
			SpeedScore:            float64Ptr(0.70),
			Metadata: map[string]interface{}{
				"family":          "gpt-4",
				"training_cutoff": "2023-04",
				"capabilities":    []string{"text_generation", "vision"},
				"context_window":  128000,
				"preview":         true,
			},
			IsDeprecated: true,
			IsEnabled:    false,
		},

		// GPT-3.5 Models
		"gpt-3.5-turbo": {
			Name:                  "gpt-3.5-turbo",
			DisplayName:           "GPT-3.5 Turbo",
			Type:                  gateway.ModelTypeText,
			InputCostPer1kTokens:  0.001,
			OutputCostPer1kTokens: 0.002,
			MaxContextTokens:      16385,
			SupportsStreaming:     true,
			SupportsFunctions:     true,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.80),
			SpeedScore:            float64Ptr(0.90),
			Metadata: map[string]interface{}{
				"family":          "gpt-3.5",
				"training_cutoff": "2023-09",
				"capabilities":    []string{"text_generation", "function_calling", "json_mode"},
				"context_window":  16385,
				"description":     "Most capable GPT-3.5 model and optimized for chat.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
		"gpt-3.5-turbo-16k": {
			Name:                  "gpt-3.5-turbo-16k",
			DisplayName:           "GPT-3.5 Turbo 16K",
			Type:                  gateway.ModelTypeText,
			InputCostPer1kTokens:  0.003,
			OutputCostPer1kTokens: 0.004,
			MaxContextTokens:      16385,
			SupportsStreaming:     true,
			SupportsFunctions:     true,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.80),
			SpeedScore:            float64Ptr(0.85),
			Metadata: map[string]interface{}{
				"family":          "gpt-3.5",
				"training_cutoff": "2023-09",
				"capabilities":    []string{"text_generation", "function_calling"},
				"context_window":  16385,
			},
			IsDeprecated: true,
			IsEnabled:    false,
		},
		"gpt-3.5-turbo-1106": {
			Name:                  "gpt-3.5-turbo-1106",
			DisplayName:           "GPT-3.5 Turbo (November 2023)",
			Type:                  gateway.ModelTypeText,
			InputCostPer1kTokens:  0.001,
			OutputCostPer1kTokens: 0.002,
			MaxContextTokens:      16385,
			SupportsStreaming:     true,
			SupportsFunctions:     true,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.80),
			SpeedScore:            float64Ptr(0.90),
			Metadata: map[string]interface{}{
				"family":          "gpt-3.5",
				"training_cutoff": "2023-09",
				"snapshot_date":   "2023-11-06",
				"capabilities":    []string{"text_generation", "function_calling", "json_mode"},
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
		"gpt-3.5-turbo-instruct": {
			Name:                  "gpt-3.5-turbo-instruct",
			DisplayName:           "GPT-3.5 Turbo Instruct",
			Type:                  gateway.ModelTypeText,
			InputCostPer1kTokens:  0.0015,
			OutputCostPer1kTokens: 0.002,
			MaxContextTokens:      4096,
			SupportsStreaming:     true,
			SupportsFunctions:     false,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.75),
			SpeedScore:            float64Ptr(0.92),
			Metadata: map[string]interface{}{
				"family":          "gpt-3.5",
				"training_cutoff": "2023-09",
				"capabilities":    []string{"text_generation", "instruct_following"},
				"context_window":  4096,
				"description":     "Similar capabilities to text-davinci-003 but compatible with legacy Completions endpoint.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},

		// Embedding Models
		"text-embedding-3-small": {
			Name:                  "text-embedding-3-small",
			DisplayName:           "Text Embedding 3 Small",
			Type:                  gateway.ModelTypeEmbedding,
			InputCostPer1kTokens:  0.00002,
			OutputCostPer1kTokens: 0.0,
			MaxContextTokens:      8191,
			SupportsStreaming:     false,
			SupportsFunctions:     false,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.85),
			SpeedScore:            float64Ptr(0.95),
			Metadata: map[string]interface{}{
				"family":       "text-embedding-3",
				"dimensions":   1536,
				"capabilities": []string{"text_embedding"},
				"description":  "Most efficient embedding model for text similarity.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
		"text-embedding-3-large": {
			Name:                  "text-embedding-3-large",
			DisplayName:           "Text Embedding 3 Large",
			Type:                  gateway.ModelTypeEmbedding,
			InputCostPer1kTokens:  0.00013,
			OutputCostPer1kTokens: 0.0,
			MaxContextTokens:      8191,
			SupportsStreaming:     false,
			SupportsFunctions:     false,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.92),
			SpeedScore:            float64Ptr(0.85),
			Metadata: map[string]interface{}{
				"family":       "text-embedding-3",
				"dimensions":   3072,
				"capabilities": []string{"text_embedding"},
				"description":  "Most powerful embedding model for text similarity.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
		"text-embedding-ada-002": {
			Name:                  "text-embedding-ada-002",
			DisplayName:           "Text Embedding Ada 002",
			Type:                  gateway.ModelTypeEmbedding,
			InputCostPer1kTokens:  0.0001,
			OutputCostPer1kTokens: 0.0,
			MaxContextTokens:      8191,
			SupportsStreaming:     false,
			SupportsFunctions:     false,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.80),
			SpeedScore:            float64Ptr(0.90),
			Metadata: map[string]interface{}{
				"family":       "ada",
				"dimensions":   1536,
				"capabilities": []string{"text_embedding"},
				"description":  "Most capable embedding model.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},

		// Legacy Models (mostly deprecated but some still available)
		"davinci-002": {
			Name:                  "davinci-002",
			DisplayName:           "Davinci 002",
			Type:                  gateway.ModelTypeText,
			InputCostPer1kTokens:  0.002,
			OutputCostPer1kTokens: 0.002,
			MaxContextTokens:      16384,
			SupportsStreaming:     true,
			SupportsFunctions:     false,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.70),
			SpeedScore:            float64Ptr(0.80),
			Metadata: map[string]interface{}{
				"family":         "davinci",
				"capabilities":   []string{"text_generation"},
				"context_window": 16384,
				"description":    "Legacy completion model.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
		"babbage-002": {
			Name:                  "babbage-002",
			DisplayName:           "Babbage 002",
			Type:                  gateway.ModelTypeText,
			InputCostPer1kTokens:  0.0004,
			OutputCostPer1kTokens: 0.0004,
			MaxContextTokens:      16384,
			SupportsStreaming:     true,
			SupportsFunctions:     false,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.60),
			SpeedScore:            float64Ptr(0.90),
			Metadata: map[string]interface{}{
				"family":         "babbage",
				"capabilities":   []string{"text_generation"},
				"context_window": 16384,
				"description":    "Legacy completion model. Capable of simple tasks, fast, and cost-effective.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},

		// Audio Models
		"whisper-1": {
			Name:                  "whisper-1",
			DisplayName:           "Whisper",
			Type:                  gateway.ModelTypeAudio,
			InputCostPer1kTokens:  0.006, // Per minute of audio
			OutputCostPer1kTokens: 0.0,
			MaxContextTokens:      25 * 1024 * 1024, // 25MB file limit
			SupportsStreaming:     false,
			SupportsFunctions:     false,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.90),
			SpeedScore:            float64Ptr(0.85),
			Metadata: map[string]interface{}{
				"family":            "whisper",
				"capabilities":      []string{"speech_to_text", "translation"},
				"supported_formats": []string{"mp3", "mp4", "mpeg", "mpga", "m4a", "wav", "webm"},
				"description":       "Speech recognition model for transcription and translation.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
		"tts-1": {
			Name:                  "tts-1",
			DisplayName:           "Text-to-Speech 1",
			Type:                  gateway.ModelTypeAudio,
			InputCostPer1kTokens:  0.015,
			OutputCostPer1kTokens: 0.0,
			MaxContextTokens:      4096,
			SupportsStreaming:     true,
			SupportsFunctions:     false,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.85),
			SpeedScore:            float64Ptr(0.95),
			Metadata: map[string]interface{}{
				"family":       "tts",
				"capabilities": []string{"text_to_speech"},
				"voices":       []string{"alloy", "echo", "fable", "onyx", "nova", "shimmer"},
				"description":  "Text-to-speech model optimized for speed.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
		"tts-1-hd": {
			Name:                  "tts-1-hd",
			DisplayName:           "Text-to-Speech 1 HD",
			Type:                  gateway.ModelTypeAudio,
			InputCostPer1kTokens:  0.030,
			OutputCostPer1kTokens: 0.0,
			MaxContextTokens:      4096,
			SupportsStreaming:     true,
			SupportsFunctions:     false,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.92),
			SpeedScore:            float64Ptr(0.80),
			Metadata: map[string]interface{}{
				"family":       "tts",
				"capabilities": []string{"text_to_speech"},
				"voices":       []string{"alloy", "echo", "fable", "onyx", "nova", "shimmer"},
				"description":  "Text-to-speech model optimized for quality.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},

		// Image Models
		"dall-e-3": {
			Name:                  "dall-e-3",
			DisplayName:           "DALL-E 3",
			Type:                  gateway.ModelTypeImage,
			InputCostPer1kTokens:  0.080, // Per standard image (1024x1024)
			OutputCostPer1kTokens: 0.0,
			MaxContextTokens:      4000, // Prompt length limit
			SupportsStreaming:     false,
			SupportsFunctions:     false,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.95),
			SpeedScore:            float64Ptr(0.70),
			Metadata: map[string]interface{}{
				"family":       "dall-e",
				"capabilities": []string{"image_generation"},
				"sizes":        []string{"1024x1024", "1792x1024", "1024x1792"},
				"quality":      []string{"standard", "hd"},
				"description":  "Most advanced image generation model with improved prompt adherence.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
		"dall-e-2": {
			Name:                  "dall-e-2",
			DisplayName:           "DALL-E 2",
			Type:                  gateway.ModelTypeImage,
			InputCostPer1kTokens:  0.020, // Per 1024x1024 image
			OutputCostPer1kTokens: 0.0,
			MaxContextTokens:      1000, // Prompt length limit
			SupportsStreaming:     false,
			SupportsFunctions:     false,
			SupportsVision:        false,
			QualityScore:          float64Ptr(0.85),
			SpeedScore:            float64Ptr(0.85),
			Metadata: map[string]interface{}{
				"family":       "dall-e",
				"capabilities": []string{"image_generation", "image_edit", "image_variation"},
				"sizes":        []string{"256x256", "512x512", "1024x1024"},
				"description":  "Image generation model with editing capabilities.",
			},
			IsDeprecated: false,
			IsEnabled:    true,
		},
	}
}

// GetActiveOpenAIModels returns only enabled, non-deprecated models
func GetActiveOpenAIModels() map[string]ModelDefinition {
	allModels := GetOpenAIModels()
	activeModels := make(map[string]ModelDefinition)

	for name, model := range allModels {
		if model.IsEnabled && !model.IsDeprecated {
			activeModels[name] = model
		}
	}

	return activeModels
}

// GetOpenAIModelsByType returns models filtered by type
func GetOpenAIModelsByType(modelType gateway.ModelType) map[string]ModelDefinition {
	allModels := GetOpenAIModels()
	filteredModels := make(map[string]ModelDefinition)

	for name, model := range allModels {
		if model.Type == modelType && model.IsEnabled && !model.IsDeprecated {
			filteredModels[name] = model
		}
	}

	return filteredModels
}

// GetChatModels returns models suitable for chat completions
func GetChatModels() map[string]ModelDefinition {
	return GetOpenAIModelsByType(gateway.ModelTypeText)
}

// GetEmbeddingModels returns models suitable for embeddings
func GetEmbeddingModels() map[string]ModelDefinition {
	return GetOpenAIModelsByType(gateway.ModelTypeEmbedding)
}

// GetImageModels returns models suitable for image generation
func GetImageModels() map[string]ModelDefinition {
	return GetOpenAIModelsByType(gateway.ModelTypeImage)
}

// GetAudioModels returns models suitable for audio processing
func GetAudioModels() map[string]ModelDefinition {
	return GetOpenAIModelsByType(gateway.ModelTypeAudio)
}

// IsModelAvailable checks if a specific model is available and enabled
func IsModelAvailable(modelName string) bool {
	models := GetOpenAIModels()
	model, exists := models[modelName]
	return exists && model.IsEnabled && !model.IsDeprecated
}

// GetModelDefinition returns the definition for a specific model
func GetModelDefinition(modelName string) (ModelDefinition, bool) {
	models := GetOpenAIModels()
	model, exists := models[modelName]
	return model, exists
}

// GetDefaultChatModel returns the default model for chat completions
func GetDefaultChatModel() string {
	return "gpt-4o-mini" // Most cost-effective GPT-4 class model
}

// GetDefaultEmbeddingModel returns the default model for embeddings
func GetDefaultEmbeddingModel() string {
	return "text-embedding-3-small" // Most cost-effective embedding model
}

// Helper function to create float64 pointer
func float64Ptr(f float64) *float64 {
	return &f
}
