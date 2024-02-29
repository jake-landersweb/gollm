package ltypes

// Content represents the base structured datatype containing multi-part content of a message.
type GemContent struct {
	Parts []GemPart `json:"parts"` // Ordered Parts that constitute a single message. Parts may have different MIME types.
	Role  string    `json:"role"`  // Optional. The producer of the content. Must be either 'user' or 'model'.
}

// Part represents a datatype containing media that is part of a multi-part Content message.
type GemPart struct {
	Text             string               `json:"text,omitempty"`             // Inline text.
	InlineData       *GemBlob             `json:"inlineData,omitempty"`       // Inline media bytes.
	FunctionCall     *GemFunctionCall     `json:"functionCall,omitempty"`     // A predicted FunctionCall returned from the model.
	FunctionResponse *GemFunctionResponse `json:"functionResponse,omitempty"` // The result output of a FunctionCall.
}

// Blob represents raw media bytes. Text should not be sent as raw bytes, use the 'text' field.
type GemBlob struct {
	MimeType string `json:"mimeType"` // The IANA standard MIME type of the source data.
	Data     string `json:"data"`     // Raw bytes for media formats. A base64-encoded string.
}

// FunctionCall represents a predicted FunctionCall returned from the model.
type GemFunctionCall struct {
	Name string                 `json:"name"` // Required. The name of the function to call.
	Args map[string]interface{} `json:"args"` // Optional. The function parameters and values in JSON object format.
}

// FunctionResponse represents the result output from a FunctionCall.
type GemFunctionResponse struct {
	Name     string                 `json:"name"`     // Required. The name of the function to call.
	Response map[string]interface{} `json:"response"` // Required. The function response in JSON object format.
}
