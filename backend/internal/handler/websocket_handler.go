package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"voice-memory/internal/pipeline"
	"voice-memory/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WSHandler WebSocket 处理器
type WSHandler struct {
	upgrader       websocket.Upgrader
	sessionManager *service.SessionManager
	sttService     service.STTService
	llmService     service.LLMService
	ttsService     service.TTSService
	intentService  service.IntentService
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(
	sm *service.SessionManager,
	stt service.STTService,
	llm service.LLMService,
	tts service.TTSService,
	intent service.IntentService,
) *WSHandler {
	return &WSHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许跨域
			},
		},
		sessionManager: sm,
		sttService:     stt,
		llmService:     llm,
		ttsService:     tts,
		intentService:  intent,
	}
}

// WSMessage WebSocket 消息结构
type WSMessage struct {
	Type string      `json:"type"` // "config", "interrupt"
	Data interface{} `json:"data,omitempty"`
}

// HandleWS 处理 WebSocket 连接
func (h *WSHandler) HandleWS(c *gin.Context) {
	// 1. 升级连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WS Upgrade Failed: %v", err)
		return
	}
	defer conn.Close()

	// 2. 获取/创建会话
	sessionID := c.Query("session_id")
	if sessionID == "" {
		sessionID = fmt.Sprintf("sess_%d", time.Now().Unix())
	}
	h.sessionManager.GetOrCreateSession(sessionID)

	log.Printf("[WS] 新连接建立 (Session: %s)", sessionID)

	// 3. 构建 Pipeline
	// 注意：这里我们为每个连接创建一个 Pipeline 实例
	pipe := pipeline.NewPipeline(
		pipeline.NewSTTProcessor(h.sttService),
		pipeline.NewIntentProcessor(h.intentService),
		pipeline.NewLLMProcessor(h.llmService, h.sessionManager),
		// pipeline.NewTTSProcessor(h.ttsService), // 开发阶段禁用 TTS，节省资源
	)

	// 4. 循环读取
	var (
		currentCancel context.CancelFunc
		mu            sync.Mutex
	)

	// 辅助函数：取消当前正在进行的任务
	cancelCurrent := func() {
		mu.Lock()
		defer mu.Unlock()
		if currentCancel != nil {
			currentCancel()
			currentCancel = nil
			log.Printf("[WS] 已触发打断，取消上一个任务")
		}
	}

	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[WS] 连接断开: %v", err)
			break
		}

		// 处理不同类型的消息
		switch messageType {
		case websocket.BinaryMessage:
			// 1. 收到新语音 -> 立即打断上一个
			cancelCurrent()

			// 2. 创建新上下文
			mu.Lock()
			ctx, cancel := context.WithCancel(context.Background())
			currentCancel = cancel
			mu.Unlock()

			// 3. 异步执行 Pipeline (关键修复：必须是 go routine)
			go func(ctx context.Context, audioData []byte) {
				// 确保任务结束时清理 cancel
				defer func() {
					mu.Lock()
					// 只有当 currentCancel 还是自己时才置空（避免把后来者的 cancel 搞丢）
					// 这里简单处理，依赖 GC，或者可以不置空，只要逻辑正确
					mu.Unlock()
				}()
				
				handleAudio(ctx, conn, pipe, sessionID, audioData)
			}(ctx, data)

		case websocket.TextMessage:
			// 收到文本指令
			var msg struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}
			if err := json.Unmarshal(data, &msg); err != nil {
				// 兼容旧的字符串指令 (比如直接发送 "interrupt")
				msgStr := string(data)
				if msgStr == "interrupt" {
					msg.Type = "interrupt"
				} else {
					log.Printf("[WS] 无法解析文本消息: %v", err)
					continue
				}
			}

			switch msg.Type {
			case "interrupt":
				cancelCurrent()
				sendJSON(conn, "state", "idle")
			case "text":
				// 处理纯文本输入
				cancelCurrent()
				mu.Lock()
				ctx, cancel := context.WithCancel(context.Background())
				currentCancel = cancel
				mu.Unlock()
				go handleText(ctx, conn, pipe, sessionID, msg.Text, h.sessionManager)
			}
			log.Printf("[WS] 收到指令: %+v", msg)
		}
	}
}

