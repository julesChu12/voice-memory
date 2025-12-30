package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"voice-memory/internal/service"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dataDir := "./data"
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}

	// 初始化数据库
	database, err := service.NewDatabase(dataDir)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.Close()

	// 初始化 AI 客户端（用于生成标题和摘要）
	apiKey := os.Getenv("GLM_API_KEY")
	if apiKey == "" {
		log.Println("未设置 GLM_API_KEY，将使用默认内容")
	}

	// 构建汾酒知识内容
	fenjiuContent := `汾酒是中国老八大名酒之一，清香型白酒的代表，有4000多年历史。

完整对话内容：
1. 用户: 给我讲解一下关于汾酒的知识把
2. AI助手: 哈，我这里关于汾酒的知识库内容不多呢~ 不过我知道汾酒是中国老八大名酒之一，清香型白酒的代表，有4000多年历史呢！你想了解它的酿造工艺、品鉴方法还是其他方面？
3. 用户: 从头给我讲讲把
4. AI助手: 哈哈，汾酒的故事要从4000多年前说起啦！它是清香型白酒鼻祖，产自山西杏花村。名字源于"汾河"，用高粱、豌豆等发酵，口感清冽甘爽，还被誉为"国酒"哦！
5. 用户: 继续
6. AI助手: 汾酒的传统工艺可是宝贝呢！地缸发酵、清蒸清烧，不沾杂味。蒸馏时酒香四溢，陈酿后口感更醇厚。最特别的是它的"一清到底"，喝起来清爽不辣喉~
7. 用户: 好的
8. AI助手: 哈哈，"好的"重复啦～看来你很期待呀！我猜你可能想先了解汾酒的经典款？比如黄盖玻汾性价比超高，青花汾更适合送礼～
9. 用户: 最高端的是什么？
10. AI助手: 汾酒的高端系列是"青花汾酒"，特别是30年、40年陈酿！酒体醇厚，包装精致，收藏送礼都超有面子～
11. 用户: 分别是什么价格？
12. AI助手: 青花30年大约800-1200元，40年要2000-3000元以上呢！
13. 用户: 去京东还是淘宝？
14. AI助手: 京东自营或天猫官方旗舰店更靠谱！物流快、保真，售后也方便～
15. 用户: 咱们之前聊过酒的知识吗？
16. AI助手: 当然聊过呢！我们刚详细聊过汾酒的历史、工艺和高端系列，还问了你对青花汾的兴趣哈～
17. 用户: 我都问了那些内容？
18. AI助手: 你问了汾酒的历史、酿造工艺、品鉴方法和高端系列价格呢！`

	// 生成知识条目
	knowledge := &service.Knowledge{
		ID:      fmt.Sprintf("kb_%d", time.Now().UnixNano()),
		Title:   "汾酒知识",
		Content: fenjiuContent,
		Summary: "汾酒是中国老八大名酒之一，清香型白酒的代表，有4000多年历史，产自山西杏花村。采用地缸发酵、清蒸清烧工艺，口感清冽甘爽。高端系列青花汾酒30年约800-1200元，40年约2000-3000元。",
		KeyPoints: []string{
			"汾酒是中国老八大名酒之一",
			"清香型白酒代表，4000多年历史",
			"产地：山西杏花村",
			"工艺：地缸发酵、清蒸清烧",
			"高端系列：青花汾酒",
			"价格：30年800-1200元，40年2000-3000元",
		},
		Category:  "生活",
		Tags:      []string{"汾酒", "白酒", "酒类知识"},
		Source:    "manual",
		SessionID: "session_1767113670786562000",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]string),
	}

	// 保存到数据库
	if err := database.SaveKnowledge(knowledge); err != nil {
		log.Fatalf("保存知识失败: %v", err)
	}

	fmt.Printf("✅ 汾酒知识已恢复！\n")
	fmt.Printf("ID: %s\n", knowledge.ID)
	fmt.Printf("标题: %s\n", knowledge.Title)
	fmt.Printf("摘要: %s\n", knowledge.Summary)
}
