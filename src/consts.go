package main

var TEST_USER_ID = "go-test"

var GPT_BASE_URL = "https://api.openai.com/v1/chat/completions"
var GPT3_MODEL = "gpt-3.5-turbo-0125"
var GPT4_MODEL = "gpt-4-0125-preview"
var GPT_MAX_TOKENS = 8096

var GEMINI_BASE_URL = "https://generativelanguage.googleapis.com/v1beta/models"
var GEMINI_SYSTEM_MESSAGE = "This is the system message of the conversation, and should be used as a general reference for the entire conversation"
var GEMINI_MODEL = "gemini-pro"

var ANTHROPIC_BASE_URL = "https://api.anthropic.com/v1/messages"
var ANTHROPIC_VERSION = "2023-06-01"
var ANTHROPIC_CLAUDE2 = "claude-2.1"
var ANTHROPIC_CLAUDE_INSTANT = "claude-instant-1.2"
var ANTHROPIC_MAX_TOKENS = 4096
