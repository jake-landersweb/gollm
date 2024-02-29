package ltypes

type GemCompletionResponse struct {
	Candidates     []GemCandidate    `json:"candidates"`
	PromptFeedback GemPromptFeedback `json:"promptFeedback"`
	Error          *GemError         `json:"error"`
}

type GemCandidate struct {
	Content          GemContent          `json:"content"`
	FinishReason     string              `json:"finishReason"`
	Index            int                 `json:"index"`
	TokenCount       int                 `json:"tokenCount"`
	CitationMetadata GemCitationMetadata `json:"citationMetadata"`
	SafetyRatings    []GemSafetyRating   `json:"safetyRatings"`
}

type GemCitationMetadata struct {
	CitationSources []GemCitationSource `json:"citationSources"`
}

type GemCitationSource struct {
	StartIndex int    `json:"startIndex"`
	EndIndex   int    `json:"endIndex"`
	Uri        string `json:"uri"`
	License    string `json:"license"`
}

type GemPromptFeedback struct {
	SafetyRatings []GemSafetyRating `json:"safetyRatings"`
}

type GemSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type GemErrorStatus string

const (
	GEM_ERROR_OK                  GemErrorStatus = "OK"
	GEM_ERROR_CANCELLED           GemErrorStatus = "CANCELLED"
	GEM_ERROR_UNKNOWN             GemErrorStatus = "UNKNOWN"
	GEM_ERROR_INVALID_ARGUMENT    GemErrorStatus = "INVALID_ARGUMENT"
	GEM_ERROR_DEADLINE_EXCEEDED   GemErrorStatus = "DEADLINE_EXCEEDED"
	GEM_ERROR_NOT_FOUND           GemErrorStatus = "NOT_FOUND"
	GEM_ERROR_ALREADY_EXISTS      GemErrorStatus = "ALREADY_EXISTS"
	GEM_ERROR_PERMISSION_DENIED   GemErrorStatus = "PERMISSION_DENIED"
	GEM_ERROR_RESOURCE_EXHAUSTED  GemErrorStatus = "RESOURCE_EXHAUSTED"
	GEM_ERROR_FAILED_PRECONDITION GemErrorStatus = "FAILED_PRECONDITION"
	GEM_ERROR_ABORTED             GemErrorStatus = "ABORTED"
	GEM_ERROR_OUT_OF_RANGE        GemErrorStatus = "OUT_OF_RANGE"
	GEM_ERROR_UNIMPLEMENTED       GemErrorStatus = "UNIMPLEMENTED"
	GEM_ERROR_INTERNAL            GemErrorStatus = "INTERNAL"
	GEM_ERROR_UNAVAILABLE         GemErrorStatus = "UNAVAILABLE"
	GEM_ERROR_DATA_LOSS           GemErrorStatus = "DATA_LOSS"
	GEM_ERROR_UNAUTHENTICATED     GemErrorStatus = "UNAUTHENTICATED"
)

type GemError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Status  GemErrorStatus `json:"status"`
}
