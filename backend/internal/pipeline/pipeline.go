package pipeline

import (
	"fmt"
	"log"
)

// Pipeline 流水线编排器，负责按顺序调度各个处理器
type Pipeline struct {
	processors []Processor
}

// NewPipeline 创建一个新的流水线，并按顺序注册处理器
func NewPipeline(processors ...Processor) *Pipeline {
	return &Pipeline{
		processors: processors,
	}
}

// Execute 顺序执行流水线中的所有处理器
func (p *Pipeline) Execute(ctx *PipelineContext) error {
	log.Printf("[Pipeline] 开始执行会话 %s 的流水线", ctx.SessionID)

	for _, proc := range p.processors {
		// 【关键：打断逻辑】在每个环节开始前，检查上下文是否已被取消
		// 如果用户在 AI 说话时插话，或者点击了停止，ctx.Ctx.Done() 就会触发
		select {
		case <-ctx.Ctx.Done():
			log.Printf("[Pipeline] 会话 %s 已被打断，停止后续执行 (当前步骤: %s)", ctx.SessionID, proc.Name())
			return ctx.Ctx.Err()
		default:
			// 正常执行当前处理器
			log.Printf("[Pipeline] [%s] 正在处理...", proc.Name())
			
			continueNext, err := proc.Process(ctx)
			if err != nil {
				log.Printf("[Pipeline] [%s] 执行失败: %v", proc.Name(), err)
				return fmt.Errorf("processor %s failed: %w", proc.Name(), err)
			}

			// 如果某个处理器决定“短路”（例如意图识别为“停止”），则提前结束
			if !continueNext {
				log.Printf("[Pipeline] [%s] 触发短路，提前结束流水线", proc.Name())
				return nil
			}
		}
	}

	log.Printf("[Pipeline] 会话 %s 执行完毕", ctx.SessionID)
	return nil
}
