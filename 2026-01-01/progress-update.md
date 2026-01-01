# Project Progress Update (2026-01-01)

## 📅 Timeline
**Date:** January 1, 2026
**Current Phase:** Phase 3 (Memory & Knowledge Enhancement)

---

## ✅ Completed Achievements

### Phase 2: Architecture & Interaction (Finished)
1.  **WebSocket Architecture**: Successfully migrated from HTTP to full-duplex WebSocket (`/ws`), enabling real-time interaction.
2.  **Cognitive Pipeline**: Implemented a modular pipeline (STT -> Intent -> LLM) supporting text/voice inputs.
3.  **Frontend Upgrade**:
    *   Added **Text Input** support (hybrid chat interface).
    *   Implemented **Real-time STT Feedback** (preview -> final bubble).
    *   Improved UI with **Markdown Rendering** (via `marked.js`) and polished styling.
4.  **LLM Optimization**:
    *   Integrated **System Prompt** for better persona control.
    *   Optimized parameters (`Temperature=0.5`, `TopP=0.8`) for accuracy.
    *   Fixed API compatibility (Anthropic-style `system` field).
    *   Implemented **Pre-generation Persistence** (User inputs saved before LLM call to prevent data loss).
5.  **Dev Experience**:
    *   Added `air` for hot-reloading.
    *   Added `test_real_audio` script for end-to-end testing.
    *   Added detailed console logs for debugging.

### Phase 3: Memory & Knowledge (In Progress)
1.  **Milestone 1: Knowledge Pipeline Integration (Done)**
    *   Created `KnowledgeProcessor` to asynchronously extract knowledge from conversations.
    *   Integrated `KnowledgeOrganizer` (Prompt v2.0) into the pipeline.
    *   Verified knowledge extraction and SQLite storage via `check_knowledge` script.
2.  **Milestone 2: Vector Database (Partial)**
    *   Defined `VectorStore` and `EmbeddingService` interfaces.
    *   Implemented `SimpleVectorStore` (Pure Go, in-memory + JSON persistence) for MVP.
    *   Implemented `EmbeddingClient` for Zhipu AI `embedding-2` model.

---

## 🚧 Pending / Blocked Tasks

1.  **Knowledge Vectorization (Paused)**:
    *   Connecting `KnowledgeProcessor` to `VectorStore` is pending.
    *   *Reason*: User requested to switch focus to Local ASR/TTS integration first.

---

## 📋 Next Steps (Immediate Priorities)

1.  **Local ASR/TTS Integration**:
    *   **Goal**: Replace Baidu Cloud APIs with local services (e.g., FunASR, GPT-SoVITS) to reduce latency and cost.
    *   **Action**: Create `LocalSTT` and `LocalTTS` service implementations once API specs are provided.
    *   **Config**: Update `config` and `server` to support service switching.

2.  **Resume Phase 3 Tasks**:
    *   Complete `KnowledgeProcessor` vectorization logic.
    *   Implement RAG (Retrieval) logic in `IntentProcessor`.

## 📝 Notes
- `.env` file handling has been resolved (ignored locally, removed from remote).
- System is stable and testable via `air`.
