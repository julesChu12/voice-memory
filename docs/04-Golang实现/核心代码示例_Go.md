# Voice Memory - 核心代码示例与使用说明 (Golang版本)

**版本：** v1.0-Go
**日期：** 2025-12-29
**语言：** Golang

---

## 目录

1. [Claude API 客户端](#一claude-api-客户端)
2. [OpenAI Whisper STT](#二openai-whisper-stt)
3. [OpenAI TTS](#三openai-tts)
4. [Gin Web 框架](#四gin-web-框架)
5. [WebSocket 实时通信](#五websocket-实时通信)
6. [SQLite 数据存储](#六sqlite-数据存储)
7. [向量搜索 (sqlite-vss)](#七向量搜索-sqlite-vss)
8. [音频处理](#八音频处理)
9. [完整集成示例](#九完整集成示例)

---

## 一、Claude API 客户端

### 1.1 基础消息发送

```go
// internal/ai/claude.go
package ai

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// ClaudeClient Claude API客户端
type ClaudeClient struct {
    apiKey     string
    baseURL    string
    httpClient *http.Client
}

// NewClaudeClient 创建Claude客户端
func NewClaudeClient(apiKey string) *ClaudeClient {
    return &ClaudeClient{
        apiKey:  apiKey,
        baseURL: "https://api.anthropic.com",
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// MessageRequest 消息请求
type MessageRequest struct {
    Model     string    `json:"model"`
    MaxTokens int       `json:"max_tokens"`
    Messages  []Message `json:"messages"`
    System    string    `json:"system,omitempty"`
    Stream    bool      `json:"stream,omitempty"`
}

// Message 消息
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// MessageResponse 消息响应
type MessageResponse struct {
    ID      string `json:"id"`
    Type    string `json:"type"`
    Role    string `json:"role"`
    Content []ContentBlock `json:"content"`
    Model   string `json:"model"`
    StopReason string `json:"stop_reason"`
}

// ContentBlock 内容块
type ContentBlock struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

// SendMessage 发送消息
func (c *ClaudeClient) SendMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
    body, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/messages", bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    // 设置请求头
    httpReq.Header.Set("x-api-key", c.apiKey)
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("anthropic-version", "2023-06-01")

    // 发送请求
    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("send request: %w", err)
    }
    defer resp.Body.Close()

    // 读取响应
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("read response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(respBody))
    }

    var result MessageResponse
    if err := json.Unmarshal(respBody, &result); err != nil {
        return nil, fmt.Errorf("unmarshal response: %w", err)
    }

    return &result, nil
}
```

### 1.2 使用示例

```go
package main

import (
    "context"
    "fmt"
    "log"
)

func main() {
    client := NewClaudeClient("your-api-key-here")

    req := MessageRequest{
        Model:     "claude-3-5-haiku-20241022",
        MaxTokens: 1024,
        Messages: []Message{
            {
                Role:    "user",
                Content: "Hello, Claude!",
            },
        },
    }

    resp, err := client.SendMessage(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

    if len(resp.Content) > 0 {
        fmt.Println(resp.Content[0].Text)
    }
}
```

### 1.3 流式响应（Phase 2准备）

```go
// SendMessageStream 发送消息并获取流式响应
func (c *ClaudeClient) SendMessageStream(ctx context.Context, req MessageRequest) (<-chan StreamChunk, <-chan error) {
    req.Stream = true
    chunkChan := make(chan StreamChunk, 10)
    errChan := make(chan error, 1)

    go func() {
        defer close(chunkChan)
        defer close(errChan)

        body, _ := json.Marshal(req)
        httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/messages", bytes.NewReader(body))
        httpReq.Header.Set("x-api-key", c.apiKey)
        httpReq.Header.Set("Content-Type", "application/json")
        httpReq.Header.Set("anthropic-version", "2023-06-01")

        resp, err := c.httpClient.Do(httpReq)
        if err != nil {
            errChan <- err
            return
        }
        defer resp.Body.Close()

        decoder := json.NewDecoder(resp.Body)
        for {
            var chunk StreamChunk
            if err := decoder.Decode(&chunk); err != nil {
                if err == io.EOF {
                    break
                }
                errChan <- err
                return
            }
            chunkChan <- chunk
        }
    }()

    return chunkChan, errChan
}

// StreamChunk 流式响应块
type StreamChunk struct {
    Type  string `json:"type"`
    Index int    `json:"index,omitempty"`
    Delta struct {
        Type string `json:"type,omitempty"`
        Text string `json:"text,omitempty"`
    } `json:"delta,omitempty"`
    Message MessageResponse `json:"message,omitempty"`
}
```

---

## 二、OpenAI Whisper STT

### 2.1 Whisper API客户端

```go
// internal/stt/whisper.go
package stt

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "time"
)

// WhisperClient Whisper API客户端
type WhisperClient struct {
    apiKey     string
    baseURL    string
    httpClient *http.Client
}

// NewWhisperClient 创建Whisper客户端
func NewWhisperClient(apiKey string) *WhisperClient {
    return &WhisperClient{
        apiKey:  apiKey,
        baseURL: "https://api.openai.com",
        httpClient: &http.Client{
            Timeout: 60 * time.Second, // 音频文件可能较大
        },
    }
}

// TranscribeRequest 转录请求
type TranscribeRequest struct {
    File      []byte
    Filename  string
    Language  string // "zh" for Chinese, "en" for English, etc.
    Prompt    string // 可选的提示词
    Temperature float32 // 0-1, 推荐0.0
}

// TranscribeResponse 转录响应
type TranscribeResponse struct {
    Text string `json:"text"`
}

// Transcribe 转录音频
func (w *WhisperClient) Transcribe(ctx context.Context, req TranscribeRequest) (*TranscribeResponse, error) {
    // 创建multipart表单
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // 添加文件
    part, err := writer.CreateFormFile("file", req.Filename)
    if err != nil {
        return nil, fmt.Errorf("create form file: %w", err)
    }
    if _, err := part.Write(req.File); err != nil {
        return nil, fmt.Errorf("write file: %w", err)
    }

    // 添加其他字段
    writer.WriteField("model", "whisper-1")
    if req.Language != "" {
        writer.WriteField("language", req.Language)
    }
    if req.Prompt != "" {
        writer.WriteField("prompt", req.Prompt)
    }
    writer.WriteField("temperature", fmt.Sprintf("%.1f", req.Temperature))

    if err := writer.Close(); err != nil {
        return nil, fmt.Errorf("close writer: %w", err)
    }

    // 创建HTTP请求
    httpReq, err := http.NewRequestWithContext(ctx, "POST", w.baseURL+"/v1/audio/transcriptions", body)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    httpReq.Header.Set("Authorization", "Bearer "+w.apiKey)
    httpReq.Header.Set("Content-Type", writer.FormDataContentType())

    // 发送请求
    resp, err := w.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("send request: %w", err)
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("read response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(respBody))
    }

    var result TranscribeResponse
    if err := json.Unmarshal(respBody, &result); err != nil {
        return nil, fmt.Errorf("unmarshal response: %w", err)
    }

    return &result, nil
}
```

### 2.2 使用示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
)

func main() {
    client := NewWhisperClient("your-openai-api-key")

    // 读取音频文件
    audioData, err := os.ReadFile("recording.wav")
    if err != nil {
        log.Fatal(err)
    }

    req := TranscribeRequest{
        File:       audioData,
        Filename:   "recording.wav",
        Language:   "zh", // 中文
        Temperature: 0.0,
    }

    resp, err := client.Transcribe(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("转录结果:", resp.Text)
}
```

---

## 三、OpenAI TTS

### 3.1 TTS API客户端

```go
// internal/tts/openai.go
package tts

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// TTSClient TTS API客户端
type TTSClient struct {
    apiKey     string
    baseURL    string
    httpClient *http.Client
}

// NewTTSClient 创建TTS客户端
func NewTTSClient(apiKey string) *TTSClient {
    return &TTSClient{
        apiKey:  apiKey,
        baseURL: "https://api.openai.com",
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// Voice 可用声音
type Voice string

const (
    VoiceAlloy   Voice = "alloy"
    VoiceEcho    Voice = "echo"
    VoiceFable   Voice = "fable"
    VoiceOnyx    Voice = "onyx"
    VoiceNova    Voice = "nova"
    VoiceShimmer Voice = "shimmer"
)

// SynthesizeRequest 合成请求
type SynthesizeRequest struct {
    Text          string `json:"text"`
    Model         string `json:"model"`          // "tts-1" or "tts-1-hd"
    Voice         Voice  `json:"voice"`          // alloy, echo, fable, onyx, nova, shimmer
    ResponseFormat string `json:"response_format"` // mp3, opus, aac, flac
    Speed         float32 `json:"speed"`          // 0.25 to 4.0
}

// Synthesize 合成语音
func (t *TTSClient) Synthesize(ctx context.Context, req SynthesizeRequest) ([]byte, error) {
    if req.Model == "" {
        req.Model = "tts-1"
    }
    if req.Voice == "" {
        req.Voice = VoiceAlloy
    }
    if req.ResponseFormat == "" {
        req.ResponseFormat = "mp3"
    }

    body, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", t.baseURL+"/v1/audio/speech", bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    httpReq.Header.Set("Authorization", "Bearer "+t.apiKey)
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := t.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("send request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        respBody, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(respBody))
    }

    audioData, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("read response: %w", err)
    }

    return audioData, nil
}
```

### 3.2 使用示例

```go
package main

import (
    "context"
    "log"
    "os"
)

func main() {
    client := NewTTSClient("your-openai-api-key")

    req := SynthesizeRequest{
        Text:          "你好，我是Voice Memory助手。",
        Model:         "tts-1",
        Voice:         VoiceNova,
        ResponseFormat: "mp3",
        Speed:         1.0,
    }

    audioData, err := client.Synthesize(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

    // 保存音频文件
    if err := os.WriteFile("output.mp3", audioData, 0644); err != nil {
        log.Fatal(err)
    }

    log.Println("音频已生成: output.mp3")
}
```

---

## 四、Gin Web 框架

### 4.1 基础服务器设置

```go
// cmd/server/main.go
package main

import (
    "log"

    "github.com/gin-gonic/gin"
)

func main() {
    // 创建Gin路由
    r := gin.Default()

    // 中间件
    r.Use(CORSMiddleware())

    // 路由
    api := r.Group("/api/v1")
    {
        api.POST("/chat", HandleChat)
        api.POST("/transcribe", HandleTranscribe)
        api.POST("/synthesize", HandleSynthesize)
        api.GET("/memories", ListMemories)
        api.POST("/memories", CreateMemory)
    }

    // WebSocket
    r.GET("/ws", HandleWebSocket)

    // 静态文件服务
    r.Static("/assets", "./web/assets")
    r.StaticFile("/", "./web/index.html")

    // 启动服务器
    log.Println("Server started on :8080")
    if err := r.Run(":8080"); err != nil {
        log.Fatal(err)
    }
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

### 4.2 API处理器

```go
// internal/api/handlers.go
package api

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// ChatRequest 聊天请求
type ChatRequest struct {
    Message string `json:"message" binding:"required"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
    Reply string `json:"reply"`
}

// HandleChat 处理聊天请求
func HandleChat(c *gin.Context) {
    var req ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 调用Claude API
    reply, err := processChat(req.Message)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, ChatResponse{Reply: reply})
}

func processChat(message string) (string, error) {
    // 实现聊天逻辑
    return "AI回复: " + message, nil
}
```

### 4.3 请求验证与错误处理

```go
// internal/api/middleware.go
package api

import (
    "log"

    "github.com/gin-gonic/gin"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // 检查是否有错误
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            log.Printf("Error: %v", err)

            c.JSON(http.StatusInternalServerError, gin.H{
                "error": err.Error(),
            })
        }
    }
}

// AuthMiddleware 认证中间件（示例）
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")

        // 验证token
        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Missing authorization header",
            })
            return
        }

        c.Next()
    }
}
```

---

## 五、WebSocket 实时通信

### 5.1 WebSocket处理器

```go
// internal/api/websocket.go
package api

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        // MVP阶段允许所有来源
        return true
    },
}

// WSMessage WebSocket消息
type WSMessage struct {
    Type string      `json:"type"` // "audio", "text", "error"
    Data interface{} `json:"data"`
}

// HandleWebSocket 处理WebSocket连接
func HandleWebSocket(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }
    defer conn.Close()

    log.Println("WebSocket client connected")

    // 消息处理循环
    for {
        var msg WSMessage
        if err := conn.ReadJSON(&msg); err != nil {
            log.Printf("WebSocket read error: %v", err)
            break
        }

        // 处理消息
        switch msg.Type {
        case "audio":
            // 处理音频数据
            if err := handleAudioMessage(conn, msg); err != nil {
                log.Printf("Audio error: %v", err)
            }
        case "text":
            // 处理文本消息
            if err := handleTextMessage(conn, msg); err != nil {
                log.Printf("Text error: %v", err)
            }
        default:
            conn.WriteJSON(WSMessage{
                Type: "error",
                Data: "Unknown message type",
            })
        }
    }
}

