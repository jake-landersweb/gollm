package tokens

import (
	"github.com/google/uuid"
	"github.com/jake-landersweb/gollm/v2/src/ltypes"
)

type TokenRecord struct {
	ID           uuid.UUID // ID to ensure usage is not reported twice
	Model        string    // for pricing calculations
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

func NewTokenRecord(model string, input int, output int, total int) *TokenRecord {
	id, _ := uuid.NewV7()
	return &TokenRecord{
		ID:           id,
		Model:        model,
		InputTokens:  input,
		OutputTokens: output,
		TotalTokens:  total,
	}
}

func NewTokenRecordFromGPTUsage(model string, usage *ltypes.GPTUsage) *TokenRecord {
	id, _ := uuid.NewV7()
	return &TokenRecord{
		ID:           id,
		Model:        model,
		InputTokens:  usage.PromptTokens,
		OutputTokens: usage.CompletionTokens,
		TotalTokens:  usage.TotalTokens,
	}
}

func NewTokenRecordFromAnthropicUsage(model string, usage *ltypes.AnthropicUsage) *TokenRecord {
	id, _ := uuid.NewV7()
	return &TokenRecord{
		ID:           id,
		Model:        model,
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		TotalTokens:  usage.InputTokens + usage.OutputTokens,
	}
}
