package pipeline

import (
	"context"
	"errors"
	"testing"
)

// MockProcessor 用于测试的模拟处理器
type MockProcessor struct {
	name          string
	shouldFail    bool
	shouldShort   bool
	processedDone bool
}

func (m *MockProcessor) Name() string { return m.name }
func (m *MockProcessor) Process(ctx *PipelineContext) (bool, error) {
	if m.shouldFail {
		return false, errors.New("forced failure")
	}
	m.processedDone = true
	// 如果 shouldShort 为 true，则返回 false (不继续下一步)
	return !m.shouldShort, nil
}

func TestPipeline_Execute(t *testing.T) {
	t.Run("正常流转测试", func(t *testing.T) {
		p1 := &MockProcessor{name: "步骤1"}
		p2 := &MockProcessor{name: "步骤2"}
		pipe := NewPipeline(p1, p2)
		ctx := NewPipelineContext(context.Background(), "sess-normal")

		err := pipe.Execute(ctx)
		if err != nil {
			t.Fatalf("不应该有错误: %v", err)
		}
		if !p1.processedDone || !p2.processedDone {
			t.Errorf("所有处理器都应该被执行")
		}
	})

	t.Run("短路机制测试", func(t *testing.T) {
		p1 := &MockProcessor{name: "停止指令", shouldShort: true}
		p2 := &MockProcessor{name: "LLM回复"}
		pipe := NewPipeline(p1, p2)
		ctx := NewPipelineContext(context.Background(), "sess-short")

		err := pipe.Execute(ctx)
		if err != nil {
			t.Fatalf("不应该有错误: %v", err)
		}
		if !p1.processedDone {
			t.Errorf("P1 应该被执行")
		}
		if p2.processedDone {
			t.Errorf("P2 不应该被执行，因为 P1 触发了短路")
		}
	})

	t.Run("上下文取消(打断)测试", func(t *testing.T) {
		p1 := &MockProcessor{name: "步骤1"}
		p2 := &MockProcessor{name: "步骤2"}
		pipe := NewPipeline(p1, p2)
		ctx := NewPipelineContext(context.Background(), "sess-cancel")

		ctx.Cancel() // 手动触发打断

		err := pipe.Execute(ctx)
		if err == nil {
			t.Fatalf("应该返回 context canceled 错误")
		}
		if p1.processedDone || p2.processedDone {
			t.Errorf("打断后不应该执行任何处理器")
		}
	})
}
