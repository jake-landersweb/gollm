package ltypes

// GenerationConfig represents configuration options for model generation and outputs.
type GemGenerationConfig struct {
	StopSequences   []string `json:"stopSequences,omitempty"`   // Optional. The set of character sequences that will stop output generation.
	CandidateCount  int      `json:"candidateCount,omitempty"`  // Optional. Number of generated responses to return.
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"` // Optional. The maximum number of tokens to include in a candidate.
	Temperature     float64  `json:"temperature,omitempty"`     // Optional. Controls the randomness of the output. From 0-1
	TopP            float64  `json:"topP,omitempty"`            // Optional. The maximum cumulative probability of tokens to consider when sampling.
	TopK            int      `json:"topK,omitempty"`            // Optional. The maximum number of tokens to consider when sampling.
}
