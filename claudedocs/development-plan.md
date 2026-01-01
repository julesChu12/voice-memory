# Voice Memory - Development Plan

**Created:** 2025-12-30
**Status:** Ready for Implementation
**Phase:** MVP (Phase 1) - Text-mediated voice interaction

---

## Current Implementation Status

### ✅ Completed Components

| Component | Status | Notes |
|-----------|--------|-------|
| Project Structure | ✅ Complete | Clean architecture with proper layering |
| Configuration Management | ✅ Complete | Environment-based config in `internal/config/` |
| Database Schema | ✅ Complete | Tables for sessions and knowledge defined |
| HTTP Routing | ✅ Complete | REST API endpoints with CORS configured |
| Service Layer Interfaces | ✅ Complete | Business logic organization established |
| Test Scripts | ✅ Complete | `test_simple.sh` and `test_session.sh` available |

### 🚧 Partial Implementation

| Component | Status | Missing |
|-----------|--------|---------|
| Main Application | 🚧 Structured | Entry point exists, needs full integration |
| Handlers | 🚧 Interfaces Only | Implementation needed for all handlers |
| GLM Client | 🚧 Stub Only | Full API integration required |
| STT Service | 🚧 Stub Only | Baidu API integration required |
| Database Operations | 🚧 Schema Only | CRUD operations need implementation |
| Knowledge Organizer | 🚧 Skeleton | AI-powered organization logic needed |

### ❌ Not Started

| Component | Priority | Notes |
|-----------|----------|-------|
| Frontend PWA | P0 | Only demo HTML exists |
| TTS Integration | P0 | Required for voice output |
| Knowledge Auto-organization | P1 | AI-powered summary/keypoints/tags |
| Build Scripts | P1 | No Makefile or build automation |
| Unit Tests | P2 | No test structure beyond shell scripts |

---

## Development Roadmap

### Phase 1: Core Backend Completion (Week 1)

#### 1.1 API Client Implementation (Days 1-2)

**Baidu STT Client** (`internal/service/baidu_stt.go`):
```go
// Required functions:
- GetAccessToken() -> Obtains Baidu API access token
- Recognize(audioData) -> Transcribes audio to text
- ValidateAudioFormat() -> Ensures supported format
```

**GLM Client** (`internal/service/glm_client.go`):
```go
// Required functions:
- Chat(messages) -> Sends conversation to GLM API
- StreamChat(messages) -> Optional: streaming responses
- FormatMessages() -> Converts internal format to GLM format
```

**Success Criteria:**
- Can obtain valid access token from Baidu
- Can transcribe audio file to text
- Can send message and receive AI reply

#### 1.2 Database Operations (Days 3-4)

**Session Store** (`internal/service/session.go`):
```go
// Required functions:
- CreateSession() -> Creates new conversation session
- AddMessage(sessionID, message) -> Appends message to session
- GetSession(sessionID) -> Retrieves session with messages
- ListSessions() -> Returns all sessions
- DeleteSession(sessionID) -> Removes session
```

**Knowledge Store** (`internal/service/knowledge_store.go`):
```go
// Required functions:
- SaveKnowledge(knowledge) -> Persists knowledge entry
- GetKnowledge(id) -> Retrieves single entry
- ListKnowledge(filters) -> Returns filtered list
- SearchKnowledge(query) -> Simple text search
- UpdateKnowledge(id, data) -> Modifies entry
```

**Success Criteria:**
- All CRUD operations work correctly
- Session-message relationships maintained
- Knowledge can be linked to sessions

#### 1.3 Handler Implementation (Days 5-7)

**Chat Handler** (`internal/handler/chat_handler.go`):
```go
// Required functions:
- HandleChat(c) -> POST /api/chat
  - Extract message and session_id from request
  - Call GLM client for AI response
  - Save both user message and AI reply to session
  - Return AI reply with session_id
```

**STT Handler** (`internal/handler/stt_handler.go`):
```go
// Required functions:
- Recognize(c) -> POST /api/stt
  - Extract audio file from form data
  - Call Baidu STT API
  - Return transcribed text
```

**Audio Chat Handler** (`internal/handler/audio_chat_handler.go`):
```go
// Required functions:
- HandleAudioChat(c) -> POST /api/audio-chat
  - Extract audio from request
  - STT -> Text -> LLM -> Text -> TTS pipeline
  - Return audio file with AI response
```

**Knowledge Handler** (`internal/handler/knowledge_handler.go`):
```go
// Required functions:
- HandleRecord(c) -> POST /api/knowledge/record
- HandleList(c) -> GET /api/knowledge/list
- HandleSearch(c) -> POST /api/knowledge/search
```

**Session Handler** (`internal/handler/session_handler.go`):
```go
// Required functions:
- HandleListSessions(c) -> GET /api/sessions
- HandleGetSession(c) -> GET /api/sessions/get
- HandleDeleteSession(c) -> DELETE /api/sessions
```

**Success Criteria:**
- All endpoints return valid responses
- Error handling works correctly
- Request/response formats match API spec

