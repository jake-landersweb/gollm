package main

import "github.com/jake-landersweb/gollm/src/ltypes"

type TokenRecord struct {
	Model        string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

func NewTokenRecordFromGPTUsage(model string, usage *ltypes.GPTUsage) *TokenRecord {
	return &TokenRecord{
		Model:        model,
		InputTokens:  usage.PromptTokens,
		OutputTokens: usage.CompletionTokens,
		TotalTokens:  usage.TotalTokens,
	}
}

func NewTokenRecordFromAnthropicUsage(model string, usage *ltypes.AnthropicUsage) *TokenRecord {
	return &TokenRecord{
		Model:        model,
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		TotalTokens:  usage.InputTokens + usage.OutputTokens,
	}
}
