package ltypes

import "github.com/pgvector/pgvector-go"

type EmbeddingsData struct {
	Raw       string
	Embedding pgvector.Vector
}
