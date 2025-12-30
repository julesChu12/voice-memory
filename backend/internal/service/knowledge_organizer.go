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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
