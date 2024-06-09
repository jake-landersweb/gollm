package gollm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGPTMessageConversion(t *testing.T) {
	messages := getTestConversation()
	returned := MessagesFromOpenAI(MessagesToOpenAI(messages))

	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}
}

func TestGeminiMessageConversion(t *testing.T) {
	messages := getTestConversation()

	// run through a cycle
	returned := MessagesFromGemini(MessagesToGemini(messages))

	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}
}

func TestAnthropicMessageConversion(t *testing.T) {
	messages := getTestConversation()

	// run through a cycle
	returned := MessagesFromAnthropic(MessagesToAnthropic(messages))

	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}
}

func TestMultiModelMessageConversion(t *testing.T) {
	messages := getTestConversation()

	// gemini -> gpt
	returned := MessagesFromOpenAI(MessagesToOpenAI(MessagesFromGemini(MessagesToGemini(messages))))
	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}

	// gpt -> gemini
	returned = MessagesFromGemini(MessagesToGemini(MessagesFromOpenAI(MessagesToOpenAI(messages))))
	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}

	// gemini -> anthropic
	returned = MessagesFromAnthropic(MessagesToAnthropic(MessagesFromGemini(MessagesToGemini(messages))))
	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}

	// anthropic -> gemini
	returned = MessagesFromGemini(MessagesToGemini(MessagesFromAnthropic(MessagesToAnthropic(messages))))
	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}

	// gpt -> anthropic
	returned = MessagesFromAnthropic(MessagesToAnthropic(MessagesFromOpenAI(MessagesToOpenAI(messages))))
	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}

	// anthropic -> gpt
	returned = MessagesFromOpenAI(MessagesToOpenAI(MessagesFromAnthropic(MessagesToAnthropic(messages))))
	for idx := range messages {
		assert.Equal(t, messages[idx].Message, returned[idx].Message)
	}
}

func getTestConversation() []*Message {
	messages := make([]*Message, 0)
	messages = append(messages, NewSystemMessage("Conduct this conversation like you are a pirate"))
	messages = append(messages, NewUserMessage("ahoy matey, how are you this fine hour"))
	messages = append(messages, &Message{
		Role:    RoleAI,
		Message: "Follow me, I'll lead ye to the palm tree where the treasure be buried. But be warned, it be guarded by the cursed spirit of Captain Blackbeard himself!",
	})
	messages = append(messages, NewUserMessage("Ah! fantastic! Why thank you matey"))
	return messages
}
