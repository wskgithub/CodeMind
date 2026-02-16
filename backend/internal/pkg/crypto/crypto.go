package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost   = 12     // bcrypt 加密强度
	apiKeyLength = 32     // API Key 随机部分长度（字节数）
	apiKeyPrefix = "cm-"  // API Key 前缀
)

// HashPassword 使用 bcrypt 加密密码
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}
	return string(hash), nil
}

// CheckPassword 验证密码是否与哈希匹配
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateAPIKey 生成新的 API Key
// 返回: (完整 Key, Key 前缀, Key 的 SHA-256 哈希)
func GenerateAPIKey() (fullKey, prefix, hash string, err error) {
	// 生成随机字节
	randomBytes := make([]byte, apiKeyLength)
	if _, err = rand.Read(randomBytes); err != nil {
		return "", "", "", fmt.Errorf("生成随机数失败: %w", err)
	}

	// 编码为十六进制字符串
	hexStr := hex.EncodeToString(randomBytes)
	fullKey = apiKeyPrefix + hexStr

	// 截取前缀用于展示（cm-xxxxxxxx）
	prefix = fullKey[:len(apiKeyPrefix)+8]

	// 计算 SHA-256 哈希用于存储
	hash = HashAPIKey(fullKey)

	return fullKey, prefix, hash, nil
}

// HashAPIKey 计算 API Key 的 SHA-256 哈希值
func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// GenerateRandomString 生成指定长度的随机十六进制字符串
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
