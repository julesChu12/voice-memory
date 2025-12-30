package handler

import (
	"voice-memory/internal/service"

	"github.com/gin-gonic/gin"
)

// SessionHandler 会话处理器
type SessionHandler struct {
	sessionManager *service.SessionManager
}

// NewSessionHandler 创建会话处理器
func NewSessionHandler(sessionManager *service.SessionManager) *SessionHandler {
	return &SessionHandler{
		sessionManager: sessionManager,
	}
}

// GetSessionResponse 获取会话响应
type GetSessionResponse struct {
	Success bool               `json:"success"`
	Session *service.Session   `json:"session,omitempty"`
	Error   string             `json:"error,omitempty"`
}

// ListSessionsResponse 列表响应
type ListSessionsResponse struct {
	Success  bool                 `json:"success"`
	Sessions []service.Session    `json:"sessions,omitempty"`
	Count    int                  `json:"count,omitempty"`
	Error    string               `json:"error,omitempty"`
}

// HandleGetSession 获取单个会话
func (h *SessionHandler) HandleGetSession(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(400, GetSessionResponse{
			Success: false,
			Error:   "缺少 session_id 参数",
		})
		return
	}

	session := h.sessionManager.GetSession(sessionID)
	if session == nil {
		c.JSON(404, GetSessionResponse{
			Success: false,
			Error:   "会话不存在",
		})
		return
	}

	c.JSON(200, GetSessionResponse{
		Success: true,
		Session: session,
	})
}

// HandleListSessions 列出所有会话
func (h *SessionHandler) HandleListSessions(c *gin.Context) {
	sessions := h.sessionManager.GetAllSessions()

	c.JSON(200, ListSessionsResponse{
		Success:  true,
		Sessions: sessions,
		Count:    len(sessions),
	})
}

// HandleDeleteSession 删除会话
func (h *SessionHandler) HandleDeleteSession(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(400, GetSessionResponse{
			Success: false,
			Error:   "缺少 session_id 参数",
		})
		return
	}

	h.sessionManager.DeleteSession(sessionID)

	c.JSON(200, GetSessionResponse{
		Success: true,
	})
}
