package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

// Encryptor AES-256-GCM 加解密器
// 用于加密存储用户的第三方服务 API Key 等敏感信息
type Encryptor struct {
	key []byte
}

// NewEncryptor 从主密钥派生 AES-256 密钥
// 使用 SHA-256 哈希 + 固定盐值派生，确保密钥长度为 32 字节
func NewEncryptor(masterSecret string) *Encryptor {
	hash := sha256.Sum256([]byte(masterSecret + ":codemind-aes-key"))
	return &Encryptor{key: hash[:]}
}

// Encrypt 使用 AES-256-GCM 加密明文
// 返回 base64 编码的 nonce + 密文（nonce 前置便于解密时分离）
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// nonce + ciphertext + tag
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt 解密 AES-256-GCM 密文
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", errors.New("密文 base64 解码失败")
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文长度不足")
	}

	nonce, sealed := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", errors.New("解密失败：密钥不匹配或数据已损坏")
	}

	return string(plaintext), nil
}
