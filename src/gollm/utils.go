package gollm

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"unicode/utf8"
)

func defaultLogger(level slog.Leveler) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

// Chunks string `s` into a list if strings into equal lengths with a max size of `n`
// If the string is not divisible by n, then the items will be as close in length as possible
func ChunkStringEqualUntilN(s string) ([]string, error) {
	if len(s) == 0 {
		return nil, fmt.Errorf("the input was empty")
	}

	totalRuneCount := utf8.RuneCountInString(s) // Count of runes instead of bytes
	numParts := int(math.Ceil(float64(totalRuneCount) / float64(embeddings_chunk_size_default)))
	evenLength := totalRuneCount / numParts
	extraChars := totalRuneCount % numParts

	var parts []string
	start := 0
	for i := 0; i < numParts; i++ {
		partLength := evenLength
		if i < extraChars {
			partLength++ // Distribute extra characters among the first few parts
		}

		end := start
		count := 0
		for count < partLength && end < len(s) {
			_, size := utf8.DecodeRuneInString(s[end:])
			end += size
			count++
		}

		parts = append(parts, s[start:end])
		start = end - embeddings_chunk_overlap_default
	}

	return parts, nil
}

// Converts a slice of one numeric type to another numeric type using generics.
func convertSlice[T any, U any](list []T, convert func(T) U) []U {
	result := make([]U, len(list))
	for i, v := range list {
		result[i] = convert(v)
	}
	return result
}

func debugPrint(input any) {
	enc, _ := json.MarshalIndent(input, "", "    ")
	fmt.Println(string(enc))
}
