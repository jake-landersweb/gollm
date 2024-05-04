# GoLanguageModel

Simple abstractions over LLM providers in Go to allow for complex LLM apps.

Currently supported LLMs:
- OpenAI GPT3.5
- OpenAI GPT4
- Google Gemini
- Anthropic Claude 2.1
- Anthropic Claude Instant 1.2

## LanguageModel Abstraction

The LLM abstraction allows for multiple LLMs to be used in the same conversation at different points. The `LanguageModel` object hosts the conversation state in a conversation object that is LLM agnostic, and when a specific LLM completion is called, the internal conversation state get transformed into the specific format for the LLM. Then, on response, the message is parsed from the LLM provider and stored in the agnostic state inside the LLM object.

This simple abstraction lets you mix and match different LLMs at any point of the conversation. For example, as seen in `llm_test.go`:

```go
model := NewLanguageModel(TEST_USER_ID, logger, "You are a pirate on a deserted island")

var err error
input1 := &CompletionInput{
    Model:       GEMINI_MODEL,
    Temperature: 0.7,
    Json:        false,
    Input:       "Where is the treasure matey?",
}
_, err = model.TokenEstimate(input1)
assert.Nil(t, err)

// run a gpt completion
_, err = model.GeminiCompletion(ctx, input1)
assert.Nil(t, err)
if err != nil {
    return
}

input2 := &CompletionInput{
    Model:       GPT3_MODEL,
    Temperature: 1.3,
    Json:        false,
    Input:       "Are you sure? You must show me now or suffer!",
}
_, err = model.TokenEstimate(input2)
assert.Nil(t, err)

// run a gemini completion
_, err = model.GPTCompletion(ctx, input2)
assert.Nil(t, err)
if err != nil {
    return
}

input3 := &CompletionInput{
    Model:       ANTHROPIC_CLAUDE2,
    Temperature: 0.7,
    Json:        false,
    Input:       "Aha! Thats more like it! Treasure for everyone!",
}
_, err = model.TokenEstimate(input3)
assert.Nil(t, err)

// run an anthropic completion
_, err = model.AnthropicCompletion(ctx, input3)
assert.Nil(t, err)
if err != nil {
    return
}

model.PrintConversation()

assert.Equal(t, 7, len(model.conversation))
```

In this example, first the conversation is started with `Gemini`. Then, the conversation is extended with `GPT 3.5`. Lastly, the conversation is finished with `Claude 2.1`. 

## Resources

### OpenAI

- [API error codes](https://platform.openai.com/docs/guides/error-codes/api-errors)
- [API error codes code](https://community.openai.com/t/openai-chat-list-of-error-codes-and-types/357791/11)
- [Tokens](https://help.openai.com/en/articles/4936856-what-are-tokens-and-how-to-count-them)
- [Estimating tokens](https://community.openai.com/t/what-is-the-openai-algorithm-to-calculate-tokens/58237/28)
- [API docs](https://platform.openai.com/docs/api-reference/audio/verbose-json-object)

### Gemini

- [Getting number of tokens](https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/get-token-count)
- [Model REST API docs](https://ai.google.dev/api/rest/v1beta/models)
- [REST API Quickstart](https://ai.google.dev/tutorials/rest_quickstart)
- [gRPC errors](https://google.aip.dev/193)
- [gRPC error codes](https://github.com/grpc/grpc/blob/master/doc/statuscodes.md)
- [Generate content API docs](https://ai.google.dev/api/rest/v1beta/models/generateContent)
- [Available endpoints](https://ai.google.dev/api/rest)
- 

### Anthropic

- [Glossary](https://docs.anthropic.com/claude/docs/glossary)
- [Avoiding hallucinations](https://docs.anthropic.com/claude/docs/let-claude-say-i-dont-know)
- [Prompting tips](https://docs.anthropic.com/claude/docs/configuring-gpt-prompts-for-claude)

### Other

- [Go tokenizer](https://github.com/sugarme/tokenizer)
- [Go tokenizer 2](https://github.com/tiktoken-go/tokenizer)
- [TikToken in Go](https://github.com/pkoukk/tiktoken-go)
- 