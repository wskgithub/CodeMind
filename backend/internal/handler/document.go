package handler

import (
	"codemind/internal/model"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/response"
	"codemind/internal/repository"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// DocumentHandler 文档处理器
type DocumentHandler struct {
	repo repository.DocumentRepository
}

// NewDocumentHandler 创建文档处理器
func NewDocumentHandler(repo repository.DocumentRepository) *DocumentHandler {
	return &DocumentHandler{repo: repo}
}

// ListDocuments 获取文档列表（公开接口，仅返回已发布文档）
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	docs, err := h.repo.List()
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Success(c, docs)
}

// GetDocument 根据 slug 获取文档详情（公开接口）
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.BadRequest(c, "文档标识不能为空")
		return
	}

	doc, err := h.repo.GetBySlug(slug)
	if err != nil {
		response.ErrorWithMsg(c, errcode.ErrRecordNotFound, "文档不存在或未发布")
		return
	}

	response.Success(c, doc)
}

// ListAllDocuments 获取所有文档（管理接口，包括未发布）
func (h *DocumentHandler) ListAllDocuments(c *gin.Context) {
	docs, err := h.repo.ListAll()
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Success(c, docs)
}

// GetDocumentByID 根据 ID 获取文档（管理接口）
func (h *DocumentHandler) GetDocumentByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文档ID")
		return
	}

	doc, err := h.repo.GetByID(id)
	if err != nil {
		response.ErrorWithMsg(c, errcode.ErrRecordNotFound, "文档不存在")
		return
	}

	response.Success(c, doc)
}

// CreateDocumentRequest 创建文档请求
type CreateDocumentRequest struct {
	Slug        string `json:"slug" binding:"required,max=50"`
	Title       string `json:"title" binding:"required,max=200"`
	Subtitle    string `json:"subtitle" binding:"max=500"`
	Icon        string `json:"icon" binding:"max=100"`
	Content     string `json:"content" binding:"required"`
	SortOrder   int    `json:"sort_order"`
	IsPublished bool   `json:"is_published"`
}

// CreateDocument 创建文档（管理接口）
func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	var req CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 转换 slug 为小写并去除空格
	slug := strings.ToLower(strings.TrimSpace(req.Slug))

	doc := &model.Document{
		Slug:        slug,
		Title:       req.Title,
		Subtitle:    req.Subtitle,
		Icon:        req.Icon,
		Content:     req.Content,
		SortOrder:   req.SortOrder,
		IsPublished: req.IsPublished,
	}

	if err := h.repo.Create(doc); err != nil {
		response.InternalError(c)
		return
	}

	response.Success(c, doc)
}

// UpdateDocumentRequest 更新文档请求
type UpdateDocumentRequest struct {
	Title       string `json:"title" binding:"required,max=200"`
	Subtitle    string `json:"subtitle" binding:"max=500"`
	Icon        string `json:"icon" binding:"max=100"`
	Content     string `json:"content" binding:"required"`
	SortOrder   int    `json:"sort_order"`
	IsPublished bool   `json:"is_published"`
}

// UpdateDocument 更新文档（管理接口）
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文档ID")
		return
	}

	var req UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	doc, err := h.repo.GetByID(id)
	if err != nil {
		response.ErrorWithMsg(c, errcode.ErrRecordNotFound, "文档不存在")
		return
	}

	doc.Title = req.Title
	doc.Subtitle = req.Subtitle
	doc.Icon = req.Icon
	doc.Content = req.Content
	doc.SortOrder = req.SortOrder
	doc.IsPublished = req.IsPublished

	if err := h.repo.Update(doc); err != nil {
		response.InternalError(c)
		return
	}

	response.Success(c, doc)
}

