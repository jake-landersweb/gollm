package gollm

import (
	"context"

	"github.com/jake-landersweb/gollm/v2/src/tokens"
)

type ChunkingFunction func(s string, n int) []string

type Embeddings interface {
	// Create the embdeddings using the provider
	Embed(ctx context.Context, input string) (*EmbedResponse, error)

	// optionally store token records state inside the object as well
	GetUsageRecords() []*tokens.UsageRecord
}
