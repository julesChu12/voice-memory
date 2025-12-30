package service

import "time"

// Knowledge 知识条目
type Knowledge struct {
	ID         string            `json:"id"`
	Content    string            `json:"content"`    // 原始内容（STT 文本）
	Summary    string            `json:"summary"`    // AI 提取的摘要
	KeyPoints  []string          `json:"key_points"` // 关键点
	Category   string            `json:"category"`   // 分类（AI 自动）
	Tags       []string          `json:"tags"`       // 标签（AI 自动）
	Source     string            `json:"source"`     // 来源：voice/file
	AudioURL   string            `json:"audio_url"`  // 音频文件路径（可选）
	SessionID  string            `json:"session_id"` // 关联的会话 ID
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Metadata   map[string]string `json:"metadata"` // 扩展元数据
}

// KnowledgeStoreData 知识库存储数据
type KnowledgeStoreData struct {
	Knowledges []Knowledge `json:"knowledges"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// KnowledgeOrganizeResult AI 整理结果
type KnowledgeOrganizeResult struct {
	Summary   string   `json:"summary"`
	KeyPoints []string `json:"key_points"`
	Category  string   `json:"category"`
	Tags      []string `json:"tags"`
}
