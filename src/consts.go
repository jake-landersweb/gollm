package main

const TEST_USER_ID = "go-test"

const GPT_BASE_URL = "https://api.openai.com/v1/chat/completions"
const GPT3_MODEL = "gpt-3.5-turbo-0125"
const GPT4_MODEL = "gpt-4-0125-preview"
const GPT_MAX_TOKENS = 8096

const GEMINI_BASE_URL = "https://generativelanguage.googleapis.com/v1beta/models"
const GEMINI_SYSTEM_MESSAGE = "This is the system message of the conversation, and should be used as a general reference for the entire conversation"
const GEMINI_MODEL = "gemini-pro"

const ANTHROPIC_BASE_URL = "https://api.anthropic.com/v1/messages"
const ANTHROPIC_VERSION = "2023-06-01"
const ANTHROPIC_CLAUDE2 = "claude-2.1"
const ANTHROPIC_CLAUDE_INSTANT = "claude-instant-1.2"
const ANTHROPIC_MAX_TOKENS = 4096
