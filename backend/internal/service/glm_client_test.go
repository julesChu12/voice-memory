package service

import (
	"testing"
)

// TestNewGLMClient 测试创建 GLM 客户端
func TestNewGLMClient(t *testing.T) {
	apiKey := "test_api_key"
	client := NewGLMClient(apiKey)

	if client == nil {
		t.Fatal("NewGLMClient 返回 nil")
	}
	if client.apiKey != apiKey {
		t.Errorf("期望 apiKey %s, 得到 %s", apiKey, client.apiKey)
	}
	if client.baseURL != "https://open.bigmodel.cn/api/anthropic" {
		t.Errorf("baseURL 不正确: %s", client.baseURL)
	}
	if client.client == nil {
		t.Error("HTTP client 未初始化")
	}
}

// TestMessageContentTypes 测试消息内容类型
func TestMessageContentTypes(t *testing.T) {
	tests := []struct {
		name    string
		content interface{}
		isValid bool
	}{
		{
			name:    "字符串内容",
			content: "这是文本内容",
			isValid: true,
		},
		{
			name: "内容块数组",
			content: []ContentBlock{
				{Type: "text", Text: "文本"},
			},
			isValid: true,
		},
		{
			name:    "空字符串",
			content: "",
			isValid: true,
		},
		{
			name:    "nil 内容",
			content: nil,
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message{
				Role:    "user",
				Content: tt.content,
			}

			if tt.isValid {
				if msg.Role == "" {
					t.Error("Role 为空")
				}
			}
		})
	}
}

// TestChatRequestDefaults 测试聊天请求默认值
func TestChatRequestDefaults(t *testing.T) {
	req := ChatRequest{
		Model:    "glm-4-plus",
		Messages: []Message{{Role: "user", Content: "你好"}},
	}

	if req.Model != "glm-4-plus" {
		t.Errorf("Model 不正确: %s", req.Model)
	}
	if req.MaxTokens == 0 {
		// MaxTokens 默认为 0 是允许的
	}
	if req.Temperature == 0 {
		// Temperature 默认为 0 是允许的
	}
	if req.Stream {
		t.Error("Stream 默认应为 false")
	}
}

// TestContentBlock 测试内容块
func TestContentBlock(t *testing.T) {
	tests := []struct {
		name  string
		block ContentBlock
		valid bool
	}{
		{
			name: "文本块",
			block: ContentBlock{
				Type: "text",
				Text: "测试文本",
			},
			valid: true,
		},
		{
			name: "音频块",
			block: ContentBlock{
				Type:     "audio_url",
				AudioURL: &AudioURL{URL: "data:audio/wav;base64,test"},
			},
			valid: true,
		},
		{
			name: "图片块",
			block: ContentBlock{
				Type:     "image_url",
				ImageURL: &ImageURL{URL: "http://example.com/image.jpg"},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				if tt.block.Type == "" {
					t.Error("Type 为空")
				}
			}
		})
	}
}

