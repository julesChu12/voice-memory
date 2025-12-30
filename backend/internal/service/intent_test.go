package service

import (
	"testing"
)

// TestNewIntentRecognizer 测试创建意图识别器
func TestNewIntentRecognizer(t *testing.T) {
	ir := NewIntentRecognizer()
	if ir == nil {
		t.Fatal("NewIntentRecognizer 返回 nil")
	}
	if ir.rules == nil {
		t.Error("rules 未初始化")
	}
}

// TestRecognizeChatIntent 测试识别普通聊天意图
func TestRecognizeChatIntent(t *testing.T) {
	ir := NewIntentRecognizer()

	tests := []struct {
		name     string
		text     string
		expected Intent
	}{
		{
			name:     "简单问候",
			text:     "你好",
			expected: IntentChat,
		},
		{
			name:     "闲聊",
			text:     "今天天气不错",
			expected: IntentChat,
		},
		{
			name:     "表情",
			text:     "哈哈",
			expected: IntentChat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ir.Recognize(tt.text)
			if result.Intent != tt.expected {
				t.Errorf("期望 %s, 得到 %s", tt.expected, result.Intent)
			}
		})
	}
}

// TestRecognizeRecordIntent 测试识别记录意图
func TestRecognizeRecordIntent(t *testing.T) {
	ir := NewIntentRecognizer()

	tests := []struct {
		name              string
		text              string
		expectIntent      Intent
		minConfidence     float64
	}{
		{
			name:          "记住命令",
			text:          "记住，我今天买了牛奶",
			expectIntent:  IntentRecord,
			minConfidence: 0.10, // 单个关键词匹配
		},
		{
			name:          "保存命令",
			text:          "帮我保存一下这个信息",
			expectIntent:  IntentRecord,
			minConfidence: 0.10,
		},
		{
			name:          "记录命令",
			text:          "记录下来，明天3点开会",
			expectIntent:  IntentRecord,
			minConfidence: 0.10,
		},
		{
			name:          "备忘命令",
			text:          "备忘：我的身份证号是123456",
			expectIntent:  IntentRecord,
			minConfidence: 0.10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ir.Recognize(tt.text)
			if result.Intent != tt.expectIntent {
				t.Errorf("期望 %s, 得到 %s", tt.expectIntent, result.Intent)
			}
			if result.Confidence < tt.minConfidence {
				t.Errorf("置信度 %.2f 低于最小值 %.2f", result.Confidence, tt.minConfidence)
			}
		})
	}
}

// TestRecognizeSearchIntent 测试识别搜索意图
func TestRecognizeSearchIntent(t *testing.T) {
	ir := NewIntentRecognizer()

	tests := []struct {
		name          string
		text          string
		expectIntent  Intent
		minConfidence float64
	}{
		{
			name:          "搜索命令",
			text:          "搜索一下我之前说的牛奶",
			expectIntent:  IntentSearch,
			minConfidence: 0.10,
		},
		{
			name:          "查找命令",
			text:          "查找我的身份证号",
			expectIntent:  IntentSearch,
			minConfidence: 0.10,
		},
		{
			name:          "有什么命令",
			text:          "我的知识库里有什么关于工作的事",
			expectIntent:  IntentSearch,
			minConfidence: 0.10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ir.Recognize(tt.text)
			if result.Intent != tt.expectIntent {
				t.Errorf("期望 %s, 得到 %s", tt.expectIntent, result.Intent)
			}
			if result.Confidence < tt.minConfidence {
				t.Errorf("置信度 %.2f 低于最小值 %.2f", result.Confidence, tt.minConfidence)
			}
		})
	}
}

// TestRecognizeQuestionIntent 测试识别问答意图
func TestRecognizeQuestionIntent(t *testing.T) {
	ir := NewIntentRecognizer()

	tests := []struct {
		name          string
		text          string
		expectIntent  Intent
		minConfidence float64
	}{
		{
			name:          "是什么问题",
			text:          "什么是人工智能",
			expectIntent:  IntentQuestion,
			minConfidence: 0.10,
		},
		{
			name:          "怎么问题",
			text:          "怎么做蛋糕",
			expectIntent:  IntentQuestion,
			minConfidence: 0.10,
		},
		{
			name:          "为什么问题",
			text:          "为什么天是蓝的",
			expectIntent:  IntentQuestion,
			minConfidence: 0.10,
		},
		{
			name:          "问号",
			text:          "这个怎么用？",
			expectIntent:  IntentQuestion,
			minConfidence: 0.10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ir.Recognize(tt.text)
			if result.Intent != tt.expectIntent {
				t.Errorf("期望 %s, 得到 %s", tt.expectIntent, result.Intent)
			}
			if result.Confidence < tt.minConfidence {
				t.Errorf("置信度 %.2f 低于最小值 %.2f", result.Confidence, tt.minConfidence)
			}
		})
	}
}

