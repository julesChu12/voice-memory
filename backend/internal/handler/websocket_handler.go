package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
		pipeline.NewTTSProcessor(h.ttsService),
	)

	// 4. 循环读取
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[WS] 连接断开: %v", err)
			break
		}

		// 处理不同类型的消息
		switch messageType {
		case websocket.BinaryMessage:
			// 收到音频包 -> 执行 Pipeline
			handleAudio(conn, pipe, sessionID, data)

		case websocket.TextMessage:
			// 收到文本指令 (如 config, interrupt)
			log.Printf("[WS] 收到文本消息: %s", string(data))
		}
	}
}

// handleAudio 处理音频输入
func handleAudio(conn *websocket.Conn, pipe *pipeline.Pipeline, sessionID string, audioData []byte) {
	// 通知客户端：收到音频，开始思考
	sendJSON(conn, "state", "processing")

	// 创建 Pipeline 上下文
	ctx := pipeline.NewPipelineContext(context.Background(), sessionID)
	ctx.InputAudio = audioData

	// 执行流水线
	if err := pipe.Execute(ctx); err != nil {
		log.Printf("[WS] Pipeline 执行错误: %v", err)
		sendJSON(conn, "error", err.Error())
		return
	}

	// 检查是否有意图短路
	if ctx.Transcript == "" {
		// 未识别到语音
		sendJSON(conn, "state", "idle")
		return
	}

	// 发送 STT 结果
	sendJSON(conn, "stt_final", ctx.Transcript)

	// 发送 TTS 音频 (如果有)
	if len(ctx.OutputAudio) > 0 {
		sendJSON(conn, "state", "speaking")
		if err := conn.WriteMessage(websocket.BinaryMessage, ctx.OutputAudio); err != nil {
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