package pipeline

import (
	"testing"
	"voice-memory/internal/service"
)

// MockIntentService 模拟的意图识别服务
type MockIntentService struct {
	Result service.IntentResult
}

func (m *MockIntentService) Recognize(text string) service.IntentResult {
	return m.Result
}

func TestIntentProcessor_Process(t *testing.T) {
	t.Run("普通聊天意图", func(t *testing.T) {
		mockIntent := &MockIntentService{
			Result: service.IntentResult{Intent: service.IntentChat},
		}
		proc := NewIntentProcessor(mockIntent)
		ctx := &PipelineContext{Transcript: "你好"}

		cont, err := proc.Process(ctx)
		if err != nil {
			t.Fatalf("意外错误: %v", err)
		}
		if !cont {
			t.Errorf("聊天意图应该继续执行(交给LLM)")
		}
	})

	t.Run("停止/清除意图", func(t *testing.T) {
		mockIntent := &MockIntentService{
			Result: service.IntentResult{Intent: service.IntentClear},
		}
		proc := NewIntentProcessor(mockIntent)
		ctx := &PipelineContext{Transcript: "清空对话"}

		cont, err := proc.Process(ctx)
		if err != nil {
			t.Fatalf("意外错误: %v", err)
		}
		if cont {
			t.Errorf("清空意图应该短路(不交给LLM)")
		}
	})
}
