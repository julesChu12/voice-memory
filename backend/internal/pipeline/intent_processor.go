package pipeline

import (
	"log"
	"voice-memory/internal/service"
)

// IntentProcessor 意图识别处理器
// 负责分析用户文本意图，并决定是否触发特定指令或继续对话
type IntentProcessor struct {
	intentService service.IntentService
}

func NewIntentProcessor(intentService service.IntentService) *IntentProcessor {
	return &IntentProcessor{
		intentService: intentService,
	}
}

func (p *IntentProcessor) Name() string {
	return "Intent"
}

func (p *IntentProcessor) Process(ctx *PipelineContext) (bool, error) {
	if ctx.Transcript == "" {
		return true, nil // 没有文本，默认继续（虽然 STT 层应该已经过滤了）
	}

	// 识别意图
	result := p.intentService.Recognize(ctx.Transcript)
	ctx.Intent = result

	log.Printf("[Intent] 识别结果: %s (置信度: %.2f)", result.Intent, result.Confidence)

	// 处理特定意图
	switch result.Intent {
	case service.IntentDelete, service.IntentClear:
		// 这些是“指令型”意图
		log.Printf("[Intent] 触发指令: %s，停止后续 LLM 生成", result.Intent)
		
		// 我们可以在这里生成一个简单的系统反馈语音（可选）
		// ctx.LLMReply = "好的，已执行" 
		// 但为了 MVP 纯粹性，我们先直接短路
		return false, nil

	case service.IntentChat, service.IntentQuestion:
		// 继续执行，交给 LLM
		return true, nil
		
	case service.IntentSearch, service.IntentRecord:
		// 知识库相关，也继续执行，交给 Memory/LLM 层处理
		return true, nil
	}

	return true, nil
}