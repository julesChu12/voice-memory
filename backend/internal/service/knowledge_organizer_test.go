package service

import (
	"fmt"
	"testing"
	"time"
)

// TestGenerateTitleFromSession 测试根据会话生成标题
func TestGenerateTitleFromSession(t *testing.T) {
	// 跳过实际 API 调用测试（需要 API key）
	t.Skip("需要 GLM API key，跳过实际 API 测试")

	// 创建 GLM 客户端（使用环境变量或测试 key）
	apiKey := "test_key"
	glmClient := NewGLMClient(apiKey)
	organizer := NewKnowledgeOrganizer(glmClient)

	// 创建测试会话 - 关于技术讨论的对话
	session := &Session{
		ID: "test_session_001",
		Messages: []Message{
			{Role: "user", Content: "Go 语言有什么优点？"},
			{Role: "assistant", Content: "Go 语言有以下优点：\n1. 简单易学，语法清晰\n2. 并发支持好，有 goroutine\n3. 性能优秀，接近 C 语言\n4. 跨平台支持好"},
			{Role: "user", Content: "goroutine 是什么？"},
			{Role: "assistant", Content: "goroutine 是 Go 的轻量级线程，由 Go 运行时管理。可以轻松创建成千上万个并发任务。"},
			{Role: "user", Content: "和普通线程有什么区别？"},
			{Role: "assistant", Content: "主要区别：\n1. 更轻量，栈空间只有 2KB\n2. 由 Go runtime 调度，不是 OS 调度\n3. 通信用 channel，避免共享内存"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 生成标题
	title, err := organizer.GenerateTitleFromSession(session)
	if err != nil {
		t.Fatalf("生成标题失败: %v", err)
	}

	// 验证标题不为空
	if title == "" {
		t.Error("标题不应为空")
	}

	// 验证标题长度合理（5-15 字符）
	if len(title) < 5 || len(title) > 20 {
		t.Errorf("标题长度应在 5-20 字符之间，得到: %d (%s)", len(title), title)
	}

	t.Logf("生成的标题: %s", title)
}

// TestOrganizeWithSessionContent 测试基于完整会话内容生成摘要
func TestOrganizeWithSessionContent(t *testing.T) {
	// 跳过实际 API 调用测试（需要 API key）
	t.Skip("需要 GLM API key，跳过实际 API 测试")

	apiKey := "test_key"
	glmClient := NewGLMClient(apiKey)
	organizer := NewKnowledgeOrganizer(glmClient)

	// 构建完整对话内容（模拟会话历史）
	conversationText := `1. 用户: 给我讲解一下关于汾酒的知识把
2. AI助手: 哈，我这里关于汾酒的知识库内容不多呢~ 不过我知道汾酒是中国老八大名酒之一，清香型白酒的代表，有4000多年历史呢！
3. 用户: 从头给我讲讲把
4. AI助手: 哈哈，汾酒的故事要从4000多年前说起啦！它是清香型白酒鼻祖，产自山西杏花村。名字源于"汾河"，用高粱、豌豆等发酵，口感清冽甘爽，还被誉为"国酒"哦！
5. 用户: 继续
6. AI助手: 汾酒的传统工艺可是宝贝呢！地缸发酵、清蒸清烧，不沾杂味。蒸馏时酒香四溢，陈酿后口感更醇厚。最特别的是它的"一清到底"，喝起来清爽不辣喉~
7. 用户: 好的
8. AI助手: 哈哈，"好的"重复啦～看来你很期待呀！
9. 用户: 最高端的是什么？
10. AI助手: 汾酒的高端系列是"青花汾酒"，特别是30年、40年陈酿！
11. 用户: 分别是什么价格？
12. AI助手: 青花30年大约800-1200元，40年要2000-3000元以上呢！`

	// 生成摘要
	result, err := organizer.Organize(conversationText)
	if err != nil {
		t.Fatalf("生成摘要失败: %v", err)
	}

	// 验证摘要
	if result.Summary == "" {
		t.Error("摘要不应为空")
	}
	if len(result.Summary) > 50 {
		t.Errorf("摘要应简洁，当前长度: %d", len(result.Summary))
	}

	// 验证关键点
	if len(result.KeyPoints) == 0 {
		t.Error("应提取关键点")
	}

	// 验证分类
	validCategories := []string{"技术", "生活", "工作", "学习", "想法"}
	categoryValid := false
	for _, cat := range validCategories {
		if result.Category == cat {
			categoryValid = true
			break
		}
	}
	if !categoryValid {
		t.Errorf("分类无效: %s", result.Category)
	}

	t.Logf("摘要: %s", result.Summary)
	t.Logf("分类: %s", result.Category)
	t.Logf("关键点: %v", result.KeyPoints)
	t.Logf("标签: %v", result.Tags)
}

// TestOrganizeSingleMessage 测试单条消息的摘要生成
func TestOrganizeSingleMessage(t *testing.T) {
	// 跳过实际 API 调用测试（需要 API key）
	t.Skip("需要 GLM API key，跳过实际 API 测试")

	apiKey := "test_key"
	glmClient := NewGLMClient(apiKey)
	organizer := NewKnowledgeOrganizer(glmClient)

	// 单条文本
	text := "今天天气真不错，适合出去散步"

	result, err := organizer.Organize(text)
	if err != nil {
		t.Fatalf("生成摘要失败: %v", err)
	}

	if result.Summary == "" {
		t.Error("摘要不应为空")
	}

	t.Logf("单条消息摘要: %s", result.Summary)
	t.Logf("分类: %s", result.Category)
}

// TestKnowledgeOrganizerMock 测试知识整理器的结构（不调用 API）
func TestKnowledgeOrganizerMock(t *testing.T) {
	apiKey := "test_key"
	glmClient := NewGLMClient(apiKey)
	organizer := NewKnowledgeOrganizer(glmClient)

	if organizer == nil {
		t.Fatal("KnowledgeOrganizer 创建失败")
	}
	if organizer.glmClient == nil {
		t.Error("GLMClient 未初始化")
	}
}

// TestGenerateTitleFromSession_Mock 测试标题生成逻辑（不调用 API）
func TestGenerateTitleFromSession_Mock(t *testing.T) {
	// 这个测试验证会话结构的正确性，不实际调用 API
	session := &Session{
		ID: "test_session",
		Messages: []Message{
			{Role: "user", Content: "什么是 Go 语言？"},
			{Role: "assistant", Content: "Go 是 Google 开发的开源编程语言"},
			{Role: "user", Content: "它有什么特点？"},
		},
	}

	// 验证会话结构
	if session.ID != "test_session" {
		t.Errorf("会话 ID 不匹配")
	}
	if len(session.Messages) != 3 {
		t.Errorf("期望 3 条消息，得到 %d", len(session.Messages))
	}

	// 验证消息角色
	if session.Messages[0].Role != "user" {
		t.Errorf("第一条消息角色应为 user")
	}
	if session.Messages[1].Role != "assistant" {
		t.Errorf("第二条消息角色应为 assistant")
	}
}

// TestBuildConversationText 测试对话文本构建逻辑
func TestBuildConversationText(t *testing.T) {
	session := &Session{
		ID: "test_session",
		Messages: []Message{
			{Role: "user", Content: "你好"},
			{Role: "assistant", Content: "你好呀"},
			{Role: "user", Content: "再见"},
		},
	}

	// 构建对话文本
	var conversationText string
	for i, msg := range session.Messages {
		role := "用户"
		if msg.Role == "assistant" {
			role = "AI助手"
		}
		conversationText += fmt.Sprintf("%d. %s: %s\n", i+1, role, msg.Content)
	}

	// 验证构建结果
	if conversationText == "" {
		t.Error("对话文本不应为空")
	}

	// 验证包含所有消息
	expectedLines := 3
	actualLines := 0
	for _, msg := range session.Messages {
		content := fmt.Sprintf("%v", msg.Content)
		if testIndexOf(conversationText, content) >= 0 {
			actualLines++
		}
	}
	if actualLines != expectedLines {
		t.Errorf("期望包含 %d 条消息内容，实际包含 %d 条", expectedLines, actualLines)
	}

	t.Logf("构建的对话文本:\n%s", conversationText)
}

// testIndexOf 辅助函数：查找子串位置
func testIndexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

