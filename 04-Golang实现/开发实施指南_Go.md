# Voice Memory - 开发实施指南 (Golang版本)

**版本：** v1.0-Go
**日期：** 2025-12-29
**语言：** Golang
**状态：** 实施阶段

---

## 目录

1. [项目概述](#一项目概述)
2. [开发环境搭建](#二开发环境搭建)
3. [项目结构](#三项目结构)
4. [开发流程](#四开发流程)
5. [测试策略](#五测试策略)
6. [部署指南](#六部署指南)
7. [常见问题](#七常见问题)
8. [开发规范](#八开发规范)

---

## 一、项目概述

### 1.1 MVP目标

**核心功能：** 简单的语音输入 → AI处理 → 语音输出

```
用户语音输入
    ↓
STT (Whisper API) → 文本
    ↓
LLM (Claude 3.5) → AI回复
    ↓
TTS (OpenAI TTS) → 语音输出
    ↓
播放给用户
```

**MVP不包含**:
- ❌ 可打断功能（Phase 2）
- ❌ 本地STT/TTS（Phase 2）
- ❌ 唤醒词检测（可选）

### 1.2 技术栈

| 类别 | 技术 |
|------|------|
| **后端** | Golang 1.21+ |
| **Web框架** | Gin |
| **数据库** | SQLite3 (go-sqlite3) |
| **WebSocket** | gorilla/websocket |
| **前端** | HTML5 + Vanilla JS |
| **STT** | OpenAI Whisper API |
| **TTS** | OpenAI TTS API |
| **LLM** | Claude 3.5 (Haiku/Sonnet) |

---

## 二、开发环境搭建

### 2.1 系统要求

- Go 1.21 或更高版本
- SQLite3
- Git
- 编辑器：VSCode / GoLand / Vim

### 2.2 安装Go

**macOS**:
```bash
# 使用Homebrew
brew install go

# 验证安装
go version
```

**Linux**:
```bash
# 下载并安装
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# 设置环境变量
export PATH=$PATH:/usr/local/go/bin
```

**Windows**:
- 下载安装器：https://go.dev/dl/
- 运行安装器

### 2.3 项目初始化

```bash
# 创建项目目录
mkdir voice-memory-go
cd voice-memory-go

# 初始化Go模块
go mod init voice-memory

# 创建目录结构
mkdir -p cmd/server
mkdir -p internal/{ai,stt,tts,memory,api,config}
mkdir -p pkg/audio
mkdir -p web
mkdir -p data

# 创建.env文件
cat > .env << EOF
CLAUDE_API_KEY=your-claude-api-key
OPENAI_API_KEY=your-openai-api-key
PORT=8080
EOF

# 创建.env.example
cp .env .env.example
```

### 2.4 安装依赖

```bash
# 安装依赖包
go get github.com/gin-gonic/gin
go get github.com/gorilla/websocket
go get github.com/mattn/go-sqlite3
go get github.com/google/uuid
go get github.com/joho/godotenv
go get golang.org/x/audio

# 整理依赖
go mod tidy
```

### 2.5 VSCode配置

创建 `.vscode/settings.json`:
```json
{
    "go.useLanguageServer": true,
    "go.buildOnSave": "workspace",
    "go.lintOnSave": "workspace",
    "go.vetOnSave": "workspace",
    "go.buildFlags": [],
    "go.lintTool": "golangci-lint",
    "go.lintFlags": [
        "--fast"
    ],
    "go.formatTool": "goimports"
}
```

创建 `.vscode/launch.json`:
```json
{
    "version": "0.10.0",
    "configurations": [
        {
            "name": "Launch Server",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/server",
            "env": {
                "CLAUDE_API_KEY": "${env:CLAUDE_API_KEY}",
                "OPENAI_API_KEY": "${env:OPENAI_API_KEY}"
            }
        }
    ]
}
```

---

## 三、项目结构

### 3.1 目录结构

```
voice-memory-go/
├── cmd/
│   └── server/
│       └── main.go                 # 入口文件
├── internal/
│   ├── ai/
│   │   ├── client.go               # Claude客户端
│   │   ├── claude.go               # Claude API封装
│   │   └── models.go               # 请求响应模型
│   ├── stt/
│   │   ├── whisper.go              # Whisper API
│   │   └── models.go
│   ├── tts/
│   │   ├── openai.go               # OpenAI TTS
│   │   └── models.go
│   ├── memory/
│   │   ├── storage.go              # SQLite存储
│   │   ├── crud.go                 # CRUD操作
│   │   ├── vector.go               # 向量搜索
│   │   └── models.go               # 数据模型
│   ├── api/
│   │   ├── handlers.go             # HTTP处理器
│   │   ├── websocket.go            # WebSocket处理
│   │   ├── middleware.go           # 中间件
│   │   └── routes.go               # 路由定义
│   └── config/
│       └── config.go               # 配置管理
├── pkg/
│   └── audio/
│       ├── processor.go            # 音频处理
│       └── converter.go            # 格式转换
├── web/
│   ├── index.html                  # 前端入口
│   ├── app.js                      # 前端逻辑
│   └── styles.css                  # 样式
├── data/
│   └── voice-memory.db             # SQLite数据库
├── scripts/
│   ├── build.sh                    # 构建脚本
│   └── migrate.sh                  # 迁移脚本
├── .env                            # 环境变量
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── Makefile                        # 构建命令
└── README.md
```

### 3.2 代码组织原则

1. **cmd/** - 应用程序入口
2. **internal/** - 私有代码，外部不可导入
3. **pkg/** - 可被外部项目导入的代码
4. **web/** - 静态前端文件
5. **data/** - 数据文件

### 3.3 关键文件说明

| 文件 | 用途 |
|------|------|
| `go.mod` | Go模块定义和依赖管理 |
| `.env` | 环境变量（不提交到Git） |
| `.gitignore` | Git忽略文件 |
| `Makefile` | 构建命令 |

---

## 四、开发流程

### 4.1 开发阶段规划

**Phase 1: 基础设施 (Week 1)**
- [x] 项目初始化
- [ ] 数据库设置
- [ ] 基础API框架
- [ ] 配置管理

**Phase 2: 核心功能 (Week 2-3)**
- [ ] STT集成
- [ ] LLM集成
- [ ] TTS集成
- [ ] 记忆存储

**Phase 3: 前端 (Week 4)**
- [ ] 录音功能
- [ ] 播放功能
- [ ] WebSocket通信
- [ ] UI设计

**Phase 4: 集成测试 (Week 5)**
- [ ] 端到端测试
- [ ] 性能优化
- [ ] Bug修复

### 4.2 开发任务清单

#### 任务1: 数据库初始化

```go
// internal/memory/storage.go
package memory

import (
    "database/sql"
    "log"

    _ "github.com/mattn/go-sqlite3"
)

func InitDB(dbPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    // 创建表
    if err := createTables(db); err != nil {
        return nil, err
    }

    return db, nil
}

func createTables(db *sql.DB) error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS memories (
            id TEXT PRIMARY KEY,
            content TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
        `CREATE TABLE IF NOT EXISTS conversations (
            id TEXT PRIMARY KEY,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
        `CREATE TABLE IF NOT EXISTS messages (
            id TEXT PRIMARY KEY,
            conversation_id TEXT,
            role TEXT NOT NULL,
            content TEXT NOT NULL,
            timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (conversation_id) REFERENCES conversations(id)
        )`,
    }

    for _, q := range queries {
        if _, err := db.Exec(q); err != nil {
            return err
        }
    }

    return nil
}
```

#### 任务2: API路由设置

```go
// internal/api/routes.go
package api

import (
    "github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, h *Handlers) {
    // CORS中间件
    r.Use(CORSMiddleware())

    // API路由
    api := r.Group("/api/v1")
    {
        api.POST("/chat", h.HandleChat)
        api.POST("/transcribe", h.HandleTranscribe)
        api.POST("/synthesize", h.HandleSynthesize)
        api.GET("/memories", h.ListMemories)
        api.POST("/memories", h.CreateMemory)
    }

    // WebSocket
    r.GET("/ws", h.HandleWebSocket)

    // 静态文件
    r.Static("/assets", "./web/assets")
    r.StaticFile("/", "./web/index.html")
}
```

#### 任务3: 简单的聊天处理器

```go
// internal/api/handlers.go
package api

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "voice-memory/internal/ai"
)

type Handlers struct {
    aiClient *ai.ClaudeClient
    // sttClient, ttsClient, storage等
}

func NewHandlers(aiClient *ai.ClaudeClient) *Handlers {
    return &Handlers{
        aiClient: aiClient,
    }
}

type ChatRequest struct {
    Message string `json:"message" binding:"required"`
}

type ChatResponse struct {
    Reply string `json:"reply"`
}

func (h *Handlers) HandleChat(c *gin.Context) {
    var req ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 调用AI
    reply, err := h.aiClient.SendMessage(c.Request.Context(), ai.MessageRequest{
        Model:     "claude-3-5-haiku-20241022",
        MaxTokens: 1024,
        Messages: []ai.Message{
            {Role: "user", Content: req.Message},
        },
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, ChatResponse{
        Reply: reply.Content[0].Text,
    })
}
```

### 4.3 开发顺序建议

1. **第一天**: 项目初始化 + 数据库
2. **第二天**: Claude API集成
3. **第三天**: Whisper API集成
4. **第四天**: TTS API集成
5. **第五天**: 前端录音播放
6. **第六天**: WebSocket通信
7. **第七天**: 集成测试

---

## 五、测试策略

### 5.1 单元测试

```go
// internal/ai/claude_test.go
package ai

import (
    "context"
    "testing"
)

func TestClaudeClient_SendMessage(t *testing.T) {
    // 跳过测试如果没有API密钥
    if testing.Short() {
        t.Skip("skipping test in short mode")
    }

    client := NewClaudeClient("test-api-key")

    tests := []struct {
        name    string
        req     MessageRequest
        wantErr bool
    }{
        {
            name: "simple message",
            req: MessageRequest{
                Model:     "claude-3-5-haiku-20241022",
                MaxTokens: 1024,
                Messages: []Message{
                    {Role: "user", Content: "Hello"},
                },
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp, err := client.SendMessage(context.Background(), tt.req)
            if (err != nil) != tt.wantErr {
                t.Errorf("SendMessage() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && resp == nil {
                t.Error("SendMessage() returned nil response")
            }
        })
    }
}
```

### 5.2 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/ai

# 详细输出
go test -v ./...

# 覆盖率
go test -cover ./...

# 覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 5.3 集成测试

```go
// tests/integration_test.go
package tests

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "encoding/json"

    "voice-memory/internal/api"
)

func TestChatEndpoint(t *testing.T) {
    // 设置测试路由
    r := setupTestRouter()

    // 创建测试请求
    reqBody := map[string]string{"message": "Hello"}
    body, _ := json.Marshal(reqBody)

    req, _ := http.NewRequest("POST", "/api/v1/chat", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    // 记录响应
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    // 验证
    if w.Code != http.StatusOK {
        t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
    }
}
```

---

## 六、部署指南

### 6.1 构建应用

**Makefile**:
```makefile
.PHONY: build run test clean

# 构建
build:
    go build -o bin/voice-memory cmd/server/main.go

# 运行
run:
    go run cmd/server/main.go

# 测试
test:
    go test -v ./...

# 清理
clean:
    rm -rf bin/

# 构建 Linux
build-linux:
    GOOS=linux GOARCH=amd64 go build -o bin/voice-memory-linux cmd/server/main.go

# 构建 macOS
build-mac:
    GOOS=darwin GOARCH=amd64 go build -o bin/voice-memory-mac cmd/server/main.go

# 构建 Windows
build-windows:
    GOOS=windows GOARCH=amd64 go build -o bin/voice-memory.exe cmd/server/main.go
```

### 6.2 本地运行

```bash
# 安装依赖
go mod download

# 运行开发服务器
go run cmd/server/main.go

# 或使用Make
make run
```

### 6.3 生产部署

#### Linux (Systemd)

```bash
# 1. 构建
make build-linux

# 2. 复制到服务器
scp bin/voice-memory-linux user@server:/opt/voice-memory/

# 3. 创建systemd服务
sudo cat > /etc/systemd/system/voice-memory.service << EOF
[Unit]
Description=Voice Memory Server
After=network.target

[Service]
Type=simple
User=voicememory
WorkingDirectory=/opt/voice-memory
Environment="CLAUDE_API_KEY=your-key"
Environment="OPENAI_API_KEY=your-key"
ExecStart=/opt/voice-memory/voice-memory-linux
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

# 4. 启动服务
sudo systemctl daemon-reload
sudo systemctl enable voice-memory
sudo systemctl start voice-memory
```

#### Docker

**Dockerfile**:
```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o voice-memory cmd/server/main.go

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/voice-memory .
COPY --from=builder /app/web ./web
COPY --from=builder /app/.env .env

EXPOSE 8080

CMD ["./voice-memory"]
```

**docker-compose.yml**:
```yaml
version: '3.8'

services:
  voice-memory:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CLAUDE_API_KEY=${CLAUDE_API_KEY}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - PORT=8080
    volumes:
      - ./data:/root/data
    restart: unless-stopped
```

**运行**:
```bash
# 构建镜像
docker build -t voice-memory .

# 运行容器
docker run -p 8080:8080 \
  -e CLAUDE_API_KEY=your-key \
  -e OPENAI_API_KEY=your-key \
  voice-memory

# 或使用docker-compose
docker-compose up -d
```

---

## 七、常见问题

### 7.1 编译错误

**CGO错误** (go-sqlite3需要):
```bash
# macOS
xcode-select --install

# Linux (Ubuntu)
sudo apt-get install build-essential

# 或使用纯Go的SQLite驱动
go get modernc.org/sqlite
```

### 7.2 API调用失败

**错误**: API key无效
```bash
# 检查环境变量
echo $CLAUDE_API_KEY

# 确保.env文件存在且正确
cat .env
```

### 7.3 WebSocket连接失败

**错误**: CORS错误
```go
// 确保CORS中间件正确设置
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

### 7.4 数据库锁定

**错误**: database is locked
```go
// 设置SQLite连接参数
db, err := sql.Open("sqlite3", "file:voice-memory.db?cache=shared&mode=rwc")
```

---

## 八、开发规范

### 8.1 代码风格

- 使用 `gofmt` 格式化代码
- 使用 `goimports` 管理import
- 遵循 Effective Go 指南

```bash
# 格式化代码
gofmt -w .

# 更新imports
goimports -w .
```

### 8.2 错误处理

```go
// 好的做法
result, err := someFunction()
if err != nil {
    return fmt.Errorf("context: %w", err)  // 包装错误
}

// 不好的做法
result, _ := someFunction()  // 忽略错误
```

### 8.3 日志

```go
import "log"

// 使用标准log
log.Printf("Processing request: %s", requestID)

// 错误日志
log.Printf("ERROR: Failed to process: %v", err)
```

### 8.4 环境变量

```go
// 使用os.Getenv
apiKey := os.Getenv("CLAUDE_API_KEY")

// 提供默认值
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
```

### 8.5 Git提交规范

```
feat: add Claude API integration
fix: resolve WebSocket connection issue
docs: update README with deployment instructions
test: add unit tests for storage layer
refactor: simplify audio processing code
```

---

## 九、快速参考

### 9.1 常用命令

```bash
# 运行
go run cmd/server/main.go

# 构建
go build -o bin/voice-memory cmd/server/main.go

# 测试
go test ./...

# 依赖管理
go mod tidy
go mod download

# 格式化
gofmt -w .
```

### 9.2 重要文件

| 文件 | 用途 |
|------|------|
| `go.mod` | 依赖管理 |
| `.env` | 环境变量（不提交） |
| `Makefile` | 构建命令 |
| `.gitignore` | Git忽略文件 |

### 9.3 端口和API

| 服务 | 端口 |
|------|------|
| HTTP API | 8080 |
| WebSocket | 8080/ws |

---

**文档版本历史**:
- v1.0-Go (2025-12-29): 创建Golang版本开发实施指南
