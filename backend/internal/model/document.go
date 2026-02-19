package model

import "time"

// Document 文档模型 - 存储开发工具接入文档
type Document struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Slug        string     `gorm:"size:50;not null;uniqueIndex" json:"slug"`      // 文档标识，如 claude, cursor
	Title       string     `gorm:"size:200;not null" json:"title"`                // 文档标题
	Subtitle    string     `gorm:"size:500" json:"subtitle"`                      // 副标题/简介
	Icon        string     `gorm:"size:100" json:"icon"`                          // 图标URL或类名
	Content     string     `gorm:"type:text;not null" json:"content"`             // Markdown内容
	SortOrder   int        `gorm:"not null;default:0" json:"sort_order"`          // 排序顺序
	IsPublished bool       `gorm:"not null;default:true" json:"is_published"`     // 是否发布
	CreatedAt   time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (Document) TableName() string {
	return "documents"
}

// DocumentSection 文档章节 - 用于展示时的结构化数据
type DocumentSection struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Level    int    `json:"level"`
	Children []DocumentSection `json:"children,omitempty"`
}

// DocumentListItem 文档列表项
type DocumentListItem struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Subtitle    string    `json:"subtitle"`
	Icon        string    `json:"icon"`
	SortOrder   int       `json:"sort_order"`
	IsPublished bool      `json:"is_published"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 预定义的工具标识常量
const (
	DocSlugClaude       = "claude"
	DocSlugClaudeIDE    = "claude-ide"
	DocSlugOpenClaw     = "openclaw"
	DocSlugOpenCode     = "opencode"
	DocSlugTrae         = "trae"
	DocSlugCline        = "cline"
	DocSlugFactoryDroid = "factory-droid"
	DocSlugKiloCode     = "kilo-code"
	DocSlugRooCode      = "roo-code"
	DocSlugCrush        = "crush"
	DocSlugGoose        = "goose"
	DocSlugCursor       = "cursor"
	DocSlugCherryStudio = "cherry-studio"
	DocSlugOthers       = "others"
)

// 默认工具列表（用于初始化）
var DefaultTools = []Document{
	{
		Slug:        DocSlugClaude,
		Title:       "Claude Code",
		Subtitle:    "智能编码工具，在终端中运行，通过自然语言命令交互",
		Icon:        "🤖",
		SortOrder:   1,
		IsPublished: true,
	},
	{
		Slug:        DocSlugClaudeIDE,
		Title:       "Claude Code IDE 插件",
		Subtitle:    "VS Code 和 JetBrains IDE 插件版本",
		Icon:        "🧩",
		SortOrder:   2,
		IsPublished: true,
	},
	{
		Slug:        DocSlugCursor,
		Title:       "Cursor",
		Subtitle:    "AI 驱动的代码编辑器，支持自定义模型配置",
		Icon:        "🎯",
		SortOrder:   3,
		IsPublished: true,
	},
	{
		Slug:        DocSlugTrae,
		Title:       "TRAE",
		Subtitle:    "面向开发者打造的智能 IDE",
		Icon:        "⚡",
		SortOrder:   4,
		IsPublished: true,
	},
	{
		Slug:        DocSlugCline,
		Title:       "Cline",
		Subtitle:    "VS Code 插件，支持 AI 代码生成和文件操作",
		Icon:        "🔧",
		SortOrder:   5,
		IsPublished: true,
	},
	{
		Slug:        DocSlugKiloCode,
		Title:       "Kilo Code",
		Subtitle:    "支持 MCP 的 VS Code 插件",
		Icon:        "💻",
		SortOrder:   6,
		IsPublished: true,
	},
	{
		Slug:        DocSlugRooCode,
		Title:       "Roo Code",
		Subtitle:    "智能 VS Code 插件，支持项目分析和代码生成",
		Icon:        "🦘",
		SortOrder:   7,
		IsPublished: true,
	},
	{
		Slug:        DocSlugOpenCode,
		Title:       "OpenCode",
		Subtitle:    "CLI + TUI AI 编程代理工具，支持 IDE 插件",
		Icon:        "🔨",
		SortOrder:   8,
		IsPublished: true,
	},
	{
		Slug:        DocSlugFactoryDroid,
		Title:       "Factory Droid",
		Subtitle:    "企业级 AI 编码代理，端到端软件开发",
		Icon:        "🏭",
		SortOrder:   9,
		IsPublished: true,
	},
	{
		Slug:        DocSlugCrush,
		Title:       "Crush",
		Subtitle:    "CLI + TUI AI 编程工具",
		Icon:        "💎",
		SortOrder:   10,
		IsPublished: true,
	},
	{
		Slug:        DocSlugGoose,
		Title:       "Goose",
		Subtitle:    "AI Agent 工具，支持桌面版和 CLI",
		Icon:        "🪿",
		SortOrder:   11,
		IsPublished: true,
	},
	{
		Slug:        DocSlugOpenClaw,
		Title:       "OpenClaw",
		Subtitle:    "个人 AI 助手，可连接多种消息平台",
		Icon:        "🦞",
		SortOrder:   12,
		IsPublished: true,
	},
	{
		Slug:        DocSlugCherryStudio,
		Title:       "Cherry Studio",
		Subtitle:    "支持多种模型配置的客户端工具",
		Icon:        "🍒",
		SortOrder:   13,
		IsPublished: true,
	},
	{
		Slug:        DocSlugOthers,
		Title:       "其他工具",
		Subtitle:    "支持 OpenAI 协议的其他工具配置方法",
		Icon:        "📦",
		SortOrder:   14,
		IsPublished: true,
	},
}
