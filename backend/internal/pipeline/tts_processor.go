package pipeline

import (
	"fmt"
	"log"
	"voice-memory/internal/service"
)

// TTSProcessor 语音合成处理器
type TTSProcessor struct {
	ttsService service.TTSService
}

func NewTTSProcessor(ttsService service.TTSService) *TTSProcessor {
	return &TTSProcessor{
		ttsService: ttsService,
	}
}

func (p *TTSProcessor) Name() string {
	return "TTS"
}

func (p *TTSProcessor) Process(ctx *PipelineContext) (bool, error) {
	if ctx.LLMReply == "" {
		return false, fmt.Errorf("llm reply is empty, nothing to synthesize")
	}

	log.Printf("[TTS] 开始合成语音 (文本长度: %d)", len(ctx.LLMReply))

	// 使用默认选项进行合成
	options := service.DefaultTTSOptions(ctx.LLMReply)
	
	audioData, err := p.ttsService.Synthesize(options)
	if err != nil {
		return false, fmt.Errorf("tts synthesis failed: %w", err)
	}

	ctx.OutputAudio = audioData
	log.Printf("[TTS] 合成完毕 (音频大小: %d bytes)", len(audioData))

	return true, nil
}
