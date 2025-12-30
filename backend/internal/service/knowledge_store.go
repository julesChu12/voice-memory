package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// KnowledgeStore 知识库存储
type KnowledgeStore struct {
	dataFile string
	data     *KnowledgeStoreData
	mu       sync.RWMutex
}

// NewKnowledgeStore 创建知识库存储
func NewKnowledgeStore(dataDir string) (*KnowledgeStore, error) {
	// 确保目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	dataFile := filepath.Join(dataDir, "knowledge.json")

	store := &KnowledgeStore{
		dataFile: dataFile,
		data: &KnowledgeStoreData{
			Knowledges: []Knowledge{},
			UpdatedAt:  time.Now(),
		},
	}

	// 加载已有数据
	if err := store.load(); err != nil {
		return nil, err
	}

	fmt.Printf("知识库初始化完成，共 %d 条记录\n", len(store.data.Knowledges))
	return store, nil
}

// load 从文件加载数据
func (s *KnowledgeStore) load() error {
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，使用默认值
			return nil
		}
		return err
	}

	return json.Unmarshal(data, s.data)
}

// save 保存数据到文件
func (s *KnowledgeStore) save() error {
	s.data.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.dataFile, data, 0644)
}

// Add 添加知识条目
func (s *KnowledgeStore) Add(knowledge *Knowledge) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	knowledge.CreatedAt = time.Now()
	knowledge.UpdatedAt = time.Now()

	s.data.Knowledges = append(s.data.Knowledges, *knowledge)

	if err := s.save(); err != nil {
		return fmt.Errorf("保存知识库失败: %w", err)
	}

	fmt.Printf("知识已保存: %s (分类: %s)\n", knowledge.ID, knowledge.Category)
	return nil
}

// List 列出所有知识
func (s *KnowledgeStore) List() []Knowledge {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回副本，避免外部修改
	result := make([]Knowledge, len(s.data.Knowledges))
	copy(result, s.data.Knowledges)

	// 按时间倒序
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// GetByID 根据 ID 获取知识
func (s *KnowledgeStore) GetByID(id string) (*Knowledge, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, k := range s.data.Knowledges {
		if k.ID == id {
			return &k, true
		}
	}

	return nil, false
}

// Search 搜索知识（简单文本匹配，后续替换为向量检索）
func (s *KnowledgeStore) Search(query string) []Knowledge {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []Knowledge

	for _, k := range s.data.Knowledges {
		// 简单匹配：标题、内容、标签
		if contains(k.Content, query) ||
			contains(k.Summary, query) ||
			contains(k.Category, query) ||
			containsAny(k.Tags, query) {
			results = append(results, k)
		}
	}

	return results
}

// GetByCategory 按分类获取
func (s *KnowledgeStore) GetByCategory(category string) []Knowledge {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []Knowledge
	for _, k := range s.data.Knowledges {
		if k.Category == category {
			results = append(results, k)
		}
	}

	return results
}

// Stats 统计信息
func (s *KnowledgeStore) Stats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]int)
	stats["total"] = len(s.data.Knowledges)

	categoryCount := make(map[string]int)
	for _, k := range s.data.Knowledges {
		categoryCount[k.Category]++
	}

	for cat, count := range categoryCount {
		stats[cat] = count
	}

	return stats
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(substr) == 0 ||
		findSubstring(s, substr))
}

func containsAny(slice []string, substr string) bool {
	for _, s := range slice {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