func handleAudioMessage(conn *websocket.Conn, msg WSMessage) error {
    // 解码音频数据
    audioData, ok := msg.Data.(string)
    if !ok {
        return nil
    }

    // 1. STT: 转录音频为文本
    // text, err := sttClient.Transcribe(audioData)

    // 2. LLM: 调用AI生成回复
    // reply, err := aiClient.SendMessage(text)

    // 3. TTS: 合成回复为音频
    // audio, err := ttsClient.Synthesize(reply)

    // 4. 发送回客户端
    return conn.WriteJSON(WSMessage{
        Type: "audio_response",
        Data: audioData,
    })
}

func handleTextMessage(conn *websocket.Conn, msg WSMessage) error {
    text, ok := msg.Data.(string)
    if !ok {
        return nil
    }

    // 调用AI获取回复
    reply := "AI: " + text

    return conn.WriteJSON(WSMessage{
        Type: "text_response",
        Data: reply,
    })
}
```

### 5.2 使用示例（前端）

```javascript
// 前端WebSocket连接
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
    console.log('WebSocket connected');
};

ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);

    switch(msg.type) {
        case 'audio_response':
            // 播放音频
            playAudio(msg.data);
            break;
        case 'text_response':
            // 显示文本
            displayText(msg.data);
            break;
    }
};

