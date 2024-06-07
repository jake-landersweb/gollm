package gollm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/jake-landersweb/gollm/v2/src/ltypes"
)

func (l *LanguageModel) geminiCompletion(ctx context.Context, logger *slog.Logger, model string, temperature float64, jsonMode bool, jsonSchema string, messages []*ltypes.GemContent) (*ltypes.GemCompletionResponse, error) {
	apiKey := l.args.GeminiApiKey
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("the environment variable `GEMINI_API_KEY` is required")
		}
	}

	comprequest := &ltypes.GemRequestBody{
		Contents: messages,
		GenerationConfig: ltypes.GemGenerationConfig{
			Temperature: temperature,
		},
		// ignore all safety settings
		SafetySettings: []*ltypes.GemSafetySetting{
			{
				Category:  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				Threshold: ltypes.BlockOnlyHigh,
			},
			{
				Category:  "HARM_CATEGORY_HATE_SPEECH",
				Threshold: ltypes.BlockOnlyHigh,
			},
			{
				Category:  "HARM_CATEGORY_HARASSMENT",
				Threshold: ltypes.BlockOnlyHigh,
			},
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: ltypes.BlockOnlyHigh,
			},
		},
	}

	// parse for json mode
	if jsonMode {
		if jsonSchema == "" {
			return nil, fmt.Errorf("please provide a valid json schema")
		}
		logger.DebugContext(ctx, "Running with json mode ENABLED")

		// add json instructions onto the end of the request
		comprequest.Contents[len(comprequest.Contents)-1].Parts[0].Text = fmt.Sprintf("%s\n\nPlease respond to this message ONLY with the given json schema. This schema should be parsed as valid json, and shall NOT contain backticks (`).\n\nJSON SCHEMA:\n%s", comprequest.Contents[len(comprequest.Contents)-1].Parts[0].Text, jsonSchema)
	} else {
		logger.DebugContext(ctx, "Running with json mode DISABLED")
	}

	// parse and encode the body
	enc, err := json.Marshal(comprequest)
	if err != nil {
		return nil, fmt.Errorf("there was an issue encoding the body: %v", err)
	}

	// create the request
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", l.args.GeminiBaseUrl, model, apiKey)
	fmt.Print(url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(enc))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	retries := 3
	backoff := 1 * time.Second

	for attempt := 0; attempt < retries; attempt++ {
		logger.InfoContext(ctx, "Sending Gemini request...")
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

		// parse the request body
		var response ltypes.GemCompletionResponse
		if err = json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("there was an issue parsing the response body: %v", err)
		}

		// if no error, return
		if response.Error == nil {
			return &response, nil
		}

		// parse the errror
		switch response.Error.Status {
		// case ltypes.GEM_ERROR_OK:
		// case ltypes.GEM_ERROR_CANCELLED:
		// case ltypes.GEM_ERROR_UNKNOWN:
		// case ltypes.GEM_ERROR_INVALID_ARGUMENT:
		// case ltypes.GEM_ERROR_DEADLINE_EXCEEDED:
		// case ltypes.GEM_ERROR_NOT_FOUND:
		// case ltypes.GEM_ERROR_ALREADY_EXISTS:
		case ltypes.GEM_ERROR_UNAUTHENTICATED:
		case ltypes.GEM_ERROR_PERMISSION_DENIED:
			return nil, fmt.Errorf("the user is not authenticated: %s", response.Error.Message)
		case ltypes.GEM_ERROR_RESOURCE_EXHAUSTED:
			logger.WarnContext(ctx, "The model is exhasted, waiting 2 seconds before trying again")
			time.Sleep(time.Second * 2)
		case ltypes.GEM_ERROR_FAILED_PRECONDITION:
			return nil, fmt.Errorf("there was a failed pre-condition: %s", response.Error.Message)
		case ltypes.GEM_ERROR_ABORTED:
			logger.WarnContext(ctx, "the response was aborted, waiting 2 seconds before trying again")
			time.Sleep(time.Second * 2)
		// case ltypes.GEM_ERROR_OUT_OF_RANGE:
		// case ltypes.GEM_ERROR_UNIMPLEMENTED:
		case ltypes.GEM_ERROR_INTERNAL:
			logger.WarnContext(ctx, "there was an internal error. waiting 2 seconds before trying again")
			time.Sleep(time.Second * 2)
		// case ltypes.GEM_ERROR_UNAVAILABLE:
		// case ltypes.GEM_ERROR_DATA_LOSS:
		default:
			return nil, fmt.Errorf("there was an unknown issue with the request: [%s]: %s", response.Error.Status, response.Error.Message)
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

func geminiTokenizerAccurate(input string, model string) (int, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" || apiKey == "null" {
		return 0, fmt.Errorf("the environment variable `GEMINI_API_KEY` is required")
	}

	// create the body
	enc, err := json.Marshal(map[string]any{
		"contents": []ltypes.GemContent{
			{
				Role: "user",
				Parts: []ltypes.GemPart{
					{
						Text: input,
					},
				},
			},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("there was an issue encoding the body: %v", err)
	}

	// create the request
	url := fmt.Sprintf("%s/%s:countTokens?key=%s", gemini_base_url, model, apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(enc))
	if err != nil {
		return 0, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("there was an issue sending the request: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("there was an issue parsing the request: %v", err)
	}

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("the response was not 200: %s", string(data))
	}

	// parse the tokens from the body
	var body map[string]int
	if err := json.Unmarshal(data, &body); err != nil {
		return 0, fmt.Errorf("there was an issue parsing the response body: %v", err)
	}

	return body["totalTokens"], nil
}

// func (l *LanguageModel) getGoogleModels() {

// }
