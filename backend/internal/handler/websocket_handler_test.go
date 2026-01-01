package handler

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"voice-memory/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// MockSTTService 模拟 STT
type MockSTTService struct{}

func (m *MockSTTService) Recognize(req *service.RecognizeRequest) ([]string, error) {
	return []string{"hello"}, nil
}

// MockLLMService 模拟 LLM
type MockLLMService struct{}

func (m *MockLLMService) SendMessage(req service.ChatRequest) (*service.ChatResponse, error) {
	return &service.ChatResponse{
		Type: "message",
		Content: []service.Content{
			{Type: "text", Text: "organized knowledge"},
		},
	}, nil
}

func (m *MockLLMService) SendMessageStream(req service.ChatRequest, callback func(service.StreamChunk)) error {
	callback(service.StreamChunk{Delta: "world"})
	callback(service.StreamChunk{Done: true})
	return nil
}

// MockTTSService 模拟 TTS
type MockTTSService struct{}

func (m *MockTTSService) Synthesize(options service.TTSOptions) ([]byte, error) {
	return []byte{1, 2, 3, 4}, nil
}

func (m *MockTTSService) SynthesizeToFile(options service.TTSOptions) (string, error) {
	return "mock_audio.mp3", nil
}

func (m *MockTTSService) ServeAudio(filename string) ([]byte, string, error) {
	return []byte{1, 2, 3, 4}, "audio/mpeg", nil
}

// MockIntentService 模拟意图
type MockIntentService struct{}

func (m *MockIntentService) Recognize(text string) service.IntentResult {
	return service.IntentResult{Intent: service.IntentChat}
}

func setupWSServer(t *testing.T) (*httptest.Server, *service.SessionManager) {
	// 创建 Mock 服务
	stt := &MockSTTService{}
	llm := &MockLLMService{}
	tts := &MockTTSService{}
	intent := &MockIntentService{}

	// 创建临时数据库和 SessionManager
	tempDir := t.TempDir()
	db, _ := service.NewDatabase(tempDir)
	sm := service.NewSessionManagerWithDB(db)

	// 创建 KnowledgeOrganizer
	organizer := service.NewKnowledgeOrganizer(llm)

	// 创建 Handler
	wsHandler := NewWSHandler(sm, stt, llm, tts, intent, organizer, db)

	// 设置路由
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ws", wsHandler.HandleWS)

	// 启动测试服务器
	ts := httptest.NewServer(r)
	return ts, sm
}

func TestWSHandler_Connection(t *testing.T) {
	ts, _ := setupWSServer(t)
	defer ts.Close()

	// 转换 URL: http:// -> ws://
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	// 连接 WS
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("连接 WebSocket 失败: %v", err)
	}
	defer conn.Close()

	// 1. 发送模拟音频 (Binary)
	audioData := []byte{0, 0, 0, 0} // 模拟静音
	if err := conn.WriteMessage(websocket.BinaryMessage, audioData); err != nil {
		t.Fatalf("发送音频失败: %v", err)
	}

	// 2. 接收响应循环
	// 我们期望收到一系列状态更新：processing -> stt_final -> llm_reply -> idle
	receivedStates := make(map[string]bool)
	receivedLLM := false
	timeout := time.After(2 * time.Second)

	for {
		select {
		case <-timeout:
			t.Fatal("测试超时，未收到完整的状态流")
		default:
			var msg map[string]interface{}
			if err := conn.ReadJSON(&msg); err != nil {
				// 如果是二进制消息(TTS音频)，虽然现在禁用了，但如果收到也不应该报错，只是跳过
				if websocket.IsUnexpectedCloseError(err) {
					t.Fatalf("连接意外关闭: %v", err)
				}
				continue
			}

			msgType, _ := msg["type"].(string)
			if msgType == "state" {
				status, _ := msg["status"].(string)
				receivedStates[status] = true
				t.Logf("收到状态: %s", status)
			} else if msgType == "stt_final" {
				text, _ := msg["text"].(string)
				if text != "hello" {
					t.Errorf("STT 结果错误: %s", text)
				}
				t.Logf("收到 STT: %s", text)
			} else if msgType == "llm_reply" {
				text, _ := msg["text"].(string)
				if text != "world" {
					t.Errorf("LLM 结果错误: %s", text)
				}
				t.Logf("收到 LLM 回复: %s", text)
				receivedLLM = true
			}
			
			// 如果收到了 idle 且收到了 LLM 回复，说明一轮对话结束
			if receivedStates["idle"] && receivedLLM {
				return // 测试通过
			}
		}
	}
}