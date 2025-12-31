package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database 数据库
type Database struct {
	db *sql.DB
}

// NewDatabase 创建数据库
func NewDatabase(dataDir string) (*Database, error) {
	// 确保目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	dbPath := filepath.Join(dataDir, "voice-memory.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 启用外键约束
	db.SetMaxOpenConns(1) // SQLite 不支持多写入

	database := &Database{db: db}

	// 初始化表结构
	if err := database.initTables(); err != nil {
		return nil, fmt.Errorf("初始化表失败: %w", err)
	}

	fmt.Printf("数据库初始化完成: %s\n", dbPath)
	return database, nil
}

// initTables 初始化表
func (d *Database) initTables() error {
	schemas := []string{
		// 会话表
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			messages TEXT,
			created_at INTEGER,
			updated_at INTEGER
		)`,

		// 知识库表
		`CREATE TABLE IF NOT EXISTS knowledge (
			id TEXT PRIMARY KEY,
			title TEXT,
			content TEXT,
			summary TEXT,
			key_points TEXT,
			category TEXT,
			tags TEXT,
			source TEXT,
			audio_url TEXT,
			session_id TEXT,
			created_at INTEGER,
			updated_at INTEGER,
			metadata TEXT,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL
		)`,

		// 添加 title 字段（如果表已存在但没有该字段）
		`ALTER TABLE knowledge ADD COLUMN title TEXT`,

		// v1.0 新增字段（实体、重要性、情感）
		`ALTER TABLE knowledge ADD COLUMN entities TEXT`,
		`ALTER TABLE knowledge ADD COLUMN importance TEXT`,
		`ALTER TABLE knowledge ADD COLUMN sentiment TEXT`,

		// 索引
		`CREATE INDEX IF NOT EXISTS idx_knowledge_category ON knowledge(category)`,
		`CREATE INDEX IF NOT EXISTS idx_knowledge_session_id ON knowledge(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_knowledge_created_at ON knowledge(created_at)`,
	}

	for _, schema := range schemas {
		if _, err := d.db.Exec(schema); err != nil {
			// 忽略 "duplicate column name" 错误（列已存在）
			if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
				return fmt.Errorf("创建表失败: %w", err)
			}
		}
	}

	return nil
}

// SaveSession 保存会话
func (d *Database) SaveSession(session *Session) error {
	messagesJSON, err := json.Marshal(session.Messages)
	if err != nil {
		return err
	}

	query := `INSERT OR REPLACE INTO sessions (id, messages, created_at, updated_at)
			  VALUES (?, ?, ?, ?)`

	createdAt := session.CreatedAt.Unix()
	updatedAt := session.UpdatedAt.Unix()

	_, err = d.db.Exec(query, session.ID, string(messagesJSON), createdAt, updatedAt)
	return err
}

// GetSession 获取会话
func (d *Database) GetSession(sessionID string) (*Session, error) {
	query := `SELECT id, messages, created_at, updated_at FROM sessions WHERE id = ?`

	var id string
	var messagesJSON string
	var createdAt, updatedAt int64

	err := d.db.QueryRow(query, sessionID).Scan(&id, &messagesJSON, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var messages []Message
	if err := json.Unmarshal([]byte(messagesJSON), &messages); err != nil {
		return nil, err
	}

	return &Session{
		ID:        id,
		Messages:  messages,
		CreatedAt: time.Unix(createdAt, 0),
		UpdatedAt: time.Unix(updatedAt, 0),
	}, nil
}

// GetAllSessions 获取所有会话
func (d *Database) GetAllSessions() ([]Session, error) {
	query := `SELECT id, messages, created_at, updated_at FROM sessions ORDER BY updated_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var id string
		var messagesJSON string
		var createdAt, updatedAt int64

		if err := rows.Scan(&id, &messagesJSON, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		var messages []Message
		if err := json.Unmarshal([]byte(messagesJSON), &messages); err != nil {
			return nil, err
		}

		sessions = append(sessions, Session{
			ID:        id,
			Messages:  messages,
			CreatedAt: time.Unix(createdAt, 0),
			UpdatedAt: time.Unix(updatedAt, 0),
		})
	}

	return sessions, nil
}

// SaveKnowledge 保存知识
func (d *Database) SaveKnowledge(knowledge *Knowledge) error {
	keyPointsJSON, _ := json.Marshal(knowledge.KeyPoints)
	tagsJSON, _ := json.Marshal(knowledge.Tags)
	entitiesJSON, _ := json.Marshal(knowledge.Entities)
	metadataJSON, _ := json.Marshal(knowledge.Metadata)

	query := `INSERT OR REPLACE INTO knowledge
			  (id, title, content, summary, key_points, entities, category, tags, importance, sentiment, source, audio_url, session_id, created_at, updated_at, metadata)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	createdAt := knowledge.CreatedAt.Unix()
	updatedAt := knowledge.UpdatedAt.Unix()

	_, err := d.db.Exec(query,
		knowledge.ID,
		knowledge.Title,
		knowledge.Content,
		knowledge.Summary,
		string(keyPointsJSON),
		string(entitiesJSON),
		knowledge.Category,
		string(tagsJSON),
		knowledge.Importance,
		knowledge.Sentiment,
		knowledge.Source,
		knowledge.AudioURL,
		knowledge.SessionID,
		createdAt,
		updatedAt,
		string(metadataJSON),
	)

	return err
}

// GetAllKnowledge 获取所有知识
func (d *Database) GetAllKnowledge() ([]Knowledge, error) {
	query := `SELECT id, COALESCE(title, '') as title, content, summary, key_points, COALESCE(entities, '{}') as entities, category, tags, COALESCE(importance, 'medium') as importance, COALESCE(sentiment, 'neutral') as sentiment, source, audio_url, session_id, created_at, updated_at, metadata
			  FROM knowledge ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var knowledges []Knowledge
	for rows.Next() {
		var k Knowledge
		var keyPointsJSON, tagsJSON, entitiesJSON, metadataJSON string
		var createdAt, updatedAt int64

		err := rows.Scan(
			&k.ID,
			&k.Title,
			&k.Content,
			&k.Summary,
			&keyPointsJSON,
			&entitiesJSON,
			&k.Category,
			&tagsJSON,
			&k.Importance,
			&k.Sentiment,
			&k.Source,
			&k.AudioURL,
			&k.SessionID,
			&createdAt,
			&updatedAt,
			&metadataJSON,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(keyPointsJSON), &k.KeyPoints)
		json.Unmarshal([]byte(tagsJSON), &k.Tags)
		json.Unmarshal([]byte(entitiesJSON), &k.Entities)
		json.Unmarshal([]byte(metadataJSON), &k.Metadata)

		k.CreatedAt = time.Unix(createdAt, 0)
		k.UpdatedAt = time.Unix(updatedAt, 0)

		knowledges = append(knowledges, k)
	}

	return knowledges, nil
}

// GetKnowledgeByCategory 按分类获取知识
func (d *Database) GetKnowledgeByCategory(category string) ([]Knowledge, error) {
	query := `SELECT id, COALESCE(title, '') as title, content, summary, key_points, category, tags, source, audio_url, session_id, created_at, updated_at, metadata
			  FROM knowledge WHERE category = ? ORDER BY created_at DESC`

	rows, err := d.db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var knowledges []Knowledge
	for rows.Next() {
		var k Knowledge
		var keyPointsJSON, tagsJSON, metadataJSON string
		var createdAt, updatedAt int64

		err := rows.Scan(
			&k.ID,
			&k.Title,
			&k.Content,
			&k.Summary,
			&keyPointsJSON,
			&k.Category,
			&tagsJSON,
			&k.Source,
			&k.AudioURL,
			&k.SessionID,
			&createdAt,
			&updatedAt,
			&metadataJSON,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(keyPointsJSON), &k.KeyPoints)
		json.Unmarshal([]byte(tagsJSON), &k.Tags)
		json.Unmarshal([]byte(metadataJSON), &k.Metadata)

		k.CreatedAt = time.Unix(createdAt, 0)
		k.UpdatedAt = time.Unix(updatedAt, 0)

		knowledges = append(knowledges, k)
	}

	return knowledges, nil
}

// SearchKnowledge 搜索知识
func (d *Database) SearchKnowledge(searchQuery string) ([]Knowledge, error) {
	query := `SELECT id, COALESCE(title, '') as title, content, summary, key_points, category, tags, source, audio_url, session_id, created_at, updated_at, metadata
			  FROM knowledge
			  WHERE content LIKE ? OR summary LIKE ? OR category LIKE ?
			  ORDER BY created_at DESC`

	pattern := "%" + searchQuery + "%"

	rows, err := d.db.Query(query, pattern, pattern, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var knowledges []Knowledge
	for rows.Next() {
		var k Knowledge
		var keyPointsJSON, tagsJSON, metadataJSON string
		var createdAt, updatedAt int64

		err := rows.Scan(
			&k.ID,
			&k.Title,
			&k.Content,
			&k.Summary,
			&keyPointsJSON,
			&k.Category,
			&tagsJSON,
			&k.Source,
			&k.AudioURL,
			&k.SessionID,
			&createdAt,
			&updatedAt,
			&metadataJSON,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(keyPointsJSON), &k.KeyPoints)
		json.Unmarshal([]byte(tagsJSON), &k.Tags)
		json.Unmarshal([]byte(metadataJSON), &k.Metadata)

		k.CreatedAt = time.Unix(createdAt, 0)
		k.UpdatedAt = time.Unix(updatedAt, 0)

		knowledges = append(knowledges, k)
	}

	return knowledges, nil
}

// DeleteKnowledge 删除知识
func (d *Database) DeleteKnowledge(id string) error {
	_, err := d.db.Exec(`DELETE FROM knowledge WHERE id = ?`, id)
	return err
}

// DeleteSession 删除会话
func (d *Database) DeleteSession(id string) error {
	_, err := d.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// Close 关闭数据库
func (d *Database) Close() error {
	return d.db.Close()
}
