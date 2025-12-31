package service

import (
	"fmt"
	"sync"
	"time"
)

// Session 对话会话
type Session struct {
	ID        string         `json:"id"`
	Messages  []Message      `json:"messages"`
	Summary   *SessionSummary `json:"summary,omitempty"`   // 会话摘要
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// SessionManager 会话管理器
type SessionManager struct {
	db *Database
	mu sync.RWMutex
}

// NewSessionManager 创建内存版会话管理器（仅用于测试，无持久化）
func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

// NewSessionManagerWithDB 创建带数据库持久化的会话管理器
func NewSessionManagerWithDB(db *Database) *SessionManager {
	return &SessionManager{
		db: db,
	}
}

// GetOrCreateSession 获取或创建会话
func (sm *SessionManager) GetOrCreateSession(sessionID string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.db != nil {
		session, err := sm.db.GetSession(sessionID)
		if err == nil && session != nil {
			return session
		}
	}

	// 创建新会话
	session := &Session{
		ID:        sessionID,
		Messages:  []Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if sm.db != nil {
		sm.db.SaveSession(session)
	}

	fmt.Printf("创建新会话: %s\n", sessionID)
	return session
}

// AddMessage 添加消息到会话
func (sm *SessionManager) AddMessage(sessionID string, role, content string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var session *Session
	var err error

	if sm.db != nil {
		session, err = sm.db.GetSession(sessionID)
	}

	if err == nil && session != nil {
		session.Messages = append(session.Messages, Message{
			Role:    role,
			Content: content,
		})
		session.UpdatedAt = time.Now()

		// 限制历史消息数量（保留最近 20 条）
		if len(session.Messages) > 20 {
			session.Messages = session.Messages[len(session.Messages)-20:]
		}

		if sm.db != nil {
			sm.db.SaveSession(session)
		}
		fmt.Printf("会话 %s 消息数: %d (已保存)\n", sessionID, len(session.Messages))
	}
}

// GetMessages 获取会话消息历史
func (sm *SessionManager) GetMessages(sessionID string) []Message {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.db != nil {
		session, err := sm.db.GetSession(sessionID)
		if err == nil && session != nil {
			return session.Messages
		}
	}
	return []Message{}
}

// GetSession 获取完整会话
func (sm *SessionManager) GetSession(sessionID string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.db != nil {
		session, _ := sm.db.GetSession(sessionID)
		return session
	}
	return nil
}

// GetAllSessions 获取所有会话
func (sm *SessionManager) GetAllSessions() []Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.db != nil {
		sessions, _ := sm.db.GetAllSessions()
		return sessions
	}
	return []Session{}
}

// GetSessionCount 获取会话总数
func (sm *SessionManager) GetSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.db != nil {
		sessions, _ := sm.db.GetAllSessions()
		return len(sessions)
	}
	return 0
}

// ClearSession 清空会话
func (sm *SessionManager) ClearSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.db != nil {
		session, _ := sm.db.GetSession(sessionID)
		if session != nil {
			session.Messages = []Message{}
			session.UpdatedAt = time.Now()
			sm.db.SaveSession(session)
			fmt.Printf("清空会话: %s\n", sessionID)
		}
	}
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.db != nil {
		sm.db.DeleteSession(sessionID)
		fmt.Printf("删除会话: %s\n", sessionID)
	}
}


// UpdateSummary 更新会话摘要
func (sm *SessionManager) UpdateSummary(sessionID string, summary *SessionSummary) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.db != nil {
		session, _ := sm.db.GetSession(sessionID)
		if session != nil {
			session.Summary = summary
			session.UpdatedAt = time.Now()
			sm.db.SaveSession(session)
			fmt.Printf("更新摘要: %s\n", sessionID)
		}
	}
}

// GetSummary 获取会话摘要
func (sm *SessionManager) GetSummary(sessionID string) *SessionSummary {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.db != nil {
		session, _ := sm.db.GetSession(sessionID)
		if session != nil {
			return session.Summary
		}
	}
	return nil
}