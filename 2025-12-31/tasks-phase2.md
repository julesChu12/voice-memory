# Phase 2 Implementation Tasks

## Milestone 1: Pipeline Skeleton (Day 1)
- [ ] **Task 1.1**: Create `internal/pipeline` directory.
- [ ] **Task 1.2**: Define `Context` and `Processor` interfaces in `pipeline/types.go`.
- [ ] **Task 1.3**: Implement the `Orchestrator` logic in `pipeline/pipeline.go`.
- [ ] **Task 1.4**: Write a simple Unit Test to verify the pipeline flow (Mock processors).

## Milestone 2: Processor Migration (Day 1-2)
- [ ] **Task 2.1**: Implement `STTProcessor` (wrapping existing `BaiduSTT` service).
- [ ] **Task 2.2**: Implement `IntentProcessor` (simple Regex for "stop/cancel").
- [ ] **Task 2.3**: Implement `LLMProcessor` (wrapping existing `GLMClient` service).
- [ ] **Task 2.4**: Implement `TTSProcessor` (wrapping existing `BaiduTTS` service).

## Milestone 3: WebSocket Layer (Day 2-3)
- [ ] **Task 3.1**: Create `internal/api/websocket.go`.
- [ ] **Task 3.2**: Implement the Connection Loop (Read Pump / Write Pump).
- [ ] **Task 3.3**: Integrate `Pipeline.Execute()` into the WebSocket loop.
- [ ] **Task 3.4**: Handle "Interrupt" signals (Context cancellation).

## Milestone 4: Frontend Integration (Day 3-4)
- [ ] **Task 4.1**: Create `web/ws-client.js` to manage WebSocket connection.
- [ ] **Task 4.2**: Update `index.html` to use WebSocket instead of HTTP POST.
- [ ] **Task 4.3**: Implement basic VAD (Voice Activity Detection) or "Push-to-Talk" logic to send audio chunks.
- [ ] **Task 4.4**: Implement Streaming Audio Player (PCM/WAV stream).

## Milestone 5: Verification (Day 5)
- [ ] **Task 5.1**: Test Barge-in (User speaks -> System stops).
- [ ] **Task 5.2**: Test Latency (Measure End-to-End time).
