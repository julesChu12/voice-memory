package main

import (
	"fmt"
	"log"
	"voice-memory/internal/service"
)

func main() {
	db, err := service.NewDatabase("./data")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	knowledges, err := db.GetAllKnowledge()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("📚 当前知识库共有 %d 条记录\n", len(knowledges))
	fmt.Println("------------------------------------------------")
	
	// 显示最近 5 条
	count := 0
	for _, k := range knowledges {
		if count >= 5 {
			break
		}
		fmt.Printf("ID: %s\n摘要: %s\n分类: %s\n标签: %v\n创建时间: %v\n", 
			k.ID, k.Summary, k.Category, k.Tags, k.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("------------------------------------------------")
		count++
	}
}

