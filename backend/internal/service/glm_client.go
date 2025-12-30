package service

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GLMClient 智谱 GLM API 客户端
type GLMClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewGLMClient 创建 GLM 客户端
func NewGLMClient(apiKey string) *GLMClient {
	return &GLMClient{
		apiKey:  apiKey,
		baseURL: "https://open.bigmodel.cn/api/anthropic",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Message 消息
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string 或 []ContentBlock
}

// ContentBlock 多模态内容块
type ContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	AudioURL  *AudioURL       `json:"audio_url,omitempty"`
	ImageURL  *ImageURL       `json:"image_url,omitempty"`
}

// AudioURL 音频 URL
type AudioURL struct {
	URL string `json:"url"`
}

// ImageURL 图片 URL
type ImageURL struct {
	URL string `json:"url"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Content 内容块
type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Delta 增量内容（流式）
type Delta struct {
	Type       string     `json:"type"`
	Text       string     `json:"text,omitempty"`
	StopReason string     `json:"stop_reason,omitempty"`
}

// Usage 使用量
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ChatResponseBlock 响应内容块
type ChatResponseBlock struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Role     string   `json:"role"`
	Content  []Content `json:"content"`
	Model    string   `json:"model"`
	StopReason string  `json:"stop_reason"`
	Usage    Usage   `json:"usage"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID           string             `json:"id"`
	Type         string             `json:"type"`
	Role         string             `json:"role"`
	Content      []Content          `json:"content"`
	Model        string             `json:"model"`
	StopReason   string             `json:"stop_reason"`
	Usage        Usage              `json:"usage"`
	Delta        Delta              `json:"delta,omitempty"`
}

// SendMessage 发送消息（非流式）
func (g *GLMClient) SendMessage(req ChatRequest) (*ChatResponse, error) {
	req.Stream = false

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	fmt.Printf("GLM API 请求: %s\n", string(jsonData))

	httpReq, err := http.NewRequest("POST", g.baseURL+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://voice-memory.app")
	httpReq.Header.Set("X-Title", "Voice Memory")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API 错误 [%d]: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("GLM API 响应: %s\n", string(body))

	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w\n原始响应: %s", err, string(body))
	}

	return &response, nil
}

// GetReplyText 提取回复文本
func (r *ChatResponse) GetReplyText() string {
	if r.Type == "message" {
		// 非流式响应
		var text string
		for _, block := range r.Content {
			if block.Type == "text" {
				text += block.Text
			}
		}
		return text
	} else if r.Type == "message_delta" {
		// 流式响应
		return r.Delta.Text
	}
	return ""
}

// SendMessageWithAudio 发送带音频的消息（GLM-4 Audio）
func (g *GLMClient) SendMessageWithAudio(audioData []byte, messages []Message) (*ChatResponse, error) {
	// 将音频编码为 base64
	audioBase64 := base64.StdEncoding.EncodeToString(audioData)
	audioURL := fmt.Sprintf("data:audio/wav;base64,%s", audioBase64)

	// 构建音频内容块
	audioContent := []ContentBlock{
		{
			Type: "audio_url",
			AudioURL: &AudioURL{
				URL: audioURL,
			},
		},
	}

	// 如果有历史消息，需要保持格式一致
	// 将所有消息转换为多模态格式
	allMessages := make([]Message, 0, len(messages)+1)

	// 添加历史消息（如果是字符串格式，转换为文本块）
	for _, msg := range messages {
		if str, ok := msg.Content.(string); ok {
			allMessages = append(allMessages, Message{
				Role: msg.Role,
				Content: []ContentBlock{
					{Type: "text", Text: str},
				},
			})
		} else if blocks, ok := msg.Content.([]ContentBlock); ok {
			allMessages = append(allMessages, Message{
				Role:    msg.Role,
				Content: blocks,
			})
		}
	}

	// 添加当前音频消息
	allMessages = append(allMessages, Message{
		Role:    "user",
		Content: audioContent,
	})

	req := ChatRequest{
		Model:       "glm-4-plus",  // GLM-4-Plus 支持多模态（包括音频）
		MaxTokens:   1024,
		Messages:    allMessages,
		Temperature: 0.85,
		Stream:      false,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	fmt.Printf("GLM Audio API 请求 (音频大小: %d bytes): %s\n", len(audioData), string(jsonData))

	httpReq, err := http.NewRequest("POST", g.baseURL+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API 错误 [%d]: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("GLM Audio API 响应: %s\n", string(body))

	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w\n原始响应: %s", err, string(body))
	}

	return &response, nil
}

// StreamChunk 流式响应数据块
type StreamChunk struct {
	Delta    string `json:"delta"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

// GLMStreamEvent GLM 流式事件
type GLMStreamEvent struct {
	Type  string      `json:"type"`
	Index int         `json:"index,omitempty"`
	Delta *GLMDelta   `json:"delta,omitempty"`
}

// GLMDelta GLM 增量内容
type GLMDelta struct {
	Type        string `json:"type"`
	Text        string `json:"text,omitempty"`
	StopReason  string `json:"stop_reason,omitempty"`
}

// SendMessageStream 发送消息（流式）
func (g *GLMClient) SendMessageStream(req ChatRequest, callback func(StreamChunk)) error {
	req.Stream = true

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", g.baseURL+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://voice-memory.app")
	httpReq.Header.Set("X-Title", "Voice Memory")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API 错误 [%d]: %s", resp.StatusCode, string(body))
	}

	// 读取流式响应
	scanner := bufio.NewScanner(resp.Body)
	lineCount := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		// 跳过空行
		if line == "" {
			continue
		}

		// 只处理 data: 行，忽略 event: 行
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// 移除 "data: " 前缀
		data := strings.TrimPrefix(line, "data: ")
		data = strings.TrimSpace(data)

		// 结束标记
		if data == "[DONE]" {
			callback(StreamChunk{Done: true})
			break
		}

		// 尝试解析 JSON
		var event GLMStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		// 处理不同类型的事件
		if event.Type == "content_block_delta" && event.Delta != nil && event.Delta.Type == "text_delta" {
			// 文本增量
			if event.Delta.Text != "" {
				callback(StreamChunk{Delta: event.Delta.Text})
			}
		} else if event.Type == "message_delta" && event.Delta != nil && event.Delta.StopReason != "" {
			// 消息结束
			callback(StreamChunk{Done: true})
			break
		} else if event.Type == "message_stop" {
			// 消息停止
			callback(StreamChunk{Done: true})
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取流失败: %w", err)
	}

	// 如果没有收到任何数据，调用完成
	if lineCount == 0 {
		callback(StreamChunk{Done: true})
	}

	return nil
}
