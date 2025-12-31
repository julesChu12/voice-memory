# Voice Memory - 技术栈决策文档 (Golang版本)

**版本：** v1.0-Go
**日期：** 2025-12-29
**语言：** Golang
**状态：** 决策确认

---

## 一、技术栈总览

### 1.1 完整技术栈

```
┌─────────────────────────────────────────────────────────────────┐
│                      Voice Memory 技术架构                       │
│                        (Golang 实现)                             │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        前端层 (Web PWA)                          │
├─────────────────────────────────────────────────────────────────┤
│  • PWA框架: 纯HTML5 + JavaScript (vanilla或React)                │
│  • UI框架: TailwindCSS 或 Bootstrap                              │
│  • 音频: Web Audio API + MediaRecorder API                       │
│  • 通信: WebSocket + HTTP REST API                               │
│  • 存储: IndexedDB (本地缓存)                                    │
└─────────────────────────────────────────────────────────────────┘
                              ↕ HTTP/WebSocket
┌─────────────────────────────────────────────────────────────────┐
│                        后端层 (Golang)                            │
├─────────────────────────────────────────────────────────────────┤
│  • Web框架: Gin / Fiber                                           │
│  • WebSocket: gorilla/websocket                                   │
│  • AI客户端: 自定义HTTP客户端 (Claude API)                         │
│  • 音频处理: 原生audio包 + 采样率转换                              │
│  • 向量搜索: sqlite-vss (SQLite扩展)                              │
└─────────────────────────────────────────────────────────────────┘
                              ↕
┌─────────────────────────────────────────────────────────────────┐
│                        数据层                                     │
├─────────────────────────────────────────────────────────────────┤
│  • 主数据库: SQLite3 (go-sqlite3)                                │
│  • 向量搜索: sqlite-vss (Virtual Table)                          │
│  • 文件存储: 本地文件系统 (Markdown笔记)                          │
│  • 会话管理: 内存 + SQLite持久化                                   │
└─────────────────────────────────────────────────────────────────┘
                              ↕
┌─────────────────────────────────────────────────────────────────┐
│                        外部服务                                   │
├─────────────────────────────────────────────────────────────────┤
│  • STT: OpenAI Whisper API                                       │
│  • TTS: OpenAI TTS API                                           │
│  • LLM: Claude 3.5 (Haiku快速响应 + Sonnet深度思考)               │
│  • 唤醒词: Picovoice Porcupine (可选)                             │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 为什么选择Golang？

| 评估维度 | Python | Golang | 决策依据 |
|---------|--------|-------|----------|
| **开发效率** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Python更简洁，但Golang足够高效 |
| **性能** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Golang原生性能更优 |
| **并发模型** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Goroutine > Python asyncio |
| **部署** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 单一二进制文件，无需依赖 |
| **团队熟悉度** | ⭐⭐ | ⭐⭐⭐⭐⭐ | **决定性因素：团队主要使用Golang** |
| **生态** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Python生态更丰富，但Go足够 |
| **总评** | 28分 | 30分 | **选择Golang** |

**核心决策理由**：
1. **团队主要使用Golang** - 这是决定性因素
2. **部署简单** - 单一二进制文件，跨平台
3. **性能优秀** - 原生并发，低延迟
4. **类型安全** - 编译时错误检查

---

## 二、前端技术栈

### 2.1 Web PWA方案

**决策：** 使用纯Web技术构建PWA (Progressive Web App)

```
技术选型对比:

方案A: 原生macOS App (Swift)
  ✅ 性能最优
  ✅ 系统集成深
  ❌ 开发周期长 (8-12周)
  ❌ 需要Swift经验
  ❌ 跨平台需重写

方案B: Electron (桌面应用)
  ✅ 跨平台
  ✅ Web技术栈
  ❌ 包体积大 (100MB+)
  ❌ 内存占用高
  ❌ 性能损耗

方案C: Web PWA ✅ 选择
  ✅ 快速开发 (2-4周)
  ✅ 跨平台 (Mac/Windows/Linux/移动端)
  ✅ 轻量级 (<5MB)
  ✅ 易于部署 (静态文件)
  ✅ 快速迭代 (刷新即更新)
  ⚠️ 性能略逊原生 (但对MVP足够)
