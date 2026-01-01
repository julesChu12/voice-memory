package service

import (
	"fmt"
	"strings"
	"sync"
)

// RAGService RAG 服务（检索增强生成）
type RAGService struct {
	embeddingClient *EmbeddingClient
	vectorStore     VectorStore
	mu              sync.RWMutex
	enabled         bool
}

// NewRAGService 创建 RAG 服务
func NewRAGService(apiKey string, vectorStore VectorStore) *RAGService {
	return &RAGService{
		embeddingClient: NewEmbeddingClient(apiKey),
		vectorStore:     vectorStore,
		enabled:         true,
	}
}

// AddKnowledge 添加知识到向量库
func (rag *RAGService) AddKnowledge(id, content string, metadata map[string]interface{}) error {
	if !rag.enabled {
		return fmt.Errorf("RAG 服务未启用")
	}

	// 1. 生成向量
	result, err := rag.embeddingClient.Embed(content)
	if err != nil {
		return fmt.Errorf("生成向量失败: %w", err)
	}

	// 2. 存储到向量库
	// Store content in metadata so we can retrieve it
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["content"] = content
	metadata["created_at"] = result.CreatedAt

	if err := rag.vectorStore.Add(id, result.Vector, metadata); err != nil {
		return fmt.Errorf("存储向量失败: %w", err)
	}

	return nil
}

// Retrieve 检索相关知识
func (rag *RAGService) Retrieve(query string, topK int) ([]*RetrievalResult, error) {
	if !rag.enabled {
		return []*RetrievalResult{}, nil
	}

	// 1. 生成查询向量
	result, err := rag.embeddingClient.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("生成查询向量失败: %w", err)
	}

	// 2. 向量搜索
	searchResults, err := rag.vectorStore.Search(result.Vector, topK)
	if err != nil {
		return nil, fmt.Errorf("向量搜索失败: %w", err)
	}

	// 3. 转换为检索结果
	retrievalResults := make([]*RetrievalResult, len(searchResults))
	for i, sr := range searchResults {
		content, _ := sr.Metadata["content"].(string)
		
		retrievalResults[i] = &RetrievalResult{
			ID:      sr.ID,
			Content: content,
			Score:   float64(sr.Score),
			Metadata: sr.Metadata,
		}
	}

	return retrievalResults, nil
}

// BuildContextWithRAG 构建 RAG 增强的上下文
func (rag *RAGService) BuildContextWithRAG(query string, topK int) (string, error) {
	// 检索相关知识
	results, err := rag.Retrieve(query, topK)
	if err != nil {
		return "", err
	}

	// 如果没有相关知识，返回空
	if len(results) == 0 {
		return "", nil
	}

	// 构建上下文
	var context strings.Builder
	context.WriteString("以下是与用户问题相关的知识库内容：\n\n")

	for i, r := range results {
		context.WriteString(fmt.Sprintf("[知识 %d] (相关度: %.2f)\n", i+1, r.Score))
		context.WriteString(r.Content)
		context.WriteString("\n\n")
	}

	return context.String(), nil
}

// DeleteKnowledge 删除指定知识
func (rag *RAGService) DeleteKnowledge(id string) bool {
	return rag.vectorStore.Delete(id) == nil
}

// Enable 启用 RAG 服务
func (rag *RAGService) Enable() {
	rag.mu.Lock()
	defer rag.mu.Unlock()
	rag.enabled = true
}

// Disable 禁用 RAG 服务
func (rag *RAGService) Disable() {
	rag.mu.Lock()
	defer rag.mu.Unlock()
	rag.enabled = false
}

// IsEnabled 检查 RAG 服务是否启用
func (rag *RAGService) IsEnabled() bool {
	rag.mu.RLock()
	defer rag.mu.RUnlock()
	return rag.enabled
}

// RetrievalResult 检索结果
type RetrievalResult struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ShouldUseRAG 判断是否应该使用 RAG（基于意图识别）
func (rag *RAGService) ShouldUseRAG(intent Intent, confidence float64) bool {
	if !rag.enabled {
		return false
	}

	// 搜索意图和问答意图可以使用 RAG
	return (intent == IntentSearch || intent == IntentQuestion) && confidence > 0.1
}