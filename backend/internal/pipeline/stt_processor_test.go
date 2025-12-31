package pipeline

import (
	"testing"
	"voice-memory/internal/service"
)

// MockSTTService 模拟的 STT 服务
type MockSTTService struct {
	Result []string
	Err    error
}

func (m *MockSTTService) Recognize(req *service.RecognizeRequest) ([]string, error) {
	return m.Result, m.Err
}

func TestSTTProcessor_Process(t *testing.T) {
	t.Run("正常识别", func(t *testing.T) {
		mockSTT := &MockSTTService{
			Result: []string{"你好", "世界"},
		}
		proc := NewSTTProcessor(mockSTT)
		ctx := &PipelineContext{
			InputAudio: []byte{1, 2, 3},
		}

		cont, err := proc.Process(ctx)
		if err != nil {
			t.Fatalf("意外错误: %v", err)
		}
		if !cont {
			t.Errorf("应该继续执行")
		}
		if ctx.Transcript != "你好世界" {
			t.Errorf("转写结果错误, 期望 '你好世界', 实际 '%s'", ctx.Transcript)
		}
	})

	t.Run("空音频报错", func(t *testing.T) {
		proc := NewSTTProcessor(&MockSTTService{})
		ctx := &PipelineContext{
			InputAudio: nil, // 空音频
		}
		_, err := proc.Process(ctx)
		if err == nil {
			t.Errorf("空音频应该报错")
		}
	})

	t.Run("识别结果为空则短路", func(t *testing.T) {
		mockSTT := &MockSTTService{
			Result: []string{}, // 空结果
		}
		proc := NewSTTProcessor(mockSTT)
		ctx := &PipelineContext{
			InputAudio: []byte{1, 2, 3},
		}

		cont, err := proc.Process(ctx)
		if err != nil {
			t.Fatalf("意外错误: %v", err)
		}
		if cont {
			t.Errorf("识别结果为空时应该短路(返回false)")
		}
	})
}