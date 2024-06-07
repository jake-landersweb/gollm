package gollm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeminiTokens(t *testing.T) {
	tokens, err := geminiTokenizerAccurate("This is an input string where I would like to know how many tokens make it up. Some grammer, can also be us'ed potentially (hopefully): yes.", gemini_model)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	fmt.Println("Tokens:", tokens)
	assert.Equal(t, 35, tokens)
}

func TestGeminiTextCompletion(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestGeminiTextCompletion")
	ctx := context.Background()

	// make the messages
	raw := make([]*LanguageModelMessage, 0)
	raw = append(raw, NewSystemMessage("You are a model that is being used to validate that method calls to your api work in a go testing environment."))
	raw = append(raw, NewUserMessage("Please respond with a single sentence."))
	messages := LLMMessagesToGemini(raw)

	llm := NewLanguageModel(test_user_id, logger, nil)
	response, err := llm.geminiCompletion(ctx, logger, gemini_model, 0.5, false, "", messages)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	fmt.Println(*response)
}

func TestGeminiJSONCompletion(t *testing.T) {
	logger := defaultLogger(slog.LevelDebug).With("test", "TestGeminiTextCompletion")
	ctx := context.Background()

	schema := `{"message": string, "date": int}`

	// make the messages
	raw := make([]*LanguageModelMessage, 0)
	raw = append(raw, NewSystemMessage("You are a model that is being used to validate that method calls to your api work in a go testing environment."))
	raw = append(raw, NewUserMessage("Please respond with a reasonable response."))
	messages := LLMMessagesToGemini(raw)

	fmt.Println(*messages[0])

	llm := NewLanguageModel(test_user_id, logger, nil)
	response, err := llm.geminiCompletion(ctx, logger, gemini_model, 0.5, true, schema, messages)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	// parse the json
	tmp := struct {
		Message string `json:"message"`
		Date    int    `json:"date"`
	}{}
	err = json.Unmarshal([]byte(response.Candidates[0].Content.Parts[0].Text), &tmp)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.NotEmpty(t, tmp.Message)
	assert.NotEmpty(t, tmp.Date)
}
