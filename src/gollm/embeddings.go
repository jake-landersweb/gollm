package gollm

import (
	"context"

	"github.com/jake-landersweb/gollm/v2/src/ltypes"
	"github.com/jake-landersweb/gollm/v2/src/tokens"
)

type ChunkingFunction func(s string, n int) []string

type Embeddings interface {
	// Create the embdeddings using the provider
	Embed(ctx context.Context, input string) ([]*ltypes.EmbeddingsData, error)

	// Get the usage records from the model
	GetTokenRecords() []*tokens.TokenRecord
}