// TestRecognizeClearIntent 测试识别清空意图
func TestRecognizeClearIntent(t *testing.T) {
	ir := NewIntentRecognizer()

	tests := []struct {
		name          string
		text          string
		expectIntent  Intent
		minConfidence float64
	}{
		{
			name:          "清空命令",
			text:          "清空对话",
			expectIntent:  IntentClear,
			minConfidence: 0.3,
		},
		{
			name:          "重置命令",
			text:          "重置我们的对话",
			expectIntent:  IntentClear,
			minConfidence: 0.3,
		},
		{
			name:          "新对话命令",
			text:          "开始新对话",
			expectIntent:  IntentClear,
			minConfidence: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ir.Recognize(tt.text)
			if result.Intent != tt.expectIntent {
				t.Errorf("期望 %s, 得到 %s", tt.expectIntent, result.Intent)
			}
		})
	}
}

// TestShouldSaveToKnowledge 测试是否应该保存到知识库
func TestShouldSaveToKnowledge(t *testing.T) {
	ir := NewIntentRecognizer()

	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "记录命令",
			text:     "记住这个信息",
			expected: true,
		},
		{
			name:     "保存命令",
			text:     "帮我保存一下",
			expected: true,
		},
		{
			name:     "普通聊天",
			text:     "你好",
			expected: false,
		},
		{
			name:     "搜索命令",
			text:     "搜索信息",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ir.ShouldSaveToKnowledge(tt.text)
			if result != tt.expected {
				t.Errorf("期望 %v, 得到 %v", tt.expected, result)
			}
		})
	}
}

// TestShouldSearchKnowledge 测试是否应该搜索知识库
func TestShouldSearchKnowledge(t *testing.T) {
	ir := NewIntentRecognizer()

	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "搜索命令",
			text:     "搜索我的信息",
			expected: true,
		},
		{
			name:     "查找命令",
			text:     "查找我的笔记",
			expected: true,
		},
		{
			name:     "普通聊天",
			text:     "你好",
			expected: false,
		},
		{
			name:     "记录命令",
			text:     "记住这个",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ir.ShouldSearchKnowledge(tt.text)
			if result != tt.expected {
				t.Errorf("期望 %v, 得到 %v", tt.expected, result)
			}
		})
	}
}

// TestExtractSearchQuery 测试提取搜索查询
func TestExtractSearchQuery(t *testing.T) {
	ir := NewIntentRecognizer()

	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "简单搜索",
			text:     "搜索牛奶",
			expected: "牛奶",
		},
		{
			name:     "复杂搜索",
			text:     "帮我找一下关于工作的笔记",
			expected: "关于工作的笔记",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ir.ExtractSearchQuery(tt.text)
			if result == "" {
				t.Error("提取的搜索查询为空")
			}
		})
	}
}

// TestGetIntentDescription 测试获取意图描述
func TestGetIntentDescription(t *testing.T) {
	tests := []struct {
		intent    Intent
		expected string
	}{
		{IntentChat, "普通聊天"},
		{IntentQuestion, "知识问答"},
		{IntentRecord, "知识记录"},
		{IntentSearch, "知识搜索"},
		{IntentDelete, "删除操作"},
		{IntentClear, "清空会话"},
		{IntentUnknown, "未知意图"},
	}

	for _, tt := range tests {
		t.Run(tt.intent.String(), func(t *testing.T) {
			result := tt.intent.GetIntentDescription()
			if result != tt.expected {
				t.Errorf("期望 %s, 得到 %s", tt.expected, result)
			}
		})
	}
}

// TestIntentString 测试 Intent String 方法
func TestIntentString(t *testing.T) {
	tests := []struct {
		intent   Intent
		expected string
	}{
		{IntentChat, "chat"},
		{IntentQuestion, "question"},
		{IntentRecord, "record"},
		{IntentSearch, "search"},
		{IntentDelete, "delete"},
		{IntentClear, "clear"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.intent.String()
			if result != tt.expected {
				t.Errorf("期望 %s, 得到 %s", tt.expected, result)
			}
		})
	}
}

// BenchmarkRecognize 性能测试
func BenchmarkRecognize(b *testing.B) {
	ir := NewIntentRecognizer()
	text := "记住，我今天下午3点有个会议"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ir.Recognize(text)
	}
}
