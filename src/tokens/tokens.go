package tokens

import (
	"github.com/google/uuid"
	"github.com/jake-landersweb/gollm/src/ltypes"
)

type TokenRecord struct {
	ID           uuid.UUID // ID to ensure usage is not reported twice
	Model        string    // for pricing calculations
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

func NewTokenRecord(model string, input int, output int, total int) *TokenRecord {
	return &TokenRecord{
		ID:           uuid.New(),
		Model:        model,
		InputTokens:  input,
		OutputTokens: output,
		TotalTokens:  total,
	}
}

func NewTokenRecordFromGPTUsage(model string, usage *ltypes.GPTUsage) *TokenRecord {
	return &TokenRecord{
		ID:           uuid.New(),
		Model:        model,
		InputTokens:  usage.PromptTokens,
		OutputTokens: usage.CompletionTokens,
		TotalTokens:  usage.TotalTokens,
	}
}

func NewTokenRecordFromAnthropicUsage(model string, usage *ltypes.AnthropicUsage) *TokenRecord {
	return &TokenRecord{
		ID:           uuid.New(),
		Model:        model,
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		TotalTokens:  usage.InputTokens + usage.OutputTokens,
	}
}
