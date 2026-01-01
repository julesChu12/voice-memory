# Phase 2 Implementation Tasks

## Milestone 1: Pipeline Skeleton (Day 1)
- [x] **Task 1.1**: Create `internal/pipeline` directory.
- [x] **Task 1.2**: Define `Context` and `Processor` interfaces in `pipeline/types.go`.
- [x] **Task 1.3**: Implement the `Orchestrator` logic in `pipeline/pipeline.go`.
- [x] **Task 1.4**: Write a simple Unit Test to verify the pipeline flow (Mock processors).

## Milestone 2: Processor Migration (Day 1-2)
- [x] **Task 2.1**: Implement `STTProcessor` (wrapping existing `BaiduSTT` service).
- [x] **Task 2.2**: Implement `IntentProcessor` (simple Regex for "stop/cancel").
- [x] **Task 2.3**: Implement `LLMProcessor` (wrapping existing `GLMClient` service).
- [x] **Task 2.4**: Implement `TTSProcessor` (wrapping existing `BaiduTTS` service).

## Milestone 3: WebSocket Layer (Day 2-3)
- [x] **Task 3.1**: Create `internal/handler/websocket_handler.go`.
- [x] **Task 3.2**: Implement the Connection Loop (Read Pump / Write Pump).
- [x] **Task 3.3**: Integrate `Pipeline.Execute()` into the WebSocket loop.
- [x] **Task 3.4**: Handle "Interrupt" signals (Context cancellation).

## Milestone 4: Frontend Integration (Day 3-4)
- [x] **Task 4.1**: Create `backend/static/ws-client.js` to manage WebSocket connection.
- [x] **Task 4.2**: Update `index.html` to use WebSocket instead of HTTP POST.
- [x] **Task 4.3**: Implement basic VAD (Voice Activity Detection) or "Push-to-Talk" logic to send audio chunks.
- [x] **Task 4.4**: Implement Streaming Audio Player (Note: Implemented using audio blob playback, not a raw PCM/WAV stream).

## Milestone 5: Verification (Day 5)
- [x] **Task 5.1**: Test Barge-in (Verified via frontend VAD silence detection and backend context cancellation).
- [x] **Task 5.2**: Test Latency (Verified via logs; optimized by disabling TTS for development).

---

## Phase 2 Summary
**Status:** Completed ✅
**Achievements:**
1. **Architecture:** Replaced HTTP with WebSocket Cognitive Pipeline.
2. **Communication:** Implemented full-duplex communication with barge-in support.
3. **Persistence:** Migrated SessionManager to SQLite.
4. **Prompt Engineering:** Upgraded Knowledge Organizer Prompt to v2.0 (Relationship-Enhanced).
5. **Frontend:** Implemented Voice/Text input and Real-time STT feedback.

**Next Steps (Phase 3):**
- Integrate `KnowledgeProcessor` into the pipeline.
- Implement RAG (Retrieval-Augmented Generation).
- Select and integrate Vector Database.
