package parser

import (
	"log"
	"sync"

	"github.com/tiktoken-go/tokenizer"
)

var (
	enc     tokenizer.Codec
	encOnce sync.Once
	encErr  error
)

func getEncoder() (tokenizer.Codec, error) {
	encOnce.Do(func() {
		enc, encErr = tokenizer.ForModel(tokenizer.GPT5)
		if encErr != nil {
			log.Printf("failed to initialize tokenizer: %v", encErr)
		}
	})
	return enc, encErr
}

// CountBodyTextTokens counts tokens in the given text using GPT-4o's o200k_base encoding.
// Returns 0 for empty text or if tokenizer initialization fails.
func CountBodyTextTokens(text string) int {
	if text == "" {
		return 0
	}

	encoder, err := getEncoder()
	if err != nil || encoder == nil {
		return 0
	}

	tokens, _, err := encoder.Encode(text)
	if err != nil {
		log.Printf("failed to encode text: %v", err)
		return 0
	}

	return len(tokens)
}
