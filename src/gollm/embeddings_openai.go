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
	"github.com/jake-landersweb/gollm/v2/src/tokens"
	"github.com/pgvector/pgvector-go"
)

// Struct to handle the creation lifecycle when using OpenAI Embeddings
type OpenAIEmbeddings struct {
	userId string
	logger *slog.Logger
	opts   *OpenAIEmbeddingsOpts

	usageRecords []*tokens.UsageRecord
}

// Optional configurations to customize the usage of the model.
// This struct can be passed in as nil, and reasonable and functional defaults will be used.
type OpenAIEmbeddingsOpts struct {
	Model                ModelOpenAIEmbeddings
	ChunkLength          int
	EmbeddingsDimentions int
	BaseUrl              string

	// Function to use when splitting the text into sections that will fit into the context window.
	// If null, the text will be split into equal-sized sections. `s` is the raw string, `n` is the max
	// length the chunks can be. Must return a list of strings that where each len(string) <= n
	ChunkingFunction func(s string, n int) []string

	// Optionally pass in an api key. If not specified, the environment variable `OPENAI_API_KEY` will be read.
	OpenAIApiKey string
}

func NewOpenAIEmbeddings(userId string, logger *slog.Logger, opts *OpenAIEmbeddingsOpts) *OpenAIEmbeddings {
	if logger == nil {
		logger = defaultLogger(slog.LevelInfo)
	}
	if opts == nil {
		opts = &OpenAIEmbeddingsOpts{}
	}
	if opts.Model == "" {
		opts.Model = OPENAI_EMBEDDINGS_MODEL
	}
	if opts.ChunkLength == 0 {
		opts.ChunkLength = openai_embeddings_chunk_size_default
	}
	if opts.EmbeddingsDimentions == 0 {
		opts.EmbeddingsDimentions = openai_embeddings_dimensions
	}
	if opts.BaseUrl == "" {
		opts.BaseUrl = openai_embeddings_base_url
	}
	if opts.ChunkingFunction == nil {
		opts.ChunkingFunction = ChunkStringEqualUntilN
	}

	return &OpenAIEmbeddings{
		userId: userId,
		logger: logger.With("userId", userId, "model", opts.Model),
		opts:   opts,
	}
}

type EmbedResponse struct {
	Embeddings []*ltypes.EmbeddingsData
	Usage      *tokens.UsageRecord
}

func (e *OpenAIEmbeddings) Embed(ctx context.Context, input string) (*EmbedResponse, error) {
	// chunk the input
	if input == "" {
		return nil, fmt.Errorf("the input cannot be empty")
	}
	chunks := e.opts.ChunkingFunction(input, e.opts.ChunkLength)
	response, err := e.openAIEmbed(ctx, chunks)
	if err != nil {
		return nil, err
	}

	// track token usage
	usageRecord := tokens.NewUsageRecordFromGPTUsage(e.opts.Model, &response.Usage)
	e.usageRecords = append(e.usageRecords, usageRecord)

	// convert openai response into pgvector data types
	list := make([]*ltypes.EmbeddingsData, 0)
	for idx := range chunks {
		list = append(list, &ltypes.EmbeddingsData{
			Raw:       chunks[idx],
			Embedding: pgvector.NewVector(convertSlice(response.Data[idx].Embedding, func(i float64) float32 { return float32(i) })),
		})
	}

	return &EmbedResponse{
		Embeddings: list,
		Usage:      usageRecord,
	}, nil
}

func (e *OpenAIEmbeddings) GetUsageRecords() []*tokens.UsageRecord {
	return e.usageRecords
}

func (e *OpenAIEmbeddings) openAIEmbed(ctx context.Context, input []string) (*ltypes.OpenAIEmbeddingResponse, error) {
	logger := e.logger.With()

	apiKey := e.opts.OpenAIApiKey
	if apiKey == "" {
		logger.DebugContext(ctx, "Reading api key from the environment")
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" || apiKey == "null" {
			return nil, fmt.Errorf("the env variable `OPENAI_API_KEY` is required to be set")
		}
	}

	// create the body
	comprequest := ltypes.OpenAIEmbeddingRequest{
		Input:      input,
		Model:      e.opts.Model,
		Dimensions: e.opts.EmbeddingsDimentions,
		User:       e.userId,
	}

	enc, err := json.Marshal(&comprequest)
	if err != nil {
		return nil, fmt.Errorf("there was an issue encoding the body into json: %v", err)
	}

	// create the request
	req, err := http.NewRequest("POST", e.opts.BaseUrl, bytes.NewBuffer(enc))
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
		logger.InfoContext(ctx, "Sending embeddings request...", "chunks", len(input))
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

		// parse into the completion response object
		var response ltypes.OpenAIEmbeddingResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, fmt.Errorf("there was an issue unmarshalling the request body: %v", err)
		}

		// act based on the error
		if response.Error == nil {
			return &response, nil

		} else {
			// act based on the error
			switch response.Error.Type {
			case ltypes.GPT_ERROR_INVALID:
				return nil, fmt.Errorf("there was a validation error: %s", string(body))
			case ltypes.GPT_ERROR_RATE_LIMIT:
				// rate limit, so wait some extra time and continue
				logger.WarnContext(ctx, "Rate limit error hit. Waiting for an additional 2 seconds...")
				time.Sleep(time.Second * 2)
			// case ltypes.GPT_ERROR_TOKENS_LIMIT:
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
