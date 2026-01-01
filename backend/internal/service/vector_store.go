package service

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// VectorItem 向量条目
type VectorItem struct {
	ID        string                 `json:"id"`
	Embedding []float32              `json:"embedding"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// SimpleVectorStore 纯 Go 实现的简单向量存储
type SimpleVectorStore struct {
	items    map[string]VectorItem
	filePath string
	mu       sync.RWMutex
}

// NewSimpleVectorStore 创建简单的向量存储
func NewSimpleVectorStore(dataDir string) (*SimpleVectorStore, error) {
	filePath := filepath.Join(dataDir, "vectors.json")
	store := &SimpleVectorStore{
		items:    make(map[string]VectorItem),
		filePath: filePath,
	}

	// 尝试加载现有数据
	if err := store.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("加载向量数据失败: %w", err)
		}
		// 如果文件不存在，初始化为空
	}

	return store, nil
}

// Add 添加向量
func (s *SimpleVectorStore) Add(id string, embedding []float32, metadata map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[id] = VectorItem{
		ID:        id,
		Embedding: embedding,
		Metadata:  metadata,
	}

	return s.save()
}

// Delete 删除向量
func (s *SimpleVectorStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, id)
	return s.save()
}

// Search 搜索相似向量
func (s *SimpleVectorStore) Search(queryVector []float32, limit int) ([]VectorResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type scoredItem struct {
		item  VectorItem
		score float32
	}

	var candidates []scoredItem

	for _, item := range s.items {
		score := cosineSimilarity(queryVector, item.Embedding)
		candidates = append(candidates, scoredItem{item: item, score: score})
	}

	// 排序 (分数从高到低)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	// 取 TopK
	if limit > len(candidates) {
		limit = len(candidates)
	}

	results := make([]VectorResult, limit)
	for i := 0; i < limit; i++ {
		results[i] = VectorResult{
			ID:       candidates[i].item.ID,
			Score:    candidates[i].score,
			Metadata: candidates[i].item.Metadata,
		}
	}

	return results, nil
}

// save 保存到文件
func (s *SimpleVectorStore) save() error {
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

// load 从文件加载
func (s *SimpleVectorStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.items)
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}