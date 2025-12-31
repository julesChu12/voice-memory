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

	// 1. 获取历史消息
	history := p.sessionManager.GetMessages(ctx.SessionID)

	// 2. 构造当前请求消息
	currentMsg := service.Message{
		Role:    "user",
		Content: ctx.Transcript,
	}

	// 3. 组装所有消息 (这里可以注入 System Prompt)
	// TODO: 可以在此处引入 PromptAssembler 优化
	allMessages := append(history, currentMsg)

	// 4. 调用流式接口，但在内部累加结果（目前是 Block Mode）
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

	// 5. 将对话存入会话管理器
	p.sessionManager.AddMessage(ctx.SessionID, "user", ctx.Transcript)
	p.sessionManager.AddMessage(ctx.SessionID, "assistant", reply)

	log.Printf("[LLM] 生成回复完毕 (长度: %d)", len(reply))

	return true, nil
}
