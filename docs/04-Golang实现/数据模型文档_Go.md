# Voice Memory - 数据模型文档 (Golang版本)

**版本：** v1.0-Go
**日期：** 2025-12-29
**语言：** Golang

---

## 目录

1. [数据模型概述](#一数据模型概述)
2. [数据库Schema](#二数据库schema)
3. [Go数据结构](#三go数据结构)
4. [向量存储](#四向量存储)
5. [数据迁移](#五数据迁移)
6. [查询优化](#六查询优化)

---

## 一、数据模型概述

### 1.1 核心实体

```
┌─────────────────────────────────────────────────────────────┐
│                      Voice Memory 数据模型                    │
└─────────────────────────────────────────────────────────────┘

Conversation (对话)
    ├── id: 主键
    ├── summary: 摘要
    ├── created_at: 创建时间
    └── Messages[]: 消息列表
        ├── id: 主键
        ├── role: user/assistant
        ├── content: 内容
        └── timestamp: 时间戳

Memory (记忆)
    ├── id: 主键
    ├── content: 内容
    ├── embedding: 向量 (1536维)
    ├── created_at: 创建时间
    ├── updated_at: 更新时间
    └── metadata: 元数据 (JSON)
```

### 1.2 数据关系

```
Conversation 1 ── N Message
    │
    └── 关联: conversation_id

Memory (独立)
    └── 通过语义搜索关联到对话
```

---

## 二、数据库Schema

### 2.1 conversations 表

```sql
CREATE TABLE IF NOT EXISTS conversations (
    id TEXT PRIMARY KEY,
    summary TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_conversations_created_at ON conversations(created_at DESC);
```

### 2.2 messages 表

```sql
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    tokens INTEGER DEFAULT 0,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_timestamp ON messages(timestamp DESC);
```

### 2.3 memories 表

```sql
CREATE TABLE IF NOT EXISTS memories (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    embedding BLOB,  -- 存储为二进制，1536个float32
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT,  -- JSON字符串
    source TEXT,    -- 来源: chat, manual, import
    is_archived BOOLEAN DEFAULT 0
);

CREATE INDEX idx_memories_created_at ON memories(created_at DESC);
CREATE INDEX idx_memories_source ON memories(source);

-- 全文搜索虚拟表 (使用FTS5)
CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
    content,
    content_rowid=memories ROWID,
    tokenize='porter unicode61'
);
```

### 2.4 memory_vectors 表 (向量搜索)

```sql
-- 使用 sqlite-vss 扩展
CREATE VIRTUAL TABLE IF NOT EXISTS memory_vectors USING vss0(
    id TEXT PRIMARY KEY,
    embedding(1536)  -- Claude使用1536维向量
);
```

### 2.5 tags 表 (标签系统)

```sql
CREATE TABLE IF NOT EXISTS tags (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    color TEXT DEFAULT '#3B82F6',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS memory_tags (
    memory_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    PRIMARY KEY (memory_id, tag_id),
    FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE INDEX idx_memory_tags_memory_id ON memory_tags(memory_id);
CREATE INDEX idx_memory_tags_tag_id ON memory_tags(tag_id);
```

---

## 三、Go数据结构

### 3.1 Conversation (对话)

```go
// internal/memory/models.go
package memory

import (
    "time"
)

// Conversation 对话
type Conversation struct {
    ID        string    `json:"id" db:"id"`
    Summary   string    `json:"summary" db:"summary"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
    Messages  []Message `json:"messages,omitempty"`
}

// NewConversation 创建新对话
func NewConversation() *Conversation {
    return &Conversation{
        ID:        generateID("conv"),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}
```

### 3.2 Message (消息)

```go
// Message 消息
type Message struct {
    ID             string    `json:"id" db:"id"`
    ConversationID string    `json:"conversation_id" db:"conversation_id"`
    Role           string    `json:"role" db:"role"` // user, assistant, system
    Content        string    `json:"content" db:"content"`
    Timestamp      time.Time `json:"timestamp" db:"timestamp"`
    Tokens         int       `json:"tokens" db:"tokens"`
}

// NewMessage 创建新消息
func NewMessage(conversationID, role, content string, tokens int) *Message {
    return &Message{
        ID:             generateID("msg"),
        ConversationID: conversationID,
        Role:           role,
        Content:        content,
        Timestamp:      time.Now(),
        Tokens:         tokens,
    }
}
```

### 3.3 Memory (记忆)

```go
// Memory 记忆
type Memory struct {
    ID         string                 `json:"id" db:"id"`
    Content    string                 `json:"content" db:"content"`
    Embedding  []float32              `json:"-" db:"embedding"` // 不序列化到JSON
    CreatedAt  time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
    Metadata   map[string]interface{} `json:"metadata" db:"metadata"`
    Source     string                 `json:"source" db:"source"` // chat, manual, import
    IsArchived bool                   `json:"is_archived" db:"is_archived"`
    Tags       []Tag                  `json:"tags,omitempty"`
}

// NewMemory 创建新记忆
func NewMemory(content string, source string) *Memory {
    return &Memory{
        ID:         generateID("mem"),
        Content:    content,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
        Metadata:   make(map[string]interface{}),
        Source:     source,
        IsArchived: false,
    }
}

// SetEmbedding 设置向量
func (m *Memory) SetEmbedding(embedding []float32) {
    m.Embedding = embedding
}

// GetEmbedding 获取向量（字节格式）
func (m *Memory) GetEmbeddingBytes() []byte {
    if len(m.Embedding) == 0 {
        return nil
    }

    // 将[]float32转换为[]byte
    buf := make([]byte, len(m.Embedding)*4)
    for i, v := range m.Embedding {
        // 使用binary包更规范
        binary.LittleEndian.PutUint32(buf[i*4:(i+1)*4], math.Float32bits(v))
    }
    return buf
}

// SetEmbeddingBytes 从字节设置向量
func (m *Memory) SetEmbeddingBytes(data []byte) {
    if len(data)%4 != 0 {
        return
    }

    m.Embedding = make([]float32, len(data)/4)
    for i := 0; i < len(m.Embedding); i++ {
        bits := binary.LittleEndian.Uint32(data[i*4 : (i+1)*4])
        m.Embedding[i] = math.Float32frombits(bits)
    }
}
```

### 3.4 Tag (标签)

```go
// Tag 标签
type Tag struct {
    ID        string    `json:"id" db:"id"`
    Name      string    `json:"name" db:"name"`
    Color     string    `json:"color" db:"color"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// NewTag 创建新标签
func NewTag(name, color string) *Tag {
    if color == "" {
        color = "#3B82F6" // 默认蓝色
    }
    return &Tag{
        ID:        generateID("tag"),
        Name:      name,
        Color:     color,
        CreatedAt: time.Now(),
    }
}
```

### 3.5 辅助函数

```go
// internal/memory/utils.go
package memory

import (
    "crypto/rand"
    "encoding/hex"
    "time"
)

// generateID 生成唯一ID
func generateID(prefix string) string {
    b := make([]byte, 8)
    rand.Read(b)
    return prefix + "_" + hex.EncodeToString(b)
}

// MetadataString 将metadata序列化为JSON字符串
func (m *Memory) MetadataString() string {
    if m.Metadata == nil {
        return "{}"
    }
    data, _ := json.Marshal(m.Metadata)
    return string(data)
}

// ParseMetadata 从JSON字符串解析metadata
func (m *Memory) ParseMetadata(data string) error {
    if data == "" || data == "{}" {
        m.Metadata = make(map[string]interface{})
        return nil
    }
    return json.Unmarshal([]byte(data), &m.Metadata)
}
```

---

## 四、向量存储

### 4.1 Embedding生成

```go
// internal/memory/embedding.go
package memory

import (
    "encoding/json"
    "fmt"
    "net/http"
)

// EmbeddingClient 向量生成客户端
type EmbeddingClient struct {
    apiKey  string
    client  *http.Client
}

// NewEmbeddingClient 创建embedding客户端
func NewEmbeddingClient(apiKey string) *EmbeddingClient {
    return &EmbeddingClient{
        apiKey: apiKey,
        client: &http.Client{},
    }
}

// GenerateEmbedding 生成文本向量
func (e *EmbeddingClient) GenerateEmbedding(text string) ([]float32, error) {
    // 调用Claude的embedding API
    // 注意: Claude可能不直接提供embedding API
    // 可能需要使用OpenAI的embedding API或本地模型

    reqBody := map[string]interface{}{
        "model": "text-embedding-ada-002",
        "input": text,
    }

    body, _ := json.Marshal(reqBody)
    req, _ := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+e.apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := e.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Data []struct {
            Embedding []float32 `json:"embedding"`
        } `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    if len(result.Data) == 0 {
        return nil, fmt.Errorf("no embedding returned")
    }

    return result.Data[0].Embedding, nil
}
```

### 4.2 向量存储操作

```go
// internal/memory/vector.go
package memory

import (
    "database/sql"
    "encoding/json"
)

// SaveEmbedding 保存向量到数据库
func (s *Storage) SaveEmbedding(id string, embedding []float32) error {
    // 1. 保存到memories表
    mem, err := s.GetMemory(id)
    if err != nil {
        return err
    }

    mem.SetEmbedding(embedding)
    memBytes := mem.GetEmbeddingBytes()

    _, err = s.db.Exec(
        "UPDATE memories SET embedding = ? WHERE id = ?",
        memBytes, id,
    )

    if err != nil {
        return err
    }

    // 2. 保存到memory_vectors虚拟表
    embeddingJSON, _ := json.Marshal(embedding)
    _, err = s.db.Exec(
        "INSERT INTO memory_vectors (id, embedding) VALUES (?, ?)",
        id, string(embeddingJSON),
    )

    return err
}

// SearchByVector 向量搜索
func (s *Storage) SearchByVector(queryEmbedding []float32, limit int) ([]*Memory, error) {
    embeddingJSON, _ := json.Marshal(queryEmbedding)

    query := `
        SELECT m.id, m.content, m.created_at, m.updated_at, m.metadata, m.source
        FROM memories m
        INNER JOIN memory_vectors v ON m.id = v.id
        WHERE v.embedding MATCH ?
        ORDER BY distance
        LIMIT ?
    `

    rows, err := s.db.Query(query, string(embeddingJSON), limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var memories []*Memory
    for rows.Next() {
        var mem Memory
        var metadataStr string

        err := rows.Scan(
            &mem.ID,
            &mem.Content,
            &mem.CreatedAt,
            &mem.UpdatedAt,
            &metadataStr,
            &mem.Source,
        )

        if err != nil {
            return nil, err
        }

        mem.ParseMetadata(metadataStr)
        memories = append(memories, &mem)
    }

    return memories, nil
}

// SemanticSearch 语义搜索（使用FTS5 + 向量）
func (s *Storage) SemanticSearch(query string, limit int) ([]*Memory, error) {
    // 1. 先用全文搜索获取候选
    ftsQuery := `
        SELECT m.id, m.content
        FROM memories m
        INNER JOIN memories_fts fts ON m.id = fts.rowid
        WHERE memories_fts MATCH ?
        LIMIT ?
    `

    rows, err := s.db.Query(ftsQuery, query, limit*5)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var candidates []*Memory
    for rows.Next() {
        var mem Memory
        if err := rows.Scan(&mem.ID, &mem.Content); err != nil {
            continue
        }
        candidates = append(candidates, &mem)
    }

    // 2. 生成query的embedding
    embedding, err := s.embeddingClient.GenerateEmbedding(query)
    if err != nil {
        // 如果embedding失败，直接返回FTS结果
        return candidates, nil
    }

    // 3. 向量搜索重新排序
    // (这里简化实现，实际应该计算cosine similarity)
    return candidates[:min(len(candidates), limit)], nil
}
```

---

## 五、数据迁移

### 5.1 迁移系统

```go
// internal/memory/migrate.go
package memory

import (
    "database/sql"
    "fmt"
)

type Migration struct {
    Version int
    Name    string
    Up      func(*sql.DB) error
    Down    func(*sql.DB) error
}

// Migrations 所有迁移
var Migrations = []Migration{
    {
        Version: 1,
        Name:    "initial_schema",
        Up:      migrateV1Up,
        Down:    migrateV1Down,
    },
    {
        Version: 2,
        Name:    "add_tags",
        Up:      migrateV2Up,
        Down:    migrateV2Down,
    },
}

// migrateV1Up V1迁移 - 创建初始表
func migrateV1Up(db *sql.DB) error {
    statements := []string{
        `CREATE TABLE IF NOT EXISTS conversations (
            id TEXT PRIMARY KEY,
            summary TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
        `CREATE TABLE IF NOT EXISTS messages (
            id TEXT PRIMARY KEY,
            conversation_id TEXT NOT NULL,
            role TEXT NOT NULL,
            content TEXT NOT NULL,
            timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
            tokens INTEGER DEFAULT 0,
            FOREIGN KEY (conversation_id) REFERENCES conversations(id)
        )`,
        `CREATE TABLE IF NOT EXISTS memories (
            id TEXT PRIMARY KEY,
            content TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            metadata TEXT,
            source TEXT
        )`,
    }

    for _, stmt := range statements {
        if _, err := db.Exec(stmt); err != nil {
            return fmt.Errorf("migration failed: %w", err)
        }
    }

    return nil
}

// migrateV1Down V1回滚
func migrateV1Down(db *sql.DB) error {
    statements := []string{
        "DROP TABLE IF EXISTS messages",
        "DROP TABLE IF EXISTS conversations",
        "DROP TABLE IF EXISTS memories",
    }

    for _, stmt := range statements {
        if _, err := db.Exec(stmt); err != nil {
            return err
        }
    }

    return nil
}

// migrateV2Up V2迁移 - 添加标签系统
func migrateV2Up(db *sql.DB) error {
    statements := []string{
        `CREATE TABLE IF NOT EXISTS tags (
            id TEXT PRIMARY KEY,
            name TEXT UNIQUE NOT NULL,
            color TEXT DEFAULT '#3B82F6',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
        `CREATE TABLE IF NOT EXISTS memory_tags (
            memory_id TEXT NOT NULL,
            tag_id TEXT NOT NULL,
            PRIMARY KEY (memory_id, tag_id),
            FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE,
            FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
        )`,
    }

    for _, stmt := range statements {
        if _, err := db.Exec(stmt); err != nil {
            return err
        }
    }

    return nil
}

// migrateV2Down V2回滚
func migrateV2Down(db *sql.DB) error {
    statements := []string{
        "DROP TABLE IF EXISTS memory_tags",
        "DROP TABLE IF EXISTS tags",
    }

    for _, stmt := range statements {
        if _, err := db.Exec(stmt); err != nil {
            return err
        }
    }

    return nil
}

// RunMigrations 运行所有迁移
func RunMigrations(db *sql.DB) error {
    // 创建迁移版本表
    if _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version INTEGER PRIMARY KEY,
            applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `); err != nil {
        return err
    }

    // 获取当前版本
    var currentVersion int
    row := db.QueryRow("SELECT MAX(version) FROM schema_migrations")
    row.Scan(&currentVersion)

    // 运行新迁移
    for _, migration := range Migrations {
        if migration.Version <= currentVersion {
            continue
        }

        if err := migration.Up(db); err != nil {
            return fmt.Errorf("migration %d failed: %w", migration.Version, err)
        }

        // 记录迁移
        if _, err := db.Exec(
            "INSERT INTO schema_migrations (version) VALUES (?)",
            migration.Version,
        ); err != nil {
            return err
        }
    }

    return nil
}
```

### 5.2 使用迁移

```go
// 初始化数据库时运行迁移
func InitDatabase(dbPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    // 运行迁移
    if err := RunMigrations(db); err != nil {
        db.Close()
        return nil, err
    }

    return db, nil
}
```

---

## 六、查询优化

### 6.1 常用查询优化

```go
// internal/memory/queries.go
package memory

// GetConversationWithMessages 获取对话及消息（使用JOIN优化）
func (s *Storage) GetConversationWithMessages(id string) (*Conversation, error) {
    // 1. 获取对话
    conv, err := s.GetConversation(id)
    if err != nil {
        return nil, err
    }

    // 2. 批量获取消息
    rows, err := s.db.Query(`
        SELECT id, role, content, timestamp, tokens
        FROM messages
        WHERE conversation_id = ?
        ORDER BY timestamp ASC
    `, id)

    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var msg Message
        if err := rows.Scan(&msg.ID, &msg.Role, &msg.Content, &msg.Timestamp, &msg.Tokens); err != nil {
            return nil, err
        }
        conv.Messages = append(conv.Messages, msg)
    }

    return conv, nil
}

// BatchCreateMemories 批量创建记忆（使用事务）
func (s *Storage) BatchCreateMemories(memories []*Memory) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }

    defer tx.Rollback()

    stmt, err := tx.Prepare(`
        INSERT INTO memories (id, content, created_at, updated_at, metadata, source)
        VALUES (?, ?, ?, ?, ?, ?)
    `)

    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, mem := range memories {
        _, err := stmt.Exec(
            mem.ID,
            mem.Content,
            mem.CreatedAt,
            mem.UpdatedAt,
            mem.MetadataString(),
            mem.Source,
        )

        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

// GetMemoriesByDateRange 按日期范围获取记忆
func (s *Storage) GetMemoriesByDateRange(start, end time.Time, limit, offset int) ([]*Memory, error) {
    query := `
        SELECT id, content, created_at, updated_at, metadata, source
        FROM memories
        WHERE created_at >= ? AND created_at <= ?
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `

    rows, err := s.db.Query(query, start, end, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    return s.scanMemories(rows)
}

// GetRecentMemories 获取最近记忆（带缓存建议）
func (s *Storage) GetRecentMemories(limit int) ([]*Memory, error) {
    // 这个查询频繁，建议在应用层缓存
    rows, err := s.db.Query(`
        SELECT id, content, created_at, updated_at, metadata, source
        FROM memories
        ORDER BY created_at DESC
        LIMIT ?
    `, limit)

    if err != nil {
        return nil, err
    }
    defer rows.Close()

    return s.scanMemories(rows)
}

// scanMemories 辅助函数：扫描内存行
func (s *Storage) scanMemories(rows *sql.Rows) ([]*Memory, error) {
    var memories []*Memory

    for rows.Next() {
        var mem Memory
        var metadataStr string

        err := rows.Scan(
            &mem.ID,
            &mem.Content,
            &mem.CreatedAt,
            &mem.UpdatedAt,
            &metadataStr,
            &mem.Source,
        )

        if err != nil {
            return nil, err
        }

        mem.ParseMetadata(metadataStr)
        memories = append(memories, &mem)
    }

    return memories, nil
}
```

### 6.2 连接池配置

```go
// internal/memory/pool.go
package memory

import (
    "database/sql"
    "time"
)

// SetConnectionPool 设置连接池参数
func SetConnectionPool(db *sql.DB) {
    // 设置最大空闲连接数
    db.SetMaxIdleConns(10)

    // 设置最大打开连接数
    db.SetMaxOpenConns(100)

    // 设置连接最大存活时间
    db.SetConnMaxLifetime(time.Hour)
}

// 使用示例
func NewStorage(dbPath string) (*Storage, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    SetConnectionPool(db)

    if err := RunMigrations(db); err != nil {
        db.Close()
        return nil, err
    }

    return &Storage{db: db}, nil
}
```

---

## 七、数据备份与恢复

### 7.1 备份

```go
// Backup 备份数据库
func (s *Storage) Backup(backupPath string) error {
    // SQLite备份：直接复制文件
    // 或使用SQLITE3的VACUUM INTO命令

    _, err := s.db.Exec("VACUUM INTO ?", backupPath)
    return err
}

// ExportMemories 导出记忆为JSON
func (s *Storage) ExportMemories() ([]byte, error) {
    memories, err := s.ListMemories()
    if err != nil {
        return nil, err
    }

    return json.MarshalIndent(memories, "", "  ")
}
```

### 7.2 恢复

```go
// ImportMemories 从JSON导入记忆
func (s *Storage) ImportMemories(data []byte) error {
    var memories []*Memory
    if err := json.Unmarshal(data, &memories); err != nil {
        return err
    }

    return s.BatchCreateMemories(memories)
}
```

---

**文档版本历史**:
- v1.0-Go (2025-12-29): 创建Golang版本数据模型文档
