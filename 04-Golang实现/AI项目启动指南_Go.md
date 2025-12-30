# Voice Memory - AI项目启动指南 (Golang版本)

**版本：** v1.0-Go
**日期：** 2025-12-29
**用途：** 指导AI助手完整实现项目

---

## 📋 给AI的完整指令

### 第一步：理解项目

**请先阅读以下文档（按顺序）**：

1. **产品文档.md** - 了解产品定位和功能
   - 重点：MVP范围（简单语音输入输出，"可打断"是Phase 2）
   - 重点：Web PWA部署方式（不是原生macOS App）

2. **技术栈决策文档_Go.md** - 理解技术选择
   - 后端：Golang + Gin
   - 前端：Web PWA (HTML5 + Vanilla JS)
   - 数据库：SQLite3
   - API：Claude + OpenAI Whisper/TTS

3. **数据模型文档_Go.md** - 了解数据库设计
   - 3张核心表：conversations, messages, memories
   - 向量搜索：sqlite-vss扩展

### 第二步：开始实现

**按以下顺序实施**：

#### 阶段1：项目初始化（第1天）

**参考文档**：开发实施指南_Go.md 第三章

```bash
# 1. 创建项目结构
mkdir -p voice-memory-go/{cmd/server,internal/{ai,stt,tts,memory,api,config},pkg/audio,web,data}

# 2. 初始化Go模块
cd voice-memory-go
go mod init voice-memory

# 3. 安装依赖
go get github.com/gin-gonic/gin
go get github.com/gorilla/websocket
go get github.com/mattn/go-sqlite3
go get github.com/google/uuid
go get github.com/joho/godotenv

# 4. 创建环境变量文件
cat > .env << 'EOF'
CLAUDE_API_KEY=your-claude-api-key
OPENAI_API_KEY=your-openai-api-key
PORT=8080
DATABASE_PATH=data/voice-memory.db
EOF
```

**参考代码**：核心代码示例_Go.md 第3.2节（项目结构）

#### 阶段2：数据库层（第2天）

**参考文档**：数据模型文档_Go.md 第二章

```go
// 创建文件：internal/memory/storage.go
// 实现以下功能：
// 1. InitDB() - 数据库初始化
// 2. createTables() - 创建表
// 3. runMigrations() - 运行迁移
```

**参考代码**：核心代码示例_Go.md 第六章（SQLite数据存储）

#### 阶段3：API客户端层（第3天）

**参考文档**：核心代码示例_Go.md 第一、二、三章

```go
// 1. internal/ai/claude.go - Claude API客户端
// 参考文档：核心代码示例_Go.md 第一节

// 2. internal/stt/whisper.go - Whisper API客户端
// 参考文档：核心代码示例_Go.md 第二节

// 3. internal/tts/openai.go - TTS API客户端
// 参考文档：核心代码示例_Go.md 第三节
```

#### 阶段4：Web服务器层（第4天）

**参考文档**：核心代码示例_Go.md 第四、五章

```go
// 1. internal/api/routes.go - 路由设置
// 参考文档：开发实施指南_Go.md 4.2节

// 2. internal/api/handlers.go - API处理器
// 参考文档：API设计文档_Go.md 第二章

// 3. internal/api/websocket.go - WebSocket处理
// 参考文档：API设计文档_Go.md 第三章

// 4. cmd/server/main.go - 入口文件
// 参考文档：核心代码示例_Go.md 9.1节
```

#### 阶段5：前端层（第5天）

**参考文档**：开发实施指南_Go.md（没有前端详细代码，需要AI实现）

```html
<!-- web/index.html -->
<!DOCTYPE html>
<html>
<head>
    <title>Voice Memory</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body>
    <div id="app">
        <!-- 录音按钮 -->
        <button id="recordBtn">🎤 开始录音</button>
        <!-- 播放区域 -->
        <audio id="audioPlayer" controls></audio>
        <!-- 聊天记录 -->
        <div id="chatLog"></div>
    </div>
    <script src="app.js"></script>
</body>
</html>
```

```javascript
// web/app.js
// 参考：API设计文档_Go.md 第三章（WebSocket消息格式）
// 实现：录音、WebSocket连接、消息处理
```

#### 阶段6：集成测试（第6天）

**参考文档**：开发实施指南_Go.md 第五章

```bash
# 1. 启动服务器
go run cmd/server/main.go

# 2. 测试API
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "你好"}'

# 3. 测试WebSocket
# 使用浏览器打开 http://localhost:8080
# 测试录音和播放功能
```

### 第三步：部署

**参考文档**：部署与运维文档_Go.md

**本地测试**：
```bash
go build -o bin/voice-memory cmd/server/main.go
./bin/voice-memory
```

**生产部署**（可选）：
- Systemd服务：参考第四章
- Docker部署：参考第四章
- Nginx反向代理：参考3.2节

---

## 🔗 文档引用关系

```
启动指南（本文）
    ↓
产品文档.md ← 了解需求
    ↓
技术栈决策文档_Go.md ← 了解技术选型
    ↓
开发实施指南_Go.md ← 项目结构
    ↓
数据模型文档_Go.md ← 数据库设计
    ↓
核心代码示例_Go.md ← 代码实现
    ↓
API设计文档_Go.md ← 接口定义
    ↓
部署与运维文档_Go.md ← 部署上线
```

---

## ✅ 检查清单

AI完成项目后，应该能够：

- [ ] 启动服务器：`go run cmd/server/main.go`
- [ ] 访问Web界面：http://localhost:8080
- [ ] 录音并转换为文字
- [ ] AI回复并转换为语音
- [ ] 播放AI的语音回复
- [ ] 查看历史记忆
- [ ] 搜索记忆

---

## 🚨 MVP范围提醒

**MVP不需要实现**：
- ❌ 可打断功能（Phase 2）
- ❌ 本地STT/TTS（Phase 2）
- ❌ 唤醒词检测（可选）
- ❌ 向量搜索（Phase 2，MVP用简单文本搜索即可）

**MVP必须实现**：
- ✅ 语音输入 → STT API → 文本
- ✅ 文本 → Claude API → AI回复
- ✅ AI回复 → TTS API → 语音输出
- ✅ 基础记忆存储（SQLite）
- ✅ 简单文本搜索

---

## 📦 最终交付物

项目完成后，应该有以下文件：

```
voice-memory-go/
├── cmd/server/main.go           ✅ 入口文件
├── internal/
│   ├── ai/claude.go              ✅ Claude客户端
│   ├── stt/whisper.go            ✅ Whisper客户端
│   ├── tts/openai.go             ✅ TTS客户端
│   ├── memory/storage.go         ✅ 数据库操作
│   └── api/
│       ├── routes.go             ✅ 路由
│       ├── handlers.go           ✅ HTTP处理器
│       └── websocket.go          ✅ WebSocket处理
├── web/
│   ├── index.html                ✅ 前端页面
│   └── app.js                    ✅ 前端逻辑
├── .env                          ✅ 环境变量
├── go.mod & go.sum               ✅ 依赖管理
└── README.md                     ✅ 使用说明
```

---

## 🎯 给AI的最后提示

1. **严格按照MVP范围**，不要过度设计
2. **先让代码跑起来**，再优化
3. **使用提供的代码示例**，不要重复造轮子
4. **遇到问题先查文档**，所有答案都在这些文档里
5. **分阶段验证**，不要等所有代码写完才测试

---

**祝开发顺利！** 🚀
