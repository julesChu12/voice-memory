package service

import (
	"testing"
)

// setupTestDB 创建一个用于测试的临时数据库和 SessionManager
func setupTestDB(t *testing.T) (*Database, *SessionManager) {
	tempDir := t.TempDir()
	db, err := NewDatabase(tempDir)
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
	sm := NewSessionManagerWithDB(db)
	return db, sm
}

// TestNewSessionManagerWithDB 测试创建会话管理器
func TestNewSessionManagerWithDB(t *testing.T) {
	_, sm := setupTestDB(t)
	if sm == nil {
		t.Fatal("NewSessionManagerWithDB 返回 nil")
	}
	if sm.GetSessionCount() != 0 {
		t.Errorf("期望 0 个会话, 得到 %d", sm.GetSessionCount())
	}
}

// TestGetOrCreateSession 测试获取或创建会话
func TestGetOrCreateSession(t *testing.T) {
	_, sm := setupTestDB(t)

	// 测试创建新会话
	sessionID := "test_session_1"
	session := sm.GetOrCreateSession(sessionID)
	if session == nil {
		t.Fatal("返回 nil")
	}
	if session.ID != sessionID {
		t.Errorf("期望 ID %s, 得到 %s", sessionID, session.ID)
	}
	if sm.GetSessionCount() != 1 {
		t.Errorf("期望 1 个会话, 得到 %d", sm.GetSessionCount())
	}

	// 测试获取已存在的会话
	sameSession := sm.GetOrCreateSession(sessionID)
	if sameSession.ID != sessionID {
		t.Errorf("期望相同 ID %s, 得到 %s", sessionID, sameSession.ID)
	}
	if sm.GetSessionCount() != 1 {
		t.Errorf("期望仍然 1 个会话, 得到 %d", sm.GetSessionCount())
	}
}

// TestAddMessage 测试添加消息
func TestAddMessage(t *testing.T) {
	_, sm := setupTestDB(t)
	sessionID := "test_session"
	sm.GetOrCreateSession(sessionID)

	// 添加用户消息
	sm.AddMessage(sessionID, "user", "你好")
	session := sm.GetSession(sessionID)
	if session == nil {
		t.Fatal("会话不存在")
	}
	if len(session.Messages) != 1 {
		t.Errorf("期望 1 条消息, 得到 %d", len(session.Messages))
	}
	if session.Messages[0].Content != "你好" {
		t.Errorf("期望 content '你好', 得到 '%s'", session.Messages[0].Content)
	}

	// 验证持久化：直接从 DB 读取
	msgs := sm.GetMessages(sessionID)
	if len(msgs) != 1 {
		t.Errorf("持久化检查失败，期望 1 条消息")
	}
}

// TestSessionPersistence 跨实例持久化测试
func TestSessionPersistence(t *testing.T) {
	tempDir := t.TempDir()
	db1, _ := NewDatabase(tempDir)
	sm1 := NewSessionManagerWithDB(db1)

	sessionID := "persist_test"
	sm1.GetOrCreateSession(sessionID)
	sm1.AddMessage(sessionID, "user", "持久化消息")
	db1.Close()

	// 重新打开数据库和管理器
	db2, _ := NewDatabase(tempDir)
	sm2 := NewSessionManagerWithDB(db2)
	defer db2.Close()

	session := sm2.GetSession(sessionID)
	if session == nil {
		t.Fatal("数据库加载后会话不存在")
	}
	if session.Messages[0].Content != "持久化消息" {
		t.Errorf("消息内容不匹配")
	}
}

// TestDeleteSession 测试删除会话
func TestDeleteSession(t *testing.T) {
	_, sm := setupTestDB(t)
	sessionID := "test_session"
	sm.GetOrCreateSession(sessionID)

	sm.DeleteSession(sessionID)

	if sm.GetSessionCount() != 0 {
		t.Errorf("删除后数量应为 0")
	}
}