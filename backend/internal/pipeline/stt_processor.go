package pipeline

import (
	"fmt"
	"strings"
	"voice-memory/internal/service"
)

// STTProcessor 语音转文字处理器
// 它是通用的，不绑定任何特定厂商（百度/OpenAI/Whisper）
type STTProcessor struct {
	sttService service.STTService
}

// NewSTTProcessor 创建 STT 处理器
// 依赖注入：传入任何实现了 STTService 接口的实例
func NewSTTProcessor(sttService service.STTService) *STTProcessor {
	return &STTProcessor{
		sttService: sttService,
	}
}

func (p *STTProcessor) Name() string {
	return "STT"
}

func (p *STTProcessor) Process(ctx *PipelineContext) (bool, error) {
	// 如果已经有文本了（例如纯文本输入场景），直接跳过 STT
	if ctx.Transcript != "" {
		return true, nil
	}

	if len(ctx.InputAudio) == 0 {
		return false, fmt.Errorf("input audio is empty")
	}

	// 调用通用 STT 接口
	results, err := p.sttService.Recognize(&service.RecognizeRequest{
		AudioData: ctx.InputAudio,
		Format:    "wav", // 未来这里可以从 ctx 中获取格式信息
		Rate:      16000,
	})

	if err != nil {
		return false, fmt.Errorf("stt service failed: %w", err)
	}

	if len(results) > 0 {
		ctx.Transcript = strings.Join(results, "")
	}

	// 如果转写结果为空，短路
	if ctx.Transcript == "" {
		return false, nil
	}

	return true, nil
}