package handler

import (
	"fmt"
	"strings"
	"time"
	"voice-memory/internal/service"

	"github.com/gin-gonic/gin"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	glmClient         *service.GLMClient
	sessionManager    *service.SessionManager
	ttsService        *service.BaiduTTS
	intentRecognizer  *service.IntentRecognizer
	contextCompressor *service.ContextCompressor
	ragService        *service.RAGService
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(glmClient *service.GLMClient) *ChatHandler {
	return &ChatHandler{
		glmClient:         glmClient,
		sessionManager:    service.NewSessionManager(),
		intentRecognizer:  service.NewIntentRecognizer(),
		contextCompressor: service.NewContextCompressor(glmClient, service.DefaultContextConfig()),
	}
}

// NewChatHandlerWithSession 创建聊天处理器（使用共享 SessionManager）
func NewChatHandlerWithSession(glmClient *service.GLMClient, sessionManager *service.SessionManager) *ChatHandler {
	return &ChatHandler{
		glmClient:         glmClient,
		sessionManager:    sessionManager,
		intentRecognizer:  service.NewIntentRecognizer(),
		contextCompressor: service.NewContextCompressor(glmClient, service.DefaultContextConfig()),
	}
}

// NewChatHandlerWithTTS 创建聊天处理器（带 TTS）
func NewChatHandlerWithTTS(glmClient *service.GLMClient, sessionManager *service.SessionManager, ttsService *service.BaiduTTS) *ChatHandler {
	return &ChatHandler{
		glmClient:         glmClient,
		sessionManager:    sessionManager,
		ttsService:        ttsService,
		intentRecognizer:  service.NewIntentRecognizer(),
		contextCompressor: service.NewContextCompressor(glmClient, service.DefaultContextConfig()),
		ragService:        nil, // 默认不启用 RAG
	}
}

// NewChatHandlerWithRAG 创建聊天处理器（带 RAG）
func NewChatHandlerWithRAG(glmClient *service.GLMClient, sessionManager *service.SessionManager, ttsService *service.BaiduTTS, ragService *service.RAGService) *ChatHandler {
	return &ChatHandler{
		glmClient:         glmClient,
		sessionManager:    sessionManager,
		ttsService:        ttsService,
		intentRecognizer:  service.NewIntentRecognizer(),
		contextCompressor: service.NewContextCompressor(glmClient, service.DefaultContextConfig()),
		ragService:        ragService,
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Message      string `json:"message" binding:"required"`
	SessionID    string `json:"session_id,omitempty"`    // 会话 ID，用于多轮对话
	IncludeAudio bool   `json:"include_audio,omitempty"` // 是否返回音频 URL
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Success   bool   `json:"success"`
	Reply     string `json:"reply,omitempty"`
	SessionID string `json:"session_id,omitempty"` // 返回会话 ID
	AudioURL  string `json:"audio_url,omitempty"` // 音频 URL
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

	// 构建响应
	resp := ChatResponse{
		Success:   true,
		Reply:     reply,
		SessionID: sessionID,
	}

	// 如果需要音频，生成音频 URL
	if req.IncludeAudio && h.ttsService != nil {
		filename, err := h.ttsService.SynthesizeToFile(service.DefaultTTSOptions(reply))
		if err == nil {
			resp.AudioURL = "/api/audio/" + filename
		}
		// TTS 失败不影响文本回复，继续返回
	}

	c.JSON(200, resp)
}

// generateID 生成随机 ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// HandleChatStream 处理流式聊天请求 (SSE)
func (h *ChatHandler) HandleChatStream(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// 获取或创建会话
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = "session_" + generateID()
	}
	_ = h.sessionManager.GetOrCreateSession(sessionID)

	// ===== 识别用户意图 =====
	intentResult := h.intentRecognizer.Recognize(req.Message)
	fmt.Printf("用户意图识别: %s (置信度: %.2f)\n", intentResult.Intent, intentResult.Confidence)

	// 发送意图信息给客户端
	fmt.Fprintf(c.Writer, "data: {\"type\":\"intent\",\"intent\":\"%s\",\"confidence\":%.2f,\"description\":\"%s\"}\n\n",
		intentResult.Intent, intentResult.Confidence, intentResult.Intent.GetIntentDescription())
	c.Writer.Flush()

	// 特殊意图处理
	if intentResult.Intent == service.IntentClear {
		h.sessionManager.ClearSession(sessionID)
		fmt.Fprintf(c.Writer, "data: {\"type\":\"content\",\"delta\":\"好的，已清空对话历史，我们可以重新开始啦！\"}\n\n")
		fmt.Fprintf(c.Writer, "data: {\"type\":\"done\",\"session_id\":\"%s\"}\n\n", sessionID)
		c.Writer.Flush()
		return
	}

	// ===== RAG 知识检索 =====
	var ragContext string
	if h.ragService != nil && h.ragService.ShouldUseRAG(intentResult.Intent, intentResult.Confidence) {
		fmt.Printf("RAG 检索中... (意图: %s)\n", intentResult.Intent)
		context, err := h.ragService.BuildContextWithRAG(req.Message, 3) // 检索 top-3
		if err != nil {
			fmt.Printf("RAG 检索失败: %v\n", err)
		} else if context != "" {
			ragContext = context
			fmt.Printf("RAG 检索成功，找到相关知识\n")

			// 发送 RAG 检索信息给客户端
			fmt.Fprintf(c.Writer, "data: {\"type\":\"rag\",\"found\":true,\"count\":3}\n\n")
			c.Writer.Flush()
		}
	}

	// 添加用户消息到历史
	h.sessionManager.AddMessage(sessionID, "user", req.Message)

	// ===== 智能上下文压缩 =====
	// 构建系统提示（包含 RAG 上下文）
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

	// 如果有 RAG 检索到的相关知识，添加到系统提示中
	if ragContext != "" {
		systemPrompt += "\n\n" + ragContext + "\n请根据以上知识库内容回答用户问题。"
	}

	// 获取历史消息并压缩
	historyMessages := h.sessionManager.GetMessages(sessionID)

	// 使用上下文压缩器
	compressedCtx, err := h.contextCompressor.Compress(historyMessages)
	if err != nil {
		fmt.Printf("上下文压缩警告: %v\n", err)
		// 压缩失败，使用原始消息
		compressedCtx = &service.CompressedContext{
			RecentMessages: historyMessages,
			TotalMessages:  len(historyMessages),
		}
	}

	// 如果有新摘要，更新会话
	if compressedCtx.Summary != nil {
		h.sessionManager.UpdateSummary(sessionID, compressedCtx.Summary)
	}

	// 构建用于 API 的消息列表
	messages := h.contextCompressor.BuildMessagesForAPI(
		compressedCtx,
		systemPrompt,
		req.Message,
	)

	// 调试日志
	summaryCount := 0
	if compressedCtx.Summary != nil {
		summaryCount = compressedCtx.Summary.MessageCount
	}
	fmt.Printf("上下文压缩: 原始 %d 条 → 摘要 %d 条 + 最近 %d 条\n",
		compressedCtx.TotalMessages,
		summaryCount,
		len(compressedCtx.RecentMessages))

	// 收集完整回复用于 TTS
	var fullReply strings.Builder

	// 发送流式请求
	err = h.glmClient.SendMessageStream(service.ChatRequest{
		Model:       "glm-4-plus",
		MaxTokens:   1024,
		Messages:    messages,
		Temperature: 0.85,
	}, func(chunk service.StreamChunk) {
		if chunk.Done {
			// 流结束
			reply := fullReply.String()

			// 添加助手回复到历史
			h.sessionManager.AddMessage(sessionID, "assistant", reply)

			// 注释：暂时不自动生成 TTS，节省资源
			// 用户需要点击"播放语音"按钮才会调用百度 TTS
			// 如果需要恢复自动 TTS，取消下面注释即可
			/*
			if req.IncludeAudio && h.ttsService != nil {
				filename, err := h.ttsService.SynthesizeToFile(service.DefaultTTSOptions(reply))
				if err == nil {
					// 发送音频 URL
					fmt.Fprintf(c.Writer, "data: {\"type\":\"audio\",\"url\":\"/api/audio/%s\",\"session_id\":\"%s\"}\n\n", filename, sessionID)
					c.Writer.Flush()
				}
			}
			*/

			// 发送完成标记
			fmt.Fprintf(c.Writer, "data: {\"type\":\"done\",\"session_id\":\"%s\"}\n\n", sessionID)
			c.Writer.Flush()
			return
		}

		if chunk.Error != "" {
			fmt.Fprintf(c.Writer, "data: {\"type\":\"error\",\"error\":\"%s\"}\n\n", chunk.Error)
			c.Writer.Flush()
			return
		}

		if chunk.Delta != "" {
			fullReply.WriteString(chunk.Delta)
			fmt.Fprintf(c.Writer, "data: {\"type\":\"content\",\"delta\":\"%s\"}\n\n", escapeJSON(chunk.Delta))
			c.Writer.Flush()
		}
	})

	if err != nil {
		fmt.Fprintf(c.Writer, "data: {\"type\":\"error\",\"error\":\"%s\"}\n\n", escapeJSON(err.Error()))
		c.Writer.Flush()
	}
}

// escapeJSON 转义 JSON 字符串
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
