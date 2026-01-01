# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Voice Memory** is an interruptible real-time voice AI assistant. The project aims to create natural, bidirectional conversations with AI where users can interrupt the AI at any time, similar to talking to a human.

**Current Status**: MVP in development - text-mediated voice interaction (STT -> LLM -> TTS)

---

## Communication Language (CRITICAL)

**All responses MUST be in Chinese (中文)**

- All user-facing messages must be in Chinese
- All code comments should use Chinese
- All documentation must be in Chinese
- Only exception: English technical terms and API endpoints

---

## Development Commands

### Running the Backend
```bash
cd backend
go run cmd/main.go
```

### Testing
```bash
# Test simple chat functionality
cd backend && ./test_simple.sh

# Test session and knowledge association
cd backend && ./test_session.sh
```

### Environment Setup
Required environment variables (set before running):
```bash
export BAIDU_API_KEY=your_api_key
export BAIDU_SECRET_KEY=your_secret_key
export GLM_API_KEY=your_glm_api_key
export SERVER_PORT=8080  # optional, defaults to 8080
```

---

## Architecture

### Clean Architecture Pattern
```
backend/
├── cmd/
│   ├── main.go              # Application entry point
│   └── migrate/main.go      # Database migration tool
├── internal/
│   ├── config/              # Configuration management (env vars)
│   ├── handler/             # HTTP request handlers (presentation layer)
│   ├── router/              # Route definitions with CORS
│   ├── server/              # Server initialization and lifecycle
│   └── service/             # Business logic layer
│       ├── database.go      # SQLite database operations
│       ├── session.go       # Session management
│       ├── knowledge_store.go  # Knowledge CRUD operations
│       ├── knowledge_organizer.go  # AI-powered knowledge organization
│       ├── knowledge_types.go     # Knowledge data models
│       ├── baidu_stt.go     # Baidu STT API client
│       └── glm_client.go    # Zhipu GLM API client (Claude-compatible)
├── go.mod
├── test_simple.sh           # Basic API tests
└── test_session.sh          # Session association tests
```

### Technology Stack
- **Backend**: Golang + Gin framework
- **Database**: SQLite3
- **AI Services**: Baidu STT API, Zhipu GLM API (Claude-compatible)
- **Frontend**: Web PWA (HTML5 + Vanilla JS + TailwindCSS) - located in `web/`

### API Endpoints
```
POST   /api/stt              # Speech-to-text recognition
POST   /api/chat             # Text chat with AI
POST   /api/audio-chat       # Audio-to-audio chat
POST   /api/knowledge/record  # Save knowledge with auto-organization
GET    /api/knowledge/list    # List all knowledge
POST   /api/knowledge/search  # Search knowledge
GET    /api/sessions          # List all sessions
GET    /api/sessions/get      # Get specific session
DELETE /api/sessions          # Delete session
GET    /health                # Health check
```

### Data Models

