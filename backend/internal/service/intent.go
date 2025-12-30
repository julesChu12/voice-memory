package service

import (
	"fmt"
	"regexp"
	"strings"
)

// Intent 用户意图类型
type Intent string

const (
	// IntentChat 普通聊天
	IntentChat Intent = "chat"
	// IntentQuestion 知识问答
	IntentQuestion Intent = "question"
	// IntentRecord 知识记录
	IntentRecord Intent = "record"
	// IntentSearch 知识搜索
	IntentSearch Intent = "search"
	// IntentDelete 删除操作
	IntentDelete Intent = "delete"
	// IntentClear 清空会话
	IntentClear Intent = "clear"
	// IntentUnknown 未知意图
	IntentUnknown Intent = "unknown"
)

// IntentResult 意图识别结果
type IntentResult struct {
	Intent    Intent             `json:"intent"`
	Confidence float64            `json:"confidence"`
	Entities  map[string]string  `json:"entities"` // 提取的实体
}

// IntentRecognizer 意图识别器
type IntentRecognizer struct {
	// 意图关键词匹配规则
	rules map[Intent][]string
}

// NewIntentRecognizer 创建意图识别器
func NewIntentRecognizer() *IntentRecognizer {
	return &IntentRecognizer{
		rules: map[Intent][]string{
			IntentRecord: {
				"记住", "保存", "记录", "备忘", "存一下", "帮我记",
				"添加到知识", "知识库添加", "新建知识",
			},
			IntentSearch: {
				"搜索", "查找", "找一下", "查询", "有什么", " recall",
				"关于.*的", "看看.*知识",
			},
			IntentQuestion: {
				"什么是", "怎么", "如何", "为什么", "?", "？",
				"解释", "介绍", "告诉我",
			},
			IntentDelete: {
				"删除", "移除", "不要", "去掉",
			},
			IntentClear: {
				"清空", "重置", "重新开始", "新对话", "忘掉",
			},
		},
	}
}

// Recognize 识别用户意图
func (ir *IntentRecognizer) Recognize(text string) IntentResult {
	text = strings.ToLower(strings.TrimSpace(text))

	// 按优先级检查意图
	intentOrder := []Intent{
		IntentClear,
		IntentDelete,
		IntentRecord,
		IntentSearch,
		IntentQuestion,
	}

	for _, intent := range intentOrder {
		if confidence, entities := ir.matchIntent(text, intent); confidence > 0.10 {
			// 降低阈值到 0.10，允许单个关键词匹配
			return IntentResult{
				Intent:     intent,
				Confidence: confidence,
				Entities:   entities,
			}
		}
	}

	// 默认为普通聊天
	return IntentResult{
		Intent:     IntentChat,
		Confidence: 0.5,
		Entities:   make(map[string]string),
	}
}

// matchIntent 匹配特定意图
func (ir *IntentRecognizer) matchIntent(text string, intent Intent) (float64, map[string]string) {
	keywords, exists := ir.rules[intent]
	if !exists {
		return 0, nil
	}

	entities := make(map[string]string)
	matchedCount := 0

	// 检查关键词匹配
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			matchedCount++
		}

		// 正则表达式匹配（用于提取实体）
		if strings.Contains(keyword, ".*") {
			re := regexp.MustCompile(keyword)
			if matches := re.FindStringSubmatch(text); len(matches) > 0 {
				matchedCount++
				// 提取实体
				if len(matches) > 1 {
					entities["extracted"] = matches[1]
				}
			}
		}
	}

	if matchedCount == 0 {
		return 0, nil
	}

	// 计算置信度：匹配数量 / 总关键词数量，最高 1.0
	confidence := float64(matchedCount) / float64(len(keywords))
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence, entities
}

// ShouldSaveToKnowledge 是否应该保存到知识库
func (ir *IntentRecognizer) ShouldSaveToKnowledge(text string) bool {
	result := ir.Recognize(text)
	// 使用较低的阈值，因为关键词匹配本身就说明意图
	return result.Intent == IntentRecord && result.Confidence > 0.1
}

// ShouldSearchKnowledge 是否应该搜索知识库
func (ir *IntentRecognizer) ShouldSearchKnowledge(text string) bool {
	result := ir.Recognize(text)
	// 使用较低的阈值，因为关键词匹配本身就说明意图
	return result.Intent == IntentSearch && result.Confidence > 0.1
}

// ExtractSearchQuery 提取搜索查询
func (ir *IntentRecognizer) ExtractSearchQuery(text string) string {
	result := ir.Recognize(text)

	// 如果有提取的实体
	if entity, exists := result.Entities["extracted"]; exists {
		return entity
	}

	// 简单启发式：移除意图关键词
	query := text
	for _, keyword := range ir.rules[IntentSearch] {
		query = strings.ReplaceAll(query, keyword, "")
	}

	return strings.TrimSpace(query)
}

// GetIntentDescription 获取意图描述
func (i Intent) GetIntentDescription() string {
	switch i {
	case IntentChat:
		return "普通聊天"
	case IntentQuestion:
		return "知识问答"
	case IntentRecord:
		return "知识记录"
	case IntentSearch:
		return "知识搜索"
	case IntentDelete:
		return "删除操作"
	case IntentClear:
		return "清空会话"
	default:
		return "未知意图"
	}
}

// String 实现 Stringer 接口
func (i Intent) String() string {
	return string(i)
}

// DebugIntent 调试意图识别
func DebugIntent(text string) {
	ir := NewIntentRecognizer()
	result := ir.Recognize(text)
	fmt.Printf("意图识别: \"%s\"\n", text)
	fmt.Printf("  → 意图: %s (%.2f)\n", result.Intent.GetIntentDescription(), result.Confidence)
	if len(result.Entities) > 0 {
		fmt.Printf("  → 实体: %v\n", result.Entities)
	}
}
