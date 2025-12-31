package pipeline

import (
	"context"
	"testing"
	"voice-memory/internal/service"
)

// MockLLMService 模拟的 LLM 服务
type MockLLMService struct {
	Reply string
}

func (m *MockLLMService) SendMessageStream(req service.ChatRequest, callback func(service.StreamChunk)) error {
	// 模拟流式发送几个词
	words := []string{"你好", "，我是", "AI", "助手"}
	for _, word := range words {
		callback(service.StreamChunk{Delta: word})
	}
	callback(service.StreamChunk{Done: true})
	return nil
}

func TestLLMProcessor_Process(t *testing.T) {
	mockLLM := &MockLLMService{}
	sm := service.NewSessionManager()
	proc := NewLLMProcessor(mockLLM, sm)

	sessionID := "test-session"
	// 【关键修复】确保会话已存在
	sm.GetOrCreateSession(sessionID)

	ctx := NewPipelineContext(context.Background(), sessionID)
	ctx.Transcript = "你好"

	cont, err := proc.Process(ctx)
	if err != nil {
		t.Fatalf("意外错误: %v", err)
	}

	if !cont {
		t.Errorf("LLM 应该继续执行")
	}

	expected := "你好，我是AI助手"
	if ctx.LLMReply != expected {
		t.Errorf("回复内容错误, 期望 '%s', 实际 '%s'", expected, ctx.LLMReply)
	}

	// 验证 SessionManager 是否存入了消息
	msgs := sm.GetMessages(sessionID)
	if len(msgs) != 2 {
		t.Errorf("SessionManager 应该存入 2 条消息，实际 %d", len(msgs))
	}
}