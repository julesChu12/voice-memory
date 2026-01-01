package pipeline

import (
	"fmt"
	"log"
	"time"
	"voice-memory/internal/service"

	"github.com/google/uuid"
)

// KnowledgeProcessor 知识整理处理器
// 负责在对话结束后，异步调用 LLM 提取知识并存入数据库
type KnowledgeProcessor struct {
	organizer *service.KnowledgeOrganizer
	db        *service.Database
}

// NewKnowledgeProcessor 创建知识整理处理器
func NewKnowledgeProcessor(organizer *service.KnowledgeOrganizer, db *service.Database) *KnowledgeProcessor {
	return &KnowledgeProcessor{
		organizer: organizer,
		db:        db,
	}
}

func (p *KnowledgeProcessor) Name() string {
	return "Knowledge"
}

func (p *KnowledgeProcessor) Process(ctx *PipelineContext) (bool, error) {
	// 只有在 LLM 产生了回复，或者用户有明确输入时才进行整理
	// 如果是单纯的命令（如清空会话），通常没有 Transcript 或 LLMReply，或者已被 IntentProcessor 短路
	if ctx.Transcript == "" && ctx.LLMReply == "" {
		return true, nil
	}

	// 异步执行知识整理，不阻塞主流程
	// 注意：这里使用 context.Background() 或者是独立的 context，
	// 因为 ctx.Ctx 可能会在 WebSocket 连接断开时被取消，而我们希望知识整理能完成。
	// 但为了避免 goroutine 泄漏，最好有一个全局的 worker pool，这里简化处理直接 go func
	go func(sessionID, userText, aiText string) {
		// 1. 构建要分析的对话片段
		// 目前我们只分析当前这一轮对话，未来可以扩展为分析整个 Session 的 buffer
		contentToAnalyze := fmt.Sprintf("User: %s\nAI: %s", userText, aiText)

		log.Printf("[Knowledge] 开始整理知识 (Session: %s)...", sessionID)
		start := time.Now()

		// 2. 调用 KnowledgeOrganizer
		// TODO: 这里应该传入 context 以便超时控制
		result, err := p.organizer.Organize(contentToAnalyze)
		if err != nil {
			log.Printf("[Knowledge] 整理失败: %v", err)
			return
		}

		// 3. 转换为存储模型
		knowledge := &service.Knowledge{
			ID:           uuid.New().String(),
			Title:        result.Summary, // 暂时用摘要当标题，或者由前端生成
			Content:      contentToAnalyze,
			Summary:      result.Summary,
			KeyPoints:    result.KeyPoints,
			Entities:     result.Entities,
			Relations:    result.Relations,
			Observations: result.Observations,
			ActionItems:  result.ActionItems,
			Category:     result.Category,
			Tags:         result.Tags,
			Importance:   result.Importance,
			Sentiment:    result.Sentiment,
			Source:       "voice_chat",
			SessionID:    sessionID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// 4. 存入数据库
		if err := p.db.SaveKnowledge(knowledge); err != nil {
			log.Printf("[Knowledge] 入库失败: %v", err)
			return
		}

		duration := time.Since(start)
		log.Printf("[Knowledge] 知识整理完成并入库 (耗时: %v, ID: %s)", duration, knowledge.ID)

	}(ctx.SessionID, ctx.Transcript, ctx.LLMReply)

	return true, nil
}
