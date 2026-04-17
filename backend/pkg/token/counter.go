// Package token 提供 LLM Token 计数功能。
package token

// EstimateTokens estimates token count using a simple heuristic.
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}

	var count int
	for _, r := range text {
		if r > 0x4E00 && r < 0x9FFF {
			count += 2
		} else {
			count++
		}
	}

	tokens := count / 4 //nolint:mnd // intentional constant.
	if tokens == 0 {
		tokens = 1
	}
	return tokens
}

// EstimateMessagesTokens estimates token count for a list of messages.
func EstimateMessagesTokens(messages []map[string]string) int {
	var total int
	for _, msg := range messages {
		total += 4
		for _, v := range msg {
			total += EstimateTokens(v)
		}
	}
	total += 2
	return total
}