**Session** - Conversation continuity:
```go
type Session struct {
    ID        string    `json:"id"`
    Messages  []Message `json:"messages"`  // role + content
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**Knowledge** - AI-organized memory storage:
```go
type Knowledge struct {
    ID        string            `json:"id"`
    Content   string            `json:"content"`
    Summary   string            `json:"summary"`        // AI-generated
    KeyPoints []string          `json:"key_points"`     // AI-extracted
    Category  string            `json:"category"`       // AI-classified
    Tags      []string          `json:"tags"`           // AI-generated
    SessionID string            `json:"session_id"`     // Links to conversation
    CreatedAt time.Time         `json:"created_at"`
    Metadata  map[string]string `json:"metadata"`
}
```

---

## Critical Development Rules

### Core Principles (CRITICAL - Never Compromise)
```
Evidence > assumptions | Code > documentation | Efficiency > verbosity
```

### Implementation Completeness
- **No partial features**: If you start implementing, MUST complete to working state
- **No TODO comments**: Never leave TODO for core functionality
- **No mock objects**: No placeholders, fake data, or stub implementations
- **No incomplete functions**: Every function must work as specified

### Git Workflow (CRITICAL)
- **Feature branches ONLY**: Create feature branches for ALL work, never work on main/master
- **Frequent commits**: Commit frequently with meaningful messages, avoid "fix", "update", "changes"
- **Always check status**: Start every session with `git status` and `git branch`
- **Review before commit**: Always `git diff` to review changes before staging

### Scope Discipline
- **Build ONLY what's asked**: No adding features beyond explicit requirements
- **MVP first**: Start with minimum viable solution, iterate based on feedback
- **No enterprise bloat**: No auth, deployment, monitoring unless explicitly requested
- **Think before build**: Understand → Plan → Build, not Build → Build more

### Code Standards
- Follow `Effective Go` best practices
- Use `gofmt` for formatting
- Package names: lowercase, no underscores
- Exported functions: uppercase first letter
- Never ignore error handling

### Task Management Pattern
For tasks with >3 steps:
1. **Understand** → Read existing code, understand patterns
2. **Plan** → Use TodoWrite to track progress
3. **Execute** → Implement with parallel operations where possible
4. **Validate** → Test before marking complete
5. **Track** → Update todos as you complete them

---

## MVP Scope Boundaries

### MUST Implement (Phase 1)
- ✅ Voice input → Baidu STT API → text
- ✅ Text → Zhipu GLM API → AI reply
- ✅ AI reply → TTS API → voice output
- ✅ Basic memory storage (SQLite)
- ✅ Simple text search
- ✅ Session-based conversation continuity
- ✅ Knowledge auto-organization (summary, key points, tags)

### NOT for MVP (Phase 2+)
- ❌ Barge-in/interruptibility functionality
- ❌ Local STT/TTS models
- ❌ Wake word detection
- ❌ Vector semantic search
- ❌ Real-time streaming STT/TTS

---

## Key Documentation

| Document | Path | Purpose |
|----------|------|---------|
| AI Implementation Guide | `04-Golang实现/AI项目启动指南_Go.md` | Primary development reference |
| Tech Stack Decisions | `04-Golang实现/技术栈决策文档_Go.md` | Technology choices rationale |
| API Design | `04-Golang实现/API设计文档_Go.md` | Interface specifications |
| Data Models | `04-Golang实现/数据模型文档_Go.md` | Database schema |
| Deployment | `04-Golang实现/部署与运维文档_Go.md` | Deployment guide |
| Product Requirements | `01-产品/产品文档.md` | Feature definitions |
| Code Examples | `04-Golang实现/核心代码示例_Go.md` | Reference implementations |

---

## Quality Standards

### Code Quality
- **SOLID Principles**: Single responsibility, Open/closed, Liskov substitution, Interface segregation, Dependency inversion
- **DRY**: Don't Repeat Yourself - abstract common functionality
- **KISS**: Keep It Simple - prefer simplicity over complexity
- **YAGNI**: You Ain't Gonna Need It - implement current requirements only

### Professional Standards
- **No marketing language**: No "blazingly fast", "100% secure", "magnificent"
- **Evidence-based claims**: All technical claims must be verifiable through testing or documentation
- **Critical assessment**: Provide honest trade-offs and potential issues
- **Realistic assessments**: State "untested", "MVP", "needs validation"

### Testing Standards
- **Never skip tests**: Never disable, comment out, or skip tests to achieve results
- **Never skip validation**: Never bypass quality checks or validation to make things work
- **Root cause analysis**: Always investigate WHY failures occur, not just that they failed

---

## Knowledge Archiving

Project-related outputs should be archived to:
```
knowledge/archive/projects/voice-memory/
```

### Session Management Pattern
- **Load**: Use `/sc:load` to resume previous session context
- **Work**: Implement features with TodoWrite tracking
- **Checkpoint**: Save progress every 30 minutes or after major tasks
- **Save**: Use `/sc:save` to persist session learnings

---

## File Organization Standards

### Test Files
Place all tests in `backend/test/` or use `*_test.go` naming convention

### Scripts
Place utility scripts in `backend/scripts/` or `tools/` directories

### Documentation
- Project docs: `04-Golang实现/` directory
- Claude-specific: `claudedocs/` directory for reports/analyses

### Separation of Concerns
- Keep tests, scripts, docs, and source code properly separated
- Organize by feature/domain, not file type

---

## Development Workflow

### Before Starting Work
1. Check git status and create feature branch
2. Read existing code to understand patterns
3. Create TodoWrite for tasks >3 steps
4. Verify environment variables are set

### During Implementation
1. Follow existing code patterns and conventions
2. Complete each feature fully before moving to next
3. Test incrementally, don't wait until end
4. Update todos as you complete tasks

### After Completing Work
1. Run all tests to verify nothing broke
2. Run lint/typecheck if available
3. Review changes with git diff
4. Commit with meaningful message describing WHAT and WHY
5. Delete temporary files and clean workspace

---

## Product Vision Context

### Target Users
- **Knowledge workers**: Capture fragmented thoughts, systematic organization
- **Researchers**: Cross-domain knowledge connections via semantic graph
- **Creators**: Capture inspiration before it's lost
- **Tech practitioners**: Project context preservation, Terminal integration
- **Privacy-sensitive users**: Local-first, fully self-hosted

### Core Value Proposition
> "Like talking to a person - natural, interruptible, persistent memory"

Differentiation from competitors:
- Voice-first design (not text-with-voice-addon)
- Silent intelligent archiving (AI auto-identifies valuable content)
- Bidirectional knowledge collaboration (human + AI co-edit knowledge base)
- Local-first privacy control

### Evolution Path
- **Phase 1 (MVP)**: STT → Text → LLM → Text → TTS, 2-3s latency
- **Phase 2**: Interruptible interaction, streaming STT/TTS, <1s response
- **Phase 3**: End-to-end voice models, <500ms latency


任务：为用户生成一个个人知识库 AI 助手 MVP（最小可运行版本），目标是文本/语音交互、知识管理、隐私优先。

前提：
1. 用户是高度重视隐私的人，每个用户有独立知识空间。
2. MVP 只支持单知识空间，多知识空间以后扩展。
3. 知识上传必须手动上传文件或文本。
4. 知识归档由 Agent 对话分析建议，用户确认归档，归档可压缩。
5. 日常对话不压缩 RAG 向量，保证对话连贯。
6. 用户可以手动删除知识，不允许自动删除。
7. MVP 阶段支持文本交互，语音可选。
8. 本地部署默认信任，云端部署需要用户空间隔离。
9. Agent 执行安全，不执行任意 shell 或系统代码。
10. 多层 Memory 简化版：Working / Episodic / Knowledge。

---

模块与职责：
1. Interaction Layer
   - 接收用户输入（文本/语音）
   - 输出结果（文本/语音）
   - MVP 可先实现文本交互
2. Session & Context Layer
   - 管理会话与上下文
   - 上下文滑窗 / 历史摘要
3. Agent Runtime & Task Controller
   - 管理 Task 生命周期（running / paused / cancelled）
   - 检查新 Task 与当前 Task 的关联性
   - 决定中断、重排或独立执行
4. Agent Cognitive Layer
   - 生成计划 / 执行步骤（LangGraph 或 Planner）
   - 选择合适的工具 / Memory 调用
5. Memory & Knowledge Layer
   - Working Memory：当前对话上下文
   - Episodic Memory：对话摘要，存储对话历史
   - Knowledge Memory：用户归档文档及向量索引
6. Tools & Capabilities Layer
   - 支持文档搜索 / 索引 / 查询
7. Feedback Layer
   - 用户显式反馈（thumbs up/down 或归档确认）
8. 安全与权限
   - 本地部署：默认信任
   - 云端部署：用户空间隔离
   - 禁止 Agent 执行 shell / system 命令

---

数据结构：
1. Task：
{
  task_id: string,
  state: "running" | "paused" | "cancelled",
  graph_node: string,
  created_at: timestamp
}

2. Memory：
{
  working_memory: [{step_id, content, timestamp}],
  episodic_memory: [{session_id, summary, timestamp}],
  knowledge_memory: [{doc_id, vector, metadata}]
}

3. Session / Context：
{
  session_id: string,
  context_buffer: [...],
  last_updated: timestamp
}

4. Feedback：
{
  user_id: string,
  task_id: string,
  feedback_type: "thumbs_up" | "thumbs_down" | "archive_confirm",
  timestamp: timestamp
}

---

MVP 开发优先级：
1. Knowledge Space + Upload + Archive + Knowledge Memory
2. Agent Runtime + Task Controller + Cognitive Layer Planner
3. Working / Episodic Memory + 用户反馈
4. Interaction Layer（文本交互，语音可选）
5. 安全 / 权限（本地默认信任，云端隔离）
6. 观测 / 运维 /自动评估可暂不实现

---

生成要求：
1. 输出一个可运行的最小 MVP 项目结构和代码模板，包含所有模块 skeleton。
2. 每个模块要有清晰职责和接口。
3. 数据结构必须实现，支持最小操作：上传知识、对话、归档、查询。
4. 支持任务中断 / 重排 / 多层 Memory。
5. 安全与权限遵循前述规则。
6. 输出格式可以是文件结构 + Python 或 Node.js 伪代码模板 + 数据结构定义。
7. 额外提供 README，说明模块边界、运行方式及可扩展点。