// DeleteDocument 删除文档（管理接口）
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文档ID")
		return
	}

	if err := h.repo.Delete(id); err != nil {
		response.InternalError(c)
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// InitializeDocuments 初始化默认文档（管理接口，仅在表为空时执行）
func (h *DocumentHandler) InitializeDocuments(c *gin.Context) {
	// 检查是否已有文档
	docs, err := h.repo.ListAll()
	if err != nil {
		response.InternalError(c)
		return
	}

	if len(docs) > 0 {
		response.BadRequest(c, "文档已存在，无法初始化")
		return
	}

	// 创建默认文档（只有基本信息，无内容）
	defaultDocs := model.DefaultTools
	for i := range defaultDocs {
		defaultDocs[i].Content = getDefaultContent(defaultDocs[i].Slug)
	}

	if err := h.repo.BatchCreate(defaultDocs); err != nil {
		response.InternalError(c)
		return
	}

	response.Success(c, gin.H{"message": "初始化成功", "count": len(defaultDocs)})
}

// getDefaultContent 获取默认文档内容
func getDefaultContent(slug string) string {
	contents := map[string]string{
		model.DocSlugClaude:       getClaudeContent(),
		model.DocSlugClaudeIDE:    getClaudeIDEContent(),
		model.DocSlugCursor:       getCursorContent(),
		model.DocSlugTrae:         getTraeContent(),
		model.DocSlugCline:        getClineContent(),
		model.DocSlugKiloCode:     getKiloCodeContent(),
		model.DocSlugRooCode:      getRooCodeContent(),
		model.DocSlugOpenCode:     getOpenCodeContent(),
		model.DocSlugFactoryDroid: getFactoryDroidContent(),
		model.DocSlugCrush:        getCrushContent(),
		model.DocSlugGoose:        getGooseContent(),
		model.DocSlugOpenClaw:     getOpenClawContent(),
		model.DocSlugCherryStudio: getCherryStudioContent(),
		model.DocSlugOthers:       getOthersContent(),
	}

	if content, ok := contents[slug]; ok {
		return content
	}
	return ""
}

// 以下是各工具的默认文档内容

func getClaudeContent() string {
	return `## 简介

Claude Code 是一个智能编码工具，可以在终端中运行，通过自然语言命令交互帮助开发者快速完成代码生成、调试、重构等任务。

## 安装步骤

### 前提条件

- 安装 Node.js 18 或更新版本环境
- MacOS 用户推荐使用 nvm 方式安装 Nodejs 或 Homebrew 方式
- Windows 用户还需安装 Git for Windows

### 安装 Claude Code

	npm install -g @anthropic-ai/claude-code

运行如下命令查看安装结果：

	claude --version

## 配置 CodeMind

### 步骤一：获取 API Key

1. 登录 CodeMind 平台
2. 进入「API Keys」页面
3. 点击「创建 API Key」
4. 复制生成的 API Key（以 cm- 开头）

### 步骤二：配置环境变量

运行以下命令配置 CodeMind：

	claude config set anthropic_api_key 你的APIKey
	claude config set anthropic_base_url https://your-domain/api/coding/paas/v4

配置示例：

![Claude Code配置](/images/docs/claude-config.png)

### 步骤三：开始使用

配置完成后，进入代码工作目录，执行 claude 命令即可开始使用：

	cd /path/to/your/project
	claude

> 若遇到「Do you want to use this API key」选择 Yes 即可

启动后选择信任 Claude Code 访问文件夹里的文件。

## 切换模型

1. 手动修改配置文件 ~/.claude/settings.json，添加或替换如下环境变量参数：

` + "```json\n" + `{
  "env": {
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7"
  }
}` + "\n```\n" + `

2. 启动一个新的命令行窗口，运行 claude 启动，在 Claude Code 中输入 /status 确认模型状态：

![模型状态确认](/images/docs/claude-status.png)

## 常见问题

### 手工修改配置不生效

- 关闭所有 Claude Code 窗口，重新打开一个新的命令行窗口
- 尝试删除 ~/.claude/settings.json 文件，然后重新配置
- 确认配置文件的 JSON 格式是否正确

### 推荐的 Claude Code 版本

建议使用最新版本的 Claude Code：

	# 升级到最新版本
	claude update
`
}

func getClaudeIDEContent() string {
	return `## 简介

Claude Code 是一个智能编码工具，可以在终端中运行，也可以通过在 VS Code、JetBrains 等 IDE 安装插件使用。

## 前置步骤

参考 **Claude Code 接入指南** 完成 Claude Code 的安装与 CodeMind 的配置。

## IDE 插件安装

### VS Code 插件

1. 打开 VS Code，进入扩展市场
2. 搜索并安装「Claude Code」插件

![VS Code搜索Claude Code插件](/images/docs/claude-ide-vscode-search.png)

3. 安装完成后，点击右上角的 Claude Code 图标进入 Claude Code 页面

![Claude Code图标入口](/images/docs/claude-ide-vscode-icon.png)

4. 在登录页面等待几秒待其初始化完成后即可使用
5. 成功进入后，在对话框输入 /config 进入设置，勾选 Disable Login Prompt 配置来关闭登录页面

![Disable Login Prompt设置](/images/docs/claude-ide-vscode-settings.png)

### JetBrains 插件

1. 打开 JetBrains IDE（如 IntelliJ IDEA、PyCharm 等）
2. 进入插件市场，搜索「Claude Code」插件并安装

![JetBrains插件安装](/images/docs/claude-ide-jetbrains-install.png)

3. 安装完成后，重启 IDE 即可使用

![JetBrains Claude Code界面](/images/docs/claude-ide-jetbrains-interface.png)

## 使用说明

配置完成后，即可在 IDE 中正常使用 Claude Code 进行开发。
`
}

func getCursorContent() string {
	return `## 简介

将 GLM 模型通过 OpenAI 协议在 Cursor 中自定义配置模型接入使用。

> **注意**：由于 Cursor 的限制，只有订阅了 Cursor 高级会员及以上的用户才支持自定义配置模型。

## 安装步骤

### 1. 安装 Cursor

访问 [Cursor 官网](https://cursor.sh) 下载并安装适合您的操作系统的版本。

### 2. 创建新 Provider/Model

1. 在 Cursor 中，打开 "Models" 部分，点击 "Add Custom Model" 按钮

![Cursor添加自定义模型](/images/docs/cursor-add-model.jpeg)

2. 选择 **OpenAI 协议**
3. 配置 **OpenAI API Key**（从 CodeMind 平台获取）
4. 在 **Override OpenAI Base URL** 中，将默认 URL 替换为：https://your-domain/api/coding/paas/v4
5. 输入您希望使用的模型，如 GLM-4.7

![Cursor OpenAI配置](/images/docs/cursor-openai-config.jpeg)

> **注意**：在 Cursor 中，需要输入模型的大写名称，如 GLM-4.7

### 3. 保存并切换模型

配置完成后，保存设置并在主页上选择您刚创建的 GLM-4.7 Provider。

![Cursor选择模型](/images/docs/cursor-select-model.jpeg)

### 4. 开始使用

通过该设置，您可以开始使用 GLM 模型进行代码生成、调试、任务分析等工作。
`
}

func getTraeContent() string {
	return `## 简介

TRAE 是一款面向开发者打造的智能 IDE，融合 AI 问答、行内代码补全及 Agentic Coding Workflow 等功能为一体，构建了新一代的编程工作流。

## 安装步骤

### 1. 安装 TRAE

访问 [TRAE 官网](https://trae.ai) 下载并安装适合您的操作系统的版本。

![TRAE官网](/images/docs/trae-website.png)

### 2. 添加自定义模型

1. 在 TRAE 中，点击 **模型切换下拉菜单** 或 **进入 Settings/Models**

![TRAE模型菜单](/images/docs/trae-model-menu.png)

2. 选择添加模型
3. 在模型服务商中选择 **Bigmodel-Plan** 或 **Custom Provider**

![TRAE选择服务商](/images/docs/trae-select-provider.png)

### 3. 配置 API 信息

- **Base URL**: https://your-domain/api/coding/paas/v4
- **API Key**: 从 CodeMind 平台获取的 API Key
- **Model**: 选择 GLM-4.7

![TRAE选择模型](/images/docs/trae-select-model.png)

![TRAE填写API Key](/images/docs/trae-api-key.png)

### 4. 保存并切换模型

配置完成后，保存设置并在自定义模型中选择您刚创建的 GLM-4.7。

![TRAE保存并切换](/images/docs/trae-save-switch.png)
`
}

func getClineContent() string {
	return `## 简介

Cline 是一个强大的 VS Code 插件，可以帮助您在编辑器中直接使用 AI 模型进行代码生成、文件操作等任务。

## 安装步骤

### 1. 打开插件市场

1. 打开 VS Code
2. 点击左侧插件市场图标
3. 在搜索框中输入 cline
4. 找到 Cline 扩展

![Cline搜索](/images/docs/cline-search.png)

### 2. 安装插件

1. 点击 Install 按钮进行安装
2. 安装完成后，选择信任开发者

![Cline安装](/images/docs/cline-install.png)

## 配置 API 设置

### 1. 选择 API Key 方式

选择 Use your own API Key

![Cline选择API Key方式](/images/docs/cline-apikey.png)

### 2. 填入配置信息

- **API Provider**：选择 OpenAI Compatible
- **Base URL**：输入 https://your-domain/api/coding/paas/v4
- **API Key**：填入您的 CodeMind API Key
- **模型**：选择"使用自定义"，并输入模型名称（如：glm-4.7）
- **其他配置**：
  - 取消勾选 Support Images
  - 调整 Context Window Size 为 200000
  - 根据您的任务需求调整 temperature 等其它参数

![Cline配置](/images/docs/cline-config.png)

## 开始使用

配置完成后，您可以在输入框中输入需求，让模型帮助您完成各种任务。

![Cline使用](/images/docs/cline-usage.png)
`
}

func getKiloCodeContent() string {
	return `## 简介

Kilo Code 是一个功能强大的 VS Code 插件，支持 MCP（Model Context Protocol），能够帮助你在编辑器中进行代码生成、调试和项目管理等任务。

## 安装步骤

### 1. 打开插件市场

1. 打开 VS Code
2. 点击左侧插件市场图标
3. 在搜索框中输入 Kilo Code
4. 找到 Kilo Code 插件

![Kilo Code搜索](/images/docs/kilo-search.png)

### 2. 安装插件

1. 点击 Install 按钮进行安装
2. 安装完成后，选择信任开发者

![Kilo Code安装](/images/docs/kilo-install.png)

## 配置 API 设置

### 1. 选择 API Key 方式

选择 使用你自己的 API 秘钥

### 2. 填入配置信息

- **API Provider**：选择 Z AI 或 OpenAI Compatible
- **Base URL**：输入 https://your-domain/api/coding/paas/v4
- **API Key**：填入您的 CodeMind API Key
- **Model**：选择 glm-4.7 或者列表中您想使用的模型

![Kilo Code选择Provider](/images/docs/kilo-provider.png)

![Kilo Code配置](/images/docs/kilo-config.png)

## 开始使用

配置完成后，您可以在输入框中输入需求，让 AI 模型帮助您完成各种任务。
`
}

func getRooCodeContent() string {
	return `## 简介

Roo Code 是一个智能的 VS Code 插件，能够帮助你完成项目分析、代码生成和重构等任务。

## 安装步骤

### 1. 打开插件市场

1. 打开 VS Code
2. 点击左侧插件市场图标
3. 在搜索框中输入 Roo Code
4. 找到 Roo Code 插件

![Roo Code搜索](/images/docs/roo-search.png)

### 2. 安装插件

1. 点击 Install 按钮进行安装
2. 安装完成后，选择信任开发者

![Roo Code安装](/images/docs/roo-install.png)

## 配置 API 设置

请按照以下配置填入相关信息：

- **API Provider**：选择 Z AI 或 OpenAI Compatible
- **Base URL**：输入 https://your-domain/api/coding/paas/v4
- **API Key**：填入您的 CodeMind API Key
- **Model**：选择 glm-4.7 或者列表中您想使用的模型

![Roo Code配置](/images/docs/roo-config.png)

## 权限设置和使用

### 1. 配置权限

根据您的具体需求，选择允许的权限：

- 文件读写操作
- 自动批准执行
- 项目访问权限

![Roo Code权限设置](/images/docs/roo-permissions.png)

### 2. 开始使用

在输入框中输入您的需求，Roo Code 可以帮助您完成各种任务。
`
}

func getOpenCodeContent() string {
	return `## 简介

OpenCode 既是一款在终端中运行的 CLI + TUI AI 编程代理工具，也提供 IDE 的插件集成。

![OpenCode界面](/images/docs/opencode-interface.png)

## 安装步骤

安装 OpenCode 最简单的方式是使用官方安装脚本：

	curl -fsSL https://opencode.ai/install | bash

或者使用 npm 安装：

	npm install -g opencode-ai

## 配置 CodeMind

### 1. 获取 API 密钥

1. 登录 CodeMind 平台
2. 进入「API Keys」页面
3. 点击「创建 API Key」
4. 复制生成的 API Key

### 2. 运行配置命令

	opencode auth login

选择 **OpenAI Compatible** 或 **Zhipu AI Coding Plan**，然后输入：

- **Base URL**: https://your-domain/api/coding/paas/v4
- **API Key**: 您的 CodeMind API Key

### 3. 启动 OpenCode

运行 opencode 启动，使用 /models 命令来选择模型。

## 专属功能

CodeMind 提供了专属的 MCP 服务器，支持多种功能扩展。
`
}

func getFactoryDroidContent() string {
	return `## 简介

Factory Droid 是一款企业级 AI 编码代理，它运行在你的终端中，负责端到端的软件开发工作流。

## 安装步骤

**macOS / Linux：**

	curl -fsSL https://app.factory.ai/cli | sh

**Windows：**

	irm https://app.factory.ai/cli/windows | iex

## 配置 CodeMind

### 配置文件位置

- macOS/Linux： ~/.factory/settings.json
- Windows： %USERPROFILE%\.factory\settings.json

### 配置内容（Anthropic 协议）

` + "```json\n" + `{
  "customModels": [
    {
      "displayName": "GLM-4.7 [CodeMind]",
      "model": "glm-4.7",
      "baseUrl": "https://your-domain/api/anthropic",
      "apiKey": "your_api_key",
      "provider": "anthropic",
      "maxOutputTokens": 131072
    }
  ]
}` + "\n```\n" + `

### 配置内容（OpenAI 协议）

` + "```json\n" + `{
  "customModels": [
    {
      "displayName": "GLM-4.7 [CodeMind] - OpenAI",
      "model": "glm-4.7",
      "baseUrl": "https://your-domain/api/coding/paas/v4",
      "apiKey": "your_api_key",
      "provider": "generic-chat-completion-api",
      "maxOutputTokens": 131072
    }
  ]
}` + "\n```\n" + `

### 启动并选择模型

1. 进入项目目录并启动 droid：
   
   	cd /path/to/your/project
   	droid
   
2. 使用 /model 命令选择 GLM 模型

你配置的 GLM 自定义模型会显示在 "Custom models（自定义模型）" 分区。
`
}

func getCrushContent() string {
	return `## 简介

Crush 既是一款在终端中运行的 CLI + TUI AI 编程工具，支持多种模型接入。

## 安装步骤

根据您的系统选择对应的安装方式：

**Homebrew（macOS 推荐）：**

	brew install charmbracelet/tap/crush

**NPM（跨平台）：**

	npm install -g @charmland/crush

## 配置 CodeMind

### 1. 获取 API 密钥

1. 登录 CodeMind 平台
2. 进入「API Keys」页面
3. 点击「创建 API Key」
4. 复制生成的 API Key

### 2. 修改 Crush 配置

配置文件位置：

- macOS/Linux： ~/.config/crush/crush.json
- Windows： %USERPROFILE%\.config\crush\crush.json

配置内容：

` + "```json\n" + `{
  "providers": {
    "codemind": {
      "id": "codemind",
      "name": "CodeMind Provider",
      "base_url": "https://your-domain/api/coding/paas/v4",
      "api_key": "your_api_key"
    }
  }
}` + "\n```\n" + `

### 3. 完成配置并启动

配置完成后，运行 crush 命令启动，选择 GLM-4.7 模型进行操作。

![Crush模型选择](/images/docs/crush-model.png)
`
}

func getGooseContent() string {
	return `## 简介

Goose 是一款 AI Agent 工具，支持在本地或桌面环境运行，也提供 CLI 形式。

## 安装步骤

1. 访问 Goose 桌面版的官方文档页面
2. 根据您的操作系统选择合适的安装方式
3. 完成 Goose 桌面版的安装

## 配置 CodeMind

### 1. 创建新 Provider

1. 打开 Goose 桌面版应用，进入主界面
2. 找到并点击左侧菜单中的 "创建新 Provider"

![Goose创建Provider](/images/docs/goose-create-provider.jpeg)

3. 按照提示输入所需信息

### 2. 选择 Anthropic 协议并配置

1. 在创建 Provider 的过程中，选择 Anthropic 协议

![Goose Anthropic配置](/images/docs/goose-anthropic-config.png)

2. 填写以下必要的配置：
   - **Base URL**: https://your-domain/api/anthropic
   - **API Key**: 您的 CodeMind API Key
   - **Model**: 选择 glm-4.7

![Goose API配置](/images/docs/goose-api-config.jpeg)

### 3. 切换模型

1. 配置完成后，回到 Goose 桌面版的主界面
2. 在主界面最底部找到并点击 "Switch Models"

![Goose切换模型](/images/docs/goose-switch-model.jpeg)

3. 在下拉列表中选择您刚才创建的新 Provider

## 开始使用

配置完成并切换模型后，您就可以开始使用 Goose 与 GLM 模型进行交互。
`
}

func getOpenClawContent() string {
	return `## 简介

OpenClaw 是一个在您自己的设备上运行的个人 AI 助手，可以连接到各种消息平台。

## 高级配置

### 模型故障转移

配置模型故障转移以确保可靠性：

` + "```json\n" + `{
  "agents": {
    "defaults": {
      "model": {
        "primary": "zai/glm-4.7",
        "fallbacks": ["zai/glm-4.5-air"]
      }
    }
  }
}` + "\n```\n" + `

### 配置 CodeMind

在 OpenClaw 配置文件中添加 CodeMind 提供商：

` + "```json\n" + `{
  "providers": {
    "codemind": {
      "baseUrl": "https://your-domain/api/coding/paas/v4",
      "apiKey": "your_api_key"
    }
  }
}` + "\n```\n" + `

## 故障排除

### 常见问题

1. **API Key 认证**
   - 确保您的 API Key 有效
   - 检查 API Key 在环境中是否正确设置

2. **连接问题**
   - 确保 OpenClaw gateway 正在运行
   - 检查到端点的网络连接
`
}

func getCherryStudioContent() string {
	return `## 简介

将 GLM 模型通过 OpenAI 协议在 Cherry Studio 中自定义配置模型接入使用。

## 安装步骤

### 1. 安装 Cherry Studio

访问 Cherry Studio 官网下载并安装适合您的操作系统的版本。

### 2. 配置 Api Key

在 Cherry Studio 中，打开 "设置 -> 模型服务" 部分：

![Cherry Studio设置](/images/docs/cherry-settings.png)

1. 选择 **智谱开放平台** 或 **OpenAI 兼容**
2. 配置 **API Key**（从 CodeMind 平台获取）
3. 变更 API 地址为：https://your-domain/api/coding/paas/v4/

![Cherry Studio API配置](/images/docs/cherry-api-config.png)

4. 点击下方管理，将 glm-4.7 模型添加

### 3. 开始使用

配置完成后，在对话中选择您刚配置的 GLM-4.7 使用。

![Cherry Studio使用模型](/images/docs/cherry-use-model.png)
`
}

func getOthersContent() string {
	return `## 简介

将 GLM 模型通过 **OpenAI 协议** 接入到兼容该协议的各种工具中。

只要是支持 **OpenAI 协议** 的工具，都可以通过替换请求的 API 链接来接入 CodeMind 平台。

## 适用工具

任何支持 **OpenAI 协议** 的工具，都可以使用 CodeMind 平台。以下是一些常见工具：

- Cursor
- Gemini CLI
- Cherry Studio
- Continue.dev
- ...

## 核心配置步骤

1. 找到适配 OpenAI 协议的 Provider
2. **添加/替换 OpenAI Base URL 为：https://your-domain/api/coding/paas/v4**
3. **输入 CodeMind API Key 并选择模型**

## 配置示例

以通用 OpenAI 兼容工具为例：

### 1. 找到 API 配置部分

在您的工具中找到 API 配置界面。

### 2. 配置参数

- **API Provider**: OpenAI Compatible / Custom
- **Base URL**: https://your-domain/api/coding/paas/v4
- **API Key**: 您的 CodeMind API Key（以 cm- 开头）
- **Model**: glm-4.7 或其他可用模型

### 3. 保存并开始使用

配置完成后，保存设置并开始使用。

## 获取 API Key

1. 登录 CodeMind 平台
2. 进入「API Keys」页面
3. 点击「创建 API Key」
4. 复制生成的 API Key

## 模型说明

CodeMind 平台支持以下模型：

- glm-4.7: 最新最强编码模型
- glm-4.5: 标准版本，适合复杂任务
- glm-4.5-air: 轻量版本，响应更快
`
}
