# Phase 2: 认知路由与中间计算层设计

**日期**: 2025-12-31
**状态**: 方案确认，待开发
**背景**: 在 MVP 阶段，STT 文本直接透传给 LLM 导致体验生硬（无法处理指令、无记忆上下文）。Phase 2 引入“中间计算层” (Cognitive Middleware) 解决此问题。

---

## 1. 核心问题

当前 (MVP) 的线性流程存在以下缺陷：
1.  **无视意图**: 用户说“停止”，LLM 可能会回答“好的，我停下来”，而不是直接停止 TTS。
2.  **无视背景**: LLM 无法获取历史记忆，导致对话缺乏连续性。
3.  **无视时间**: LLM 缺乏当前时间上下文。

## 2. 解决方案：认知流水线 (Cognitive Pipeline)

在 STT 和 LLM 之间引入一个“认知路由层”，将处理流程从“直通车”改为“可编排的流水线”。

### 2.1 架构图

```mermaid
graph TD
    User[用户输入 Audio] --> STT[STT 语音转文本]
    STT --> Text[原始文本]
    
    subgraph "中间计算层 (Cognitive Middleware)"
        Text --> PreProcessor{1. 预处理 & 意图识别}
        
        %% 分支 A: 快速指令 (Fast Path)
        PreProcessor -- "命中指令 (停止/大声点)" --> FastAction[⚡️ 执行本地指令]
        FastAction --> End((流程结束))
        
        %% 分支 B: 正常对话 (Slow Path)
        PreProcessor -- "正常对话" --> Memory{2. 记忆检索 (RAG)}
        
        Memory -- "无相关记忆" --> Assembler[3. Prompt 组装]
        Memory -- "命中向量索引" --> Context[检索到的记忆片段]
        Context --> Assembler
        
        Assembler --> FinalPrompt[最终 Prompt: 角色 + 时间 + 记忆 + 问题]
    end
    
    FinalPrompt --> LLM[LLM 大模型]
    LLM --> TTS[TTS 语音合成]
```

---

## 3. 详细模块设计

### 3.1 预处理与意图识别 (Pre-processor)

**目标**: 建立“脊髓反射”，拦截不需要 LLM 处理的高频指令。
**实现策略**: **规则/正则匹配 (Regex)**
**优势**: 0 延迟，零成本，绝对可靠。

| 意图类型 | 关键词示例 | 动作 |
| :--- | :--- | :--- |
| **Stop** | "停止", "别说了", "闭嘴", "停" | 立即终止 TTS，清空播放队列 |
| **Cancel** | "取消", "算了" | 停止当前生成任务 |
| **System** | "声音大点", "几点了" | (可选) 调用本地系统 API 或简单回复 |

### 3.2 记忆检索 (Memory Retriever)

**目标**: 让 AI 拥有长期记忆。
**技术选型**: **Qdrant (向量数据库)**
**流程**:
1.  **Embedding**: 调用 OpenAI `text-embedding-3-small` 将用户文本转为向量。
2.  **Search**: 在 Qdrant 中搜索 Cosine Distance 最近的 Top 3 记录。
3.  **Threshold**: 设定相似度阈值 (e.g., > 0.7)，低于阈值则认为无相关记忆。

### 3.3 Prompt 组装器 (Assembler)

**目标**: 动态构建 System Prompt。
**模板结构**:

```text
[System]
Current Time: 2025-12-31 10:30:00 (注入当前时间)
Context: (如果检索到记忆，在此插入)
- 2025-12-29: 用户提到装修方案V2，核心是无主灯设计...

[Instruction]
你是一个可打断的语音助手。
- 如果包含 Context，请结合 Context 回答。
- 保持回答简练，除非用户要求展开。
```

---

## 4. 方案对比与决策

### 4.1 意图识别方案

| 方案 | 延迟 | 成本 | 复杂度 | 决策 |
| :--- | :--- | :--- | :--- | :--- |
| **规则匹配 (Regex)** | < 1ms | 0 | 低 | **✅ Phase 2 采用** |
| **Embedding 匹配** | ~200ms | 低 | 中 | Phase 3 考虑 |
| **LLM Function Call** | > 1s | 高 | 高 | 仅用于复杂任务 |

### 4.2 记忆检索方案

| 方案 | 语义理解 | 部署难度 | 决策 |
| :--- | :--- | :--- | :--- |
| **关键词搜索 (SQL)** | 差 | 低 | ❌ |
| **向量搜索 (Qdrant)** | 优 | 中 (Docker) | **✅ Phase 2 采用** |

---

## 5. 下一步行动

1.  **Docker 环境**: 部署 Qdrant 容器。
2.  **后端重构**: 将 Golang 的处理逻辑重构为 Pipeline 模式。
3.  **Prompt 优化**: 设计支持动态 Context 注入的 System Prompt。
