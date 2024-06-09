package tokens

import (
	"github.com/google/uuid"
	"github.com/jake-landersweb/gollm/v2/src/ltypes"
)

type UsageRecord struct {
	ID           uuid.UUID // ID to ensure usage is not reported twice
	Model        string    // for pricing calculations
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

func NewUsageRecord(model string, input int, output int, total int) *UsageRecord {
	id, _ := uuid.NewV7()
	return &UsageRecord{
		ID:           id,
		Model:        model,
		InputTokens:  input,
		OutputTokens: output,
		TotalTokens:  total,
	}
}

func NewUsageRecordFromGPTUsage(model string, usage *ltypes.GPTUsage) *UsageRecord {
	id, _ := uuid.NewV7()
	return &UsageRecord{
		ID:           id,
		Model:        model,
		InputTokens:  usage.PromptTokens,
		OutputTokens: usage.CompletionTokens,
		TotalTokens:  usage.TotalTokens,
	}
}

func NewUsageRecordFromAnthropicUsage(model string, usage *ltypes.AnthropicUsage) *UsageRecord {
	id, _ := uuid.NewV7()
	return &UsageRecord{
		ID:           id,
		Model:        model,
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		TotalTokens:  usage.InputTokens + usage.OutputTokens,
	}
}
