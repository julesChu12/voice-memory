package service

import (
	"encoding/json"
	"fmt"
)

// KnowledgeOrganizer 知识整理器
type KnowledgeOrganizer struct {
	glmClient *GLMClient
}

// NewKnowledgeOrganizer 创建知识整理器
func NewKnowledgeOrganizer(glmClient *GLMClient) *KnowledgeOrganizer {
	return &KnowledgeOrganizer{
		glmClient: glmClient,
	}
}

// Organize 自动整理知识内容 (v1.0 improved)
func (o *KnowledgeOrganizer) Organize(content string) (*KnowledgeOrganizeResult, error) {
	systemPrompt := `你是 Voice Memory 的知识整理助手，负责将对话内容转化为结构化、可检索的知识条目。

【核心任务】从完整对话中提取核心信息，创建易于检索的知识记录。

【输出格式】JSON：
{
  "summary": "核心要点摘要（30字以内，一句话概括主题）",
  "key_points": [
    "关键点1（具体、可操作的信息）",
    "关键点2（数据、结论或建议）",
    "关键点3（重要细节或注意事项）"
  ],
  "entities": {
    "people": ["人名1", "人名2"],
    "products": ["产品名1", "工具名2", "框架名3"],
    "companies": ["公司名1", "品牌名2"],
    "locations": ["地点1", "地区2"]
  },
  "category": "分类（见下方分类体系）",
  "tags": ["标签1", "标签2", "标签3"],
  "importance": "high/medium/low",
  "sentiment": "positive/neutral/negative"
}

【分类体系 - 二级分类】

### 技术 (technology)
- 编程开发 (coding): 编程语言、算法、代码实现
- 架构设计 (architecture): 系统架构、设计模式、技术选型
- 开发工具 (tools): IDE、框架、库、开发环境
- 运维部署 (devops): CI/CD、容器、云服务、监控
- 技术方案 (solution): 具体问题的技术解决方案

### 生活 (life)
- 健康养生 (health): 运动、医疗、饮食、作息
- 购物消费 (shopping): 产品对比、价格、购买决策
- 美食烹饪 (food): 菜谱、餐厅、营养、食材
- 旅行出行 (travel): 景点、交通、住宿、行程
- 生活技巧 (tips): 生活小妙招、经验总结

### 工作 (work)
- 会议记录 (meeting): 会议内容、决策、行动项
- 项目管理 (project): 项目进展、里程碑、风险管理
- 任务计划 (task): 待办事项、提醒、目标
- 职业发展 (career): 求职、技能提升、职业规划
- 工作决策 (decision): 工作相关的选择和判断

### 学习 (learning)
- 学习笔记 (note): 知识总结、概念理解
- 教程指南 (tutorial): 操作步骤、how-to、入门指南
- 学习资源 (resource): 书籍、课程、文档、链接
- 问题答疑 (qa): 疑难点、常见错误、易错点
- 考试考证 (exam): 考试准备、知识点、复习策略

### 想法 (thought)
- 灵感创意 (inspiration): 创意点子、灵感记录
- 思考总结 (reflection): 复盘、反思、感悟
- 目标计划 (goal): 目标设定、计划安排
- 观点看法 (opinion): 对某事的评价、观点

【实体提取规则】
- 人名: 真实人名、昵称、角色
- 产品: 软件、硬件、工具、框架、库
- 公司: 企业、品牌、组织、机构
- 地点: 城市、国家、地址、场所

【关键点提取标准】
- 每个关键点应该：独立、具体、有价值
- 优先提取：数据、结论、建议、操作步骤
- 避免泛泛而谈，要具体可执行
- 最多 5 个关键点，按重要性排序

【标签生成原则】
- 3-5 个标签，覆盖核心主题
- 包含：实体名、二级分类、关键属性
- 使用简短词语（2-4 字）

【重要性判断】
- high: 重要决策、关键数据、高价值信息
- medium: 有用知识、经验总结
- low: 随手记录、临时信息

【情感判断】
- positive: 积极、正面、满意、成功
- neutral: 中性、客观、事实陈述
- negative: 消极、问题、失败、抱怨`

	messages := []Message{
		{
			Role:    "user",
			Content: systemPrompt + "\n\n内容：" + content,
		},
	}

	req := ChatRequest{
		Model:       "glm-4-flash", // 使用快速模型
		MaxTokens:   512,
		Messages:    messages,
		Temperature: 0.3, // 低温度，更稳定
	}

	resp, err := o.glmClient.SendMessage(req)
	if err != nil {
		return nil, fmt.Errorf("AI 整理失败: %w", err)
	}

	// 解析 JSON 响应
	reply := resp.GetReplyText()

	// 清理可能的 markdown 代码块标记
	reply = cleanMarkdownCode(reply)

	var result KnowledgeOrganizeResult
	if err := json.Unmarshal([]byte(reply), &result); err != nil {
		// 如果解析失败，返回基础整理
		return &KnowledgeOrganizeResult{
			Summary:   content[:min(len(content), 20)],
			KeyPoints: []string{},
			Category:  "想法",
			Tags:      []string{},
		}, nil
	}

	return &result, nil
}

