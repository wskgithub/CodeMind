package llm

import "net/http"

// AuthScheme 决定向上游发送哪种鉴权头。
// 鉴权头型与协议格式解耦：部分上游（如 vLLM 的 /v1/messages）说 Anthropic
// 协议却只认 Authorization: Bearer，需要显式指定而非按格式推断。
type AuthScheme string

const (
	// AuthSchemeAuto 按协议格式推断：anthropic 发 x-api-key，其余发 Bearer。
	AuthSchemeAuto AuthScheme = "auto"
	// AuthSchemeBearer 强制发送 Authorization: Bearer。
	AuthSchemeBearer AuthScheme = "bearer"
	// AuthSchemeAPIKey 强制发送 x-api-key。
	AuthSchemeAPIKey AuthScheme = "x-api-key"
)

// NormalizeAuthScheme 规范化鉴权头类型，空值与未知值回退 auto。
func NormalizeAuthScheme(s string) AuthScheme {
	switch AuthScheme(s) {
	case AuthSchemeBearer, AuthSchemeAPIKey:
		return AuthScheme(s)
	default:
		return AuthSchemeAuto
	}
}

// ApplyAuthHeader 按 scheme 设置鉴权头；auto 时按 format 推断
// （anthropic → x-api-key，其余 → Authorization: Bearer）。
// 与客户端原有行为保持一致：apiKey 为空时不设置任何鉴权头。
func ApplyAuthHeader(h http.Header, scheme AuthScheme, format string, apiKey string) {
	if apiKey == "" {
		return
	}
	if scheme == AuthSchemeAuto {
		if format == "anthropic" {
			scheme = AuthSchemeAPIKey
		} else {
			scheme = AuthSchemeBearer
		}
	}
	switch scheme {
	case AuthSchemeAPIKey:
		h.Set("x-api-key", apiKey)
	default:
		h.Set("Authorization", "Bearer "+apiKey)
	}
}
