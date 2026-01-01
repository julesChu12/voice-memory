# Phase 3 Implementation Tasks: Memory & Knowledge Enhancement

**Goal:** Transform "Voice Chat" into "Voice Memory" by enabling automatic knowledge extraction, long-term storage, and retrieval-augmented generation (RAG).

## Milestone 1: Knowledge Pipeline Integration (Day 1)
- [ ] **Task 1.1**: Define `KnowledgeProcessor` struct implementing the `Processor` interface.
- [ ] **Task 1.2**: Implement async execution logic (extract knowledge *after* or *parallel to* LLM response).
- [ ] **Task 1.3**: Integrate `KnowledgeOrganizer` (v2.0) into the processor to extract entities/facts.
- [ ] **Task 1.4**: Update `Pipeline` to include `KnowledgeProcessor`.

## Milestone 2: Vector Database & Storage (Day 2)
- [ ] **Task 2.1**: Finalize Vector DB decision (ChromaDB service vs. Pure Go solution).
- [ ] **Task 2.2**: Define `VectorStore` interface in `internal/service/interfaces.go`.
- [ ] **Task 2.3**: Implement the chosen Vector Store adapter.
- [ ] **Task 2.4**: Update `KnowledgeService` to save extracted knowledge embeddings to Vector Store.

## Milestone 3: RAG Implementation (Day 3)
- [ ] **Task 3.1**: Update `IntentProcessor` to reliably detect `IntentSearch` vs `IntentChat`.
- [ ] **Task 3.2**: Implement `RetrievalService` to query Vector Store based on user query.
- [ ] **Task 3.3**: Update `LLMProcessor` to inject retrieved context into the System Prompt.
- [ ] **Task 3.4**: Verify RAG loop (Ask -> Search -> Generate -> Reply).

## Milestone 4: Frontend Experience (Day 4)
- [ ] **Task 4.1**: Implement Markdown rendering for LLM responses (e.g., using `marked.js`).
- [ ] **Task 4.2**: Simulate "Typing" effect for text responses (improve visual pacing).
- [ ] **Task 4.3**: Add UI for "Memory" visualization (e.g., "I remembered this..." toast notification).

## Milestone 5: System Polish (Day 5)
- [ ] **Task 5.1**: Performance tuning (async knowledge extraction shouldn't affect chat latency).
- [ ] **Task 5.2**: Token usage optimization (summary compression).
