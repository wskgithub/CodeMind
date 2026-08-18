package llm

import (
	"net/http"
	"testing"
)

// TestApplyAuthHeader 验证鉴权头按 scheme 发送，auto 按协议格式推断。
// 关键场景：vLLM 的 /v1/messages 说 Anthropic 协议但只认 Authorization，
// 需通过显式 bearer 覆盖按格式推断的 x-api-key。
func TestApplyAuthHeader(t *testing.T) {
	tests := []struct {
		name       string
		scheme     AuthScheme
		format     string
		wantAuth   string
		wantAPIKey string
	}{
		{
			name:       "auto + anthropic 格式推断为 x-api-key",
			scheme:     AuthSchemeAuto,
			format:     "anthropic",
			wantAPIKey: "sk-test",
		},
		{
			name:     "auto + openai 格式推断为 Bearer",
			scheme:   AuthSchemeAuto,
			format:   "openai",
			wantAuth: "Bearer sk-test",
		},
		{
			name:     "强制 bearer 覆盖 anthropic 格式推断",
			scheme:   AuthSchemeBearer,
			format:   "anthropic",
			wantAuth: "Bearer sk-test",
		},
		{
			name:       "强制 x-api-key 覆盖 openai 格式推断",
			scheme:     AuthSchemeAPIKey,
			format:     "openai",
			wantAPIKey: "sk-test",
		},
		{
			name:     "未知 scheme 回退 auto 推断",
			scheme:   AuthScheme("bogus"),
			format:   "openai",
			wantAuth: "Bearer sk-test",
		},
	}

	for _, tt := range tests {
		h := http.Header{}
		ApplyAuthHeader(h, tt.scheme, tt.format, "sk-test")
		if got := h.Get("Authorization"); got != tt.wantAuth {
			t.Errorf("%s: Authorization = %q, want %q", tt.name, got, tt.wantAuth)
		}
		if got := h.Get("x-api-key"); got != tt.wantAPIKey {
			t.Errorf("%s: x-api-key = %q, want %q", tt.name, got, tt.wantAPIKey)
		}
	}

	// 空 key 不设置任何鉴权头，与客户端原有行为一致
	h := http.Header{}
	ApplyAuthHeader(h, AuthSchemeBearer, "openai", "")
	if h.Get("Authorization") != "" || h.Get("x-api-key") != "" {
		t.Error("empty api key should not set any auth header")
	}
}

// TestNormalizeAuthScheme 验证鉴权头类型规范化：空值与未知值回退 auto。
func TestNormalizeAuthScheme(t *testing.T) {
	cases := map[string]AuthScheme{
		"":          AuthSchemeAuto,
		"auto":      AuthSchemeAuto,
		"bearer":    AuthSchemeBearer,
		"x-api-key": AuthSchemeAPIKey,
		"Bearer":    AuthSchemeAuto, // 大小写敏感，未知值回退
		"bogus":     AuthSchemeAuto,
	}
	for in, want := range cases {
		if got := NormalizeAuthScheme(in); got != want {
			t.Errorf("NormalizeAuthScheme(%q) = %q, want %q", in, got, want)
		}
	}
}