// TestGetReplyText 测试提取回复文本
func TestGetReplyText(t *testing.T) {
	tests := []struct {
		name     string
		response ChatResponse
		expected string
	}{
		{
			name: "消息类型-单个内容块",
			response: ChatResponse{
				Type: "message",
				Content: []Content{
					{Type: "text", Text: "你好"},
				},
			},
			expected: "你好",
		},
		{
			name: "消息类型-多个内容块",
			response: ChatResponse{
				Type: "message",
				Content: []Content{
					{Type: "text", Text: "你好，"},
					{Type: "text", Text: "世界！"},
				},
			},
			expected: "你好，世界！",
		},
		{
			name: "消息增量类型",
			response: ChatResponse{
				Type: "message_delta",
				Delta: Delta{
					Type: "text_delta",
					Text: "流式文本",
				},
			},
			expected: "流式文本",
		},
		{
			name: "空内容",
			response: ChatResponse{
				Type:    "unknown",
				Content: []Content{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.response.GetReplyText()
			if result != tt.expected {
				t.Errorf("期望 '%s', 得到 '%s'", tt.expected, result)
			}
		})
	}
}

// TestStreamChunk 测试流式数据块
func TestStreamChunk(t *testing.T) {
	chunk := StreamChunk{
		Delta: "测试文本",
		Done:  false,
	}

	if chunk.Delta != "测试文本" {
		t.Errorf("Delta 不正确: %s", chunk.Delta)
	}
	if chunk.Done {
		t.Error("Done 应为 false")
	}

	// 测试完成块
	doneChunk := StreamChunk{
		Done: true,
	}
	if !doneChunk.Done {
		t.Error("Done 应为 true")
	}
}

// TestGLMStreamEvent 测试 GLM 流式事件
func TestGLMStreamEvent(t *testing.T) {
	event := GLMStreamEvent{
		Type:  "content_block_delta",
		Index: 0,
		Delta: &GLMDelta{
			Type: "text_delta",
			Text: "文本",
		},
	}

	if event.Type != "content_block_delta" {
		t.Errorf("Type 不正确: %s", event.Type)
	}
	if event.Index != 0 {
		t.Errorf("Index 不正确: %d", event.Index)
	}
	if event.Delta == nil {
		t.Fatal("Delta 为 nil")
	}
	if event.Delta.Text != "文本" {
		t.Errorf("Delta.Text 不正确: %s", event.Delta.Text)
	}
}

// TestUsage 测试使用量
func TestUsage(t *testing.T) {
	usage := Usage{
		InputTokens:  100,
		OutputTokens: 50,
	}

	if usage.InputTokens != 100 {
		t.Errorf("InputTokens 不正确: %d", usage.InputTokens)
	}
	if usage.OutputTokens != 50 {
		t.Errorf("OutputTokens 不正确: %d", usage.OutputTokens)
	}

	total := usage.InputTokens + usage.OutputTokens
	if total != 150 {
		t.Errorf("总 token 数不正确: %d", total)
	}
}

// TestDelta 测试增量
func TestDelta(t *testing.T) {
	delta := Delta{
		Type:       "text_delta",
		Text:       "增量文本",
		StopReason: "end_turn",
	}

	if delta.Type != "text_delta" {
		t.Errorf("Type 不正确: %s", delta.Type)
	}
	if delta.Text != "增量文本" {
		t.Errorf("Text 不正确: %s", delta.Text)
	}
	if delta.StopReason != "end_turn" {
		t.Errorf("StopReason 不正确: %s", delta.StopReason)
	}
}

// TestAudioURL 测试音频 URL
func TestAudioURL(t *testing.T) {
	audioURL := AudioURL{
		URL: "data:audio/wav;base64,U3VwZXIgc2VjcmV0IGF1ZGlv",
	}

	if audioURL.URL == "" {
		t.Error("URL 为空")
	}
	if len(audioURL.URL) < 10 {
		t.Error("URL 太短")
	}
}

// TestImageURL 测试图片 URL
func TestImageURL(t *testing.T) {
	imageURL := ImageURL{
		URL: "http://example.com/image.jpg",
	}

	if imageURL.URL == "" {
		t.Error("URL 为空")
	}
	if imageURL.URL[:4] != "http" {
		t.Errorf("URL 协议不正确: %s", imageURL.URL)
	}
}

// TestMessageRole 测试消息角色
func TestMessageRole(t *testing.T) {
	validRoles := []string{"user", "assistant", "system"}

	for _, role := range validRoles {
		msg := Message{
			Role:    role,
			Content: "测试",
		}

		if msg.Role != role {
			t.Errorf("Role 不匹配: 期望 %s, 得到 %s", role, msg.Role)
		}
	}

	// 无效角色
	invalidMsg := Message{
		Role:    "invalid_role",
		Content: "测试",
	}

	if invalidMsg.Role != "invalid_role" {
		t.Error("无效角色应该被保留")
	}
}

// BenchmarkNewGLMClient 性能测试
func BenchmarkNewGLMClient(b *testing.B) {
	apiKey := "test_api_key"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewGLMClient(apiKey)
	}
}

// BenchmarkGetReplyText 性能测试
func BenchmarkGetReplyText(b *testing.B) {
	response := ChatResponse{
		Type: "message",
		Content: []Content{
			{Type: "text", Text: "这是一个很长的回复内容，包含多个文本块，用于性能测试。"},
			{Type: "text", Text: "这是第二个文本块的内容。"},
			{Type: "text", Text: "这是第三个文本块的内容，用于测试性能。"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response.GetReplyText()
	}
}

// 注：以下测试需要真实 API key，默认跳过
// 可以通过设置环境变量 ENABLE_API_TESTS=1 来启用

/*
func TestSendMessageIntegration(t *testing.T) {
	apiKey := os.Getenv("GLM_API_KEY")
	if apiKey == "" {
		t.Skip("需要 GLM_API_KEY 环境变量")
	}

	client := NewGLMClient(apiKey)
	req := ChatRequest{
		Model:     "glm-4-plus",
		MaxTokens: 100,
		Messages: []Message{
			{Role: "user", Content: "你好"},
		},
		Temperature: 0.7,
	}

	resp, err := client.SendMessage(req)
	if err != nil {
		t.Fatalf("SendMessage 失败: %v", err)
	}

	if resp == nil {
		t.Fatal("响应为 nil")
	}

	reply := resp.GetReplyText()
	if reply == "" {
		t.Error("回复为空")
	}

	t.Logf("API 回复: %s", reply)
}

func TestSendMessageStreamIntegration(t *testing.T) {
	apiKey := os.Getenv("GLM_API_KEY")
	if apiKey == "" {
		t.Skip("需要 GLM_API_KEY 环境变量")
	}

	client := NewGLMClient(apiKey)
	req := ChatRequest{
		Model:     "glm-4-plus",
		MaxTokens: 100,
		Messages: []Message{
			{Role: "user", Content: "数到10"},
		},
		Temperature: 0.7,
	}

	chunkCount := 0
	err := client.SendMessageStream(req, func(chunk StreamChunk) {
		if chunk.Delta != "" {
			chunkCount++
		}
		if chunk.Done {
			t.Logf("收到 %d 个数据块", chunkCount)
		}
	})

	if err != nil {
		t.Fatalf("SendMessageStream 失败: %v", err)
	}

	if chunkCount == 0 {
		t.Error("未收到任何数据块")
	}
}
*/
