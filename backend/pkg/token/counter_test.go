package token

import "testing"

// TestEstimateTokens tests token estimation.
func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minToken int
		maxToken int
	}{
		{"empty string", "", 0, 0},
		{"short English", "hello", 1, 2},
		{"English sentence", "The quick brown fox jumps over the lazy dog", 8, 15},
		{"pure Chinese", "你好世界", 1, 4},
		{"mixed Chinese-English", "Hello 世界", 1, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateTokens(tt.text)
			if result < tt.minToken || result > tt.maxToken {
				t.Errorf("EstimateTokens(%q) = %d, expected range [%d, %d]",
					tt.text, result, tt.minToken, tt.maxToken)
			}
		})
	}
}

// TestEstimateMessagesTokens tests message list token estimation.
func TestEstimateMessagesTokens(t *testing.T) {
	messages := []map[string]string{
		{"role": "user", "content": "Hello, how are you?"},
		{"role": "assistant", "content": "I'm doing well, thank you!"},
	}

	result := EstimateMessagesTokens(messages)
	if result <= 0 {
		t.Errorf("message list token count should be positive, got: %d", result)
	}

	// Two messages should have more tokens than one
	single := []map[string]string{
		{"role": "user", "content": "Hello"},
	}
	singleResult := EstimateMessagesTokens(single)
	if result <= singleResult {
		t.Errorf("more messages should produce more tokens: %d <= %d", result, singleResult)
	}
}

// TestEstimateTokensEmpty tests empty input.
func TestEstimateTokensEmpty(t *testing.T) {
	if EstimateTokens("") != 0 {
		t.Error("empty string should return 0")
	}
}
