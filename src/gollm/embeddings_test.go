package gollm

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAIEmbeddings(t *testing.T) {
	ctx := context.TODO()
	logger := defaultLogger(slog.LevelInfo)
	input := "Hello world, this is a string that I am going to convert into an embedding!"

	// send the embeddings request
	embeddings := NewOpenAIEmbeddings(test_user_id, logger, nil)
	response, err := embeddings.Embed(ctx, input)
	require.Nil(t, err)

	require.Equal(t, 1, len(response.Embeddings))
	require.NotNil(t, response.Usage)
}
