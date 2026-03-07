-- ============================================================
-- CodeMind Seed Data
-- ============================================================
-- Initial data for first deployment.
-- Creates default admin account and system configurations.
-- ============================================================

-- Default super admin account
-- Username: admin
-- Password: Admin@123456 (bcrypt hash, cost=12)
INSERT INTO users (username, password_hash, display_name, email, role, status)
VALUES (
    'admin',
    '$2a$12$AIv9JVUf369wHF5lIszVROJV/05hM6w4KQyha7fnFVEvB1NVcPb/W',
    'System Administrator',
    'admin@company.com',
    'super_admin',
    1
) ON CONFLICT (username) DO NOTHING;

-- Default system configurations
INSERT INTO system_configs (config_key, config_value, description) VALUES
    ('llm.base_url', '"http://llm-server:8080"', 'LLM service base URL'),
    ('llm.api_key', '""', 'LLM service API key'),
    ('llm.models', '["deepseek-coder-v2"]', 'Available LLM models'),
    ('llm.default_model', '"deepseek-coder-v2"', 'Default LLM model'),
    ('system.max_keys_per_user', '10', 'Maximum API keys per user'),
    ('system.default_concurrency', '5', 'Default max concurrent requests per user'),
    ('system.force_change_password', 'true', 'Force password change on first login'),
    ('system.training_data_collection', 'true', '是否开启 LLM 请求/响应数据采集（用于模型训练）')
ON CONFLICT (config_key) DO NOTHING;

-- Default global rate limits
INSERT INTO rate_limits (target_type, target_id, period, max_tokens, max_requests, max_concurrency, alert_threshold)
VALUES
    ('global', 0, 'daily', 1000000, 0, 5, 80),
    ('global', 0, 'monthly', 20000000, 0, 5, 80)
ON CONFLICT (target_type, target_id, period) DO NOTHING;

-- ============================================================
-- 默认接入文档
-- ============================================================
INSERT INTO documents (slug, title, subtitle, icon, content, sort_order, is_published) VALUES

('openai-sdk', 'OpenAI SDK 接入', '使用 OpenAI 兼容格式调用平台', '🤖',
E'## 接入说明\n\n本平台完全兼容 OpenAI API 格式，支持所有主流语言的 OpenAI SDK 直接接入，无需修改业务代码。\n\n## 接入端点\n\n| 接口 | 路径 |\n|------|------|\n| 对话补全 | `POST /v1/chat/completions` |\n| 文本补全 | `POST /v1/completions` |\n| 向量嵌入 | `POST /v1/embeddings` |\n| 模型列表 | `GET /v1/models` |\n| Responses API | `POST /v1/responses` |\n\n## Python 接入示例\n\n```python\nfrom openai import OpenAI\n\nclient = OpenAI(\n    base_url="http://YOUR_PLATFORM_HOST/v1",  # 注意：base_url 需包含 /v1\n    api_key="cm-your-api-key"                 # 您的 API Key\n)\n\nresponse = client.chat.completions.create(\n    model="your-model-name",\n    messages=[{"role": "user", "content": "你好"}]\n)\nprint(response.choices[0].message.content)\n```\n\n## Node.js 接入示例\n\n```javascript\nimport OpenAI from "openai";\n\nconst client = new OpenAI({\n    baseURL: "http://YOUR_PLATFORM_HOST/v1",  // 注意：baseURL 需包含 /v1\n    apiKey: "cm-your-api-key"\n});\n\nconst response = await client.chat.completions.create({\n    model: "your-model-name",\n    messages: [{ role: "user", content: "你好" }]\n});\nconsole.log(response.choices[0].message.content);\n```\n\n## 注意事项\n\n- OpenAI SDK 的 `base_url` / `baseURL` **必须包含 `/v1` 路径后缀**\n- API Key 格式为 `cm-` 开头的 64 位十六进制字符串\n- 支持流式输出（`stream: true`）',
1, true),