// handleText 处理纯文本输入
func handleText(ctx context.Context, conn *websocket.Conn, pipe *pipeline.Pipeline, sessionID, text string, sm *service.SessionManager) {
	sendJSON(conn, "state", "processing")

	pCtx := pipeline.NewPipelineContext(ctx, sessionID)
	pCtx.Transcript = text // 直接设置文本，跳过 STT

	// 我们需要一个不含 STT 的 Pipeline，或者让 STTProcessor 发现有文本时自动跳过
	// 为了简单起见，我们直接执行现有的 pipe，并在执行前确保 Transcript 已存在
	// STTProcessor 应该在有文本时返回 true (继续) 且不做任何事
	
	if err := pipe.Execute(pCtx); err != nil {
		if ctx.Err() == context.Canceled {
			return
		}
		sendJSON(conn, "error", err.Error())
		return
	}

	// 发送 LLM 回复
	if pCtx.LLMReply != "" {
		log.Printf("[WS] AI 回复 (Session: %s): %s", sessionID, pCtx.LLMReply)
		sendJSON(conn, "llm_reply", pCtx.LLMReply)
	}

	sendJSON(conn, "state", "idle")
}

// handleAudio 处理音频输入
func handleAudio(ctx context.Context, conn *websocket.Conn, pipe *pipeline.Pipeline, sessionID string, audioData []byte) {
	// 通知客户端：收到音频，开始思考
	sendJSON(conn, "state", "processing")

	// 创建 Pipeline 上下文 (使用传入的可取消 Context)
	pCtx := pipeline.NewPipelineContext(ctx, sessionID)
	pCtx.InputAudio = audioData

	// 执行流水线
	if err := pipe.Execute(pCtx); err != nil {
		// 如果是 context canceled，说明是正常打断，不需要报错
		if ctx.Err() == context.Canceled {
			log.Printf("[WS] Pipeline 被打断")
			return
		}
		log.Printf("[WS] Pipeline 执行错误: %v", err)
		sendJSON(conn, "error", err.Error())
		return
	}

	// 检查是否有意图短路或被取消
	if pCtx.Transcript == "" || ctx.Err() != nil {
		sendJSON(conn, "state", "idle")
		return
	}

	// 发送 STT 结果
	log.Printf("[WS] 用户输入 (Session: %s): %s", sessionID, pCtx.Transcript)
	sendJSON(conn, "stt_final", pCtx.Transcript)

	// 发送 LLM 回复文本 (替代 TTS)
	if pCtx.LLMReply != "" {
		log.Printf("[WS] AI 回复 (Session: %s): %s", sessionID, pCtx.LLMReply)
		sendJSON(conn, "llm_reply", pCtx.LLMReply)
	}

	// 发送 TTS 音频 (如果有)
	if len(pCtx.OutputAudio) > 0 {
		// 在发送音频前再次检查是否被打断
		if ctx.Err() != nil {
			return
		}
		sendJSON(conn, "state", "speaking")
		if err := conn.WriteMessage(websocket.BinaryMessage, pCtx.OutputAudio); err != nil {
			log.Printf("[WS] 发送音频失败: %v", err)
		}
	}

	sendJSON(conn, "state", "idle")
}

// sendJSON 发送 JSON 消息辅助函数
func sendJSON(conn *websocket.Conn, msgType string, payload interface{}) {
	msg := map[string]interface{}{
		"type": msgType,
	}
	if str, ok := payload.(string); ok {
		// 如果 payload 是字符串，放入 text 或 status 字段
		if msgType == "state" {
			msg["status"] = str
		} else if msgType == "error" {
			msg["error"] = str
		} else {
			msg["text"] = str
		}
	} else {
		msg["data"] = payload
	}

	conn.WriteJSON(msg)
}