```

### 2.2 前端技术栈

| 类别 | 技术选择 | 理由 |
|------|----------|------|
| **核心框架** | 纯HTML5 + Vanilla JS | MVP最简单，无构建步骤 |
| **可选升级** | React / Vue | 后续UI复杂时可引入 |
| **样式** | TailwindCSS | 快速开发，响应式 |
| **图标** | Lucide Icons | 轻量、现代 |
| **音频处理** | Web Audio API | 浏览器原生 |
| **录音** | MediaRecorder API | 浏览器原生 |
| **通信** | WebSocket + Fetch API | 实时 + REST |
| **本地存储** | IndexedDB | 持久化缓存 |

### 2.3 关键依赖

```json
// package.json (如果使用npm)
{
  "dependencies": {
    "tailwindcss": "^3.4.0",        // 样式框架
    "lucide": "^0.300.0"            // 图标库
  }
}
```

**MVP阶段无需构建步骤**：
- 直接引用CDN的TailwindCSS
- 使用纯ES6 JavaScript
- 无需Webpack/Vite等构建工具

---

## 三、后端技术栈 (Golang)

### 3.1 Web框架选择

**Gin vs Fiber 对比**:

| 特性 | Gin | Fiber | 选择 |
|------|-----|-------|------|
| **性能** | 极快 | 更快 | Gin足够 |
| **API设计** | Express风格 | Express风格 | 平局 |
| **生态** | 成熟 | 较新 | ✅ Gin |
| **文档** | 完善 | 完善 | 平局 |
| **社区** | 大 | 中等 | ✅ Gin |
| **中间件** | 丰富 | 丰富 | 平局 |
| **GitHub Stars** | 77k | 32k | ✅ Gin |

**决策：使用Gin**

**理由**：
1. 生态成熟，社区活跃
2. 文档完善，示例丰富
3. 性能足够优秀
4. 团队更熟悉

### 3.2 核心依赖包

```go
// go.mod
module voice-memory

go 1.21

require (
    // Web框架
    github.com/gin-gonic/gin v1.10.0

    // WebSocket
    github.com/gorilla/websocket v1.5.1

    // 数据库
    github.com/mattn/go-sqlite3 v1.14.22
    github.com/google/uuid v1.6.0

    // HTTP客户端 (用于API调用)
    net/http 标准库

    // JSON处理
    encoding/json 标准库

    // 音频处理
    golang.org/x/audio v0.15.0

    // 环境变量
    github.com/joho/godotenv v1.5.1
)
```

### 3.3 项目结构

```
voice-memory-go/
├── cmd/
│   └── server/
│       └── main.go                 # 入口文件
├── internal/
│   ├── ai/
│   │   ├── client.go               # Claude API客户端
│   │   ├── claude.go               # Claude调用封装
│   │   └── models.go               # 请求响应模型
│   ├── stt/
│   │   ├── whisper.go              # OpenAI Whisper API
│   │   └── models.go
│   ├── tts/
│   │   ├── openai.go               # OpenAI TTS API
│   │   └── models.go
│   ├── memory/
│   │   ├── storage.go              # SQLite存储
│   │   ├── vector.go               # 向量搜索封装
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
│       ├── processor.go            # 音频处理工具
│       └── converter.go            # 格式转换
├── web/
│   ├── index.html                  # 前端入口
│   ├── app.js                      # 前端逻辑
│   └── styles.css                  # 样式
├── data/
│   ├── voice-memory.db             # SQLite数据库
│   └── memories/                   # Markdown文件存储
├── .env                            # 环境变量
├── .env.example
├── go.mod
├── go.sum
├── Makefile                        # 构建脚本
└── README.md
```

---

## 四、数据层技术栈

### 4.1 数据库选择：SQLite3

**为什么选择SQLite？**

| 方案 | 优势 | 劣势 | 适用场景 |
|------|------|------|----------|
| **SQLite** | ✅ 零配置<br>✅ 单文件<br>✅ 本地优先 | ⚠️ 单写者 | MVP/个人使用 |
| PostgreSQL | ✅ 功能完整<br>✅ 并发优秀 | ❌ 需要部署<br>❌ 复杂 | 企业版 |
| MySQL | ✅ 流行 | ❌ 需要部署 | 企业版 |

**决策：SQLite3 (使用 go-sqlite3)**

### 4.2 向量搜索方案

**方案对比**:

| 方案 | 实现复杂度 | 性能 | 部署 | 选择 |
|------|-----------|------|------|------|
| **sqlite-vss** | 低 | 中 | 极简 | ✅ MVP |
| **chromadb** | 中 | 高 | 需Python | Phase 2 |
| **pgvector** | 中 | 高 | 需Postgres | 企业版 |

**决策：sqlite-vss (SQLite扩展)**

**集成方式**:
```go
// 加载sqlite-vss扩展
db.Exec("SELECT load_extension('vss')")
```

### 4.3 数据模型设计

```go
// internal/memory/models.go
type Memory struct {
    ID          string    `json:"id" db:"id"`
    Content     string    `json:"content" db:"content"`
    Embedding   []float32 `json:"-" db:"embedding"`  // 向量
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
    Metadata    string    `json:"metadata" db:"metadata"`  // JSON
}

