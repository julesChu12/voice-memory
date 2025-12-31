# Voice Memory - API设计文档

**版本：** v1.0
**日期：** 2025-12-29
**状态：** 最终设计版本

---

## 目录

1. [API设计原则](#一api设计原则)
2. [基础规范](#二基础规范)
3. [核心API](#三核心api)
4. [记忆管理API](#四记忆管理api)
5. [WebSocket API](#五websocket-api)
6. [错误处理](#六错误处理)
7. [认证授权](#七认证授权)
8. [限流策略](#八限流策略)

---

## 一、API设计原则

### 1.1 设计理念

```yaml
RESTful设计:
  ✅ 资源导向的URL设计
  ✅ 标准HTTP方法（GET/POST/PUT/DELETE）
  ✅ 语义化的状态码
  ✅ 统一的响应格式

开发者友好:
  ✅ 清晰的错误信息
  ✅ 详细的API文档（OpenAPI/Swagger）
  ✅ 一致的命名规范
  ✅ 合理的默认值

性能优先:
  ✅ 异步处理（asyncio）
  ✅ 流式响应（streaming）
  ✅ 分页支持
  ✅ 缓存策略
```

### 1.2 URL设计规范

```
┌─────────────────────────────────────────────────────────────┐
│                    URL结构规范                               │
└─────────────────────────────────────────────────────────────┘

基础路径: /api/v1

资源命名: 复数名词
  ✅ /api/v1/memories
  ✅ /api/v1/conversations
  ❌ /api/v1/memory
  ❌ /api/v1/conversation

层级关系: 最多3层
  ✅ /api/v1/memories/{permalink}/observations
  ✅ /api/v1/conversations/{id}/messages
  ❌ /api/v1/memories/{permalink}/observations/{id}/relations

查询参数: snake_case
  ✅ /api/v1/memories?search_type=semantic
  ✅ /api/v1/memories?tags=ai,memory
  ❌ /api/v1/memories?searchType=semantic
```

---

## 二、基础规范

### 2.1 请求规范

#### 请求头

```http
# 标准请求头
Accept: application/json
Content-Type: application/json
Authorization: Bearer {access_token}
User-Agent: voice-memory-client/1.0.0

# 可选请求头
X-Request-ID: {unique_request_id}
X-Client-Version: 1.0.0
X-Device-ID: {unique_device_id}
```

#### 请求格式

```json
{
  "data": {
    // 业务数据
  },
  "meta": {
    // 元数据（可选）
  }
}
```

### 2.2 响应规范

#### 成功响应

```json
{
  "success": true,
  "data": {
    // 业务数据
  },
  "meta": {
    "request_id": "req_123456",
    "timestamp": "2025-12-29T12:00:00Z",
    "version": "1.0.0"
  }
}
```

#### 错误响应

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数验证失败",
    "details": [
      {
        "field": "message",
        "message": "消息内容不能为空"
      }
    ]
  },
  "meta": {
    "request_id": "req_123456",
    "timestamp": "2025-12-29T12:00:00Z"
  }
}
```

### 2.3 HTTP状态码

| 状态码 | 说明 | 使用场景 |
|--------|------|----------|
| **200 OK** | 请求成功 | GET/PUT/PATCH成功 |
| **201 Created** | 创建成功 | POST创建资源成功 |
| **204 No Content** | 无内容 | DELETE成功 |
| **400 Bad Request** | 请求错误 | 参数验证失败 |
| **401 Unauthorized** | 未授权 | Token无效/过期 |
| **403 Forbidden** | 禁止访问 | 权限不足 |
| **404 Not Found** | 资源不存在 | 资源未找到 |
| **409 Conflict** | 冲突 | 资源已存在 |
| **422 Unprocessable Entity** | 无法处理 | 业务逻辑验证失败 |
| **429 Too Many Requests** | 请求过多 | 触发限流 |
| **500 Internal Server Error** | 服务器错误 | 服务器内部错误 |
| **503 Service Unavailable** | 服务不可用 | 服务维护/过载 |

---

## 三、核心API

### 3.1 健康检查

#### GET /health

**描述：** 检查服务健康状态

**请求：** 无需参数

**响应：**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2025-12-29T12:00:00Z",
  "services": {
    "database": "ok",
    "ai_service": "ok",
    "storage": "ok"
  }
}
```

**状态码：** 200 OK

---

### 3.2 对话API

#### POST /api/v1/chat

**描述：** 单次对话（Phase 1：文本中介模式）

**请求体：**
```json
{
  "message": "帮我查一下今天的天气",
  "conversation_id": "conv_123456",  // 可选，用于上下文连续
  "options": {
    "use_local_stt": false,         // 是否使用本地STT
    "prefer_local_stt": false,      // 优先本地STT
    "model": "auto",                // auto, haiku, sonnet
    "stream": false,                // 是否流式响应（Phase 2）
    "save_memory": true             // 是否自动保存到记忆
  }
}
```

**请求参数验证：**
```yaml
message:
  type: string
  required: true
  min_length: 1
  max_length: 5000

conversation_id:
  type: string
  required: false
  pattern: ^[a-zA-Z0-9-]{1,50}$

options:
  type: object
  required: false
  properties:
    use_local_stt: boolean
    prefer_local_stt: boolean
    model: enum(auto, haiku, sonnet)
    stream: boolean
    save_memory: boolean
```

**响应（非流式）：**
```json
{
  "success": true,
  "data": {
    "response": "好的，我帮你查询今天北京的天气。今天北京晴转多云，气温-3到7度，空气质量良。",
    "conversation_id": "conv_123456",
    "message_id": "msg_789012",
    "usage": {
      "input_tokens": 150,
      "output_tokens": 200,
      "total_tokens": 350,
      "model": "claude-3-5-haiku-20241022",
      "estimated_cost": 0.0006
    },
    "transcription": {
      "text": "帮我查一下今天的天气",
      "confidence": 0.95,
      "language": "zh",
      "duration": 2.3
    },
    "memory_saved": {
      "saved": true,
      "permalink": "weather-query-2025-12-29",
      "auto_summarized": false
    },
    "timings": {
      "stt_duration_ms": 500,
      "ai_duration_ms": 1200,
      "tts_duration_ms": 800,
      "total_duration_ms": 2500
    }
  },
  "meta": {
    "request_id": "req_123456",
    "timestamp": "2025-12-29T12:00:00Z"
  }
}
```

**响应（流式 - Phase 2）：**
```http
HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

data: {"type": "transcription", "text": "帮我查一下今天的天气", "confidence": 0.95}

data: {"type": "ai_start", "model": "claude-3-5-haiku-20241022"}

data: {"type": "ai_chunk", "content": "好的，"}

data: {"type": "ai_chunk", "content": "我帮你查询"}

data: {"type": "ai_chunk", "content": "今天北京的天气。"}

data: {"type": "ai_end", "usage": {"input_tokens": 150, "output_tokens": 200}}

data: {"type": "tts_start", "voice": "alloy"}

data: {"type": "tts_chunk", "audio": "base64_encoded_audio_chunk"}

data: {"type": "tts_end", "duration_ms": 800}

data: {"type": "done", "total_duration_ms": 2500}
```

**状态码：**
- 200 OK: 成功
- 400 Bad Request: 参数错误
- 429 Too Many Requests: 超出速率限制
- 500 Internal Server Error: 服务错误

---

### 3.3 会话管理API

#### POST /api/v1/conversations

**描述：** 创建新会话

**请求体：**
```json
{
  "title": "关于AI架构的讨论",  // 可选，自动生成
  "system_prompt": "你是一个专业的技术顾问",  // 可选
  "metadata": {
    "source": "voice",
    "device": "macos"
  }
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "conversation_id": "conv_123456",
    "title": "关于AI架构的讨论",
    "created_at": "2025-12-29T12:00:00Z",
    "message_count": 0
  }
}
```

---

#### GET /api/v1/conversations

**描述：** 获取会话列表

**查询参数：**
```
?page=1                    // 页码，默认1
?limit=20                  // 每页数量，默认20，最大100
?sort=created_at           // 排序字段：created_at, updated_at, title
?order=desc                // 排序方向：asc, desc
?search=关键词              // 搜索标题
```

**响应：**
```json
{
  "success": true,
  "data": {
    "conversations": [
      {
        "conversation_id": "conv_123456",
        "title": "关于AI架构的讨论",
        "created_at": "2025-12-29T12:00:00Z",
        "updated_at": "2025-12-29T12:30:00Z",
        "message_count": 15,
        "last_message": "我们下次继续讨论"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "total_pages": 3
    }
  }
}
```

---

#### GET /api/v1/conversations/{conversation_id}

**描述：** 获取会话详情

**路径参数：**
- `conversation_id`: 会话ID

**响应：**
```json
{
  "success": true,
  "data": {
    "conversation_id": "conv_123456",
    "title": "关于AI架构的讨论",
    "system_prompt": "你是一个专业的技术顾问",
    "created_at": "2025-12-29T12:00:00Z",
    "updated_at": "2025-12-29T12:30:00Z",
    "messages": [
      {
        "message_id": "msg_001",
        "role": "user",
        "content": "我们讨论一下微服务架构",
        "timestamp": "2025-12-29T12:00:00Z"
      },
      {
        "message_id": "msg_002",
        "role": "assistant",
        "content": "好的，微服务架构有以下几个关键点...",
        "timestamp": "2025-12-29T12:00:05Z",
        "usage": {
          "input_tokens": 20,
          "output_tokens": 150
        }
      }
    ],
    "metadata": {
      "source": "voice",
      "device": "macos",
      "total_messages": 15,
      "total_tokens": 3500
    }
  }
}
```

---

#### DELETE /api/v1/conversations/{conversation_id}

**描述：** 删除会话

**路径参数：**
- `conversation_id`: 会话ID

**响应：**
```json
{
  "success": true,
  "data": {
    "deleted": true,
    "conversation_id": "conv_123456"
  }
}
```

**状态码：**
- 204 No Content: 删除成功
- 404 Not Found: 会话不存在

---

### 3.4 语音处理API

#### POST /api/v1/audio/transcribe

**描述：** 语音转文字（STT）

**请求：**
- Content-Type: `multipart/form-data`
- Body: 音频文件

**表单参数：**
```
audio: 音频文件（mp3, wav, m4a, ogg）
language: 语言代码（zh, en），默认auto
model: 模型选择（whisper, whisper-local），默认whisper
timestamp: 是否返回时间戳，默认false
```

**响应：**
```json
{
  "success": true,
  "data": {
    "text": "帮我查一下今天的天气",
    "language": "zh",
    "confidence": 0.95,
    "duration": 2.3,
    "segments": [
      {
        "text": "帮我查一下",
        "start": 0.0,
        "end": 1.2,
        "confidence": 0.96
      },
      {
        "text": "今天的天气",
        "start": 1.2,
        "end": 2.3,
        "confidence": 0.94
      }
    ],
    "model": "whisper",
    "processing_time_ms": 500
  }
}
```

---

#### POST /api/v1/audio/synthesize

**描述：** 文字转语音（TTS）

**请求体：**
```json
{
  "text": "好的，我帮你查询今天北京的天气",
  "voice": "alloy",          // alloy, echo, fable, onyx, nova, shimmer
  "speed": 1.0,              // 0.25 - 4.0，默认1.0
  "format": "mp3",           // mp3, opus, aac, flac
  "stream": false            // 是否流式返回
}
```

**响应（非流式）：**
```json
{
  "success": true,
  "data": {
    "audio": "base64_encoded_audio_data",
    "format": "mp3",
    "duration": 3.2,
    "size_bytes": 25600,
    "voice": "alloy",
    "processing_time_ms": 800
  }
}
```

**响应（流式）：**
```http
HTTP/1.1 200 OK
Content-Type: audio/mpeg
Transfer-Encoding: chunked

[Audio binary data]
```

---

## 四、记忆管理API

### 4.1 记忆CRUD

#### POST /api/v1/memories

**描述：** 保存记忆（笔记）

**请求体：**
```json
{
  "title": "笔记标题",
  "content": "笔记内容...",
  "type": "note",             // note, observation, idea
  "tags": ["ai", "architecture"],
  "categories": ["技术笔记"],
  "relations": [
    {
      "target_permalink": "related-note",
      "relation_type": "relates_to",
      "context": "相关讨论"
    }
  ],
  "metadata": {
    "source": "conversation",
    "conversation_id": "conv_123456"
  }
}
```

**请求参数验证：**
```yaml
title:
  type: string
  required: true
  min_length: 1
  max_length: 200

content:
  type: string
  required: true
  min_length: 1
  max_length: 50000

type:
  type: string
  required: false
  enum: [note, observation, idea]
  default: note

tags:
  type: array
  required: false
  max_items: 10
  items:
    type: string
    pattern: ^[a-zA-Z0-9-_]{1,30}$

relations:
  type: array
  required: false
  max_items: 20
```

**响应：**
```json
{
  "success": true,
  "data": {
    "permalink": "note-permalink-123",
    "title": "笔记标题",
    "content": "笔记内容...",
    "type": "note",
    "tags": ["ai", "architecture"],
    "categories": ["技术笔记"],
    "created_at": "2025-12-29T12:00:00Z",
    "updated_at": "2025-12-29T12:00:00Z",
    "observations": [
      {
        "id": 1,
        "category": "技术要点",
        "content": "微服务架构的关键考虑",
        "tags": ["microservices"],
        "created_at": "2025-12-29T12:00:00Z"
      }
    ],
    "relations": [
      {
        "id": 1,
        "target_permalink": "related-note",
        "target_title": "相关笔记",
        "relation_type": "relates_to",
        "context": "相关讨论"
      }
    ]
  }
}
```

---

#### GET /api/v1/memories/{permalink}

**描述：** 获取单个记忆

**路径参数：**
- `permalink`: 记忆永久链接

**响应：**
```json
{
  "success": true,
  "data": {
    "permalink": "note-permalink-123",
    "title": "笔记标题",
    "content": "笔记内容...",
    "type": "note",
    "tags": ["ai", "architecture"],
    "categories": ["技术笔记"],
    "created_at": "2025-12-29T12:00:00Z",
    "updated_at": "2025-12-29T12:00:00Z",
    "markdown_path": "/storage/markdown/2025-12-29/note-permalink-123.md"
  }
}
```

---

#### GET /api/v1/memories

**描述：** 获取记忆列表

**查询参数：**
```
?page=1                    // 页码
?limit=20                  // 每页数量
?type=note                 // 过滤类型
?tags=ai,architecture      // 过滤标签（逗号分隔）
?search=关键词              // 搜索标题和内容
?sort=created_at           // 排序字段
?order=desc                // 排序方向
```

**响应：**
```json
{
  "success": true,
  "data": {
    "memories": [
      {
        "permalink": "note-1",
        "title": "微服务架构设计",
        "type": "note",
        "tags": ["architecture", "microservices"],
        "created_at": "2025-12-29T12:00:00Z",
        "updated_at": "2025-12-29T12:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "total_pages": 3
    }
  }
}
```

---

#### PUT /api/v1/memories/{permalink}

**描述：** 更新记忆

**路径参数：**
- `permalink`: 记忆永久链接

**请求体：** 同POST

**响应：** 同POST响应

---

#### DELETE /api/v1/memories/{permalink}

**描述：** 删除记忆

**路径参数：**
- `permalink`: 记忆永久链接

**响应：**
```json
{
  "success": true,
  "data": {
    "deleted": true,
    "permalink": "note-permalink-123"
  }
}
```

---

### 4.2 记忆搜索API

#### GET /api/v1/memories/search

**描述：** 搜索记忆（语义搜索）

**查询参数：**
```
?q=搜索查询                // 搜索关键词
?search_type=semantic      // 搜索类型：keyword, semantic, hybrid
?limit=10                  // 返回数量，默认10，最大50
?type=note                 // 过滤类型
?tags=ai,architecture      // 过滤标签
```

**响应（语义搜索）：**
```json
{
  "success": true,
  "data": {
    "query": "微服务架构的最佳实践",
    "search_type": "semantic",
    "results": [
      {
        "permalink": "microservices-best-practices",
        "title": "微服务架构最佳实践",
        "type": "note",
        "similarity": 0.92,
        "highlighted_content": "微服务架构的**最佳实践**包括...",
        "tags": ["architecture", "microservices"],
        "created_at": "2025-12-29T12:00:00Z"
      },
      {
        "permalink": "api-design-patterns",
        "title": "API设计模式",
        "type": "note",
        "similarity": 0.85,
        "highlighted_content": "在**微服务**架构中，API设计...",
        "tags": ["api", "architecture"],
        "created_at": "2025-12-28T15:30:00Z"
      }
    ],
    "total": 2,
    "search_time_ms": 150
  }
}
```

---

### 4.3 观察记录API

#### POST /api/v1/memories/{permalink}/observations

**描述：** 添加观察记录到笔记

**路径参数：**
- `permalink`: 记忆永久链接

**请求体：**
```json
{
  "category": "技术要点",
  "content": "微服务需要考虑服务发现和配置管理",
  "tags": ["microservices", "config"],
  "context": "从讨论中总结"
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": 123,
    "entity_id": 456,
    "category": "技术要点",
    "content": "微服务需要考虑服务发现和配置管理",
    "tags": ["microservices", "config"],
    "context": "从讨论中总结",
    "created_at": "2025-12-29T12:00:00Z"
  }
}
```

---

### 4.4 关系管理API

#### POST /api/v1/memories/{permalink}/relations

**描述：** 添加笔记关系

**路径参数：**
- `permalink`: 源记忆永久链接

**请求体：**
```json
{
  "target_permalink": "related-note",
  "relation_type": "relates_to",    // relates_to, builds_on, contradicts, extends
  "context": "相关讨论"
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": 789,
    "source_entity_id": 456,
    "target_entity_id": 789,
    "target_permalink": "related-note",
    "target_title": "相关笔记",
    "relation_type": "relates_to",
    "context": "相关讨论",
    "created_at": "2025-12-29T12:00:00Z"
  }
}
```

---

## 五、WebSocket API

### 5.1 实时对话WebSocket

#### WebSocket /api/v1/ws/chat

**描述：** 实时对话WebSocket（Phase 2: 可打断交互）

**连接参数：**
```
?conversation_id=conv_123456  // 可选，会话ID
?token=access_token            // 认证令牌
```

**消息格式：**

客户端 → 服务器：
```json
{
  "type": "audio_chunk",
  "data": "base64_encoded_audio",
  "sequence": 1,
  "timestamp": "2025-12-29T12:00:00Z"
}
```

```json
{
  "type": "user_interrupt",
  "timestamp": "2025-12-29T12:00:00Z"
}
```

服务器 → 客户端：
```json
{
  "type": "transcription",
  "text": "帮我查一下天气",
  "is_final": true,
  "confidence": 0.95
}
```

```json
{
  "type": "ai_start",
  "model": "claude-3-5-haiku-20241022"
}
```

```json
{
  "type": "ai_chunk",
  "content": "好的，我帮你",
  "sequence": 1
}
```

```json
{
  "type": "ai_end",
  "usage": {
    "input_tokens": 150,
    "output_tokens": 200
  }
}
```

```json
{
  "type": "tts_start",
  "voice": "alloy"
}
```

```json
{
  "type": "tts_chunk",
  "audio": "base64_encoded_audio",
  "sequence": 1
}
```

```json
{
  "type": "tts_end",
  "duration_ms": 800
}
```

```json
{
  "type": "interrupted",
  "reason": "user_spoke"
}
```

```json
{
  "type": "error",
  "code": "STT_ERROR",
  "message": "语音识别失败"
}
```

**状态码：**
- 101 Switching Protocols: 连接成功
- 401 Unauthorized: 认证失败
- 429 Too Many Requests: 连接数超限

---

### 5.2 状态同步WebSocket

#### WebSocket /api/v1/ws/sync

**描述：** 多端状态同步

**连接参数：**
```
?token=access_token
?device_id=device_123
```

**消息格式：**

客户端 → 服务器：
```json
{
  "type": "subscribe",
  "channels": ["memories", "conversations"]
}
```

服务器 → 客户端：
```json
{
  "type": "memory_created",
  "data": {
    "permalink": "new-note",
    "title": "新笔记",
    "created_at": "2025-12-29T12:00:00Z"
  }
}
```

```json
{
  "type": "memory_updated",
  "data": {
    "permalink": "updated-note",
    "updated_at": "2025-12-29T12:05:00Z"
  }
}
```

```json
{
  "type": "conversation_updated",
  "data": {
    "conversation_id": "conv_123456",
    "last_message": "最新消息",
    "updated_at": "2025-12-29T12:05:00Z"
  }
}
```

---

## 六、错误处理

### 6.1 错误码规范

```yaml
# 通用错误
UNKNOWN_ERROR: 未知错误
VALIDATION_ERROR: 参数验证失败
NOT_FOUND: 资源不存在
CONFLICT: 资源冲突
RATE_LIMIT_EXCEEDED: 超出速率限制
SERVER_ERROR: 服务器错误

# 认证授权
UNAUTHORIZED: 未授权
INVALID_TOKEN: Token无效
TOKEN_EXPIRED: Token过期
INSUFFICIENT_PERMISSIONS: 权限不足

# AI服务
AI_SERVICE_ERROR: AI服务错误
AI_RATE_LIMIT: AI服务速率限制
AI_TIMEOUT: AI服务超时
AI_INVALID_RESPONSE: AI响应无效

# 语音服务
STT_ERROR: 语音识别错误
TTS_ERROR: 语音合成错误
INVALID_AUDIO_FORMAT: 音频格式不支持
AUDIO_TOO_SHORT: 音频过短
AUDIO_TOO_LONG: 音频过长

# 记忆管理
MEMORY_NOT_FOUND: 记忆不存在
DUPLICATE_PERMALINK: Permalink重复
INVALID_RELATION: 无效关系
CIRCULAR_RELATION: 循环关系

# 会话管理
CONVERSATION_NOT_FOUND: 会话不存在
MESSAGE_NOT_FOUND: 消息不存在
CONTEXT_TOO_LONG: 上下文过长
```

### 6.2 错误响应示例

**参数验证错误：**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数验证失败",
    "details": [
      {
        "field": "message",
        "message": "消息内容不能为空",
        "constraint": "min_length=1"
      },
      {
        "field": "options.model",
        "message": "模型选择无效",
        "constraint": "enum(auto,haiku,sonnet)"
      }
    ]
  },
  "meta": {
    "request_id": "req_123456",
    "timestamp": "2025-12-29T12:00:00Z"
  }
}
```

**AI服务错误：**
```json
{
  "success": false,
  "error": {
    "code": "AI_SERVICE_ERROR",
    "message": "AI服务暂时不可用",
    "details": {
      "provider": "anthropic",
      "original_error": "Rate limit exceeded",
      "retry_after": 60
    }
  },
  "meta": {
    "request_id": "req_123456",
    "timestamp": "2025-12-29T12:00:00Z"
  }
}
```

---

## 七、认证授权

### 7.1 认证方式

**Phase 1 (MVP):** 简单Token认证

```yaml
认证流程:
  1. 客户端使用API Key
  2. 在请求头中携带: Authorization: Bearer {api_key}
  3. 服务器验证Token有效性

Token获取:
  - 在配置文件中设置
  - 或通过环境变量
```

**Phase 2+:** JWT Token认证

```yaml
认证流程:
  1. 用户登录获取JWT Token
  2. Token包含: user_id, exp, iat
  3. 每次请求携带Token
  4. Token过期后刷新

Token结构:
  header: {"alg": "HS256", "typ": "JWT"}
  payload: {"user_id": "123", "exp": 1735689600}
  signature: HMACSHA256(secret, header + payload)
```

### 7.2 权限控制

```yaml
资源访问权限:
  memories:
    create: own_only
    read: own_only
    update: own_only
    delete: own_only

  conversations:
    create: authenticated
    read: own_only
    update: own_only
    delete: own_only

  admin:
    access: admin_only
```

---

## 八、限流策略

### 8.1 速率限制

```yaml
# 基于IP的限流
IP-based:
  window: 1分钟
  limits:
    default: 60 requests/minute
    chat: 20 requests/minute
    audio: 10 requests/minute

# 基于用户的限流
User-based:
  window: 1小时
  limits:
    free_tier: 1000 requests/hour
    pro_tier: 10000 requests/hour

# 基于API Key的限流
API Key-based:
  window: 1天
  limits:
    tokens: 1000000 tokens/day
    audio: 1000 requests/day
```

### 8.2 限流响应

**响应头：**
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1735689660
```

**错误响应：**
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "请求过于频繁，请稍后再试",
    "details": {
      "limit": 100,
      "remaining": 0,
      "reset_at": "2025-12-29T12:01:00Z",
      "retry_after": 30
    }
  }
}
```

---

## 九、API版本管理

### 9.1 版本策略

```yaml
URL版本控制:
  - 当前版本: /api/v1/
  - 下一版本: /api/v2/
  - 旧版本支持: 至少6个月

向后兼容:
  - 新增字段不影响旧客户端
  - 废弃字段至少提前3个月通知
  - 重大变更需要新版本
```

### 9.2 废弃通知

**响应头（废弃API）：**
```http
Deprecation: true
Sunset: Sun, 29 Jun 2026 12:00:00 GMT
Link: </api/v2/endpoint>; rel="successor-version"
```

**响应体：**
```json
{
  "success": true,
  "data": {...},
  "warnings": [
    {
      "code": "DEPRECATED_API",
      "message": "此API已废弃，将于2026年6月29日停用",
      "migration_guide": "https://docs.example.com/migration-v1-to-v2"
    }
  ]
}
```

---

## 十、OpenAPI规范

### 10.1 基础信息

```yaml
openapi: 3.0.3
info:
  title: Voice Memory API
  version: 1.0.0
  description: 可打断的实时语音AI助手API
  contact:
    name: API Support
    email: support@voicememory.com

servers:
  - url: https://api.voicememory.com/api/v1
    description: Production
  - url: https://staging-api.voicememory.com/api/v1
    description: Staging
  - url: http://localhost:8000/api/v1
    description: Development
```

### 10.2 安全定义

```yaml
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - BearerAuth: []
```

---

**文档版本历史：**
- v1.0 (2025-12-29): 初始版本，完整API设计
