package token

// EstimateTokens 估算文本的 Token 数
// 使用简单的启发式方法：英文约 4 个字符 = 1 token，中文约 1.5 个字符 = 1 token
// 这只是粗略估算，精确计量应以 LLM 响应中的 usage 字段为准
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}

	var count int
	for _, r := range text {
		if r > 0x4E00 && r < 0x9FFF {
			// CJK 统一表意文字（中日韩字符）
			count += 2
		} else {
			count++
		}
	}

	// 大约 4 个字符 = 1 个 token
	tokens := count / 4
	if tokens == 0 {
		tokens = 1
	}
	return tokens
}

// EstimateMessagesTokens 估算消息列表的 Token 数
func EstimateMessagesTokens(messages []map[string]string) int {
	var total int
	for _, msg := range messages {
		// 每条消息有固定开销（约 4 token）
		total += 4
		for _, v := range msg {
			total += EstimateTokens(v)
		}
	}
	// 回复的固定 token 开销
	total += 2
	return total
}
