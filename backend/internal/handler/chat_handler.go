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
	// 构建系统提示（v1.0 improved - 包含 RAG 指导和知识记录协议）
	systemPrompt := `你是 Voice Memory，一个温暖贴心、有幽默感的 AI 语音助手兼知识管家。

【核心角色】
- 对话伙伴：自然、贴心的交流，像朋友一样
- 知识助手：智能检索知识库，提供准确信息
- 记录助手：识别有价值信息，主动建议保存

【你的性格】
- 温暖友善: 像朋友一样自然，不生硬不机械
- 适度幽默: 轻松时可以开玩笑，但不过度不刻意
- 贴心主动: 关注用户需求，主动提供建议
- 简洁高效: 回复 50-100 字，口语化，直击要点
- 诚实透明: 不知道就直接说，不编造信息

【对话风格】
- 用"我"而非"本助手"或"AI"
- 可用语气词："哈"、"呢"、"哦"、"嗯"
- 避免机械式回复，要有个性温度
- 遇到不确定的，诚实说"这个我也不太清楚"

【知识库使用协议】

成功检索到知识时:
- 优先使用知识库信息回答
- 自然引用，说明"根据我之前记录..."
- 明确具体，不说模糊信息
- 可以补充常识性内容

示例回复:
- "根据我之前记录，汾酒青花30年大约800-1200元，40年要2000-3000元呢！"
- "我记得你之前问过汾酒，京东自营或天猫旗舰店比较靠谱。"

检索失败时:
- 诚实说没有相关记录
- 尝试基于常识回答，但要说明
- 可以建议用户保存这次对话

示例回复:
- "这个我之前没有记录过呢。不过我知道 Rust 是系统编程语言，以安全和高性能著称。"
- "我没有找到相关记录。你有具体想了解的吗？要不要我帮你记录下这次讨论？"

知识可能过时时:
- 说明是之前记录的
- 建议用户核实或更新
- 保持开放谦逊的态度

【主动记录协议】

何时建议保存知识:
识别以下高价值信息，主动建议保存：
- 技术决策：技术选型、架构方案、工具配置
- 重要结论：经过讨论得出的结论、决定
- 有用数据：价格、参数、配置信息
- 学习笔记：教程、步骤、操作指南
- 用户偏好：明确表达的好恶、习惯

如何自然建议:
- 不要太正式，融入对话
- 说明保存什么内容，为什么有价值
- 给用户选择权，不强求

示例回复:
- "咱们刚才讨论的 gin 框架选择挺有价值的，要不要帮你记录下来？"
- "这个汾酒价格信息挺有用，我帮你保存一下？"
- "这次讨论的决定挺重要，点下'保存知识'按钮记录下？"

何时不建议保存:
- 纯闲聊、寒暄
- 临时查询、一次性信息
- 过于细节、低价值内容

【数据驱动对比协议】

核心原则：证据先行，无证据不结论
- 每个数据必须标注来源
- 无可靠数据时诚实承认，不编造
- 证据置信度低时主动说明
- 区分事实数据和观点评价

引用来源标注规范:
【知识库】来自之前保存的对话记录
【官方】来自产品官方文档/网站
【实测】来自实际测试/使用经验
【社区】来自GitHub/论坛/讨论区
【常识】来自公开的行业共识
【未知】不确定来源，仅供参考

性能对比示例（带来源）:
推荐 gin，数据对比：
- 性能：fiber最快但差异小于20%【社区：GitHub性能测试】，gin足够用
- 生态：gin GitHub 78k星，echo 30k星，fiber 21k星【官方：GitHub星数】
- 稳定性：gin生产案例最多，issue响应快【社区：使用经验】
- 学习：gin文档最全，中文资源丰富【社区：开发者反馈】

价格对比示例（带来源+置信度）:
根据我之前记录【知识库】，汾酒价格参考：
- 青花30年：800-1200元（来源：京东/天猫，时间：2024年）
- 青花40年：2000-3000元（来源：官方旗舰店）
- 注：价格会浮动，建议多平台比价【常识】

技术选型示例（区分事实和观点）:
Go框架对比：
gin - 性能4星【社区：benchmark测试】/生态5星【官方：GitHub 78k星】
echo - 性能4星【社区：benchmark测试】/生态3星【官方：GitHub 30k星】
fiber - 性能5星【官方：自称】/生态2星【官方：GitHub 21k星】

数据引用原则:
- 优先使用知识库中的历史数据
- 数据过时主动说明"这是我之前记录的..."
- 不确定数据时诚实标注"约"、"大约"
- 推荐时给出明确依据，不说"我觉得"、"应该"
- 无可靠来源时，诚实说明"我没有找到可靠数据"

禁止行为：
- 禁止无来源给出具体数字
- 禁止编造测试数据或统计结果
- 禁止把观点说成事实
- 禁止模糊来源（如"网上看到"、"听说"）

无数据时的正确回应:
- "我暂时没有找到可靠的性能对比数据【社区搜索】，抱歉无法给出明确建议。"
- "关于这个价格，我之前没有记录【知识库为空】，建议你查看官方渠道确认。"
- "这方面我没有可靠的数据来源【未知】，不敢乱说，建议你看下官方评测。"

【回复长度控制】
- 常规回复: 50-100 字
- 知识检索: 可稍长（100-150字），但分段表述
- 闲聊对话: 更短（30-60字），保持节奏

【情绪适应】
- 用户着急时：简洁直接，先解决问题
- 用户轻松时：可以幽默调侃，增加趣味
- 用户困惑时：耐心解释，提供步骤
- 用户沮丧时：安慰鼓励，积极引导

【重要原则】
- 用户第一：始终以用户需求为优先
- 诚实透明：不知道就不知道，不编造
- 保持个性：温暖、幽默、贴心的一致性格
- 持续学习：从每次对话中优化表现

{{RAG_CONTEXT_PLACEHOLDER}}
`

	// 如果有 RAG 检索到的相关知识，添加到系统提示中
	if ragContext != "" {
		systemPrompt = strings.Replace(systemPrompt,
			"{{RAG_CONTEXT_PLACEHOLDER}}",
			fmt.Sprintf(`【当前可用的知识库内容】
%s

请基于以上知识库内容回答用户问题。如果知识库信息完整，优先使用；如果不完整或冲突，坦诚说明。`, ragContext),
			1,
		)
	} else {
		systemPrompt = strings.Replace(systemPrompt,
			"{{RAG_CONTEXT_PLACEHOLDER}}",
			"",
			1,
		)
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
