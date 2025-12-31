# Voice Memory - API设计文档 (Golang版本)

**版本：** v1.0-Go
**日期：** 2025-12-29
**语言：** Golang + Gin

---

## 目录

1. [API概述](#一api概述)
2. [REST API](#二rest-api)
3. [WebSocket API](#三websocket-api)
4. [数据模型](#四数据模型)
5. [错误处理](#五错误处理)
6. [认证授权](#六认证授权)
7. [限流策略](#七限流策略)

---

## 一、API概述

### 1.1 基础信息

```
Base URL: http://localhost:8080
API Version: v1
Content-Type: application/json
```

### 1.2 API分类

| 类型 | 端点 | 用途 |
|------|------|------|
| **HTTP REST** | `/api/v1/*` | 标准API调用 |
| **WebSocket** | `/ws` | 实时双向通信 |

### 1.3 通用响应格式

```go
// 成功响应
{
    "success": true,
    "data": { ... }
}

// 错误响应
{
    "success": false,
    "error": {
        "code": "ERROR_CODE",
        "message": "Human readable message"
    }
}
```

---

## 二、REST API

### 2.1 聊天 API

#### POST /api/v1/chat

发送文本消息给AI，获取回复。

**请求**:
```json
{
    "message": "你好，请介绍一下你自己",
    "model": "claude-3-5-haiku-20241022",  // 可选
    "max_tokens": 1024,                      // 可选
    "stream": false                          // 可选
}
```

**响应**:
```json
{
    "success": true,
    "data": {
        "id": "msg_abc123",
        "role": "assistant",
        "content": "你好！我是Voice Memory，一个可打断的AI语音助手...",
        "model": "claude-3-5-haiku-20241022",
        "usage": {
            "input_tokens": 20,
            "output_tokens": 50
        }
    }
}
```

**Gin处理函数**:
```go
func (h *Handlers) HandleChat(c *gin.Context) {
    var req ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error": gin.H{
                "code":    "INVALID_REQUEST",
                "message": err.Error(),
            },
        })
        return
    }

    resp, err := h.aiClient.SendMessage(c.Request.Context(), ai.MessageRequest{
        Model:     req.Model,
        MaxTokens: req.MaxTokens,
        Messages: []ai.Message{
            {Role: "user", Content: req.Message},
        },
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error": gin.H{
                "code":    "AI_ERROR",
                "message": err.Error(),
            },
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "id":      resp.ID,
            "role":    resp.Role,
            "content": resp.Content[0].Text,
            "model":   resp.Model,
        },
    })
}
```

### 2.2 语音转文本 API

#### POST /api/v1/transcribe

将音频文件转录为文本。

**请求**:
```json
{
    "audio": "base64_encoded_wav_data",
    "language": "zh",     // 可选: "zh", "en", "ja"等
    "prompt": "Voice Memory"  // 可选: 提示词
}
```

**响应**:
```json
{
    "success": true,
    "data": {
        "text": "你好，请介绍一下你自己",
        "language": "zh",
        "duration": 2.5
    }
}
```

**Gin处理函数**:
```go
func (h *Handlers) HandleTranscribe(c *gin.Context) {
    var req TranscribeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse(err))
        return
    }

    // 解码Base64音频
    audioData, err := base64.StdEncoding.DecodeString(req.Audio)
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse(err))
        return
    }

    // 调用Whisper API
    resp, err := h.sttClient.Transcribe(c.Request.Context(), stt.TranscribeRequest{
        File:       audioData,
        Filename:   "audio.wav",
        Language:   req.Language,
        Prompt:     req.Prompt,
        Temperature: 0.0,
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse(err))
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "text": resp.Text,
        },
    })
}
```

### 2.3 文本转语音 API

#### POST /api/v1/synthesize

将文本转换为语音音频。

**请求**:
```json
{
    "text": "你好！我是Voice Memory",
    "voice": "nova",           // 可选: alloy, echo, fable, onyx, nova, shimmer
    "model": "tts-1",          // 可选: tts-1, tts-1-hd
    "response_format": "mp3"   // 可选: mp3, opus, aac
}
```

**响应**:
```json
{
    "success": true,
    "data": {
        "audio": "base64_encoded_mp3_data",
        "format": "mp3",
        "duration": 3.2
    }
}
```

**Gin处理函数**:
```go
func (h *Handlers) HandleSynthesize(c *gin.Context) {
    var req SynthesizeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse(err))
        return
    }

    // 调用TTS API
    audioData, err := h.ttsClient.Synthesize(c.Request.Context(), tts.SynthesizeRequest{
        Text:          req.Text,
        Model:         req.Model,
        Voice:         tts.Voice(req.Voice),
        ResponseFormat: req.ResponseFormat,
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse(err))
        return
    }

    // 编码为Base64
    audioBase64 := base64.StdEncoding.EncodeToString(audioData)

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "audio":  audioBase64,
            "format": req.ResponseFormat,
        },
    })
}
```

### 2.4 记忆管理 API

#### GET /api/v1/memories

获取所有记忆列表。

**查询参数**:
- `limit`: 限制返回数量 (默认: 50)
- `offset`: 偏移量 (默认: 0)

**请求**:
```
GET /api/v1/memories?limit=10&offset=0
```

**响应**:
```json
{
    "success": true,
    "data": {
        "memories": [
            {
                "id": "mem_abc123",
                "content": "用户讨论了项目架构设计...",
                "created_at": "2025-12-29T10:30:00Z",
                "updated_at": "2025-12-29T10:30:00Z"
            }
        ],
        "total": 100,
        "limit": 10,
        "offset": 0
    }
}
```

#### POST /api/v1/memories

创建新记忆。

**请求**:
```json
{
    "content": "这是一个重要的讨论内容...",
    "metadata": {
        "source": "chat",
        "conversation_id": "conv_abc123"
    }
}
```

**响应**:
```json
{
    "success": true,
    "data": {
        "id": "mem_xyz789",
        "content": "这是一个重要的讨论内容...",
        "created_at": "2025-12-29T11:00:00Z"
    }
}
```

#### GET /api/v1/memories/:id

获取单个记忆详情。

**请求**:
```
GET /api/v1/memories/mem_abc123
```

**响应**:
```json
{
    "success": true,
    "data": {
        "id": "mem_abc123",
        "content": "用户讨论了项目架构设计...",
        "created_at": "2025-12-29T10:30:00Z",
        "updated_at": "2025-12-29T10:30:00Z",
        "metadata": {
            "source": "chat",
            "tags": ["architecture", "design"]
        }
    }
}
```

#### DELETE /api/v1/memories/:id

删除记忆。

**请求**:
```
DELETE /api/v1/memories/mem_abc123
```

**响应**:
```json
{
    "success": true
}
```

### 2.5 搜索 API

#### GET /api/v1/memories/search

搜索记忆。

**查询参数**:
- `q`: 搜索关键词
- `limit`: 返回数量限制 (默认: 10)

**请求**:
```
GET /api/v1/memories/search?q=架构设计&limit=5
```

**响应**:
```json
{
    "success": true,
    "data": {
        "results": [
            {
                "id": "mem_abc123",
                "content": "用户讨论了项目架构设计...",
                "score": 0.95,
                "created_at": "2025-12-29T10:30:00Z"
            }
        ],
        "total": 15
    }
}
```

### 2.6 对话管理 API

#### GET /api/v1/conversations

获取对话列表。

**响应**:
```json
{
    "success": true,
    "data": {
        "conversations": [
            {
                "id": "conv_abc123",
                "summary": "讨论了Voice Memory的技术选型",
                "created_at": "2025-12-29T09:00:00Z",
                "message_count": 15
            }
        ]
    }
}
```

#### GET /api/v1/conversations/:id

获取对话详情（包含所有消息）。

**请求**:
```
GET /api/v1/conversations/conv_abc123
```

**响应**:
```json
{
    "success": true,
    "data": {
        "id": "conv_abc123",
        "summary": "讨论了Voice Memory的技术选型",
        "created_at": "2025-12-29T09:00:00Z",
        "messages": [
            {
                "id": "msg_001",
                "role": "user",
                "content": "你好",
                "timestamp": "2025-12-29T09:00:00Z"
            },
            {
                "id": "msg_002",
                "role": "assistant",
                "content": "你好！有什么可以帮助你的？",
                "timestamp": "2025-12-29T09:00:01Z"
            }
        ]
    }
}
```

#### POST /api/v1/conversations

创建新对话。

**请求**:
```json
{
    "summary": "技术讨论"
}
```

**响应**:
```json
{
    "success": true,
    "data": {
        "id": "conv_xyz789",
        "summary": "技术讨论",
        "created_at": "2025-12-29T12:00:00Z"
    }
}
```

---

## 三、WebSocket API

### 3.1 连接

**端点**: `/ws`

**连接示例**:
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
    console.log('WebSocket connected');
};

ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};

ws.onclose = () => {
    console.log('WebSocket closed');
};
```

### 3.2 消息格式

**客户端 → 服务器**:

```json
{
    "type": "audio",       // audio | text | ping
    "data": "base64_audio_data_or_text"
}
```

**服务器 → 客户端**:

```json
{
    "type": "audio_response",  // audio_response | text_response | error | pong
    "data": "..."
}
```

### 3.3 消息类型

#### 发送音频

**客户端发送**:
```json
{
    "type": "audio",
    "data": "base64_encoded_wav_audio"
}
```

**服务器响应**:
```json
{
    "type": "transcription",
    "data": {
        "text": "你好，请介绍一下你自己"
    }
}
```

```json
{
    "type": "audio_response",
    "data": "base64_encoded_mp3_audio"
}
```

#### 发送文本

**客户端发送**:
```json
{
    "type": "text",
    "data": "你好"
}
```

**服务器响应**:
```json
{
    "type": "text_response",
    "data": "你好！有什么可以帮助你的？"
}
```

#### 心跳

**客户端发送**:
```json
{
    "type": "ping"
}
```

**服务器响应**:
```json
{
    "type": "pong",
    "data": "2025-12-29T12:00:00Z"
}
```

### 3.4 WebSocket处理器实现

```go
// internal/api/websocket.go
package api

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // MVP阶段
    },
}

type WSMessage struct {
    Type string      `json:"type"`
    Data interface{} `json:"data"`
}

func (h *Handlers) HandleWebSocket(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }
    defer conn.Close()

    log.Println("WebSocket connected")

    for {
        var msg WSMessage
        if err := conn.ReadJSON(&msg); err != nil {
            log.Printf("Read error: %v", err)
            break
        }

        switch msg.Type {
        case "audio":
            go h.handleAudio(conn, msg)
        case "text":
            go h.handleText(conn, msg)
        case "ping":
            conn.WriteJSON(WSMessage{Type: "pong"})
        default:
            conn.WriteJSON(WSMessage{
                Type: "error",
                Data: "Unknown message type",
            })
        }
    }
}

func (h *Handlers) handleAudio(conn *websocket.Conn, msg WSMessage) {
    // 1. 解码音频
    audioStr, _ := msg.Data.(string)
    audioData, _ := base64.StdEncoding.DecodeString(audioStr)

    // 2. STT
    transcribeResp, err := h.sttClient.Transcribe(context.Background(), stt.TranscribeRequest{
        File:     audioData,
        Filename: "audio.wav",
    })

    if err != nil {
        conn.WriteJSON(WSMessage{Type: "error", Data: err.Error()})
        return
    }

    // 发送转录结果
    conn.WriteJSON(WSMessage{
        Type: "transcription",
        Data: transcribeResp.Text,
    })

    // 3. LLM
    aiResp, err := h.aiClient.SendMessage(context.Background(), ai.MessageRequest{
        Model:     "claude-3-5-haiku-20241022",
        MaxTokens: 1024,
        Messages: []ai.Message{
            {Role: "user", Content: transcribeResp.Text},
        },
    })

    if err != nil {
        conn.WriteJSON(WSMessage{Type: "error", Data: err.Error()})
        return
    }

    reply := aiResp.Content[0].Text

    // 4. TTS
    ttsResp, err := h.ttsClient.Synthesize(context.Background(), tts.SynthesizeRequest{
        Text:  reply,
        Model: "tts-1",
        Voice: tts.VoiceNova,
    })

    if err != nil {
        conn.WriteJSON(WSMessage{Type: "error", Data: err.Error()})
        return
    }

    // 发送音频响应
    audioBase64 := base64.StdEncoding.EncodeToString(ttsResp)
    conn.WriteJSON(WSMessage{
        Type: "audio_response",
        Data: audioBase64,
    })
}

func (h *Handlers) handleText(conn *websocket.Conn, msg WSMessage) {
    text, _ := msg.Data.(string)

    aiResp, err := h.aiClient.SendMessage(context.Background(), ai.MessageRequest{
        Model:     "claude-3-5-haiku-20241022",
        MaxTokens: 1024,
        Messages: []ai.Message{
            {Role: "user", Content: text},
        },
    })

    if err != nil {
        conn.WriteJSON(WSMessage{Type: "error", Data: err.Error()})
        return
    }

    conn.WriteJSON(WSMessage{
        Type: "text_response",
        Data: aiResp.Content[0].Text,
    })
}
```

---

## 四、数据模型

### 4.1 请求模型

```go
// internal/api/models.go
package api

// ChatRequest 聊天请求
type ChatRequest struct {
    Message   string `json:"message" binding:"required"`
    Model     string `json:"model"`
    MaxTokens int    `json:"max_tokens"`
    Stream    bool   `json:"stream"`
}

// TranscribeRequest 转录请求
type TranscribeRequest struct {
    Audio     string `json:"audio" binding:"required"`
    Language  string `json:"language"`
    Prompt    string `json:"prompt"`
}

// SynthesizeRequest 合成请求
type SynthesizeRequest struct {
    Text          string `json:"text" binding:"required"`
    Voice         string `json:"voice"`
    Model         string `json:"model"`
    ResponseFormat string `json:"response_format"`
}

// CreateMemoryRequest 创建记忆请求
type CreateMemoryRequest struct {
    Content  string                 `json:"content" binding:"required"`
    Metadata map[string]interface{} `json:"metadata"`
}
```

### 4.2 响应模型

```go
// ChatResponse 聊天响应
type ChatResponse struct {
    ID      string `json:"id"`
    Role    string `json:"role"`
    Content string `json:"content"`
    Model   string `json:"model"`
}

// TranscribeResponse 转录响应
type TranscribeResponse struct {
    Text     string  `json:"text"`
    Language string  `json:"language"`
    Duration float64 `json:"duration"`
}

// SynthesizeResponse 合成响应
type SynthesizeResponse struct {
    Audio    string  `json:"audio"`
    Format   string  `json:"format"`
    Duration float64 `json:"duration"`
}

// MemoryResponse 记忆响应
type MemoryResponse struct {
    ID        string                 `json:"id"`
    Content   string                 `json:"content"`
    CreatedAt string                 `json:"created_at"`
    UpdatedAt string                 `json:"updated_at"`
    Metadata  map[string]interface{} `json:"metadata"`
}
```

---

## 五、错误处理

### 5.1 错误代码

| 代码 | HTTP状态 | 说明 |
|------|---------|------|
| `INVALID_REQUEST` | 400 | 请求参数无效 |
| `UNAUTHORIZED` | 401 | 未授权 |
| `FORBIDDEN` | 403 | 禁止访问 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `AI_ERROR` | 500 | AI服务错误 |
| `STT_ERROR` | 500 | 语音识别错误 |
| `TTS_ERROR` | 500 | 语音合成错误 |
| `DATABASE_ERROR` | 500 | 数据库错误 |

### 5.2 错误响应格式

```json
{
    "success": false,
    "error": {
        "code": "AI_ERROR",
        "message": "Failed to connect to Claude API",
        "details": "connection timeout"
    }
}
```

### 5.3 错误处理函数

```go
// ErrorResponse 创建错误响应
func ErrorResponse(err error) gin.H {
    return gin.H{
        "success": false,
        "error": gin.H{
            "code":    "INTERNAL_ERROR",
            "message": err.Error(),
        },
    }
}

// ErrorCodeResponse 创建带代码的错误响应
func ErrorCodeResponse(code, message string) gin.H {
    return gin.H{
        "success": false,
        "error": gin.H{
            "code":    code,
            "message": message,
        },
    }
}
```

---

## 六、认证授权

### 6.1 API Key认证 (MVP)

**Header**:
```
Authorization: Bearer YOUR_API_KEY
```

**中间件**:
```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")

        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error": gin.H{
                    "code":    "UNAUTHORIZED",
                    "message": "Missing authorization header",
                },
            })
            return
        }

        // 验证token
        if !validateToken(token) {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error": gin.H{
                    "code":    "UNAUTHORIZED",
                    "message": "Invalid token",
                },
            })
            return
        }

        c.Next()
    }
}
```

---

## 七、限流策略

### 7.1 限流中间件

```go
// internal/api/rate_limit.go
package api