// 发送音频
function sendAudio(audioBase64) {
    ws.send(JSON.stringify({
        type: 'audio',
        data: audioBase64
    }));
}
```

---

## 六、SQLite 数据存储

### 6.1 数据库初始化

```go
// internal/memory/storage.go
package memory

import (
    "database/sql"
    "embed"
    "fmt"
    "log"

    _ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Storage struct {
    db *sql.DB
}

// NewStorage 创建存储实例
func NewStorage(dbPath string) (*Storage, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }

    // 启用外键约束
    db.Exec("PRAGMA foreign_keys = ON")

    // 运行迁移
    if err := runMigrations(db); err != nil {
        return nil, fmt.Errorf("run migrations: %w", err)
    }

    return &Storage{db: db}, nil
}

func runMigrations(db *sql.DB) error {
    // 创建memories表
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS memories (
            id TEXT PRIMARY KEY,
            content TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            metadata TEXT
        )
    `)
    if err != nil {
        return err
    }

    // 创建conversations表
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS conversations (
            id TEXT PRIMARY KEY,
            summary TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        return err
    }

    // 创建messages表
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS messages (
            id TEXT PRIMARY KEY,
            conversation_id TEXT NOT NULL,
            role TEXT NOT NULL,
            content TEXT NOT NULL,
            timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (conversation_id) REFERENCES conversations(id)
        )
    `)

    return err
}