type Conversation struct {
    ID        string         `json:"id" db:"id"`
    Messages  []Message      `json:"messages"`
    Summary   string         `json:"summary"`
    CreatedAt time.Time      `json:"created_at"`
}

type Message struct {
    ID        string    `json:"id" db:"id"`
    Role      string    `json:"role" db:"role"`        // "user" | "assistant"
    Content   string    `json:"content" db:"content"`
    Timestamp time.Time `json:"timestamp" db:"timestamp"`
}
```

---

## 五、外部服务集成

### 5.1 AI服务：Claude 3.5

**模型选择策略**:

| 场景 | 模型 | 理由 |
|------|------|------|
| **实时对话** | Claude 3.5 Haiku | 快速、便宜 |
| **深度思考** | Claude 3.5 Sonnet | 质量高 |
| **归档总结** | Claude 3.5 Haiku | 成本效益 |

**Golang实现方式**:
```go
// internal/ai/client.go
type ClaudeClient struct {
    apiKey  string
    baseURL string
    httpClient *http.Client
}

func (c *ClaudeClient) SendMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
    // 实现Claude API调用
}
```

### 5.2 STT服务：OpenAI Whisper API

**选择理由**:
- 准确率最高 (行业领先)
- API简单
- 支持多语言
- 延迟可接受 (2-5秒)

**Golang实现**:
```go
// internal/stt/whisper.go
func Transcribe(audio []byte, language string) (string, error) {
    // 调用OpenAI Whisper API
}
```

### 5.3 TTS服务：OpenAI TTS

**选择理由**:
- 语音自然
- API简单
- 多种声音选择 (alloy, echo, fable, onyx, nova, shimmer)
- 延迟低 (1-2秒)

**Golang实现**:
```go
// internal/tts/openai.go
func Synthesize(text string, voice string) ([]byte, error) {
    // 调用OpenAI TTS API
}
```

### 5.4 唤醒词检测（可选）

**Porcupine by Picovoice**:
- 有免费版 (最多3个唤醒词)
- 跨平台支持
- 延迟极低
- **注意**: 可能需要CGO或HTTP API

**MVP阶段**:
- 可以先用按钮触发，不需要唤醒词
- Phase 2再添加

---

## 六、关键代码决策

### 6.1 Claude API调用（Golang）

**HTTP客户端模式**:

```go
// internal/ai/claude.go
package ai

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type ClaudeClient struct {
    apiKey  string
    baseURL string
}

