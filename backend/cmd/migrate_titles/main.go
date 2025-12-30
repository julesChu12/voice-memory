package main

import (
	"fmt"
	"log"
	"os"
	"voice-memory/internal/service"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 获取数据目录
	dataDir := "."
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}

	fmt.Printf("开始迁移知识条目标题和摘要...\n")
	fmt.Printf("数据目录: %s\n\n", dataDir)

	// 初始化数据库
	database, err := service.NewDatabase(dataDir)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.Close()

	// 初始化 AI 客户端
	apiKey := os.Getenv("GLM_API_KEY")
	if apiKey == "" {
		log.Fatalf("请设置 GLM_API_KEY 环境变量")
	}

	glmClient := service.NewGLMClient(apiKey)
	organizer := service.NewKnowledgeOrganizer(glmClient)

	// 获取所有知识
	knowledges, err := database.GetAllKnowledge()
	if err != nil {
		log.Fatalf("获取知识列表失败: %v", err)
	}

	fmt.Printf("共有 %d 条知识条目\n\n", len(knowledges))

	// 统计
	updatedCount := 0
	errorCount := 0

	for i, knowledge := range knowledges {
		fmt.Printf("[%d/%d] 处理知识: %s\n", i+1, len(knowledges), knowledge.ID)

		var title string
		var summary string
		var keyPoints []string
		var category string
		var tags []string

		// 如果有 session_id，从会话生成标题和摘要
		if knowledge.SessionID != "" {
			session, err := database.GetSession(knowledge.SessionID)
			if err != nil {
				fmt.Printf("    ! 获取会话失败: %v\n", err)
			}

			if session != nil && len(session.Messages) > 0 {
				fmt.Printf("    → 会话有 %d 条消息，正在生成标题和摘要...\n", len(session.Messages))

				// 生成标题
				if knowledge.Title == "" {
					title, err = organizer.GenerateTitleFromSession(session)
					if err != nil {
						fmt.Printf("    ! 生成标题失败: %v\n", err)
					}
				} else {
					title = knowledge.Title // 保持已有标题
				}

				// 构建完整对话内容用于摘要生成
				var conversationText string
				for i, msg := range session.Messages {
					role := "用户"
					if msg.Role == "assistant" {
						role = "AI助手"
					}
					conversationText += fmt.Sprintf("%d. %s: %s\n", i+1, role, msg.Content)
				}

				// 生成摘要（使用完整会话内容）
				organizeResult, err := organizer.Organize(conversationText)
				if err != nil {
					fmt.Printf("    ! 生成摘要失败: %v，使用原始内容\n", err)
					summary = knowledge.Content
				} else {
					summary = organizeResult.Summary
					keyPoints = organizeResult.KeyPoints
					category = organizeResult.Category
					tags = organizeResult.Tags
					fmt.Printf("    → 摘要: %s\n", summary)
				}
			}
		}

		// 如果没有会话，使用内容生成
		if summary == "" && knowledge.Content != "" {
			fmt.Printf("    → 使用知识内容生成标题和摘要...\n")

			// 生成标题
			if title == "" {
				contentForTitle := knowledge.Content
				if len(contentForTitle) > 200 {
					contentForTitle = contentForTitle[:200] + "..."
				}

				prompt := `请为以下内容生成一个简洁的标题（5-15个字符）：

内容：` + contentForTitle + `

【要求】
1. 标题长度：5-15个字符
2. 简洁明了，一眼就能看出内容主题
3. 使用简体中文
4. 直接输出标题，不要其他文字`

				messages := []service.Message{
					{Role: "user", Content: prompt},
				}

				req := service.ChatRequest{
					Model:       "glm-4-flash",
					MaxTokens:   100,
					Messages:    messages,
					Temperature: 0.3,
				}

				resp, err := glmClient.SendMessage(req)
				if err == nil {
					title = resp.GetReplyText()
					if len(title) > 3 && title[:3] == "```" {
						title = title[3:]
					}
					if len(title) > 3 && title[len(title)-3:] == "```" {
						title = title[:len(title)-3]
					}
					title = trimQuotes(title)
				}
			}

			// 生成摘要
			organizeResult, err := organizer.Organize(knowledge.Content)
			if err != nil {
				fmt.Printf("    ! 生成摘要失败: %v\n", err)
				summary = knowledge.Content
			} else {
				summary = organizeResult.Summary
				keyPoints = organizeResult.KeyPoints
				category = organizeResult.Category
				tags = organizeResult.Tags
			}
		}

		if title == "" && summary == "" {
			fmt.Printf("    ✗ 生成失败：标题和摘要都为空\n\n")
			errorCount++
			continue
		}

		// 更新知识
		if title != "" {
			knowledge.Title = title
		}
		if summary != "" {
			knowledge.Summary = summary
		}
		if len(keyPoints) > 0 {
			knowledge.KeyPoints = keyPoints
		}
		if category != "" {
			knowledge.Category = category
		}
		if len(tags) > 0 {
			knowledge.Tags = tags
		}

		if err := database.SaveKnowledge(&knowledge); err != nil {
			fmt.Printf("    ✗ 保存失败: %v\n\n", err)
			errorCount++
			continue
		}

		fmt.Printf("    ✓ 已更新: 标题='%s', 摘要='%s'\n\n", title, summary)
		updatedCount++
	}

	// 输出统计
	fmt.Println("\n========== 迁移完成 ==========")
	fmt.Printf("总计: %d 条\n", len(knowledges))
	fmt.Printf("更新: %d 条\n", updatedCount)
	fmt.Printf("失败: %d 条\n", errorCount)
	fmt.Println("==============================")
}

// trimQuotes 去除字符串两端的引号
func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
