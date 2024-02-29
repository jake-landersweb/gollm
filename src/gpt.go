package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jake-landersweb/gollm/src/ltypes"
)

func gptCompletion(ctx context.Context, logger *slog.Logger, userId string, model string, temperature float64, jsonMode bool, jsonSchema string, messages []*ltypes.GPTCompletionMessage) (*ltypes.GPTCompletionResponse, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" || apiKey == "null" {
		return nil, fmt.Errorf("the env variable `OPENAI_API_KEY` is required to be set")
	}

	// create the body
	comprequest := ltypes.GPTCompletionRequest{
		Messages:    messages,
		Model:       model,
		Temperature: temperature,
		User:        userId,
		N:           1,
		Stream:      false,
	}

	if jsonMode {
		if jsonSchema == "" {
			return nil, fmt.Errorf("please provide a valid json schema for the model to follow")
		}
		logger.DebugContext(ctx, "Running with json mode ENABLED")

		// add the required json validation text and schema for the model to follow
		comprequest.ResponseFormat = ltypes.GPTRespFormat{Type: "json_object"}
		comprequest.Messages[len(comprequest.Messages)-1].Content = fmt.Sprintf("%s\n\nPlease respond to this message ONLY with the given json schema.\n\nJSON SCHEMA:\n%s", comprequest.Messages[len(comprequest.Messages)-1].Content, jsonSchema)
	} else {
		logger.DebugContext(ctx, "Running with json mode DISABLED")
		comprequest.ResponseFormat = ltypes.GPTRespFormat{Type: "text"}
	}

	enc, err := json.Marshal(&comprequest)
	if err != nil {
		return nil, fmt.Errorf("there was an issue encoding the body into json: %v", err)
	}

	// create the request
	req, err := http.NewRequest("POST", GPT_BASE_URL, bytes.NewBuffer(enc))
	if err != nil {
		return nil, fmt.Errorf("there was an issue creating the http request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// send the request
	client := &http.Client{}

	retries := 3
	backoff := 1 * time.Second

	for attempt := 0; attempt < retries; attempt++ {
		logger.InfoContext(ctx, "Sending Request...")
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("there was an unknown issue with the request: %v", err)
		}

		// read the body
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("there was an issue reading the body: %v", err)
		}

		logger.InfoContext(ctx, "Completed request", "statusCode", resp.StatusCode)
		logger.DebugContext(ctx, "Response body", "body", string(body))

		// parse into the completion response object
		var completion ltypes.GPTCompletionResponse
		err = json.Unmarshal(body, &completion)
		if err != nil {
			return nil, fmt.Errorf("there was an issue unmarshalling the request body: %v", err)
		}

		// act based on the error
		if completion.Error == nil {
			// success. Relay to the user
			if len(completion.Choices) == 0 {
				return nil, fmt.Errorf("the completion list was 0")
			}
			return &completion, nil

		} else {
			// act based on the error
			switch completion.Error.Type {
			case ltypes.GPT_ERROR_INVALID:
				return nil, fmt.Errorf("there was a validation error: %s", string(body))
			case ltypes.GPT_ERROR_RATE_LIMIT:
				// rate limit, so wait some extra time and continue
				logger.WarnContext(ctx, "Rate limit error hit. Waiting for an additional 2 seconds...")
				time.Sleep(time.Second * 2)
			case ltypes.GPT_ERROR_TOKENS_LIMIT:
				// too many tokens, trim the message and try again
				tmp := messages[len(messages)-1]
				tmp.Content = tmp.Content[:GPT_MAX_TOKENS]
			case ltypes.GPT_ERROR_AUTH:
				return nil, fmt.Errorf("the user is not authenticated: %s", string(body))
			case ltypes.GPT_ERROR_NOT_FOUND:
				return nil, fmt.Errorf("the requested resource was not found: %s", string(body))
			case ltypes.GPT_ERROR_SERVER:
				// internal server error, wait and try again
				logger.WarnContext(ctx, "There was an issue on OpenAI's side. Waiting 2 seconds and trying again ...", "body", string(body))
				time.Sleep(time.Second * 2)
			case ltypes.GPT_ERROR_PERMISSION:
				return nil, fmt.Errorf("the requested resource was not found: %s", string(body))
			default:
				return nil, fmt.Errorf("there was an unknown error: %s", string(body))
			}

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

func gptTokenizerApproximate(method string, input string) (int, error) {
	// Split text into words and count characters
	wordCount := len(strings.Fields(input))
	charCount := len(input)

	// Calculate token counts based on words and characters
	tokensCountWordEst := float64(wordCount) / 0.75
	tokensCountCharEst := float64(charCount) / 4.0

	// Include additional tokens for spaces and punctuation marks
	additionalTokens := float64(len(strings.FieldsFunc(input, func(r rune) bool {
		return strings.ContainsRune(" .,!?;", r)
	})))

	tokensCountWordEst += additionalTokens
	tokensCountCharEst += additionalTokens

	var output float64
	switch method {
	case "avg":
		output = (tokensCountWordEst + tokensCountCharEst) / 2
	case "words":
		output = tokensCountWordEst
	case "chars":
		output = tokensCountCharEst
	case "max":
		output = math.Max(tokensCountWordEst, tokensCountCharEst)
	case "min":
		output = math.Min(tokensCountWordEst, tokensCountCharEst)
	default:
		return 0, fmt.Errorf("invlaid method: %s", method)
	}

	return int(output), nil
}
