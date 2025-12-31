package pipeline

import (
	"context"
	"voice-memory/internal/service"
)

// PipelineContext 贯穿整个流水线的上下文数据
// 它载着原始音频，经过 STT 变成文本，经过 Intent 变成意图，经过 LLM 变成回复，最后经过 TTS 变成音频。
type PipelineContext struct {
	Ctx       context.Context    // 用于控制生命周期和取消（打断信号会触发这个 Context 的 Done）
	Cancel    context.CancelFunc // 取消函数，用于手动打断
	SessionID string             // 当前会话 ID

	// 数据槽位
	InputAudio  []byte               // 输入音频原始数据
	Transcript  string               // STT 转写后的文本内容
	Intent      service.IntentResult // 意图识别结果
	LLMReply    string               // LLM 生成的文本回复内容
	OutputAudio []byte               // TTS 合成后的音频数据（可选，如果是流式播放则可能在 Processor 内部直接发送）
}

// NewPipelineContext 创建一个新的流水线上下文
func NewPipelineContext(parent context.Context, sessionID string) *PipelineContext {
	ctx, cancel := context.WithCancel(parent)
	return &PipelineContext{
		Ctx:       ctx,
		Cancel:    cancel,
		SessionID: sessionID,
	}
}

// Processor 处理器接口。每一个环节（如 STTProcessor, LLMProcessor）都要实现这个接口。
type Processor interface {
	// Name 返回处理器的名称
	Name() string

	// Process 执行具体的处理逻辑
	// ctx: 流水线上下文，可以在这里读取上一步的结果，并写入这一步的结果
	// 返回值 (bool, error): 
	//   - bool: 是否继续执行后续的处理器。返回 false 表示“短路”（例如识别到停止意图，不需要再走 LLM）
	//   - error: 执行过程中的错误
	Process(ctx *PipelineContext) (bool, error)
}
