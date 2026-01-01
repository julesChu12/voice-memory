package service

import "time"

// Knowledge 知识条目
type Knowledge struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Content      string            `json:"content"`
	Summary      string            `json:"summary"`
	KeyPoints    []string          `json:"key_points"`
	Entities     Entities          `json:"entities"`
	Relations    []Relation        `json:"relations"`    // v2.0 new field
	Observations []Observation     `json:"observations"` // v2.0 new field
	ActionItems  []string          `json:"action_items"` // v2.0 new field
	Category     string            `json:"category"`
	Tags         []string          `json:"tags"`
	Importance   string            `json:"importance"`
	Sentiment    string            `json:"sentiment"`
	Source       string            `json:"source"`
	AudioURL     string            `json:"audio_url"`
	SessionID    string            `json:"session_id"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Metadata     map[string]string `json:"metadata"`
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
	Concepts  []string `json:"concepts"`  // 概念/术语
}

// Relation 实体关系
type Relation struct {
	Type    string `json:"type"`    // 关系类型
	Target  string `json:"target"`  // 目标实体
	Context string `json:"context"` // 关系说明
}

// Observation 观察点
type Observation struct {
	Category string `json:"category"` // 观察类型
	Content  string `json:"content"`  // 观察内容
	Context  string `json:"context"`  // 上下文
}

// KnowledgeOrganizeResult AI 整理结果 (v2.0 - 关系增强)
type KnowledgeOrganizeResult struct {
	Summary      string        `json:"summary"`
	KeyPoints    []string      `json:"key_points"`
	Entities     Entities      `json:"entities"`
	Category     string        `json:"category"`
	Tags         []string      `json:"tags"`
	Relations    []Relation    `json:"relations"`
	Observations []Observation `json:"observations"`
	Sentiment    string        `json:"sentiment"`
	Importance   string        `json:"importance"`
	ActionItems  []string      `json:"action_items"`
}
