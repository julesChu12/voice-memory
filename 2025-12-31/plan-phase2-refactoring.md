# Phase 2 Kickoff & Architecture Refactoring Plan

## Decisions
- **Client Strategy:** Shift focus to **Web PWA** as the primary interface. Cancel Native/Mini-program plans for now.
- **Tech Stack:** Confirm usage of Baidu (STT/TTS) and GLM (LLM) for localization.
- **Action:** Skip MVP polish; start backend refactoring immediately.

## Refactoring Plan: Cognitive Pipeline

### 1. Goal
Transform the current `Handler -> Service` linear call into a modular `Pipeline` architecture.

### 2. New Directory Structure
```
backend/internal/
├── pipeline/           # [NEW] Core pipeline logic
│   ├── pipeline.go     # Orchestrator
│   ├── context.go      # Request/Session context
│   └── processors/     # Middleware components
│       ├── stt.go
│       ├── intent.go   # "Stop", "System Command"
│       ├── memory.go   # RAG/Vector search
│       ├── llm.go
│       └── tts.go
```

### 3. Core Interface
```go
type Context struct {
    SessionID string
    AudioData []byte
    Text      string
    Intent    IntentType
    // ...
}

type Processor interface {
    Process(ctx *Context) error
    Name() string
}
```

### 4. Implementation Steps
1. **Define Interfaces:** Create `pipeline/types.go`.
2. **Migrate STT:** Move Baidu STT logic into a processor.
3. **Implement Intent:** Add simple Regex matching (e.g., "Stop").
4. **Migrate LLM/TTS:** Wrap existing services into processors.
5. **Wire up:** Replace `AudioChatHandler` logic with `Pipeline.Execute()`.
