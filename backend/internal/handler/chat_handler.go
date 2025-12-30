package handler

import (
	"fmt"
	"time"
	"voice-memory/internal/service"

	"github.com/gin-gonic/gin"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	glmClient       *service.GLMClient
	sessionManager  *service.SessionManager
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(glmClient *service.GLMClient) *ChatHandler {
	return &ChatHandler{
		glmClient:      glmClient,
		sessionManager: service.NewSessionManager(),
	}
}

// NewChatHandlerWithSession 创建聊天处理器（使用共享 SessionManager）
func NewChatHandlerWithSession(glmClient *service.GLMClient, sessionManager *service.SessionManager) *ChatHandler {
	return &ChatHandler{
		glmClient:      glmClient,
		sessionManager: sessionManager,
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Message   string `json:"message" binding:"required"`
	SessionID string `json:"session_id,omitempty"` // 会话 ID，用于多轮对话
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Success   bool   `json:"success"`
	Reply     string `json:"reply,omitempty"`
	SessionID string `json:"session_id,omitempty"` // 返回会话 ID
	Error     string `json:"error,omitempty"`
}

// HandleChat 处理聊天请求
func (h *ChatHandler) HandleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ChatResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取或创建会话
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = "session_" + generateID()
	}
	_ = h.sessionManager.GetOrCreateSession(sessionID)

	// 添加用户消息到历史
	h.sessionManager.AddMessage(sessionID, "user", req.Message)

	// 构建消息列表（包含历史）
	systemPrompt := `你是 Voice Memory，一个温暖贴心、有幽默感的 AI 语音助手。

【你的性格】
- 温暖友善：像朋友一样自然交流，不生硬
- 适度幽默：轻松时可以开玩笑，但不刻意
- 贴心细致：关注用户需求，主动提供帮助
- 简洁自然：回复 50-100 字，口语化表达

【对话风格】
- 用"我"而不是"本助手"或"AI"
- 可以用语气词如"哈"、"呢"、"哦"
- 避免机械式回复，要有个性
- 遇到不知道的，诚实说"这个我也不太清楚"

【重要】记住用户的偏好和之前对话的上下文，保持对话连贯性。`

	// 获取历史消息
	messages := []service.Message{
		{Role: "user", Content: systemPrompt},
	}
	messages = append(messages, h.sessionManager.GetMessages(sessionID)...)

	response, err := h.glmClient.SendMessage(service.ChatRequest{
		Model:       "glm-4-plus",
		MaxTokens:   1024,
		Messages:    messages,
		Temperature: 0.85,
	})

	if err != nil {
		c.JSON(500, ChatResponse{
			Success: false,
			Error:   "AI 调用失败: " + err.Error(),
		})
		return
	}

	reply := response.GetReplyText()

	// 添加助手回复到历史
	h.sessionManager.AddMessage(sessionID, "assistant", reply)

	c.JSON(200, ChatResponse{
		Success:   true,
		Reply:     reply,
		SessionID: sessionID,
	})
}

// generateID 生成随机 ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
