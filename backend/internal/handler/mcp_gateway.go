package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"codemind/internal/middleware"
	"codemind/internal/service"
	mcpPkg "codemind/pkg/mcp"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MCPGatewayHandler MCP 网关协议处理器
// 处理来自 AI 编码工具的 MCP 协议请求
type MCPGatewayHandler struct {
	mcpService     *service.MCPService
	sessionManager *mcpPkg.SessionManager
	logger         *zap.Logger
}

// NewMCPGatewayHandler 创建 MCP 网关 Handler
func NewMCPGatewayHandler(mcpService *service.MCPService, logger *zap.Logger) *MCPGatewayHandler {
	return &MCPGatewayHandler{
		mcpService:     mcpService,
		sessionManager: mcpPkg.NewSessionManager(),
		logger:         logger,
	}
}

// SSEConnect 建立 SSE 连接（MCP SSE 传输）
// GET /mcp/sse
func (h *MCPGatewayHandler) SSEConnect(c *gin.Context) {
	// 创建 SSE 会话
	messageBaseURL := h.getMessageBaseURL(c)
	session, err := mcpPkg.NewSSESession(c.Writer, messageBaseURL)
	if err != nil {
		h.logger.Error("创建 SSE 会话失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "不支持 SSE"})
		return
	}

	// 注册会话
	h.sessionManager.Add(session)
	defer h.sessionManager.Remove(session.ID)

	h.logger.Info("MCP SSE 连接已建立",
		zap.String("session_id", session.ID),
		zap.Int64("user_id", middleware.GetUserID(c)),
	)

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)

	// 发送端点 URL 事件
	if err := session.SendEndpoint(); err != nil {
		h.logger.Error("发送 endpoint 事件失败", zap.Error(err))
		return
	}

	// 保持连接直到客户端断开
	select {
	case <-session.Done:
		h.logger.Info("MCP SSE 会话结束", zap.String("session_id", session.ID))
	case <-c.Request.Context().Done():
		h.logger.Info("MCP SSE 客户端断开", zap.String("session_id", session.ID))
	}
}

// HandleMessage 处理 JSON-RPC 消息
// POST /mcp/message
func (h *MCPGatewayHandler) HandleMessage(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 sessionId 参数"})
		return
	}

	// 验证会话
	session, ok := h.sessionManager.Get(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在或已过期"})
		return
	}

	// 解析 JSON-RPC 请求
	var request mcpPkg.JSONRPCRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errResp := mcpPkg.NewErrorResponse(nil, mcpPkg.ErrCodeParseError, "JSON 解析失败")
		c.JSON(http.StatusOK, errResp)
		return
	}

	// 验证 JSON-RPC 版本
	if request.JSONRPC != mcpPkg.JSONRPCVersion {
		errResp := mcpPkg.NewErrorResponse(request.ID, mcpPkg.ErrCodeInvalidRequest, "不支持的 JSON-RPC 版本")
		c.JSON(http.StatusOK, errResp)
		return
	}

	h.logger.Debug("收到 MCP 请求",
		zap.String("session_id", sessionID),
		zap.String("method", request.Method),
	)

	// 处理请求
	resp, err := h.handleMCPRequest(c, &request)
	if err != nil {
		errResp := mcpPkg.NewErrorResponse(request.ID, mcpPkg.ErrCodeInternalError, err.Error())
		// 通过 SSE 通道返回响应
		_ = session.SendMessage(errResp)
		c.Status(http.StatusAccepted)
		return
	}

	// 通过 SSE 通道返回响应
	if err := session.SendMessage(resp); err != nil {
		h.logger.Error("发送 SSE 响应失败", zap.Error(err))
		// 降级为 HTTP 直接返回
		c.JSON(http.StatusOK, resp)
		return
	}

	c.Status(http.StatusAccepted)
}

// HandleStreamableHTTP 处理 Streamable HTTP 传输
// POST /mcp/
func (h *MCPGatewayHandler) HandleStreamableHTTP(c *gin.Context) {
	// 解析请求
	var request mcpPkg.JSONRPCRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errResp := mcpPkg.NewErrorResponse(nil, mcpPkg.ErrCodeParseError, "JSON 解析失败")
		c.JSON(http.StatusOK, errResp)
		return
	}

	if request.JSONRPC != mcpPkg.JSONRPCVersion {
		errResp := mcpPkg.NewErrorResponse(request.ID, mcpPkg.ErrCodeInvalidRequest, "不支持的 JSON-RPC 版本")
		c.JSON(http.StatusOK, errResp)
		return
	}

	h.logger.Debug("收到 MCP Streamable HTTP 请求",
		zap.String("method", request.Method),
	)

	resp, err := h.handleMCPRequest(c, &request)
	if err != nil {
		errResp := mcpPkg.NewErrorResponse(request.ID, mcpPkg.ErrCodeInternalError, err.Error())
		c.JSON(http.StatusOK, errResp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// handleMCPRequest 处理 MCP 请求的核心逻辑
func (h *MCPGatewayHandler) handleMCPRequest(c *gin.Context, request *mcpPkg.JSONRPCRequest) (*mcpPkg.JSONRPCResponse, error) {
	// 处理 initialized 通知
	if request.Method == mcpPkg.MethodInitialized {
		// 通知消息无需响应
		return mcpPkg.NewResponse(request.ID, struct{}{}), nil
	}

	// 获取可用服务信息
	serviceInfos, err := h.mcpService.GetServiceInfosForGateway()
	if err != nil {
		return nil, fmt.Errorf("获取 MCP 服务列表失败: %w", err)
	}

	// 检查用户访问权限，过滤用户有权限访问的服务
	userID := middleware.GetUserID(c)
	deptID := middleware.GetDepartmentID(c)
	roleStr, _ := c.Get(middleware.CtxKeyRole)
	role, _ := roleStr.(string)

	allowedServices := h.mcpService.FilterAccessibleServices(serviceInfos, userID, deptID, role)

	// 通过代理处理请求
	proxy := h.mcpService.GetProxy()
	return proxy.HandleRequest(request, allowedServices)
}

// getMessageBaseURL 获取消息端点的完整 URL
func (h *MCPGatewayHandler) getMessageBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/mcp/message", scheme, c.Request.Host)
}

// GetSessionCount 获取当前活跃会话数（用于监控）
func (h *MCPGatewayHandler) GetSessionCount() int {
	return h.sessionManager.Count()
}

// encodeJSON 辅助 JSON 编码
func encodeJSON(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
