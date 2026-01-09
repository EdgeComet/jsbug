package parser

import (
	"strings"
	"testing"
)

func TestCountBodyTextTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty string", "", 0},
		{"simple two words", "Hello world", 2},
		{"sentence with punctuation", "The quick brown fox jumps over the lazy dog.", 10},
		{"whitespace only", "   \n\t   ", 2},
		{"longer paragraph", "This is a longer paragraph that contains multiple sentences. It should produce more tokens than a simple phrase. The tokenizer splits text into subword units.", 29},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountBodyTextTokens(tt.input)
			if result != tt.expected {
				t.Errorf("CountBodyTextTokens(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCountBodyTextTokens_Deterministic(t *testing.T) {
	text := "This is a test of deterministic tokenization."
	expected := 9

	first := CountBodyTextTokens(text)
	second := CountBodyTextTokens(text)

	if first != expected {
		t.Errorf("CountBodyTextTokens() = %d, want %d", first, expected)
	}
	if first != second {
		t.Errorf("Non-deterministic results: first=%d, second=%d", first, second)
	}
}

func TestCountBodyTextTokens_Unicode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"chinese characters", "你好世界", 2},
		{"emoji", "Hello 👋 World 🌍", 6},
		{"mixed scripts cyrillic and chinese", "Hello мир 世界", 3},
		{"arabic text", "مرحبا بالعالم", 4},
		{"japanese hiragana", "こんにちは", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountBodyTextTokens(tt.input)
			if result != tt.expected {
				t.Errorf("CountBodyTextTokens(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCountBodyTextTokens_HTMLEntities(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"ampersand entity", "&amp;", 2},
		{"nbsp entity", "&nbsp;", 2},
		{"numeric entity", "&#60;", 3},
		{"mixed with text", "Hello &amp; World", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountBodyTextTokens(tt.input)
			if result != tt.expected {
				t.Errorf("CountBodyTextTokens(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCountBodyTextTokens_LargeText(t *testing.T) {
	// Test with ~500KB of text (100,000 repetitions of "word ")
	largeText := strings.Repeat("word ", 100000)
	expected := 100001 // "word " repeated produces n+1 tokens due to spacing

	result := CountBodyTextTokens(largeText)
	if result != expected {
		t.Errorf("CountBodyTextTokens(large text) = %d, want %d", result, expected)
	}
}