type MessageRequest struct {
    Model     string    `json:"model"`
    MaxTokens int       `json:"max_tokens"`
    Messages  []Message `json:"messages"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

func (c *ClaudeClient) SendMessage(req MessageRequest) (*Response, error) {
    body, _ := json.Marshal(req)

    httpReq, _ := http.NewRequest("POST", c.baseURL+"/v1/messages", bytes.NewReader(body))
    httpReq.Header.Set("x-api-key", c.apiKey)
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("anthropic-version", "2023-06-01")

    client := &http.Client{}
    resp, err := client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result Response
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}
```

### 6.2 WebSocket实时通信

```go
// internal/api/websocket.go
package api

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // MVP阶段允许所有来源
    },
}

func HandleWebSocket(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()

    for {
        var msg WSMessage
        err := conn.ReadJSON(&msg)
        if err != nil {
            break
        }

        // 处理消息
        switch msg.Type {
        case "audio":
            // 处理音频数据
        case "text":
            // 处理文本消息
        }
    }
}
```

### 6.3 音频处理

```go
// pkg/audio/processor.go
package audio

import (
    "bytes"
    "encoding/base64"
)

func ConvertWAVToBase64(wavData []byte) string {
    return base64.StdEncoding.EncodeToString(wavData)
}

func ConvertBase64ToWAV(base64Data string) ([]byte, error) {
    return base64.StdEncoding.DecodeString(base64Data)
}
```

---

## 七、部署与运维

### 7.1 开发环境

**要求**:
- Go 1.21+
- SQLite3
- Make (可选)

**初始化**:
```bash
# 克隆项目
git clone https://github.com/yourusername/voice-memory-go.git
cd voice-memory-go

# 安装依赖
go mod download

# 复制环境变量
cp .env.example .env
# 编辑.env，填入API密钥

# 运行
go run cmd/server/main.go
```

### 7.2 构建与部署

**单文件部署**:
```makefile
# Makefile
build:
    go build -o bin/voice-memory cmd/server/main.go

build-linux:
    GOOS=linux GOARCH=amd64 go build -o bin/voice-memory-linux cmd/server/main.go

build-mac:
    GOOS=darwin GOARCH=amd64 go build -o bin/voice-memory-mac cmd/server/main.go

build-windows:
    GOOS=windows GOARCH=amd64 go build -o bin/voice-memory.exe cmd/server/main.go
```

**部署**:
```bash
# 构建
make build

# 运行
./bin/voice-memory

# 或直接go run（开发）
go run cmd/server/main.go
```

### 7.3 生产环境（可选）

**使用Systemd (Linux)**:
```ini
[Unit]
Description=Voice Memory Server
After=network.target

[Service]
Type=simple
User=voicememory
WorkingDirectory=/opt/voice-memory
ExecStart=/opt/voice-memory/bin/voice-memory
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

### 7.4 Docker部署（可选）

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o voice-memory cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/voice-memory .
COPY --from=builder /app/web ./web
COPY --from=builder /app/.env .env
EXPOSE 8080
CMD ["./voice-memory"]
```

---

## 八、成本估算

### 8.1 API成本（每月）

| 服务 | 使用量 | 单价 | 月成本 |
|------|--------|------|--------|
| **Claude Haiku** | 10万token | $0.25/百万 | $0.025 |
| **Claude Sonnet** | 5万token | $3/百万 | $0.15 |
| **Whisper** | 100分钟 | $0.006/分钟 | $0.60 |
| **TTS** | 100万字 | $15/百万 | $1.50 |
| **总计** | - | - | **~$2.30/月** |

### 8.2 优化策略

1. **Haiku优先** - 实时对话用Haiku，复杂任务才用Sonnet
2. **本地缓存** - 常见问题本地缓存，减少API调用
3. **批量处理** - 归档任务批量执行，利用off-peak时段

---

## 九、技术风险与缓解

### 9.1 风险矩阵

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| **OpenAI API限流** | 中 | 高 | 实现重试+队列，必要时切换到本地Whisper |
| **Claude API限流** | 低 | 中 | 实现重试机制，使用指数退避 |
| **SQLite并发限制** | 低 | 中 | MVP单用户足够，企业版升级PostgreSQL |
| **前端兼容性** | 中 | 低 | 使用标准Web API，测试主流浏览器 |

### 9.2 技术债务管理

**当前技术债务**:
- MVP阶段使用API方式STT/TTS，延迟较高（2-5秒）
- 单一SQLite实例，无法多实例部署

**Phase 2改进计划**:
- 考虑本地Whisper模型（降低STT延迟）
- 升级到PostgreSQL + pgvector（支持多用户）

---

## 十、后续演进路径

### 10.1 技术栈演进

```
Phase 1 (MVP): 当前技术栈
  ├─ 前端: Web PWA
  ├─ 后端: Golang + Gin
  ├─ 数据: SQLite + sqlite-vss
  ├─ STT/TTS: OpenAI API
  └─ AI: Claude 3.5

Phase 2 (可打断):
  ├─ 前端: WebRTC低延迟音频
  ├─ 后端: 添加WebSocket流式处理
  ├─ STT: 本地Whisper (降低延迟)
  ├─ VAD: sherpa-onnx本地检测
  └─ 打断检测: 实时VAD

Phase 3 (端到端语音):
  ├─ 模型: GPT-4o Realtime API
  ├─ 音频: WebRTC原生支持
  └─ 延迟: <500ms
```

### 10.2 扩展性考虑

**数据库迁移路径**:
```
SQLite (MVP) → PostgreSQL (企业版)
    ↓
  迁移工具: pgloader / 自定义脚本
```

**微服务化 (未来)**:
```
单体应用 → 服务拆分
    ↓
  - API Gateway
  - AI Service
  - Memory Service
  - STT/TTS Service
```

---

## 十一、参考资源

### 11.1 官方文档

- [Gin Web框架](https://gin-gonic.com/docs/)
- [go-sqlite3](https://github.com/mattn/go-sqlite3)
- [Claude API文档](https://docs.anthropic.com/claude/reference/)
- [OpenAI Whisper API](https://platform.openai.com/docs/guides/speech-to-text)
- [OpenAI TTS API](https://platform.openai.com/docs/guides/text-to-speech)

### 11.2 代码示例

- [Gin WebSocket示例](https://github.com/gin-gonic/examples/tree/master/websocket)
- [SQLite Go教程](https://www.riptutorial.com/go/example/17635/introduction-to-sqlite)

### 11.3 最佳实践

- [Go项目布局](https://github.com/golang-standards/project-layout)
- [Effective Go](https://go.dev/doc/effective_go)

---

**文档版本历史**:
- v1.0-Go (2025-12-29): 创建Golang版本技术栈文档
