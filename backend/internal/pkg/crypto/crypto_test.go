package crypto

import (
	"strings"
	"testing"
)

// TestHashPassword 测试密码哈希
func TestHashPassword(t *testing.T) {
	password := "Test@12345"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("密码哈希失败: %v", err)
	}

	if hash == "" {
		t.Fatal("哈希结果不能为空")
	}

	// 哈希值不应与原文相同
	if hash == password {
		t.Fatal("哈希值不应与原文相同")
	}
}

// TestCheckPassword 测试密码校验
func TestCheckPassword(t *testing.T) {
	password := "Test@12345"
	hash, _ := HashPassword(password)

	// 正确密码
	if !CheckPassword(password, hash) {
		t.Error("正确密码应通过校验")
	}

	// 错误密码
	if CheckPassword("WrongPassword", hash) {
		t.Error("错误密码不应通过校验")
	}
}

// TestGenerateAPIKey 测试 API Key 生成
func TestGenerateAPIKey(t *testing.T) {
	fullKey, prefix, hash, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("生成 API Key 失败: %v", err)
	}

	// 检查完整 Key 的前缀格式
	if !strings.HasPrefix(fullKey, "cm-") {
		t.Errorf("Key 应以 'cm-' 开头，实际: %s", fullKey)
	}

	// 检查 Key 长度：cm- + 64 hex = 67 字符
	if len(fullKey) != 67 {
		t.Errorf("Key 长度应为 67，实际: %d", len(fullKey))
	}

	// 检查前缀格式：cm-前8位
	if !strings.HasPrefix(prefix, "cm-") {
		t.Errorf("前缀格式不正确: %s", prefix)
	}

	// 检查哈希不为空
	if hash == "" {
		t.Fatal("哈希结果不能为空")
	}

	// 用同一个 Key 重新哈希应得到相同结果
	rehash := HashAPIKey(fullKey)
	if rehash != hash {
		t.Error("相同 Key 的哈希值应一致")
	}
}

// TestHashAPIKey 测试 API Key 哈希
func TestHashAPIKey(t *testing.T) {
	key1 := "cm-abcdef1234567890abcdef12345678"
	key2 := "cm-abcdef1234567890abcdef12345679"

	hash1 := HashAPIKey(key1)
	hash2 := HashAPIKey(key2)

	// 不同 Key 应产生不同哈希
	if hash1 == hash2 {
		t.Error("不同 Key 应产生不同哈希")
	}

	// 相同 Key 应产生相同哈希
	if HashAPIKey(key1) != hash1 {
		t.Error("相同 Key 应产生相同哈希")
	}
}

// TestGenerateAPIKeyUniqueness 测试 API Key 唯一性
func TestGenerateAPIKeyUniqueness(t *testing.T) {
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, _, _, err := GenerateAPIKey()
		if err != nil {
			t.Fatalf("生成 API Key 失败: %v", err)
		}
		if keys[key] {
			t.Fatalf("生成了重复的 Key: %s", key)
		}
		keys[key] = true
	}
}
