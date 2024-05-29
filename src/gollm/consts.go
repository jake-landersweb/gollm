package gollm

const test_user_id = "go-test"

const default_system_message = "You are a friendly and helpful AI assistant. Please respond with helpful and useful answers."

const gpt_base_url = "https://api.openai.com/v1/chat/completions"
const gpt3_model = "gpt-3.5-turbo-0125"

// const gpt4_model = "gpt-4-0125-preview"
const gpt_max_tokens = 8096

const gemini_base_url = "https://generativelanguage.googleapis.com/v1/models"
const gemini_system_message = "This is the system message of the conversation, and should be used as a general reference for the entire conversation"
const gemini_model = "gemini-pro"

const anthropic_base_url = "https://api.anthropic.com/v1/messages"
const anthropic_version = "2023-06-01"
const anthropic_claude2 = "claude-2.1"

// const anthropic_claude_instant = "claude-instant-1.2"
const anthropic_max_tokens = 4096

const openai_embeddings_base_url = "https://api.openai.com/v1/embeddings"
const openai_embeddings_dimensions = 512
const openai_embeddings_chunk_size_default = 512
const OPENAI_EMBEDDINGS_INPUT_MAX = 8191

type ModelOpenAIEmbeddings = string

const (
	OPENAI_EMBEDDINGS_MODEL ModelOpenAIEmbeddings = "text-embedding-3-small"
)
