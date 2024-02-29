package ltypes

// SafetySetting represents the safety setting affecting the safety-blocking behavior.
type GemSafetySetting struct {
	Category  GemHarmCategory       `json:"category"`  // Required. The category for this setting.
	Threshold GemHarmBlockThreshold `json:"threshold"` // Required. Controls the probability threshold at which harm is blocked.
}

// HarmCategory represents the categories of harm that can be used to adjust safety settings.
type GemHarmCategory string // Extend this as necessary with enum values.

// HarmBlockThreshold represents the thresholds at which harm is blocked.
type GemHarmBlockThreshold string

const (
	// HarmBlockThresholdUnspecified indicates an unspecified threshold.
	HarmBlockThresholdUnspecified GemHarmBlockThreshold = "HARM_BLOCK_THRESHOLD_UNSPECIFIED"
	// BlockLowAndAbove indicates that content with NEGLIGIBLE will be allowed.
	BlockLowAndAbove GemHarmBlockThreshold = "BLOCK_LOW_AND_ABOVE"
	// BlockMediumAndAbove indicates that content with NEGLIGIBLE and LOW will be allowed.
	BlockMediumAndAbove GemHarmBlockThreshold = "BLOCK_MEDIUM_AND_ABOVE"
	// BlockOnlyHigh indicates that content with NEGLIGIBLE, LOW, and MEDIUM will be allowed.
	BlockOnlyHigh GemHarmBlockThreshold = "BLOCK_ONLY_HIGH"
	// BlockNone indicates that all content will be allowed.
	BlockNone GemHarmBlockThreshold = "BLOCK_NONE"
)
