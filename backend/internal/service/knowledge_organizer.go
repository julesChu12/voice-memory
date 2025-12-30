package service

import (
	"encoding/json"
	"fmt"
)

// KnowledgeOrganizer 知识整理器
type KnowledgeOrganizer struct {
	glmClient *GLMClient
}

// NewKnowledgeOrganizer 创建知识整理器
func NewKnowledgeOrganizer(glmClient *GLMClient) *KnowledgeOrganizer {
	return &KnowledgeOrganizer{
		glmClient: glmClient,
	}
}

// Organize 自动整理知识内容
func (o *KnowledgeOrganizer) Organize(content string) (*KnowledgeOrganizeResult, error) {
	systemPrompt := `你是 Voice Memory 的知识整理助手。

【任务】分析用户输入的内容，提取结构化信息。

【输出格式】JSON：
{
  "summary": "一句话摘要（20字以内）",
  "key_points": ["关键点1", "关键点2"],
  "category": "分类（技术/生活/工作/学习/想法）",
  "tags": ["标签1", "标签2"]
}

【分类规则】
- 技术：编程、架构、工具、技术方案
- 生活：日常记录、购物、健康
- 工作：会议、项目、任务
- 学习：笔记、教程、概念
- 想法：灵感、创意、思考`

	messages := []Message{
		{
			Role:    "user",
			Content: systemPrompt + "\n\n内容：" + content,
		},
	}

	req := ChatRequest{
		Model:       "glm-4-flash", // 使用快速模型
		MaxTokens:   512,
		Messages:    messages,
		Temperature: 0.3, // 低温度，更稳定
	}

	resp, err := o.glmClient.SendMessage(req)
	if err != nil {
		return nil, fmt.Errorf("AI 整理失败: %w", err)
	}

	// 解析 JSON 响应
	reply := resp.GetReplyText()

	// 清理可能的 markdown 代码块标记
	reply = cleanMarkdownCode(reply)

	var result KnowledgeOrganizeResult
	if err := json.Unmarshal([]byte(reply), &result); err != nil {
		// 如果解析失败，返回基础整理
		return &KnowledgeOrganizeResult{
			Summary:   content[:min(len(content), 20)],
			KeyPoints: []string{},
			Category:  "想法",
			Tags:      []string{},
		}, nil
	}

	return &result, nil
}

// cleanMarkdownCode 清理 markdown 代码块标记
func cleanMarkdownCode(s string) string {
	// 去除开头的 ```json 或 ```
	if len(s) > 7 && s[0:7] == "```json" {
		s = s[7:]
	} else if len(s) > 3 && s[0:3] == "```" {
		s = s[3:]
	}

	// 去除结尾的 ```
	if len(s) > 3 && s[len(s)-3:] == "```" {
		s = s[:len(s)-3]
	}

	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateTitleFromSession 根据会话内容生成标题
func (o *KnowledgeOrganizer) GenerateTitleFromSession(session *Session) (string, error) {
	// 构建会话摘要提示
	var conversationText string
	for i, msg := range session.Messages {
		role := "用户"
		if msg.Role == "assistant" {
			role = "AI助手"
		}
		conversationText += fmt.Sprintf("%d. %s: %s\n", i+1, role, msg.Content)
	}

	systemPrompt := `你是 Voice Memory 的会话标题生成助手。

【任务】根据对话内容生成一个简洁的标题。

【要求】
1. 标题长度：5-15个字符
2. 简洁明了，一眼就能看出对话主题
3. 使用简体中文
4. 如果对话涉及具体主题（如技术、产品、事件等），标题应体现该主题

【输出格式】直接输出标题，不要任何其他文字或标点符号包裹

【对话内容】`

	messages := []Message{
		{
			Role:    "user",
			Content: systemPrompt + "\n\n" + conversationText,
		},
	}

	req := ChatRequest{
		Model:       "glm-4-flash",
		MaxTokens:   100,
		Messages:    messages,
		Temperature: 0.3,
	}

	resp, err := o.glmClient.SendMessage(req)
	if err != nil {
		return "", fmt.Errorf("AI 生成标题失败: %w", err)
	}

	title := resp.GetReplyText()
	// 清理可能的 markdown 代码块标记
	if len(title) > 3 {
		if title[0:3] == "```" {
			title = title[3:]
		}
		if len(title) > 3 && title[len(title)-3:] == "```" {
			title = title[:len(title)-3]
		}
	}
	// 去除可能的引号
	title = trimQuotes(title)

	return title, nil
}

// trimQuotes 去除字符串两端的引号
func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
