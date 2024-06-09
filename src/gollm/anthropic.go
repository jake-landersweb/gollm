package gollm

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/jake-landersweb/gollm/v2/src/ltypes"
)

func (l *LanguageModel) anthropicCompletion(
	ctx context.Context,
	logger *slog.Logger,
	model string,
	temperature float64,
	jsonMode bool,
	jsonSchema string,
	messages []*ltypes.AnthropicMessage,
	tools []*ltypes.AnthropicTool,
) (*ltypes.AnthropicResponse, error) {
	apiKey := l.args.AnthropicApiKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" || apiKey == "null" {
			return nil, fmt.Errorf("the environment variable `GEMINI_API_KEY` is required")
		}
	}

	// parse a system message if exists
	var msgs []*ltypes.AnthropicMessage
	systemMsg := ""
	if messages[0].Role == "system" {
		systemMsg = messages[0].Content[0].Text
		msgs = messages[1:] // trim off the first message
	} else {
		msgs = messages
	}

	// compose the request body
	comprequest := &ltypes.AnthropicRequest{
		Model:       model,
		Messages:    msgs,
		System:      systemMsg,
		MaxTokens:   l.args.AnthropicMaxTokens,
		Temperature: temperature,
		Tools:       tools,
	}

	// do not add any extra options for tool use
	if messages[len(messages)-1].Content[0].Type != "tool_result" {
		// add instructions for json mode
		if jsonMode {
			if jsonSchema == "" {
				return nil, fmt.Errorf("invalid json schema provided")
			}

			logger.DebugContext(ctx, "Running with json mode ENABLED")

			// add json instructions onto the end of the request
			comprequest.Messages[len(comprequest.Messages)-1].Content[0].Text = fmt.Sprintf("%s\n\nPlease place your JSON output inside <json></json> xml tags, following this schema: %s", comprequest.Messages[len(comprequest.Messages)-1].Content[0].Text, jsonSchema)
		} else {
			logger.DebugContext(ctx, "Running with json mode DISABLED")
		}
	}

	debugPrint(comprequest)

	// parse and encode the body
	enc, err := json.Marshal(comprequest)
	if err != nil {
		return nil, fmt.Errorf("there was an issue encoding the body: %v", err)
	}

	// create the request
	req, err := http.NewRequest("POST", l.args.AnthropicBaseUrl, bytes.NewBuffer(enc))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", l.args.AnthropicVersion)

	retries := 3
	backoff := 1 * time.Second

	for attempt := 0; attempt < retries; attempt++ {
		logger.InfoContext(ctx, "Sending Anthropic request...")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("there was an issue sending the request: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("there was an issue parsing the request: %v", err)
		}

		logger.InfoContext(ctx, "Completed request", "statusCode", resp.StatusCode)
		logger.DebugContext(ctx, "Response body", "body", string(body))
		fmt.Println(string(body))

		// parse the request body
		var response ltypes.AnthropicResponse
		if err = json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("there was an issue parsing the response body: %v", err)
		}

		// if no error, return
		if response.Error == nil {
			// if there are function calls, do not parse the response
			if len(response.Content) > 1 || !jsonMode {
				return &response, nil
			}

			logger.DebugContext(ctx, "Parsing json ...")

			// parse the xml response
			response.Content[0].Text = fmt.Sprintf("<root>%s</root>", response.Content[0].Text)
			type rtag struct {
				Content string `xml:",innerxml"`
			}

			type root struct {
				Json rtag `xml:"json"`
			}

			var tmp root
			if err := xml.Unmarshal([]byte(response.Content[0].Text), &tmp); err != nil {
				return nil, fmt.Errorf("failed to parse the xml: %s", err)
			}
			response.Content[0].Text = tmp.Json.Content
			return &response, nil
		}

		logger.ErrorContext(ctx, "there was an api error", "type", response.Error.Type, "message", response.Error.Message)

		// parse the errror
		switch response.Error.Type {
		// case ltypes.ANTHROPIC_API_ERROR:
		case ltypes.ANTHROPIC_PERMISSION_ERROR:
			fallthrough
		case ltypes.ANTHROPIC_AUTHENTICATION_ERROR:
			return nil, fmt.Errorf("there was an issue authenticating: %s", response.Error.Message)
		// case ltypes.ANTHROPIC_INVALID_REQUEST_ERROR:
		// case ltypes.ANTHROPIC_NOT_FOUND_ERROR:
		case ltypes.ANTHROPIC_OVERLOADED_ERROR:
			logger.WarnContext(ctx, "The api is overloaded, waiting 2 seconds then trying again ...")
			time.Sleep(time.Second * 2)
		case ltypes.ANTHROPIC_RATE_LIMIT_ERROR:
			logger.WarnContext(ctx, "Rate limit hit, waiting 2 seconds then trying again ...")
			time.Sleep(time.Second * 2)
		default:
			return nil, fmt.Errorf("there was an unknown issue with the request: [%s]: %s", response.Error.Type, response.Error.Message)
		}

		if attempt < retries-1 {
			sleep := backoff + time.Duration(rand.Intn(1000))*time.Millisecond // Add jitter
			time.Sleep(sleep)
			backoff *= 2 // Double the backoff interval
		} else if resp != nil && resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("there was an issue with the request and could not recover: %s", string(body))
		}
	}

	return nil, err
}

// Calculated from https://docs.anthropic.com/claude/docs/glossary#tokens
func anthropicTokenizerAproximate(input string) int {
	return int(float64(len(input)) / 3.5)
}