---

### Phase 2: Knowledge Organization (Week 2)

#### 2.1 AI-Powered Knowledge Organizer

**Knowledge Organizer** (`internal/service/knowledge_organizer.go`):
```go
// Required functions:
- OrganizeKnowledge(content, sessionID) -> {
    summary: "AI-generated summary",
    keyPoints: ["extracted", "key", "points"],
    category: "AI-classified category",
    tags: ["auto-generated", "tags"]
  }
```

**Implementation Notes:**
- Use GLM API for summarization
- Extract key points from content
- Classify into categories: "idea", "task", "reference", "decision"
- Generate relevant tags

**Success Criteria:**
- Knowledge is automatically organized when saved
- Summaries are concise and accurate
- Categories make sense for content type

---

### Phase 3: Frontend PWA (Week 3)

#### 3.1 Core UI Structure

**web/index.html**:
- Recording interface with microphone button
- Real-time transcription display
- Chat message history
- Knowledge browser panel
- Settings panel for API keys

**web/app.js**:
```javascript
// Required modules:
- AudioRecorder -> Handles microphone input
- APIClient -> Communicates with backend
- ChatUI -> Manages message display
- KnowledgeUI -> Displays knowledge entries
```

**Features:**
- Record audio and send to `/api/stt`
- Display transcription in real-time
- Send text to `/api/chat`
- Play AI response audio
- Browse and search knowledge

**Success Criteria:**
- Can record audio and see transcription
- Can have text conversation with AI
- Can view saved knowledge entries
- Works on mobile browsers (PWA)

---

### Phase 4: Integration & Testing (Week 4)

#### 4.1 End-to-End Testing

**Test Scenarios:**
1. **Voice Chat Flow**: Record → STT → Chat → TTS → Play
2. **Session Continuity**: Multiple messages in same session
3. **Knowledge Saving**: Auto-save from conversation
4. **Knowledge Search**: Find saved information
5. **Session Management**: List, view, delete sessions

**Performance Targets:**
- End-to-end latency: <3 seconds
- STT accuracy: >90%
- API response time: <500ms

#### 4.2 Deployment Preparation

**Deliverables:**
- Build script (Makefile or `build.sh`)
- Environment configuration template (`.env.example`)
- Production-ready main.go
- Basic deployment documentation

---

## Implementation Order Summary

```
Week 1: Backend Core
├─ Day 1-2: API Clients (Baidu STT, GLM)
├─ Day 3-4: Database Operations (Sessions, Knowledge)
└─ Day 5-7: HTTP Handlers (Chat, STT, Knowledge, Sessions)

Week 2: Knowledge Features
├─ Day 1-3: AI Knowledge Organizer
├─ Day 4-5: Knowledge Search & Filtering
└─ Day 6-7: Testing & Refinement

Week 3: Frontend PWA
├─ Day 1-2: Core UI Structure
├─ Day 3-4: Audio Recording & Playback
├─ Day 5-6: Chat Interface
└─ Day 7: Knowledge Browser

Week 4: Integration & Polish
├─ Day 1-2: End-to-End Integration
├─ Day 3-4: Testing & Bug Fixes
├─ Day 5-6: Performance Optimization
└─ Day 7: Deployment Preparation
```

---

## Next Steps (Immediate Actions)

1. **Start with API Clients**: Implement Baidu STT and GLM client first
2. **Test API Integration**: Verify external APIs work correctly
3. **Build Database Layer**: Create session and knowledge stores
4. **Implement Handlers**: Wire everything together with HTTP handlers
5. **Create Simple Frontend**: Basic HTML/JS for testing
6. **End-to-End Test**: Run through complete user journey
7. **Refine & Polish**: Improve UX and fix bugs

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Baidu API rate limits | Can't process audio | Implement queuing, fallback to other STT |
| GLM API downtime | No AI responses | Graceful error handling, clear user messages |
| Audio format issues | STT failures | Validate format early, provide clear error messages |
| Database corruption | Data loss | Regular backups, migration system |
| Browser compatibility | PWA doesn't work | Test on multiple browsers, progressive enhancement |

---

## Success Metrics

### Technical Metrics
- All API endpoints functional and tested
- End-to-end latency <3 seconds for voice chat
- STT accuracy >90% on clear speech
- Zero data loss (all sessions/knowledge saved)

### User Experience Metrics
- Can complete voice chat without errors
- Knowledge is correctly auto-organized
- Session history persists across app restarts
- Search returns relevant results

### Code Quality Metrics
- All handlers have proper error handling
- Database operations use transactions
- No TODO comments in production code
- Follow Go best practices throughout

---

## Documentation Updates

After implementation, update:
1. **README.md** - Add usage instructions
2. **API Design Document** - Document any API changes
3. **Deployment Guide** - Add actual deployment steps
4. **CLAUDE.md** - Update implementation status

---

**Ready to begin implementation! Start with Week 1, Day 1: API Client Implementation.**
