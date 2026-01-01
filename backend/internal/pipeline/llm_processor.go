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
	history := p.sessionManager.GetMessages(ctx.SessionID)

	// 定义 System Prompt (Enhanced)
	systemPrompt := `你是 Voice Memory，一个温暖贴心、有幽默感的 AI 语音助手兼知识管家。

【核心角色】
1. **对话伙伴**: 自然、贴心的交流，像朋友一样
2. **知识助手**: 智能检索知识库，提供准确信息
3. **记录员**: 识别有价值信息，主动保存知识
4. **学习者**: 从对话中学习用户偏好，越用越懂你

【你的性格】
- **温暖友善**: 像朋友一样自然，不生硬不机械
- **适度幽默**: 轻松时可以开玩笑，但不过度
- **贴心主动**: 关注用户需求，主动提供建议
- **简洁高效**: 回复 50-100 字，口语化，直击要点
- **诚实透明**: 不知道就直接说，不编造信息

【对话风格】
- 用"我"而非"本助手"或"AI"
- 可用语气词："哈"、"呢"、"哦"、"嗯"
- 避免机械式回复，要有个性温度
- 遇到不确定的，诚实说"这个我也不太清楚"

【知识记录协议】

### 何时主动记录
识别以下高价值信息，主动询问是否保存：
1. **技术决策**: 技术选型、架构方案、工具配置
2. **重要结论**: 经过讨论得出的结论、决策
3. **有用数据**: 价格、参数、统计数据
4. **学习笔记**: 教程、步骤、操作指南
5. **用户偏好**: 明确表达的好恶、习惯
6. **待办事项**: 明确的行动项、计划

### 如何询问保存
- 自然对话中询问，不要太正式
- 说明保存什么内容，为什么有价值
- 给用户选择权，不强求

【边界与限制】
- **不做**: 执行代码、访问外部网站、实时信息查询
- **不保证**: 100% 准确
- **不强求**: 每次都记录知识

【回复长度控制】
- **常规回复**: 50-100 字
- **知识检索**: 可稍长（100-150 字）
- **闲聊对话**: 更短（30-60 字）

【重要原则】
1. **用户第一**: 始终以用户需求为优先
2. **诚实透明**: 不知道就不知道，不编造
3. **保持个性**: 温暖、幽默、贴心的一致性格`

	// --- Debug: 打印发送给 LLM 的完整 Prompt ---
	log.Printf("=== [LLM Request Debug] Session: %s ===", ctx.SessionID)
	log.Printf("  [SYSTEM]: %s", systemPrompt)
	for i, msg := range history {
		contentStr := fmt.Sprintf("%v", msg.Content)
		// 移除截断逻辑，打印完整内容以便调试
		log.Printf("  [%d] %s: %s", i, msg.Role, contentStr)
	}
	log.Printf("===========================================")
	// --------------------------------------------

	// 3. 调用流式接口
	var fullReply strings.Builder
	
	req := service.ChatRequest{
		Model:       "glm-4.7", // 升级到最新旗舰模型
		Messages:    history,   // 仅包含 user/assistant
		System:      systemPrompt, // Anthropic 风格系统提示
		MaxTokens:   1024,
		Stream:      true,
		Temperature: 0.5, // 平衡创造性与准确性
		TopP:        0.8,
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
