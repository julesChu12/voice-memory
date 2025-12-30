package service

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ContextConfig 上下文配置
type ContextConfig struct {
	MaxRecentMessages  int           // 保留最近多少条原始消息
	MaxTotalTokens     int           // 最大 token 数量
	SummaryThreshold   int           // 触发压缩的消息数量阈值
	SummaryMaxAge      time.Duration // 摘要最大有效期
}

// DefaultContextConfig 默认上下文配置
func DefaultContextConfig() ContextConfig {
	return ContextConfig{
		MaxRecentMessages: 6,  // 保留最近 6 条原始消息
		MaxTotalTokens:     4000, // 最大 4000 tokens
		SummaryThreshold:   10,  // 超过 10 条消息触发压缩
		SummaryMaxAge:      1 * time.Hour,
	}
}

// SessionSummary 会话摘要
type SessionSummary struct {
	Content       string    `json:"content"`        // 摘要内容
	KeyPoints     []string  `json:"key_points"`     // 关键点
	Topics        []string  `json:"topics"`         // 涉及的主题
	MessageCount  int       `json:"message_count"`  // 摘要的消息数量
	CreatedAt     time.Time `json:"created_at"`     // 创建时间
	UpdatedAt     time.Time `json:"updated_at"`     // 更新时间
}

// CompressedContext 压缩后的上下文
type CompressedContext struct {
	Summary      *SessionSummary `json:"summary,omitempty"`       // 历史摘要
	RecentMessages []Message     `json:"recent_messages"`          // 最近消息
	TotalMessages  int           `json:"total_messages"`           // 总消息数
}

// ContextCompressor 上下文压缩器
type ContextCompressor struct {
	config   ContextConfig
	glmClient *GLMClient
}

// NewContextCompressor 创建上下文压缩器
func NewContextCompressor(glmClient *GLMClient, config ContextConfig) *ContextCompressor {
	return &ContextCompressor{
		config:    config,
		glmClient: glmClient,
	}
}

// ShouldCompress 判断是否需要压缩
func (cc *ContextCompressor) ShouldCompress(messages []Message) bool {
	return len(messages) > cc.config.SummaryThreshold
}

// Compress 压缩上下文
func (cc *ContextCompressor) Compress(messages []Message) (*CompressedContext, error) {
	if len(messages) == 0 {
		return &CompressedContext{
			RecentMessages: []Message{},
			TotalMessages:  0,
		}, nil
	}

	// 不需要压缩，直接返回最近消息
	if !cc.ShouldCompress(messages) {
		return &CompressedContext{
			RecentMessages: messages,
			TotalMessages:  len(messages),
		}, nil
	}

	// 分离历史和最近消息
	splitIndex := len(messages) - cc.config.MaxRecentMessages
	if splitIndex < 0 {
		splitIndex = 0
	}

	historyMessages := messages[:splitIndex]
	recentMessages := messages[splitIndex:]

	// 生成摘要
	summary, err := cc.generateSummary(historyMessages)
	if err != nil {
		// 摘要生成失败，返回原始消息
		return &CompressedContext{
			RecentMessages: messages,
			TotalMessages:  len(messages),
		}, nil
	}

	return &CompressedContext{
		Summary:         summary,
		RecentMessages:  recentMessages,
		TotalMessages:   len(messages),
	}, nil
}

