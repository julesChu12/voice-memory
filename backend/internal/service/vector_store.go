package service

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// VectorDocument 向量文档
type VectorDocument struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Vector    []float32              `json:"vector"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Document   *VectorDocument `json:"document"`
	Score      float64         `json:"score"`
	Distance   float64         `json:"distance"`
}

// VectorStore 向量存储（内存版）
type VectorStore struct {
	mu        sync.RWMutex
	documents map[string]*VectorDocument
	dimension int
}

// NewVectorStore 创建向量存储
func NewVectorStore() *VectorStore {
	return &VectorStore{
		documents: make(map[string]*VectorDocument),
		dimension: 1024, // embedding-2 模型的维度
	}
}

// Add 添加文档到向量存储
func (vs *VectorStore) Add(doc *VectorDocument) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	// 验证向量维度
	if len(doc.Vector) != vs.dimension {
		return fmt.Errorf("向量维度不匹配: 期望 %d, 得到 %d", vs.dimension, len(doc.Vector))
	}

	// 设置创建时间
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now()
	}

	vs.documents[doc.ID] = doc
	return nil
}

// AddBatch 批量添加文档
func (vs *VectorStore) AddBatch(docs []*VectorDocument) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	for _, doc := range docs {
		if len(doc.Vector) != vs.dimension {
			return fmt.Errorf("文档 %s 向量维度不匹配", doc.ID)
		}
		if doc.CreatedAt.IsZero() {
			doc.CreatedAt = time.Now()
		}
		vs.documents[doc.ID] = doc
	}

	return nil
}

// Search 搜索最相似的文档（余弦相似度）
func (vs *VectorStore) Search(queryVector []float32, topK int) ([]*SearchResult, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	// 验证查询向量维度
	if len(queryVector) != vs.dimension {
		return nil, fmt.Errorf("查询向量维度不匹配: 期望 %d, 得到 %d", vs.dimension, len(queryVector))
	}

	if len(vs.documents) == 0 {
		return []*SearchResult{}, nil
	}

	// 计算所有文档的相似度
	results := make([]*SearchResult, 0, len(vs.documents))
	for _, doc := range vs.documents {
		score := cosineSimilarity(queryVector, doc.Vector)
		results = append(results, &SearchResult{
			Document: doc,
			Score:    score,
			Distance: 1 - score, // 距离 = 1 - 相似度
		})
	}

	// 按相似度降序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 返回 top-K 结果
	if topK > len(results) {
		topK = len(results)
	}
	return results[:topK], nil
}

// Get 根据 ID 获取文档
func (vs *VectorStore) Get(id string) (*VectorDocument, bool) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	doc, exists := vs.documents[id]
	return doc, exists
}

// Delete 删除文档
func (vs *VectorStore) Delete(id string) bool {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if _, exists := vs.documents[id]; exists {
		delete(vs.documents, id)
		return true
	}
	return false
}

// Count 获取文档数量
func (vs *VectorStore) Count() int {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	return len(vs.documents)
}

// Clear 清空所有文档
func (vs *VectorStore) Clear() {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	vs.documents = make(map[string]*VectorDocument)
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct float64
	var normA float64
	var normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// UpdateMetadata 更新文档元数据
func (vs *VectorStore) UpdateMetadata(id string, metadata map[string]interface{}) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	doc, exists := vs.documents[id]
	if !exists {
		return fmt.Errorf("文档 %s 不存在", id)
	}

	doc.Metadata = metadata
	return nil
}

// ListAll 列出所有文档（谨慎使用，数据量大时性能差）
func (vs *VectorStore) ListAll() []*VectorDocument {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	docs := make([]*VectorDocument, 0, len(vs.documents))
	for _, doc := range vs.documents {
		docs = append(docs, doc)
	}
	return docs
}
