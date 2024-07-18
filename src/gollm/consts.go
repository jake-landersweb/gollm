package gollm

const test_user_id = "go-test"

const gpt_base_url = "https://api.openai.com/v1/chat/completions"
const gpt3_model = "gpt-3.5-turbo"
const gpt_max_tokens = 8096

const gemini_base_url = "https://generativelanguage.googleapis.com/v1beta/models"
const gemini_model = "gemini-1.5-flash"

const anthropic_base_url = "https://api.anthropic.com/v1/messages"
const anthropic_version = "2023-06-01"
const anthropic_claude3 = "claude-3-haiku-20240307"

// const anthropic_claude_instant = "claude-instant-1.2"
const anthropic_max_tokens = 4096

const openai_embeddings_base_url = "https://api.openai.com/v1/embeddings"
const openai_embeddings_dimensions = 512
const embeddings_chunk_size_default = 1024
const embeddings_chunk_overlap_default = 200
const OPENAI_EMBEDDINGS_INPUT_MAX = 8191

type ModelOpenAIEmbeddings = string

const (
	OPENAI_EMBEDDINGS_MODEL ModelOpenAIEmbeddings = "text-embedding-3-small"
)