// Close 关闭数据库连接
func (s *Storage) Close() error {
    return s.db.Close()
}
```

### 6.2 CRUD操作

```go
// internal/memory/crud.go
package memory

import (
    "database/sql"
    "encoding/json"
    "time"
)

// CreateMemory 创建记忆
func (s *Storage) CreateMemory(mem *Memory) error {
    query := `
        INSERT INTO memories (id, content, created_at, updated_at, metadata)
        VALUES (?, ?, ?, ?, ?)
    `
    metadata, _ := json.Marshal(mem.Metadata)
    _, err := s.db.Exec(
        query,
        mem.ID,
        mem.Content,
        time.Now(),
        time.Now(),
        string(metadata),
    )
    return err
}

// GetMemory 获取记忆
func (s *Storage) GetMemory(id string) (*Memory, error) {
    query := `SELECT id, content, created_at, updated_at, metadata FROM memories WHERE id = ?`
    row := s.db.QueryRow(query, id)

    var mem Memory
    var metadataStr string
    err := row.Scan(&mem.ID, &mem.Content, &mem.CreatedAt, &mem.UpdatedAt, &metadataStr)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }

    json.Unmarshal([]byte(metadataStr), &mem.Metadata)
    return &mem, nil
}

// ListMemories 列出所有记忆
func (s *Storage) ListMemories() ([]*Memory, error) {
    query := `SELECT id, content, created_at, updated_at, metadata FROM memories ORDER BY created_at DESC`
    rows, err := s.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var memories []*Memory
    for rows.Next() {
        var mem Memory
        var metadataStr string
        if err := rows.Scan(&mem.ID, &mem.Content, &mem.CreatedAt, &mem.UpdatedAt, &metadataStr); err != nil {
            return nil, err
        }
        json.Unmarshal([]byte(metadataStr), &mem.Metadata)
        memories = append(memories, &mem)
    }

    return memories, nil
}

