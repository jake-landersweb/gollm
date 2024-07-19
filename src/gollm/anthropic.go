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
	"strings"
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
	prohibitTool bool,
	toolChoice string,
) (*ltypes.AnthropicResponse, error) {
	apiKey := l.args.AnthropicApiKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" || apiKey == "null" {
			return nil, fmt.Errorf("the environment variable `ANTHROPIC_API_KEY` is required")
		}
	}

	// parse a system message if exists
	var msgs []*ltypes.AnthropicMessage
	systemMsg := ""
	if messages[0].Role == "system" {
		systemMsg = fmt.Sprintf("%s\n\nFormatting Instructions:\nYou MUST place your response to this message inside <response></response> XML tags. Any context or extra information shall be placed outside these tags, with the <response> XML tag containing exactly what was requested.", messages[0].Content[0].Text)
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

	if !prohibitTool && toolChoice != "" {
		comprequest.ToolChoice = &ltypes.AnthropicToolChoice{
			Type: "tool",
			Name: toolChoice,
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
			if len(response.Content) > 1 {
				return &response, nil
			}

			logger.DebugContext(ctx, "Parsing the AI message result")

			// parse the xml response
			tmpContent := response.Content[0].Text
			response.Content[0].Text = fmt.Sprintf("<root>%s</root>", response.Content[0].Text)
			type rtag struct {
				Content string `xml:",innerxml"`
			}

			type root struct {
				Response rtag `xml:"response"`
			}

			var tmp root
			if err := xml.Unmarshal([]byte(response.Content[0].Text), &tmp); err != nil {
				logger.Warn("failed to parse the xml, using the raw content", "error", err)
				response.Content[0].Text = tmpContent
			} else {
				if strings.Trim(tmp.Response.Content, "") != "" {
					response.Content[0].Text = tmp.Response.Content
				} else {
					logger.Warn("there was an issue parsing the xml, using the raw response")
					response.Content[0].Text = tmpContent
				}
			}
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
