package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGPTMessageConversion(t *testing.T) {
	messages := getTestConversation()
	returned := LLMMessagesFromGPT(LLMMessagesToGPT(messages))

	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}
}

func TestGeminiMessageConversion(t *testing.T) {
	messages := getTestConversation()

	// run through a cycle
	returned := LLMMessagesFromGemini(LLMMessagesToGemini(messages))

	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}
}

func TestAnthropicMessageConversion(t *testing.T) {
	messages := getTestConversation()

	// run through a cycle
	returned := LLMMessagesFromAnthropic(LLMMessagesToAnthropic(messages))

	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}
}

func TestMultiModelMessageConversion(t *testing.T) {
	messages := getTestConversation()

	// gemini -> gpt
	returned := LLMMessagesFromGPT(LLMMessagesToGPT(LLMMessagesFromGemini(LLMMessagesToGemini(messages))))
	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}

	// gpt -> gemini
	returned = LLMMessagesFromGemini(LLMMessagesToGemini(LLMMessagesFromGPT(LLMMessagesToGPT(messages))))
	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}

	// gemini -> anthropic
	returned = LLMMessagesFromAnthropic(LLMMessagesToAnthropic(LLMMessagesFromGemini(LLMMessagesToGemini(messages))))
	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}

	// anthropic -> gemini
	returned = LLMMessagesFromGemini(LLMMessagesToGemini(LLMMessagesFromAnthropic(LLMMessagesToAnthropic(messages))))
	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}

	// gpt -> anthropic
	returned = LLMMessagesFromAnthropic(LLMMessagesToAnthropic(LLMMessagesFromGPT(LLMMessagesToGPT(messages))))
	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}

	// anthropic -> gpt
	returned = LLMMessagesFromGPT(LLMMessagesToGPT(LLMMessagesFromAnthropic(LLMMessagesToAnthropic(messages))))
	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}
}

func getTestConversation() []*LanguageModelMessage {
	messages := make([]*LanguageModelMessage, 0)
	messages = append(messages, NewSystemMessage("Conduct this conversation like you are a pirate"))
	messages = append(messages, NewUserMessage("ahoy matey, how are you this fine hour"))
	messages = append(messages, &LanguageModelMessage{
		Role:    RoleAI,
		Message: "Follow me, I'll lead ye to the palm tree where the treasure be buried. But be warned, it be guarded by the cursed spirit of Captain Blackbeard himself!",
	})
	messages = append(messages, NewUserMessage("Ah! fantastic! Why thank you matey"))
	return messages
}
