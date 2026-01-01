package pipeline

import (
	"fmt"
	"log"
	"strings"
	"voice-memory/internal/service"
)

// LLMProcessor 大模型处理器
type LLMProcessor struct {
	llmService     service.LLMService
	sessionManager *service.SessionManager
}

func NewLLMProcessor(llmService service.LLMService, sessionManager *service.SessionManager) *LLMProcessor {
	return &LLMProcessor{
		llmService:     llmService,
		sessionManager: sessionManager,
	}
}

func (p *LLMProcessor) Name() string {
	return "LLM"
}

func (p *LLMProcessor) Process(ctx *PipelineContext) (bool, error) {
	if ctx.Transcript == "" {
		return false, fmt.Errorf("transcript is empty, nothing to ask LLM")
	}

	// 1. 立即保存用户消息 (防止 LLM 失败导致数据丢失)
	p.sessionManager.AddMessage(ctx.SessionID, "user", ctx.Transcript)

	// 2. 获取包含最新消息的历史记录
	// 这里获取到的 messages 已经包含了刚刚存入的 user message
	allMessages := p.sessionManager.GetMessages(ctx.SessionID)

	// 3. 调用流式接口
	var fullReply strings.Builder
	
	req := service.ChatRequest{
		Model:     "glm-4-plus", // 默认模型
		Messages:  allMessages,
		MaxTokens: 1024,
		Stream:    true,
	}

	log.Printf("[LLM] 开始请求 LLM (Session: %s)", ctx.SessionID)

	err := p.llmService.SendMessageStream(req, func(chunk service.StreamChunk) {
		if chunk.Error != "" {
			log.Printf("[LLM] 流式响应出错: %s", chunk.Error)
			return
		}
		if chunk.Delta != "" {
			fullReply.WriteString(chunk.Delta)
			// 未来这里可以触发“句子级” TTS 合成
		}
	})

	if err != nil {
		return false, fmt.Errorf("llm request failed: %w", err)
	}

	reply := fullReply.String()
	ctx.LLMReply = reply

	// 4. 将 AI 回复存入会话管理器
	p.sessionManager.AddMessage(ctx.SessionID, "assistant", reply)

	log.Printf("[LLM] 生成回复完毕 (长度: %d)", len(reply))

	return true, nil
}