// SearchMemories 搜索记忆（全文搜索）
func (s *Storage) SearchMemories(query string) ([]*Memory, error) {
    sqlQuery := `
        SELECT id, content, created_at, updated_at, metadata
        FROM memories
        WHERE content LIKE ?
        ORDER BY created_at DESC
    `
    rows, err := s.db.Query(sqlQuery, "%"+query+"%")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var memories []*Memory
    for rows.Next() {
        var mem Memory
        var metadataStr string
        if err := rows.Scan(&mem.ID, &mem.Content, &mem.CreatedAt, &mem.UpdatedAt, &metadataStr); err != nil {
            return nil, err
        }
        json.Unmarshal([]byte(metadataStr), &mem.Metadata)
        memories = append(memories, &mem)
    }

    return memories, nil
}
```

---

## 七、向量搜索 (sqlite-vss)

### 7.1 向量搜索设置

```go
// internal/memory/vector.go
package memory

import (
    "database/sql"
    "encoding/json"
    "fmt"
)

// InitVectorSearch 初始化向量搜索
func (s *Storage) InitVectorSearch() error {
    // 加载sqlite-vss扩展
    _, err := s.db.Exec("SELECT load_extension('vss0')")
    if err != nil {
        return fmt.Errorf("load vss extension: %w", err)
    }

    // 创建虚拟表用于向量搜索
    _, err = s.db.Exec(`
        CREATE VIRTUAL TABLE IF NOT EXISTS memory_vectors USING vss0(
            id TEXT PRIMARY KEY,
            embedding(1536)  -- Claude使用1536维向量
        )
    `)

    return err
}

// AddVector 添加向量
func (s *Storage) AddVector(id string, embedding []float32) error {
    embeddingJSON, _ := json.Marshal(embedding)

    _, err := s.db.Exec(`
        INSERT INTO memory_vectors (id, embedding)
        VALUES (?, ?)
    `, id, string(embeddingJSON))

    return err
}