('anthropic-sdk', 'Anthropic SDK 接入', '使用 Anthropic 原生格式调用平台', '🧠',
E'## 接入说明\n\n本平台支持 Anthropic 原生 API 格式，可直接使用 Anthropic SDK 接入。平台会自动处理底层模型的协议转换。\n\n## 接入端点\n\n| 接口 | 路径 |\n|------|------|\n| Messages API | `POST /v1/messages` |\n\n## Python 接入示例\n\n```python\nfrom anthropic import Anthropic\n\nclient = Anthropic(\n    base_url="http://YOUR_PLATFORM_HOST",  # 注意：base_url 不能包含 /v1\n    api_key="cm-your-api-key"              # 您的 API Key\n)\n\nresponse = client.messages.create(\n    model="your-model-name",\n    max_tokens=1024,\n    messages=[{"role": "user", "content": "你好"}]\n)\nprint(response.content)\n```\n\n## 使用 Tools（Function Calling）\n\n```python\nfrom anthropic import Anthropic\n\nclient = Anthropic(\n    base_url="http://YOUR_PLATFORM_HOST",  # 注意：不含 /v1\n    api_key="cm-your-api-key"\n)\n\nresponse = client.messages.create(\n    model="your-model-name",\n    max_tokens=1024,\n    messages=[{"role": "user", "content": "北京今天天气怎么样？"}],\n    tools=[{\n        "name": "get_weather",\n        "description": "获取指定城市的天气信息",\n        "input_schema": {\n            "type": "object",\n            "properties": {\n                "city": {"type": "string", "description": "城市名称"}\n            },\n            "required": ["city"]\n        }\n    }]\n)\nprint(response.content)\n```\n\n## ⚠️ 重要：base_url 配置区别\n\n| SDK | base_url 写法 | 实际请求路径 |\n|-----|--------------|-------------|\n| **Anthropic SDK** | `http://host:port`（**不含 /v1**） | `http://host:port/v1/messages` ✅ |\n| **Anthropic SDK** | `http://host:port/v1`（含 /v1） | `http://host:port/v1/v1/messages` ❌ 404 错误 |\n| **OpenAI SDK** | `http://host:port/v1`（**含 /v1**） | `http://host:port/v1/chat/completions` ✅ |\n\nAnthropic SDK 会在 `base_url` 后**自动追加** `/v1` 前缀，而 OpenAI SDK 不会。\n\n## Node.js 接入示例\n\n```javascript\nimport Anthropic from "@anthropic-ai/sdk";\n\nconst client = new Anthropic({\n    baseURL: "http://YOUR_PLATFORM_HOST",  // 不含 /v1\n    apiKey: "cm-your-api-key"\n});\n\nconst response = await client.messages.create({\n    model: "your-model-name",\n    max_tokens: 1024,\n    messages: [{ role: "user", content: "你好" }]\n});\nconsole.log(response.content);\n```',
2, true),

('cursor-ide', 'Cursor IDE 接入', '在 Cursor 编辑器中使用平台模型', '⌨️',
E'## 接入说明\n\nCursor IDE 支持自定义 OpenAI 兼容的 API 端点，可将平台模型集成到编辑器中。\n\n## 配置步骤\n\n1. 打开 Cursor 设置（`Cmd+,` 或 `Ctrl+,`）\n2. 进入 **Models** 标签页\n3. 在 **OpenAI API Key** 填写您的平台 API Key：`cm-your-api-key`\n4. 在 **Override OpenAI Base URL** 填写：`http://YOUR_PLATFORM_HOST/v1`\n5. 点击 **Verify** 验证连接\n6. 在模型列表中选择您的模型\n\n## 注意事项\n\n- Cursor 使用 OpenAI 格式，`base_url` **必须包含 `/v1`**\n- 确保平台服务对 Cursor 所在网络可访问\n- 如果使用代理，请确认代理支持 SSE（Server-Sent Events）流式传输',
3, true),

('vscode-extension', 'VS Code 扩展接入', '在 VS Code 中使用平台（Continue / Copilot Chat）', '💻',
E'## Continue 扩展接入\n\n[Continue](https://continue.dev) 是主流的 VS Code AI 编程助手扩展，支持自定义 LLM 端点。\n\n### 配置方法\n\n在 `~/.continue/config.json` 中添加：\n\n```json\n{\n  "models": [\n    {\n      "title": "CodeMind 平台",\n      "provider": "openai",\n      "model": "your-model-name",\n      "apiKey": "cm-your-api-key",\n      "apiBase": "http://YOUR_PLATFORM_HOST/v1"\n    }\n  ]\n}\n```\n\n### 注意事项\n\n- `apiBase` **必须包含 `/v1`** 路径\n- `provider` 填写 `openai` 以使用 OpenAI 兼容格式\n\n## Cline / RooCode 扩展接入\n\n在扩展设置中：\n\n- **API Provider**：选择 `OpenAI Compatible`\n- **Base URL**：`http://YOUR_PLATFORM_HOST/v1`（含 `/v1`）\n- **API Key**：`cm-your-api-key`\n- **Model**：填写模型名称',
4, true)

ON CONFLICT (slug) DO NOTHING;