// generateSummary 生成历史摘要
func (cc *ContextCompressor) generateSummary(messages []Message) (*SessionSummary, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("没有消息可摘要")
	}

	// 构建对话文本
	var dialogText strings.Builder
	dialogText.WriteString("请总结以下对话内容（包含用户和AI的交流）：\n\n")

	for i, msg := range messages {
		role := "用户"
		if msg.Role == "assistant" {
			role = "AI助手"
		}
		dialogText.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, role, msg.Content))
	}

	dialogText.WriteString("\n请按以下格式输出：\n")
	dialogText.WriteString("1. 对话摘要（2-3句话）\n")
	dialogText.WriteString("2. 关键信息（3-5个要点）\n")
	dialogText.WriteString("3. 涉及主题（标签）")

	// 调用 GLM 生成摘要
	response, err := cc.glmClient.SendMessage(ChatRequest{
		Model:       "glm-4-flash", // 使用快速模型生成摘要
		MaxTokens:   512,
		Messages: []Message{
			{Role: "user", Content: dialogText.String()},
		},
		Temperature: 0.3, // 低温度保证稳定性
	})

	if err != nil {
		return nil, fmt.Errorf("生成摘要失败: %w", err)
	}

	// 解析摘要响应
	summaryText := response.GetReplyText()
	summary := &SessionSummary{
		Content:       summaryText,
		MessageCount:  len(messages),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 提取关键点和主题（简单实现）
	cc.extractSummaryComponents(summary)

	return summary, nil
}

// extractSummaryComponents 从摘要文本中提取结构化信息
func (cc *ContextCompressor) extractSummaryComponents(summary *SessionSummary) {
	text := summary.Content

	// 提取关键点（查找数字列表）
	keyPoints := make([]string, 0)
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 匹配 "1. xxx" 或 "- xxx" 格式
		if matched, _ := regexpMatchString(`^\d+\.\s+(.+)`, line); matched != "" {
			keyPoints = append(keyPoints, matched)
		} else if matched, _ := regexpMatchString(`^-\s+(.+)`, line); matched != "" {
			keyPoints = append(keyPoints, matched)
		}
	}
	summary.KeyPoints = keyPoints

	// 提取主题标签（简单实现）
	topics := make([]string, 0)
	if strings.Contains(text, "技术") || strings.Contains(text, "代码") {
		topics = append(topics, "技术")
	}
	if strings.Contains(text, "生活") || strings.Contains(text, "日常") {
		topics = append(topics, "生活")
	}
	if strings.Contains(text, "工作") || strings.Contains(text, "项目") {
		topics = append(topics, "工作")
	}
	summary.Topics = topics
}

// BuildMessagesForAPI 构建用于 API 调用的消息列表
func (cc *ContextCompressor) BuildMessagesForAPI(
	compressedCtx *CompressedContext,
	systemPrompt string,
	currentUserMessage string,
) []Message {
	messages := make([]Message, 0)

	// 添加系统提示（GLM-4 不支持 system 角色，使用 user 替代）
	if systemPrompt != "" {
		messages = append(messages, Message{
			Role:    "user",
			Content: systemPrompt,
		})
	}

	// 添加历史摘要（如果有）
	if compressedCtx.Summary != nil && compressedCtx.Summary.Content != "" {
		messages = append(messages, Message{
			Role:    "user",
			Content: fmt.Sprintf("【历史对话摘要】\n%s\n（共 %d 条消息已摘要）",
				compressedCtx.Summary.Content,
				compressedCtx.Summary.MessageCount),
		})
	}

	// 添加最近消息
	messages = append(messages, compressedCtx.RecentMessages...)

	// 添加当前用户消息
	messages = append(messages, Message{
		Role:    "user",
		Content: currentUserMessage,
	})

	return messages
}

// EstimateTokenCount 估算 token 数量
func (cc *ContextCompressor) EstimateTokenCount(messages []Message) int {
	totalChars := 0
	for _, msg := range messages {
		if content, ok := msg.Content.(string); ok {
			totalChars += len(content)
		}
	}
	// 粗略估算：1 token ≈ 2 字符（中文）或 4 字符（英文）
	return totalChars / 2
}

// ShouldRegenerateSummary 判断是否需要重新生成摘要
func (cc *ContextCompressor) ShouldRegenerateSummary(summary *SessionSummary) bool {
	if summary == nil {
		return false
	}
	// 摘要超过一定时间需要更新
	return time.Since(summary.UpdatedAt) > cc.config.SummaryMaxAge
}

// regexpMatchString 简单正则匹配
func regexpMatchString(pattern, text string) (string, bool) {
	// 简化实现：使用 strings 包
	// 完整实现可以使用 regexp 包
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}
