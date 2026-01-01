package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"voice-memory/internal/service"
)

// SessionData 本地定义旧数据结构
type SessionData struct {
	Sessions  map[string]service.Session `json:"sessions"`
	UpdatedAt time.Time                  `json:"updated_at"`
}

func main() {
	dataDir := "./data"

	// 创建数据库
	database, err := service.NewDatabase(dataDir)
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	defer database.Close()

	fmt.Println("开始迁移 JSON 数据到 SQLite...")

	// 迁移会话
	sessionFile := filepath.Join(dataDir, "sessions", "sessions.json")
	if _, err := os.Stat(sessionFile); err == nil {
		fmt.Println("\n迁移会话...")
		if err := migrateSessions(database, sessionFile); err != nil {
			fmt.Printf("会话迁移失败: %v\n", err)
		} else {
			fmt.Println("会话迁移完成")
		}
	}

	// 迁移知识库
	knowledgeFile := filepath.Join(dataDir, "knowledge", "knowledge.json")
	if _, err := os.Stat(knowledgeFile); err == nil {
		fmt.Println("\n迁移知识库...")
		if err := migrateKnowledge(database, knowledgeFile); err != nil {
			fmt.Printf("知识库迁移失败: %v\n", err)
		} else {
			fmt.Println("知识库迁移完成")
		}
	}

	fmt.Println("\n✅ 数据迁移完成!")
	fmt.Println("\n备份文件位置:")
	fmt.Println("  - sessions.json.bak")
	fmt.Println("  - knowledge.json.bak")
}

func migrateSessions(db *service.Database, sessionFile string) error {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return err
	}

	var sessionData SessionData
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return err
	}

	count := 0
	for _, sess := range sessionData.Sessions {
		// 将 map 转为 slice
		messages := make([]service.Message, len(sess.Messages))
		copy(messages, sess.Messages)

		session := &service.Session{
			ID:        sess.ID,
			Messages:  messages,
			CreatedAt: sess.CreatedAt,
			UpdatedAt: sess.UpdatedAt,
		}

		if err := db.SaveSession(session); err != nil {
			fmt.Printf("  警告: 迁移会话 %s 失败: %v\n", sess.ID, err)
		} else {
			count++
		}
	}

	fmt.Printf("  迁移了 %d 个会话\n", count)

	// 备份原文件
	os.Rename(sessionFile, sessionFile+".bak")

	return nil
}

func migrateKnowledge(db *service.Database, knowledgeFile string) error {
	data, err := os.ReadFile(knowledgeFile)
	if err != nil {
		return err
	}

	var storeData service.KnowledgeStoreData
	if err := json.Unmarshal(data, &storeData); err != nil {
		return err
	}

	count := 0
	for _, k := range storeData.Knowledges {
		if err := db.SaveKnowledge(&k); err != nil {
			fmt.Printf("  警告: 迁移知识 %s 失败: %v\n", k.ID, err)
		} else {
			count++
		}
	}

	fmt.Printf("  迁移了 %d 条知识\n", count)

	// 备份原文件
	os.Rename(knowledgeFile, knowledgeFile+".bak")

	return nil
}