import (
    "sync"
    "time"
)

type RateLimiter struct {
    visitors map[string]*Visitor
    mu       sync.RWMutex
    rate     int           // 每分钟请求数
    burst    int           // 突发请求数
}

type Visitor struct {
    tokens    int
    lastSeen  time.Time
}

func NewRateLimiter(rate, burst int) *RateLimiter {
    rl := &RateLimiter{
        visitors: make(map[string]*Visitor),
        rate:     rate,
        burst:    burst,
    }
    go rl.cleanup()
    return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    v, exists := rl.visitors[ip]
    if !exists {
        rl.visitors[ip] = &Visitor{tokens: rl.burst - 1, lastSeen: time.Now()}
        return true
    }

    v.lastSeen = time.Now()
    if v.tokens > 0 {
        v.tokens--
        return true
    }

    return false
}

func (rl *RateLimiter) cleanup() {
    for {
        time.Sleep(time.Minute)
        rl.mu.Lock()
        for ip, v := range rl.visitors {
            if time.Since(v.lastSeen) > 3*time.Minute {
                delete(rl.visitors, ip)
            }
        }
        rl.mu.Unlock()
    }
}

// RateLimitMiddleware 限流中间件
func (h *Handlers) RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()

        if !h.rateLimiter.Allow(ip) {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "success": false,
                "error": gin.H{
                    "code":    "RATE_LIMIT_EXCEEDED",
                    "message": "Too many requests",
                },
            })
            return
        }

        c.Next()
    }
}
```

### 7.2 应用限流

```go
// 在路由中应用
api := r.Group("/api/v1")
api.Use(h.RateLimitMiddleware())
{
    api.POST("/chat", h.HandleChat)
    // ...
}
```

---

## 八、API使用示例

### 8.1 cURL示例

```bash
# 聊天
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "你好"}'

# 转录音频
curl -X POST http://localhost:8080/api/v1/transcribe \
  -H "Content-Type: application/json" \
  -d '{"audio": "base64_audio_data"}'

# 合成语音
curl -X POST http://localhost:8080/api/v1/synthesize \
  -H "Content-Type: application/json" \
  -d '{"text": "你好"}'

# 获取记忆列表
curl http://localhost:8080/api/v1/memories

# 搜索记忆
curl "http://localhost:8080/api/v1/memories/search?q=架构"
```

### 8.2 Go客户端示例

```go
package main

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
)

func main() {
    // 创建请求
    reqBody := map[string]string{"message": "你好"}
    body, _ := json.Marshal(reqBody)

    // 发送请求
    resp, err := http.Post("http://localhost:8080/api/v1/chat", "application/json", bytes.NewReader(body))
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    // 读取响应
    respBody, _ := io.ReadAll(resp.Body)

    var result map[string]interface{}
    json.Unmarshal(respBody, &result)

    println(result["data"].(map[string]interface{})["content"].(string))
}
```

---

**文档版本历史**:
- v1.0-Go (2025-12-29): 创建Golang版本API设计文档