// SearchByVector 向量搜索
func (s *Storage) SearchByVector(queryEmbedding []float32, limit int) ([]*Memory, error) {
    embeddingJSON, _ := json.Marshal(queryEmbedding)

    query := `
        SELECT m.id, m.content, m.created_at, m.updated_at, m.metadata
        FROM memories m
        INNER JOIN memory_vectors v ON m.id = v.id
        WHERE v.embedding MATCH ?
        ORDER BY distance
        LIMIT ?
    `

    rows, err := s.db.Query(query, string(embeddingJSON), limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var memories []*Memory
    for rows.Next() {
        var mem Memory
        var metadataStr string
        if err := rows.Scan(&mem.ID, &mem.Content, &mem.CreatedAt, &mem.UpdatedAt, &metadataStr); err != nil {
            return nil, err
        }
        json.Unmarshal([]byte(metadataStr), &mem.Metadata)
        memories = append(memories, &mem)
    }

    return memories, nil
}
```

---

## 八、音频处理

### 8.1 音频转换

```go
// pkg/audio/converter.go
package audio

import (
    "bytes"
    "encoding/base64"
    "io"
)

// WAVToBase64 WAV转Base64
func WAVToBase64(wavData []byte) string {
    return base64.StdEncoding.EncodeToString(wavData)
}

// Base64ToWAV Base64转WAV
func Base64ToWAV(base64Data string) ([]byte, error) {
    return base64.StdEncoding.DecodeString(base64Data)
}

// ValidateWAV 验证WAV格式
func ValidateWAV(data []byte) error {
    if len(data) < 12 {
        return fmt.Errorf("invalid WAV: too short")
    }

    // 检查RIFF头
    if string(data[0:4]) != "RIFF" {
        return fmt.Errorf("invalid WAV: missing RIFF header")
    }

    // 检查WAVE标识
    if string(data[8:12]) != "WAVE" {
        return fmt.Errorf("invalid WAV: missing WAVE identifier")
    }

    return nil
}

// ConvertAudioFormat 音频格式转换（示例）
func ConvertAudioFormat(input []byte, fromFormat, toFormat string) ([]byte, error) {
    // MVP阶段：直接返回原始数据
    // 实际实现需要使用ffmpeg或其他音频库
    return input, nil
}
```

### 8.2 音频处理工具

```go
// pkg/audio/processor.go
package audio

import (
    "bytes"
    "encoding/binary"
)

// AudioInfo 音频信息
type AudioInfo struct {
    SampleRate   int
    Channels     int
    BitsPerSample int
    Duration     float64
}

// GetAudioInfo 获取音频信息
func GetAudioInfo(wavData []byte) (*AudioInfo, error) {
    if len(wavData) < 44 {
        return nil, fmt.Errorf("invalid WAV: too short")
    }

    info := &AudioInfo{}

    // 采样率 (字节24-27)
    info.SampleRate = int(binary.LittleEndian.Uint32(wavData[24:28]))

    // 声道数 (字节22-23)
    info.Channels = int(binary.LittleEndian.Uint16(wavData[22:24]))

    // 位深 (字节34-35)
    info.BitsPerSample = int(binary.LittleEndian.Uint16(wavData[34:36]))

    // 计算时长
    dataSize := binary.LittleEndian.Uint32(wavData[40:44])
    byteRate := binary.LittleEndian.Uint32(wavData[28:32])
    info.Duration = float64(dataSize) / float64(byteRate)

    return info, nil
}

// Resample 重采样（简化版）
func Resample(wavData []byte, targetSampleRate int) ([]byte, error) {
    info, err := GetAudioInfo(wavData)
    if err != nil {
        return nil, err
    }

    if info.SampleRate == targetSampleRate {
        return wavData, nil
    }

    // MVP阶段：返回错误提示需要外部工具
    return nil, fmt.Errorf("resampling from %d to %d requires external tool", info.SampleRate, targetSampleRate)
}
```

---

## 九、完整集成示例

### 9.1 主服务器

```go
// cmd/server/main.go
package main

import (
    "context"
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"

    "voice-memory/internal/ai"
    "voice-memory/internal/api"
    "voice-memory/internal/memory"
    "voice-memory/internal/stt"
    "voice-memory/internal/tts"
)

func main() {
    // 加载环境变量
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    // 初始化组件
    claudeClient := ai.NewClaudeClient(os.Getenv("CLAUDE_API_KEY"))
    whisperClient := stt.NewWhisperClient(os.Getenv("OPENAI_API_KEY"))
    ttsClient := tts.NewTTSClient(os.Getenv("OPENAI_API_KEY"))

    storage, err := memory.NewStorage("data/voice-memory.db")
    if err != nil {
        log.Fatal(err)
    }
    defer storage.Close()

    // 创建Gin路由
    r := gin.Default()

    // 设置中间件
    r.Use(api.CORSMiddleware())

    // 创建处理器
    handlers := api.NewHandlers(claudeClient, whisperClient, ttsClient, storage)

    // 注册路由
    api := r.Group("/api/v1")
    {
        api.POST("/chat", handlers.HandleChat)
        api.POST("/transcribe", handlers.HandleTranscribe)
        api.POST("/synthesize", handlers.HandleSynthesize)
        api.GET("/memories", handlers.ListMemories)
        api.POST("/memories", handlers.CreateMemory)
        api.GET("/memories/search", handlers.SearchMemories)
    }

    // WebSocket
    r.GET("/ws", handlers.HandleWebSocket)

    // 静态文件
    r.Static("/assets", "./web/assets")
    r.StaticFile("/", "./web/index.html")

    // 启动服务器
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Server started on :%s", port)
    if err := r.Run(":" + port); err != nil {
        log.Fatal(err)
    }
}
```

### 9.2 环境变量

```bash
# .env
CLAUDE_API_KEY=your-claude-api-key
OPENAI_API_KEY=your-openai-api-key
PORT=8080

# 可选
DATABASE_PATH=data/voice-memory.db
LOG_LEVEL=debug
```

### 9.3 构建和运行

```bash
# 安装依赖
go mod download

# 运行开发服务器
go run cmd/server/main.go

# 构建生产版本
go build -o bin/voice-memory cmd/server/main.go

# 运行生产版本
./bin/voice-memory
```

---

**文档版本历史**:
- v1.0-Go (2025-12-29): 创建Golang版本核心代码文档
