package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewSessionManager 测试创建会话管理器
func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager()
	if sm == nil {
		t.Fatal("NewSessionManager 返回 nil")
	}
	if sm.sessions == nil {
		t.Error("sessions 未初始化")
	}
	if sm.GetSessionCount() != 0 {
		t.Errorf("期望 0 个会话, 得到 %d", sm.GetSessionCount())
	}
}

// TestNewSessionManagerWithStorage 测试创建带存储的会话管理器
func TestNewSessionManagerWithStorage(t *testing.T) {
	tempDir := t.TempDir()
	sm, err := NewSessionManagerWithStorage(tempDir)
	if err != nil {
		t.Fatalf("NewSessionManagerWithStorage 失败: %v", err)
	}
	if sm == nil {
		t.Fatal("返回 nil")
	}
	if sm.storageFile == "" {
		t.Error("storageFile 未设置")
	}
}

// TestGetOrCreateSession 测试获取或创建会话
func TestGetOrCreateSession(t *testing.T) {
	sm := NewSessionManager()

	// 测试创建新会话
	sessionID := "test_session_1"
	session := sm.GetOrCreateSession(sessionID)
	if session == nil {
		t.Fatal("返回 nil")
	}
	if session.ID != sessionID {
		t.Errorf("期望 ID %s, 得到 %s", sessionID, session.ID)
	}
	if len(session.Messages) != 0 {
		t.Errorf("期望 0 条消息, 得到 %d", len(session.Messages))
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
	sm := NewSessionManager()
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
	if session.Messages[0].Role != "user" {
		t.Errorf("期望 role 'user', 得到 '%s'", session.Messages[0].Role)
	}
	if session.Messages[0].Content != "你好" {
		t.Errorf("期望 content '你好', 得到 '%s'", session.Messages[0].Content)
	}

	// 添加助手消息
	sm.AddMessage(sessionID, "assistant", "你好呀")
	session = sm.GetSession(sessionID)
	if len(session.Messages) != 2 {
		t.Errorf("期望 2 条消息, 得到 %d", len(session.Messages))
	}

	// 测试消息数量限制（超过 20 条应该截断）
	for i := 0; i < 25; i++ {
		sm.AddMessage(sessionID, "user", "测试消息")
	}
	session = sm.GetSession(sessionID)
	if len(session.Messages) > 20 {
		t.Errorf("期望最多 20 条消息, 得到 %d", len(session.Messages))
	}
}

// TestGetMessages 测试获取消息列表
func TestGetMessages(t *testing.T) {
	sm := NewSessionManager()
	sessionID := "test_session"
	sm.GetOrCreateSession(sessionID)

	// 空会话
	messages := sm.GetMessages(sessionID)
	if len(messages) != 0 {
		t.Errorf("期望 0 条消息, 得到 %d", len(messages))
	}

	// 添加消息后
	sm.AddMessage(sessionID, "user", "消息1")
	sm.AddMessage(sessionID, "assistant", "消息2")
	messages = sm.GetMessages(sessionID)
	if len(messages) != 2 {
		t.Errorf("期望 2 条消息, 得到 %d", len(messages))
	}

	// 不存在的会话
	emptyMessages := sm.GetMessages("non_existent")
	if len(emptyMessages) != 0 {
		t.Errorf("不存在的会话应返回空列表, 得到 %d 条", len(emptyMessages))
	}
}

// TestGetSession 测试获取会话
func TestGetSession(t *testing.T) {
	sm := NewSessionManager()
	sessionID := "test_session"
	sm.GetOrCreateSession(sessionID)
	sm.AddMessage(sessionID, "user", "测试")

	session := sm.GetSession(sessionID)
	if session == nil {
		t.Fatal("期望返回会话, 得到 nil")
	}
	if session.ID != sessionID {
		t.Errorf("期望 ID %s, 得到 %s", sessionID, session.ID)
	}

	// 不存在的会话
	nilSession := sm.GetSession("non_existent")
	if nilSession != nil {
		t.Error("不存在的会话应返回 nil")
	}
}

// TestGetAllSessions 测试获取所有会话
func TestGetAllSessions(t *testing.T) {
	sm := NewSessionManager()

	// 空管理器
	all := sm.GetAllSessions()
	if len(all) != 0 {
		t.Errorf("期望 0 个会话, 得到 %d", len(all))
	}

	// 添加多个会话
	sm.GetOrCreateSession("session_1")
	sm.AddMessage("session_1", "user", "消息1")
	sm.GetOrCreateSession("session_2")
	sm.AddMessage("session_2", "user", "消息2")
	sm.GetOrCreateSession("session_3")

	all = sm.GetAllSessions()
	if len(all) != 3 {
		t.Errorf("期望 3 个会话, 得到 %d", len(all))
	}
}

// TestClearSession 测试清空会话
func TestClearSession(t *testing.T) {
	sm := NewSessionManager()
	sessionID := "test_session"
	sm.GetOrCreateSession(sessionID)
	sm.AddMessage(sessionID, "user", "消息1")
	sm.AddMessage(sessionID, "assistant", "消息2")

	// 清空前
	session := sm.GetSession(sessionID)
	if len(session.Messages) != 2 {
		t.Errorf("清空前期望 2 条消息, 得到 %d", len(session.Messages))
	}

	// 清空
	sm.ClearSession(sessionID)
	session = sm.GetSession(sessionID)
	if len(session.Messages) != 0 {
		t.Errorf("清空后期望 0 条消息, 得到 %d", len(session.Messages))
	}
}

// TestDeleteSession 测试删除会话
func TestDeleteSession(t *testing.T) {
	sm := NewSessionManager()
	sessionID := "test_session"
	sm.GetOrCreateSession(sessionID)

	// 删除前
	if sm.GetSessionCount() != 1 {
		t.Errorf("删除前期望 1 个会话, 得到 %d", sm.GetSessionCount())
	}

	// 删除
	sm.DeleteSession(sessionID)

	// 删除后
	if sm.GetSessionCount() != 0 {
		t.Errorf("删除后期望 0 个会话, 得到 %d", sm.GetSessionCount())
	}
	session := sm.GetSession(sessionID)
	if session != nil {
		t.Error("删除后获取会话应返回 nil")
	}
}

// TestGetSessionCount 测试获取会话数量
func TestGetSessionCount(t *testing.T) {
	sm := NewSessionManager()

	if sm.GetSessionCount() != 0 {
		t.Errorf("期望 0 个会话, 得到 %d", sm.GetSessionCount())
	}

	sm.GetOrCreateSession("session_1")
	if sm.GetSessionCount() != 1 {
		t.Errorf("期望 1 个会话, 得到 %d", sm.GetSessionCount())
	}

	sm.GetOrCreateSession("session_2")
	if sm.GetSessionCount() != 2 {
		t.Errorf("期望 2 个会话, 得到 %d", sm.GetSessionCount())
	}

	sm.DeleteSession("session_1")
	if sm.GetSessionCount() != 1 {
		t.Errorf("删除后期望 1 个会话, 得到 %d", sm.GetSessionCount())
	}
}

// TestSessionPersistence 测试会话持久化
func TestSessionPersistence(t *testing.T) {
	tempDir := t.TempDir()

	// 创建管理器并添加会话
	sm1, err := NewSessionManagerWithStorage(tempDir)
	if err != nil {
		t.Fatalf("创建会话管理器失败: %v", err)
	}

	sessionID := "persist_test"
	sm1.GetOrCreateSession(sessionID)
	sm1.AddMessage(sessionID, "user", "持久化测试消息")
	sm1.AddMessage(sessionID, "assistant", "回复消息")

	// 验证文件存在
	storageFile := filepath.Join(tempDir, "sessions.json")
	if _, err := os.Stat(storageFile); os.IsNotExist(err) {
		t.Error("会话文件未创建")
	}

	// 创建新管理器并加载
	sm2, err := NewSessionManagerWithStorage(tempDir)
	if err != nil {
		t.Fatalf("加载会话管理器失败: %v", err)
	}

	session := sm2.GetSession(sessionID)
	if session == nil {
		t.Fatal("加载后会话不存在")
	}
	if len(session.Messages) != 2 {
		t.Errorf("加载后期望 2 条消息, 得到 %d", len(session.Messages))
	}
	if session.Messages[0].Content != "持久化测试消息" {
		t.Errorf("消息内容不匹配: %s", session.Messages[0].Content)
	}
}

// TestUpdateSummary 测试更新摘要
func TestUpdateSummary(t *testing.T) {
	sm := NewSessionManager()
	sessionID := "test_session"
	sm.GetOrCreateSession(sessionID)

	summary := &SessionSummary{
		Content:      "这是一个测试摘要",
		KeyPoints:    []string{"要点1", "要点2"},
		Topics:       []string{"技术"},
		MessageCount: 10,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	sm.UpdateSummary(sessionID, summary)

	session := sm.GetSession(sessionID)
	if session.Summary == nil {
		t.Fatal("摘要未设置")
	}
	if session.Summary.Content != "这是一个测试摘要" {
		t.Errorf("摘要内容不匹配: %s", session.Summary.Content)
	}
	if len(session.Summary.KeyPoints) != 2 {
		t.Errorf("关键点数量不匹配: %d", len(session.Summary.KeyPoints))
	}
}

// TestGetSummary 测试获取摘要
func TestGetSummary(t *testing.T) {
	sm := NewSessionManager()
	sessionID := "test_session"
	sm.GetOrCreateSession(sessionID)

	// 无摘要
	summary := sm.GetSummary(sessionID)
	if summary != nil {
		t.Error("期望返回 nil")
	}

	// 设置摘要
	testSummary := &SessionSummary{
		Content:     "测试摘要",
		MessageCount: 5,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	sm.UpdateSummary(sessionID, testSummary)

	// 获取摘要
	summary = sm.GetSummary(sessionID)
	if summary == nil {
		t.Fatal("期望返回摘要, 得到 nil")
	}
	if summary.Content != "测试摘要" {
		t.Errorf("摘要内容不匹配: %s", summary.Content)
	}
}

// TestConcurrentAccess 并发测试
func TestConcurrentAccess(t *testing.T) {
	sm := NewSessionManager()
	sessionID := "concurrent_test"
	sm.GetOrCreateSession(sessionID)

	// 并发写入（每个 goroutine 添加 5 条消息，总共 50 条）
	// 由于限制为 20 条，最终应该只有 20 条
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 5; j++ {
				sm.AddMessage(sessionID, "user", "消息")
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证消息数量（由于限制最多 20 条）
	session := sm.GetSession(sessionID)
	if len(session.Messages) != 20 {
		t.Errorf("期望最多 20 条消息, 得到 %d", len(session.Messages))
	}
}

// BenchmarkAddMessage 性能测试
func BenchmarkAddMessage(b *testing.B) {
	sm := NewSessionManager()
	sessionID := "bench_session"
	sm.GetOrCreateSession(sessionID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.AddMessage(sessionID, "user", "测试消息")
	}
}

// BenchmarkGetMessages 性能测试
func BenchmarkGetMessages(b *testing.B) {
	sm := NewSessionManager()
	sessionID := "bench_session"
	sm.GetOrCreateSession(sessionID)

	// 预填充消息
	for i := 0; i < 100; i++ {
		sm.AddMessage(sessionID, "user", "测试消息")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.GetMessages(sessionID)
	}
}