// cleanMarkdownCode 清理 markdown 代码块标记
func cleanMarkdownCode(s string) string {
	// 去除开头的 ```json 或 ```
	if len(s) > 7 && s[0:7] == "```json" {
		s = s[7:]
	} else if len(s) > 3 && s[0:3] == "```" {
		s = s[3:]
	}

	// 去除结尾的 ```
	if len(s) > 3 && s[len(s)-3:] == "```" {
		s = s[:len(s)-3]
	}

	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateTitleFromSession 根据会话内容生成标题 (v1.0 improved)
func (o *KnowledgeOrganizer) GenerateTitleFromSession(session *Session) (string, error) {
	// 构建会话摘要提示
	var conversationText string
	for i, msg := range session.Messages {
		role := "用户"
		if msg.Role == "assistant" {
			role = "AI助手"
		}
		conversationText += fmt.Sprintf("%d. %s: %s\n", i+1, role, msg.Content)
	}

	systemPrompt := `你是 Voice Memory 的会话标题生成助手。

【核心任务】为对话生成简洁、准确、易识别的标题，便于快速浏览和检索。

【标题要求】
1. 长度: 8-20 个字符（最佳 12-15 字）
2. 风格: 简洁明确，包含核心主题
3. 优先级: 具体概念 > 抽象主题
4. 格式: 直接输出标题文本，无引号、无标点包裹
5. 语言: 简体中文，可含常见英文术语（如 gin、Redis）

【标题生成策略】

### 技术类对话
公式: 技术/框架 + 核心目的
示例:
- gin框架性能优化
- Redis缓存设计
- Go vs Python性能对比

避免: "关于框架选择"、"技术讨论"

### 产品/工具讨论
公式: 产品名 + 用途/特点
示例:
- Notion知识管理
- iPhone拍照技巧
- ChatGPT使用指南

避免: "某产品使用"、"工具推荐"

### 决策/选择类
公式: 动词 + 对象 + 可选补充
示例:
- 选购云服务器方案
- 选择Go Web框架
- 机械键盘对比选购

避免: "关于选择"、"决策讨论"

### 学习/教程类
公式: 主题 + 类型
示例:
- Docker容器入门
- RAG架构学习
- Go并发编程教程

避免: "学习记录"、"知识总结"

### 生活/购物类
公式: 商品/主题 + 核心关注点
示例:
- 办公机械键盘选购指南
- 汾酒品牌与价格
- 健身计划制定

避免: "购物记录"、"生活琐事"

### 多主题对话
选择: 最重要的主题，或综合概括
示例:
- Go微服务架构设计（综合）
- 汾酒选购与品鉴（综合）

【避免模式】
- "关于XX"、"XX讨论" - 过于泛化
- "XX记录"、"XX笔记" - 缺少具体主题
- "XX问题"、"XX相关" - 不够具体
- 纯疑问句 - 使用陈述式更明确

【示例对照表】

| 对话主题 | 好的标题 |
|---------|---------|
| 讨论Go框架对比 | Go框架选型：gin vs echo |
| 学习Docker | Docker容器化入门 |
| 购买机械键盘 | 办公机械键盘选购指南 |
| 汾酒知识 | 汾酒品牌与选购指南 |
| 重构用户服务 | 用户服务Go重构方案 |
| 性能优化讨论 | API服务性能优化方案 |

【输出格式】
仅输出标题文本，不包含引号、括号、书名号，不添加任何解释。

【对话内容】`

	messages := []Message{
		{
			Role:    "user",
			Content: systemPrompt + "\n\n" + conversationText,
		},
	}

	req := ChatRequest{
		Model:       "glm-4-flash",
		MaxTokens:   100,
		Messages:    messages,
		Temperature: 0.3,
	}

	resp, err := o.glmClient.SendMessage(req)
	if err != nil {
		return "", fmt.Errorf("AI 生成标题失败: %w", err)
	}

	title := resp.GetReplyText()
	// 清理可能的 markdown 代码块标记
	if len(title) > 3 {
		if title[0:3] == "```" {
			title = title[3:]
		}
		if len(title) > 3 && title[len(title)-3:] == "```" {
			title = title[:len(title)-3]
		}
	}
	// 去除可能的引号
	title = trimQuotes(title)

	return title, nil
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
