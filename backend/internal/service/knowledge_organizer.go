package service

import (
	"encoding/json"
	"fmt"
)

// KnowledgeOrganizer 知识整理器
type KnowledgeOrganizer struct {
	llmService LLMService
}

// NewKnowledgeOrganizer 创建知识整理器
func NewKnowledgeOrganizer(llmService LLMService) *KnowledgeOrganizer {
	return &KnowledgeOrganizer{
		llmService: llmService,
	}
}

// Organize 自动整理知识内容 (v2.0 - 关系增强)
func (o *KnowledgeOrganizer) Organize(content string) (*KnowledgeOrganizeResult, error) {
	systemPrompt := `你是 Voice Memory 的知识整理助手，负责创建结构化的语义知识图谱。

【核心任务】将对话转化为知识条目，提取实体、关系和上下文，构建可连接的知识网络。

【输出格式】JSON：
{
  "summary": "核心要点（30字以内）",
  "key_points": ["关键点1", "关键点2"],
  "entities": {
    "people": ["人名"],
    "products": ["产品/工具"],
    "companies": ["公司/品牌"],
    "locations": ["地点"],
    "concepts": ["概念/术语"]
  },
  "category": "分类",
  "tags": ["标签1", "标签2"],
  "relations": [
    {
      "type": "relates_to|requires|implements|improves|contrasts_with",
      "target": "相关实体或概念",
      "context": "关系说明"
    }
  ],
  "observations": [
    {
      "category": "fact|decision|preference|technique|question",
      "content": "观察内容",
      "context": "上下文"
    }
  ],
  "sentiment": "positive/neutral/negative",
  "importance": "high/medium/low",
  "action_items": ["可能的行动项1", "行动项2"]
}

【关系类型定义】
- **relates_to**: 一般关联，主题相关
- **requires**: 依赖关系，需要X才能Y
- **implements**: 实现关系，Y是X的具体实现
- **improves**: 改进关系，Y是对X的优化
- **contrasts_with**: 对比关系，Y与X形成对比
- **part_of**: 包含关系，Y是X的一部分
- **inspired_by**: 灵感来源，Y受X启发
- **alternative_to**: 替代关系，Y可替代X

【观察类型分类】
- **fact**: 事实、数据、信息
- **decision**: 决策、选择、决定
- **preference**: 偏好、喜好、习惯
- **technique**: 技术、方法、技巧
- **question**: 问题、疑问、未解决

【实体增强 - 概念提取】
- 提取专业术语、技术概念、业务名词
- 用于构建知识图谱的核心节点
- 例如："微服务"、"CI/CD","敏捷开发"

【关系构建原则】
- 引用对话中明确提及的相关实体
- 推测隐含的关联关系
- 创建前向引用（可能尚未存在的实体）
- 双向关系优先（如果A与B相关，B也与A相关）

【观察提取原则】
- 每条观察独立、语义完整
- 包含类型标记，便于后续查询
- 添加上下文说明，理解更准确
- 最多 7 条观察，按相关性排序

【行动项提取】
- 提取对话中的待办事项
- 明确下一步行动
- 可选字段，非必要`

	messages := []Message{
		{
			Role:    "user",
			Content: systemPrompt + "\n\n内容：" + content,
		},
	}

	req := ChatRequest{
		Model:       "glm-4.7", // 升级到最新模型，提升提取质量
		MaxTokens:   1024,          // 增加 Token 数量以适应更复杂的输出
		Messages:    messages,
		Temperature: 0.2, // 更低的温度，确保结构化输出的稳定性
	}

	resp, err := o.llmService.SendMessage(req)
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
			Summary:   content[:min(len(content), 30)],
			Category:  "想法",
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
		Model:       "glm-4.7",
		MaxTokens:   100,
		Messages:    messages,
		Temperature: 0.3,
	}

	resp, err := o.llmService.SendMessage(req)
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
