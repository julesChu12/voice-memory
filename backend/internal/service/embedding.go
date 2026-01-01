package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EmbeddingClient Embedding 客户端
type EmbeddingClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewEmbeddingClient 创建 Embedding 客户端
func NewEmbeddingClient(apiKey string) *EmbeddingClient {
	return &EmbeddingClient{
		apiKey:  apiKey,
		baseURL: "https://open.bigmodel.cn/api/paas/v4/embeddings",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// EmbeddingRequest Embedding 请求
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse Embedding 响应
type EmbeddingResponse struct {
	Model string `json:"model"`
	Data  []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// EmbeddingResult Embedding 结果
type EmbeddingResult struct {
	Vector    []float32 `json:"vector"`
	Tokens    int       `json:"tokens"`
	Duration  int64     `json:"duration_ms"`
	CreatedAt time.Time `json:"created_at"`
}

// Embed 生成单个文本的向量
func (c *EmbeddingClient) Embed(text string) (*EmbeddingResult, error) {
	start := time.Now()

	req := EmbeddingRequest{
		Model: "embedding-2", // 智谱 Embedding 模型
		Input: []string{text},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	// 构建请求
	httpReq, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 发送请求
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API 调用失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API 错误 [%d]: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result EmbeddingResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("未返回向量数据")
	}

	duration := time.Since(start).Milliseconds()

	return &EmbeddingResult{
		Vector:    result.Data[0].Embedding,
		Tokens:    result.Usage.TotalTokens,
		Duration:  duration,
		CreatedAt: time.Now(),
	}, nil
}

// GetEmbedding 实现 EmbeddingService 接口
func (c *EmbeddingClient) GetEmbedding(text string) ([]float32, error) {
	result, err := c.Embed(text)
	if err != nil {
		return nil, err
	}
	return result.Vector, nil
}

// EmbedBatch 批量生成向量（优化性能）
func (c *EmbeddingClient) EmbedBatch(texts []string) ([]*EmbeddingResult, error) {
	req := EmbeddingRequest{
		Model: "embedding-2",
		Input: texts,
	}

	jsonData, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API 错误 [%d]: %s", resp.StatusCode, string(body))
	}

	var result EmbeddingResponse
	json.Unmarshal(body, &result)

	results := make([]*EmbeddingResult, len(result.Data))
	now := time.Now()
	for i, data := range result.Data {
		results[i] = &EmbeddingResult{
			Vector:    data.Embedding,
			Tokens:    result.Usage.TotalTokens / len(result.Data), // 估算
			CreatedAt: now,
		}
	}

	return results, nil
}
