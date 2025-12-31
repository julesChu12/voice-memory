package pipeline

import (
	"context"
	"testing"
	"voice-memory/internal/service"
)

// MockTTSService 模拟的 TTS 服务
type MockTTSService struct{}

func (m *MockTTSService) Synthesize(options service.TTSOptions) ([]byte, error) {
	return []byte("fake-audio-data"), nil
}

func TestTTSProcessor_Process(t *testing.T) {
	mockTTS := &MockTTSService{}
	proc := NewTTSProcessor(mockTTS)

	ctx := NewPipelineContext(context.Background(), "test-tts")
	ctx.LLMReply = "你好"

	cont, err := proc.Process(ctx)
	if err != nil {
		t.Fatalf("意外错误: %v", err)
	}

	if !cont {
		t.Errorf("TTS 应该继续执行")
	}

	if string(ctx.OutputAudio) != "fake-audio-data" {
		t.Errorf("音频数据错误")
	}
}
