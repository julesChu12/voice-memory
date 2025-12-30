package handler

import (
	"io"
	"voice-memory/internal/service"

	"github.com/gin-gonic/gin"
)

// AudioChatHandler 音频聊天处理器
type AudioChatHandler struct {
	glmClient       *service.GLMClient
	sessionManager  *service.SessionManager
}

// NewAudioChatHandler 创建音频聊天处理器
func NewAudioChatHandler(glmClient *service.GLMClient) *AudioChatHandler {
	return &AudioChatHandler{
		glmClient:      glmClient,
		sessionManager: service.NewSessionManager(),
	}
}

// AudioChatRequest 音频聊天请求
type AudioChatRequest struct {
	SessionID string `form:"session_id"` // 会话 ID
}

// AudioChatResponse 音频聊天响应
type AudioChatResponse struct {
	Success   bool   `json:"success"`
	Reply     string `json:"reply,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// HandleAudioChat 处理音频聊天请求
func (h *AudioChatHandler) HandleAudioChat(c *gin.Context) {
	// 获取会话 ID
	sessionID := c.PostForm("session_id")
	if sessionID == "" {
		sessionID = c.Query("session_id")
	}
	if sessionID == "" {
		sessionID = "audio_session_" + generateID()
	}
	_ = h.sessionManager.GetOrCreateSession(sessionID)

	// 读取上传的音频文件
	fileHeader, err := c.FormFile("audio")
	if err != nil {
		c.JSON(400, AudioChatResponse{
			Success: false,
			Error:   "音频文件上传失败: " + err.Error(),
		})
		return
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, AudioChatResponse{
			Success: false,
			Error:   "打开音频文件失败: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// 读取音频数据
	audioData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(500, AudioChatResponse{
			Success: false,
			Error:   "读取音频数据失败: " + err.Error(),
		})
		return
	}

	// 获取历史消息
	historyMessages := h.sessionManager.GetMessages(sessionID)

	// 构建系统 prompt
	systemPrompt := service.Message{
		Role: "user",
		Content: `你是 Voice Memory，一个温暖贴心、有幽默感的 AI 语音助手。

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

【重要】用户发送的是语音，你需要理解语音内容并回复。记住对话上下文。`,
	}

	// 将系统 prompt 添加到历史消息前面
	allMessages := append([]service.Message{systemPrompt}, historyMessages...)

	// 调用 GLM Audio API
	response, err := h.glmClient.SendMessageWithAudio(audioData, allMessages)
	if err != nil {
		c.JSON(500, AudioChatResponse{
			Success: false,
			Error:   "GLM Audio API 调用失败: " + err.Error(),
		})
		return
	}

	reply := response.GetReplyText()

	// 添加用户音频消息到历史（用一个占位符表示）
	h.sessionManager.AddMessage(sessionID, "user", "[语音输入]")

	// 添加助手回复到历史
	h.sessionManager.AddMessage(sessionID, "assistant", reply)

	c.JSON(200, AudioChatResponse{
		Success:   true,
		Reply:     reply,
		SessionID: sessionID,
	})
}
