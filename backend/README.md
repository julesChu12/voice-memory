# Voice Memory

智能语音助手 - 语音记录 → AI 整理 → 知识库 (RAG)

## 功能特性

### 核心功能
- **语音识别** - 百度 STT 语音转文字
- **智能对话** - 基于 GLM-4 的 AI 助手
- **知识整理** - AI 自动生成标题、摘要、关键点、分类、标签
- **RAG 检索** - 智谱 AI 向量搜索，从知识库中智能检索
- **语音播报** - 百度 TTS 文字转语音
- **会话管理** - 自动跟踪对话历史，上下文压缩

### AI 能力
- **意图识别** - 自动判断是普通对话还是知识检索
- **标题生成** - 从完整对话生成 5-15 字简洁标题
- **内容摘要** - 基于会话上下文生成结构化摘要
- **知识分类** - 自动分类：技术/生活/工作/学习/想法
- **标签提取** - 智能提取关键词标签

## 快速开始

### 1. 获取 API Keys

#### 百度 AI (语音识别 + 语音合成)
1. 访问 [百度 AI 开放平台](https://console.bce.baidu.com/ai/#/ai/speech/app/list)
2. 创建应用 → 选择"语音识别"和"语音合成"
3. 获取 API Key 和 Secret Key
4. 免费额度: 5-15万次/月

#### 智谱 AI (GLM-4 对话 + 向量搜索)
1. 访问 [智谱 AI 开放平台](https://open.bigmodel.cn/usercenter/apikeys)
2. 创建 API Key
3. 获取 `GLM_API_KEY`

### 2. 设置环境变量

创建 `.env` 文件:

```bash
# 服务器端口
SERVER_PORT=8080

# 百度语音识别配置
BAIDU_API_KEY=你的_BAIDU_API_KEY
BAIDU_SECRET_KEY=你的_BAIDU_SECRET_KEY

# GLM 智谱 AI 配置
GLM_API_KEY=你的_GLM_API_KEY
```

### 3. 安装依赖

```bash
cd backend
go mod download
```

### 4. 运行服务

```bash
go run cmd/main.go
```

服务将在 http://localhost:8080 启动，直接访问即可使用 Web 界面。

## API 接口

### 语音识别
```
POST /api/stt
- Content-Type: multipart/form-data
- audio: 音频文件
- format: wav/pcm

响应: {"success": true, "result": ["识别文本"]}
```

### 流式对话
```
POST /api/chat/completions
- Content-Type: application/json
- Body: {"session_id": "xxx", "message": "用户输入"}

响应: SSE 流式输出
```

### 知识管理
```
# 保存知识
POST /api/knowledge/record
- Content-Type: multipart/form-data
- audio: 音频文件 (可选)
- text: 文本内容
- session_id: 会话ID (用于生成完整摘要)
- auto_organize: true/false

# 知识列表
GET /api/knowledge/list?category=分类

# 搜索知识
POST /api/knowledge/search
- Body: {"query": "搜索关键词"}
```

### 语音合成
```
POST /api/tts
- Content-Type: application/json
- Body: {"text": "要播报的文字"}

响应: audio/mp3 文件流
```

## 项目结构

```
backend/
├── cmd/                      # 主程序和工具
│   ├── main.go              # 服务入口
│   ├── migrate_titles/      # 标题迁移工具
│   └── restore_fenjiu/      # 数据恢复工具
├── internal/
│   ├── handler/             # HTTP 处理器
│   │   ├── chat_handler.go  # 对话处理
│   │   ├── knowledge_handler.go  # 知识管理
│   │   └── tts_handler.go   # 语音合成
│   ├── service/             # 业务服务
│   │   ├── rag_service.go        # RAG 检索服务
│   │   ├── knowledge_organizer.go # 知识整理
│   │   ├── session.go            # 会话管理
│   │   ├── intent.go             # 意图识别
│   │   ├── vector_store.go       # 向量存储
│   │   ├── embedding.go          # 向量化
│   │   ├── baidu_stt.go          # 百度STT
│   │   ├── baidu_tts.go          # 百度TTS
│   │   ├── glm_client.go         # GLM-4客户端
│   │   ├── context_compressor.go # 上下文压缩
│   │   └── database.go           # 数据库
│   ├── router/              # 路由
│   └── server/              # 服务器
├── static/                   # 前端静态文件（嵌入）
├── data/                     # 数据目录
│   ├── voice-memory.db     # SQLite 数据库
│   ├── audio/              # 音频文件
│   └── sessions/           # 会话备份
└── .env                     # 环境配置
```

## 技术栈

- **Go** - 后端语言
- **Gin** - HTTP Web 框架
- **SQLite** - 本地数据库
- **百度 STT** - 语音识别
- **百度 TTS** - 语音合成
- **智谱 GLM-4** - AI 对话
- **智谱 Embedding** - 向量化
- **SSE** - 流式响应

## 开发进度

### Phase 1: MVP + RAG (已完成)
- ✅ 语音识别 (百度 STT)
- ✅ AI 对话 (GLM-4)
- ✅ 知识整理 (标题、摘要、分类、标签)
- ✅ RAG 检索 (向量搜索 + 意图识别)
- ✅ 会话管理 (自动跟踪、上下文压缩)
- ✅ 语音播报 (百度 TTS)
- ✅ 前端嵌入 (静态文件集成)

### Phase 2: 规划中
- ⏳ 知识图谱
- ⏳ 多轮对话优化
- ⏳ 知识导出 (Markdown/JSON)
- ⏳ 用户认证
- ⏳ 数据同步 (云端备份)

## 数据管理

### 数据库位置
```
data/voice-memory.db  # SQLite 数据库
```

### 备份会话
```
data/sessions/sessions.json  # 会话历史备份
```

### 迁移工具
```bash
# 标题迁移 - 为历史知识生成 AI 标题
GLM_API_KEY=xxx ./migrate_titles ./data

# 数据恢复 - 恢复丢失的会话
./restore_fenjiu
```

## 测试

```bash
# 运行所有测试
go test ./internal/service/...

# 运行特定测试
go test ./internal/service -run TestKnowledgeOrganizer

# 查看覆盖率
go test ./internal/service/... -cover
```

## 环境要求

- Go 1.21+
- 有效 API Keys (百度 + 智谱)

## License

MIT
