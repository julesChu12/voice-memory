package service

import "time"

// Knowledge 知识条目
type Knowledge struct {
	ID         string            `json:"id"`
	Title      string            `json:"title"`       // 会话标题（AI 自动生成）
	Content    string            `json:"content"`     // 原始内容（STT 文本）
	Summary    string            `json:"summary"`     // AI 提取的摘要
	KeyPoints  []string          `json:"key_points"`  // 关键点
	Entities   Entities          `json:"entities"`    // 实体信息
	Category   string            `json:"category"`    // 分类（AI 自动）
	Tags       []string          `json:"tags"`        // 标签（AI 自动）
	Importance string            `json:"importance"`  // 重要性 (high/medium/low)
	Sentiment  string            `json:"sentiment"`   // 情感 (positive/neutral/negative)
	Source     string            `json:"source"`      // 来源：voice/file
	AudioURL   string            `json:"audio_url"`   // 音频文件路径（可选）
	SessionID  string            `json:"session_id"`  // 关联的会话 ID
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Metadata   map[string]string `json:"metadata"`    // 扩展元数据
}

// KnowledgeStoreData 知识库存储数据
type KnowledgeStoreData struct {
	Knowledges []Knowledge `json:"knowledges"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// Entities 实体信息
type Entities struct {
	People    []string `json:"people"`    // 人名
	Products  []string `json:"products"`  // 产品/工具
	Companies []string `json:"companies"` // 公司/品牌
	Locations []string `json:"locations"` // 地点
}

// KnowledgeOrganizeResult AI 整理结果 (v1.0)
type KnowledgeOrganizeResult struct {
	Summary   string   `json:"summary"`
	KeyPoints []string `json:"key_points"`
	Entities  Entities `json:"entities"`  // 实体信息
	Category  string   `json:"category"`
	Tags      []string `json:"tags"`
	Importance string  `json:"importance"` // high/medium/low
	Sentiment  string  `json:"sentiment"`  // positive/neutral/negative
}
