# Phase 2 Technical Design Document

**Version:** 2.1 (Updated: Tech Stack Rationale)
**Date:** 2025-12-31
**Status:** Approved for Implementation

## 1. System Overview

Phase 2 transforms the backend from a linear REST API to a **WebSocket-based, Event-Driven Pipeline**.

### 1.1 Architecture Diagram

```mermaid
graph TD
    Client[Web PWA] <-->|WebSocket (Opus/PCM)| WSHandler[WebSocket Handler]
    
    subgraph "Go Backend (Memory Scope)"
        WSHandler -->|Event: Audio/Text| Pipeline[Pipeline Orchestrator]
        
        Pipeline -->|1. Raw Audio| VAD[VAD Processor]
        VAD -->|Voice Detected| STT[STT Processor]
        STT -->|Transcript| Intent[Intent Router]
        
        Intent --"Stop/Cancel"--> CmdExec[Command Executor]
        Intent --"Conversation"--> LLM[LLM Processor]
        
        LLM -->|Stream Text| TTS[TTS Processor]
        TTS -->|Stream Audio| Pipeline
        
        Pipeline -->|Output Event| WSHandler
    end
```

## 2. Technology Stack & Rationale

| Component | Technology | Rationale |
| :--- | :--- | :--- |
| **Protocol** | **WebSocket** (`github.com/gorilla/websocket`) | Low latency bidirectional comms for Barge-in. |
| **Pipeline** | **Native Go Channels** | **vs. Python Asyncio**: Go offers true parallelism (multi-core) essential for audio processing + networking without blocking. Go compiles to a single binary, solving deployment/dependency hell compared to Python. |
| **Web Server** | **Gin** | Existing stack, high performance, good middleware support. |
| **Audio** | **Pure Go** (Header parsing) + **FFmpeg** (if transcoding needed) | Avoid CGO stability issues. |
| **State** | **In-Memory Map** (`sync.Map`) | Zero-dependency, microsecond access time. |
| **LLM Adapter** | **SSE Client** (`bufio.Scanner`) | Universal compatibility (OpenAI/GLM/DeepSeek). |

## 3. Core Interface Definitions

### 3.1 The Pipeline Context
Data flowing through the pipe.

```go
package pipeline

import (
    "context"
    "voice-memory/internal/domain"
)

type Stage string

const (
    StageInput  Stage = "input"
    StageSTT    Stage = "stt"
    StageIntent Stage = "intent"
    StageLLM    Stage = "llm"
    StageTTS    Stage = "tts"
)

// PipelineContext carries data and control signals
type PipelineContext struct {
    Ctx       context.Context    // Go context for cancellation
    Cancel    context.CancelFunc // Function to stop the pipeline immediately
    SessionID string
    
    // Data slots (Thread-safe access required)
    InputAudio []byte
    Transcript string
    Intent     domain.Intent
    LLMReply   string
    OutputAudio []byte
}
```

### 3.2 The Processor Interface
Every building block implements this.

```go
type Processor interface {
    // Name returns the processor identifier
    Name() string
    
    // Process executes the logic. 
    // It returns true if the pipeline should continue, false if it should stop (short-circuit).
    Process(ctx *PipelineContext) (bool, error)
}
```

### 3.3 The Pipeline Orchestrator

```go
type Pipeline struct {
    processors []Processor
}

func (p *Pipeline) Execute(ctx *PipelineContext) error {
    for _, proc := range p.processors {
        // Check for cancellation before each step
        select {
        case <-ctx.Ctx.Done():
            return ctx.Ctx.Err()
        default:
            continueNext, err := proc.Process(ctx)
            if err != nil {
                return err
            }
            if !continueNext {
                return nil // Short-circuit (e.g., Intent = Stop)
            }
        }
    }
    return nil
}
```

## 4. WebSocket Protocol (JSON-based Control + Binary Audio)

To keep it simple, we use a single WebSocket connection sending Text (JSON) and Binary messages.

### 4.1 Client -> Server
*   **Binary Message**: Raw Audio Chunk (Int16 PCM / Float32).
*   **Text Message (JSON)**:
    ```json
    { "type": "config", "data": { "sample_rate": 16000 } }
    { "type": "interrupt" }  // User clicked "Stop" or VAD triggered
    ```

### 4.2 Server -> Client
*   **Binary Message**: TTS Audio Chunk.
*   **Text Message (JSON)**:
    ```json
    { "type": "stt_intermediate", "text": "I am..." }
    { "type": "stt_final", "text": "I am listening." }
    { "type": "llm_token", "text": "Hello" }
    { "type": "state", "status": "listening/thinking/speaking" }
    ```

## 5. Implementation Strategy (Step-by-Step)

1.  **Skeleton**: Create `internal/pipeline` and define Interfaces.
2.  **Websocket**: Create `internal/api/websocket.go` to handle connection upgrade and loop.
3.  **Migration**:
    *   Move Baidu STT logic into `internal/pipeline/processors/stt_baidu.go`.
    *   Move Intent logic (regex) into `internal/pipeline/processors/intent.go`.
    *   Move GLM logic into `internal/pipeline/processors/llm_glm.go`.
4.  **Integration**: Wire them up in `main.go`.

## 6. Future Proofing (Phase 3/4)
*   **Memory**: Just add `internal/pipeline/processors/memory.go` after Intent and before LLM.
*   **Tools**: Add logic inside `llm.go` to handle Function Call loops.