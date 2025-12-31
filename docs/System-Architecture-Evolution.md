# 系统架构演进与现状概览

**日期**: 2025-12-31
**视角**: 全局架构视角 (Interaction -> Memory)

---

## 1. 架构分层全景图

Voice Memory 的系统架构分为六层，每一层在不同阶段有不同的成熟度。

```
┌──────────────────────────┐
│ 1. Interaction Layer     │  Input/Output, VAD, AEC
└───────────────▲──────────┘
                │
┌───────────────┴──────────┐
│ 2. Session & Context     │  Conversation History, State Machine
└───────────────▲──────────┘
                │
┌───────────────┴──────────┐
│ 3. Cognitive Middleware   │  Router, Intent, RAG (新增核心层)
└───────────────▲──────────┘
                │
┌───────────────┴──────────┐
│ 4. Agent Runtime          │  LLM Caller, Task Planner
└───────────────▲──────────┘
                │
┌───────────────┴──────────┐
│ 5. Memory & Knowledge     │  Vector DB, Knowledge Graph
└───────────────▲──────────┘
                │
┌───────────────┴──────────┐
│ 6. Tools & Infrastructure │  APIs, Docker, Deployment
└───────────────▲──────────┘
```

---

## 2. 分阶段演进状态 (Status Matrix)

### 🟢 已完成 (MVP) | 🟡 进行中 (Phase 2) | ⚪️ 规划中 (Phase 3)

| 架构层级 | MVP (Walking Skeleton) <br> *现状: 能跑通，体验差* | Phase 2 (Robust Core) <br> *目标: 好用，可打断，有记性* | Phase 3 (Advanced) <br> *目标: 智能，多模态* |
| :--- | :--- | :--- | :--- |
| **1. Interaction** | 🟢 **Web/CLI**<br>- 简单的录音上传<br>- 无 VAD，无打断 | 🟡 **PWA + Streaming**<br>- WebRTC VAD (本地检测)<br>- 浏览器级 AEC (回声消除)<br>- 全双工流式传输 | ⚪️ **Native / Realtime**<br>- 端到端语音模型<br>- 毫秒级延迟 |
| **2. Session** | 🟢 **Linear List**<br>- 简单的数组存储对话<br>- 容易混淆上下文 | 🟡 **FSM (有限状态机)**<br>- 明确状态 (Listening/Speaking)<br>- 处理打断信号 | ⚪️ **Dynamic Context**<br>- 自动压缩历史上下文<br>- 多话题管理 |
| **3. Middleware** | ❌ **None**<br>- STT 直连 LLM<br>- 无法处理指令 | 🟡 **Cognitive Router**<br>- 正则意图识别 (停止/取消)<br>- 记忆检索路由 | ⚪️ **Agent Planner**<br>- 复杂任务规划 (LangGraph)<br>- 多步推理 |
| **4. Runtime** | 🟢 **Simple API**<br>- 每次调用一次 LLM | 🟡 **Dynamic Prompting**<br>- 注入时间、记忆<br>- 动态控制回复长度 | ⚪️ **Tool Use**<br>- 自动调用搜索、日历等工具 |
| **5. Memory** | 🟢 **File/SQLite**<br>- 仅存文本<br>- 仅支持关键词搜索 | 🟡 **Vector DB (Qdrant)**<br>- 语义向量索引<br>- 长期记忆检索 (RAG) | ⚪️ **Knowledge Graph**<br>- 知识图谱 (实体关系)<br>- 自动归纳整理 |
| **6. Tools** | 🟢 **Local File**<br>- 本地读写 | 🟡 **Docker**<br>- 容器化部署<br>- 服务编排 | ⚪️ **External APIs**<br>- 联网搜索、邮件集成 |

---

## 3. 当前阶段 (Phase 2) 重点攻坚

根据上述分析，Phase 2 的核心任务集中在 **中间三层** 的重构：

1.  **Interaction (Layer 1)**: 实现 **VAD + AEC**，支持打断信号发送。
2.  **Session (Layer 2)**: 在后端实现 **FSM**，处理打断逻辑。
3.  **Middleware (Layer 3)**: 搭建 **Pipeline**，引入 **Qdrant** 实现记忆检索。

---

## 4. 决策记录

- **2025-12-31**: 确认 Phase 2 放弃小程序 RTC 方案，专注于 **PWA (Progressive Web App)** 以确保最佳的 Layer 1 体验。
- **2025-12-31**: 确认引入 **"中间计算层"**，在 STT 和 LLM 之间增加意图识别和 RAG 检索。
- **2025-12-31**: 确认 Layer 5 选用 **Qdrant** 作为向量数据库。
