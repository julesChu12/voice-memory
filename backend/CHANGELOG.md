# Changelog

All notable changes to Voice Memory will be documented in this file.

## [0.2.0] - 2024-12-31

### Added
- **RAG (Retrieval-Augmented Generation)** - 智谱 AI 向量搜索
  - 意图识别：自动判断普通对话 vs 知识检索
  - 向量化存储：知识自动同步到向量数据库
  - 语义搜索：智能检索相关知识

- **会话管理**
  - 自动跟踪对话历史
  - 上下文压缩：Token 超限时自动摘要会话
  - 会话持久化：保存到 sessions.json

- **知识整理增强**
  - AI 生成标题：从完整对话生成 5-15 字标题
  - AI 生成摘要：基于会话上下文生成结构化摘要
  - 自动分类：技术/生活/工作/学习/想法
  - 标签提取：智能提取关键词

- **语音播报 (TTS)**
  - 百度 TTS 集成
  - 自动播放 AI 回复
  - 流式音频支持

- **前端优化**
  - 静态文件嵌入 Go 服务
  - 保存后自动重置会话
  - 输入防重复提交
  - 显示 AI 生成的标题

- **测试覆盖**
  - 知识整理器测试
  - 意图识别测试
  - GLM 客户端测试
  - 会话管理测试

- **迁移工具**
  - `migrate_titles`: 为历史知识生成 AI 标题
  - `restore_fenjiu`: 恢复丢失的会话数据

### Fixed
- Nil pointer in chat context compression
- GLM-4 API 不支持 system role（仅 user/assistant）
- 数据库迁移时的 duplicate column name 错误
- NULL title 处理（使用 COALESCE）
- JSON 解析失败（AI 返回 markdown 代码块）

### Changed
- 会话摘要现在使用完整对话内容，而非单条输入
- 知识列表显示 AI 生成的标题

## [0.1.0] - 2024-12-30

### Added
- 语音识别（百度 STT）
- AI 对话（智谱 GLM-4）
- 知识保存和检索
- 基础 Web 界面
- SQLite 数据库存储

---

**Note:** Version 0.2.0 represents Phase 1 completion with full RAG capabilities.
