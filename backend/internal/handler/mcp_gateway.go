package handler

import (
	"fmt"
	"net/http"

	"codemind/internal/middleware"
	"codemind/internal/service"

	mcpPkg "codemind/pkg/mcp"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MCPGatewayHandler handles MCP protocol requests from AI coding tools.
type MCPGatewayHandler struct {
	mcpService     *service.MCPService
	sessionManager *mcpPkg.SessionManager
	logger         *zap.Logger
}

// NewMCPGatewayHandler creates a new MCP gateway handler.
func NewMCPGatewayHandler(mcpService *service.MCPService, logger *zap.Logger) *MCPGatewayHandler {
	return &MCPGatewayHandler{
		mcpService:     mcpService,
		sessionManager: mcpPkg.NewSessionManager(),
		logger:         logger,
	}
}

// SSEConnect handles GET /mcp/sse requests.
func (h *MCPGatewayHandler) SSEConnect(c *gin.Context) {
	messageBaseURL := h.getMessageBaseURL(c)
	session, err := mcpPkg.NewSSESession(c.Writer, messageBaseURL)
	if err != nil {
		h.logger.Error("failed to create SSE session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSE not supported"})
		return
	}

	h.sessionManager.Add(session)
	defer h.sessionManager.Remove(session.ID)

	h.logger.Info("MCP SSE connection established",
		zap.String("session_id", session.ID),
		zap.Int64("user_id", middleware.GetUserID(c)),
	)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)

	if err := session.SendEndpoint(); err != nil {
		h.logger.Error("failed to send endpoint event", zap.Error(err))
		return
	}

	select {
	case <-session.Done:
		h.logger.Info("MCP SSE session ended", zap.String("session_id", session.ID))
	case <-c.Request.Context().Done():
		h.logger.Info("MCP SSE client disconnected", zap.String("session_id", session.ID))
	}
}

// HandleMessage handles POST /mcp/message requests.
func (h *MCPGatewayHandler) HandleMessage(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing sessionId parameter"})
		return
	}

	session, ok := h.sessionManager.Get(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found or expired"})
		return
	}

	var request mcpPkg.JSONRPCRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errResp := mcpPkg.NewErrorResponse(nil, mcpPkg.ErrCodeParseError, "JSON parse failed")
		c.JSON(http.StatusOK, errResp)
		return
	}

	if request.JSONRPC != mcpPkg.JSONRPCVersion {
		errResp := mcpPkg.NewErrorResponse(request.ID, mcpPkg.ErrCodeInvalidRequest, "unsupported JSON-RPC version")
		c.JSON(http.StatusOK, errResp)
		return
	}

	h.logger.Debug("received MCP request",
		zap.String("session_id", sessionID),
		zap.String("method", request.Method),
	)

	resp, err := h.handleMCPRequest(c, &request)
	if err != nil {
		h.logger.Error("MCP request handling failed", zap.Error(err), zap.String("method", request.Method))
		errResp := mcpPkg.NewErrorResponse(request.ID, mcpPkg.ErrCodeInternalError, "internal processing error")
		_ = session.SendMessage(errResp)
		c.Status(http.StatusAccepted)
		return
	}

	if err := session.SendMessage(resp); err != nil {
		h.logger.Error("failed to send SSE response", zap.Error(err))
		c.JSON(http.StatusOK, resp)
		return
	}

	c.Status(http.StatusAccepted)
}

// HandleStreamableHTTP handles POST /mcp/ requests.
func (h *MCPGatewayHandler) HandleStreamableHTTP(c *gin.Context) {
	var request mcpPkg.JSONRPCRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errResp := mcpPkg.NewErrorResponse(nil, mcpPkg.ErrCodeParseError, "JSON parse failed")
		c.JSON(http.StatusOK, errResp)
		return
	}

	if request.JSONRPC != mcpPkg.JSONRPCVersion {
		errResp := mcpPkg.NewErrorResponse(request.ID, mcpPkg.ErrCodeInvalidRequest, "unsupported JSON-RPC version")
		c.JSON(http.StatusOK, errResp)
		return
	}

	h.logger.Debug("received MCP Streamable HTTP request",
		zap.String("method", request.Method),
	)

	resp, err := h.handleMCPRequest(c, &request)
	if err != nil {
		h.logger.Error("MCP Streamable HTTP request failed", zap.Error(err), zap.String("method", request.Method))
		errResp := mcpPkg.NewErrorResponse(request.ID, mcpPkg.ErrCodeInternalError, "internal processing error")
		c.JSON(http.StatusOK, errResp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *MCPGatewayHandler) handleMCPRequest(c *gin.Context, request *mcpPkg.JSONRPCRequest) (*mcpPkg.JSONRPCResponse, error) {
	if request.Method == mcpPkg.MethodInitialized {
		return mcpPkg.NewResponse(request.ID, struct{}{}), nil
	}

	serviceInfos, err := h.mcpService.GetServiceInfosForGateway()
	if err != nil {
		return nil, fmt.Errorf("failed to get MCP service list: %w", err)
	}

	userID := middleware.GetUserID(c)
	deptID := middleware.GetDepartmentID(c)
	roleStr, _ := c.Get(middleware.CtxKeyRole)
	role, _ := roleStr.(string)

	allowedServices := h.mcpService.FilterAccessibleServices(serviceInfos, userID, deptID, role)

	proxy := h.mcpService.GetProxy()
	return proxy.HandleRequest(request, allowedServices)
}

func (h *MCPGatewayHandler) getMessageBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/mcp/message", scheme, c.Request.Host)
}

// GetSessionCount returns the current active session count.
func (h *MCPGatewayHandler) GetSessionCount() int {
	return h.sessionManager.Count()
}
