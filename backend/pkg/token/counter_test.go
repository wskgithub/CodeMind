package token

import "testing"

// TestEstimateTokens 测试 Token 估算.
func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minToken int
		maxToken int
	}{
		{"空字符串", "", 0, 0},
		{"短英文", "hello", 1, 2},
		{"一句英文", "The quick brown fox jumps over the lazy dog", 8, 15},
		{"纯中文", "你好世界", 1, 4},
		{"中英混合", "Hello 世界", 1, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateTokens(tt.text)
			if result < tt.minToken || result > tt.maxToken {
				t.Errorf("EstimateTokens(%q) = %d, 预期范围 [%d, %d]",
					tt.text, result, tt.minToken, tt.maxToken)
			}
		})
	}
}

// TestEstimateMessagesTokens 测试消息列表 Token 估算.
func TestEstimateMessagesTokens(t *testing.T) {
	messages := []map[string]string{
		{"role": "user", "content": "Hello, how are you?"},
		{"role": "assistant", "content": "I'm doing well, thank you!"},
	}

	result := EstimateMessagesTokens(messages)
	if result <= 0 {
		t.Errorf("消息列表的 Token 数应为正数，实际: %d", result)
	}

	// 两条消息应该比单条多
	single := []map[string]string{
		{"role": "user", "content": "Hello"},
	}
	singleResult := EstimateMessagesTokens(single)
	if result <= singleResult {
		t.Errorf("更多消息应产生更多 Token: %d <= %d", result, singleResult)
	}
}

// TestEstimateTokensEmpty 测试空输入.
func TestEstimateTokensEmpty(t *testing.T) {
	if EstimateTokens("") != 0 {
		t.Error("空字符串应返回 0")
	}
}
