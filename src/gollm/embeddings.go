package gollm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jake-landersweb/gollm/v2/src/tokens"
)

type ChunkingFunction func(s string, n int) []string

type EmbedArgs struct {
	Input            string
	InputChunks      []string
	ChunkingFunction func(input string) ([]string, error)
}

func (args *EmbedArgs) IsValid() error {
	if args == nil {
		return fmt.Errorf("cannot be nil")
	}
	if args.Input == "" && len(args.InputChunks) == 0 {
		return fmt.Errorf("either input or inputChunks cannot be empty")
	}
	if args.ChunkingFunction == nil {
		args.ChunkingFunction = ChunkStringEqualUntilN
	}
	return nil
}

type Embeddings interface {
	// Create the embdeddings using the provider
	Embed(
		ctx context.Context,
		logger *slog.Logger,
		args *EmbedArgs,
	) (*EmbedResponse, error)

	// optionally store token records state inside the object as well
	GetUsageRecords() []*tokens.UsageRecord
}
