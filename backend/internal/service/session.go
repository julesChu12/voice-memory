package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Session 对话会话
type Session struct {
	ID        string    `json:"id"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 兼容旧格式的大写字段名
	IDCompat        string    `json:"ID,omitempty"`
	MessagesCompat  []Message `json:"Messages,omitempty"`
	CreatedAtCompat time.Time `json:"CreatedAt,omitempty"`
	UpdatedAtCompat time.Time `json:"UpdatedAt,omitempty"`
}

// UnmarshalJSON 自定义 JSON 反序列化，兼容旧格式
func (s *Session) UnmarshalJSON(data []byte) error {
	// 定义别名类型避免递归调用
	type Alias Session
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 如果新字段为空，从旧字段复制
	if s.ID == "" && s.IDCompat != "" {
		s.ID = s.IDCompat
	}
	if len(s.Messages) == 0 && len(s.MessagesCompat) > 0 {
		s.Messages = s.MessagesCompat
	}
	if s.CreatedAt.IsZero() && !s.CreatedAtCompat.IsZero() {
		s.CreatedAt = s.CreatedAtCompat
	}
	if s.UpdatedAt.IsZero() && !s.UpdatedAtCompat.IsZero() {
		s.UpdatedAt = s.UpdatedAtCompat
	}

	return nil
}

// SessionData 会话数据（用于存储）
type SessionData struct {
	Sessions  map[string]Session `json:"sessions"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions   map[string]*Session
	mu         sync.RWMutex
	storageFile string
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// NewSessionManagerWithStorage 创建带持久化的会话管理器
func NewSessionManagerWithStorage(dataDir string) (*SessionManager, error) {
	// 确保目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建会话目录失败: %w", err)
	}

	storageFile := filepath.Join(dataDir, "sessions.json")
	sm := &SessionManager{
		sessions:    make(map[string]*Session),
		storageFile: storageFile,
	}

	// 加载已有数据
	if err := sm.load(); err != nil {
		fmt.Printf("会话加载警告: %v\n", err)
	}

	fmt.Printf("会话管理器初始化完成，共 %d 个会话\n", len(sm.sessions))
	return sm, nil
}

// load 从文件加载会话
func (sm *SessionManager) load() error {
	if sm.storageFile == "" {
		return nil // 不使用持久化
	}

	data, err := os.ReadFile(sm.storageFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，使用默认值
		}
		return err
	}

	var sessionData SessionData
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return err
	}

	// 转换为指针类型，确保使用 key 作为 ID（优先于对象内部的 ID 字段）
	sm.sessions = make(map[string]*Session)
	for id, sess := range sessionData.Sessions {
		// 使用 map 的 key 作为正确的 ID，而不是对象内部可能不一致的 ID 字段
		sess.ID = id
		sm.sessions[id] = &sess
	}

	return nil
}

// save 保存会话到文件
func (sm *SessionManager) save() error {
	if sm.storageFile == "" {
		return nil // 不使用持久化
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessionData := SessionData{
		Sessions:  make(map[string]Session),
		UpdatedAt: time.Now(),
	}

	for id, sess := range sm.sessions {
		sessionData.Sessions[id] = *sess
	}

	data, err := json.MarshalIndent(sessionData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sm.storageFile, data, 0644)
}

// GetOrCreateSession 获取或创建会话
func (sm *SessionManager) GetOrCreateSession(sessionID string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		return session
	}

	// 创建新会话
	session := &Session{
		ID:        sessionID,
		Messages:  []Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	sm.sessions[sessionID] = session

	// 自动保存
	sm.saveUnsafe()

	fmt.Printf("创建新会话: %s\n", sessionID)
	return session
}

// AddMessage 添加消息到会话
func (sm *SessionManager) AddMessage(sessionID string, role, content string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.Messages = append(session.Messages, Message{
			Role:    role,
			Content: content,
		})
		session.UpdatedAt = time.Now()

		// 限制历史消息数量（保留最近 20 条）
		if len(session.Messages) > 20 {
			session.Messages = session.Messages[len(session.Messages)-20:]
		}

		fmt.Printf("会话 %s 消息数: %d\n", sessionID, len(session.Messages))

		// 自动保存
		sm.saveUnsafe()
	}
}

// GetMessages 获取会话消息历史
func (sm *SessionManager) GetMessages(sessionID string) []Message {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if session, exists := sm.sessions[sessionID]; exists {
		return session.Messages
	}
	return []Message{}
}

// GetSession 获取完整会话
func (sm *SessionManager) GetSession(sessionID string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if session, exists := sm.sessions[sessionID]; exists {
		return session
	}
	return nil
}

// GetAllSessions 获取所有会话
func (sm *SessionManager) GetAllSessions() []Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]Session, 0, len(sm.sessions))
	for id, sess := range sm.sessions {
		fmt.Printf("[DEBUG GetAllSessions] id=%s, sess.ID=%s\n", id, sess.ID)
		result = append(result, *sess)
	}
	return result
}

// ClearSession 清空会话
func (sm *SessionManager) ClearSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.Messages = []Message{}
		session.UpdatedAt = time.Now()
		fmt.Printf("清空会话: %s\n", sessionID)

		// 自动保存
		sm.saveUnsafe()
	}
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[sessionID]; exists {
		delete(sm.sessions, sessionID)
		fmt.Printf("删除会话: %s\n", sessionID)

		// 自动保存
		sm.saveUnsafe()
	}
}

// GetSessionCount 获取会话数量
func (sm *SessionManager) GetSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// saveUnsafe 不加锁的保存（内部使用）
func (sm *SessionManager) saveUnsafe() {
	if sm.storageFile == "" {
		return
	}

	sessionData := SessionData{
		Sessions:  make(map[string]Session),
		UpdatedAt: time.Now(),
	}

	for id, sess := range sm.sessions {
		sessionData.Sessions[id] = *sess
	}

	data, err := json.MarshalIndent(sessionData, "", "  ")
	if err != nil {
		fmt.Printf("保存会话失败: %v\n", err)
		return
	}

	if err := os.WriteFile(sm.storageFile, data, 0644); err != nil {
		fmt.Printf("保存会话失败: %v\n", err)
	}
}
