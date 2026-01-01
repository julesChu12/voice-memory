package service

// STTService 语音转文字服务接口
type STTService interface {
	Recognize(req *RecognizeRequest) ([]string, error)
}

// LLMService 大语言模型服务接口
type LLMService interface {
	// SendMessage 发送消息（非流式）
	SendMessage(req ChatRequest) (*ChatResponse, error)
	// SendMessageStream 流式发送消息
	// callback: 每收到一个 chunk 就回调一次
	SendMessageStream(req ChatRequest, callback func(StreamChunk)) error
}

// TTSService 文字转语音服务接口
type TTSService interface {
	// Synthesize 合成语音
	Synthesize(options TTSOptions) ([]byte, error)
	// SynthesizeToFile 合成语音并保存到文件
	SynthesizeToFile(options TTSOptions) (string, error)
	// ServeAudio 提供音频文件
	ServeAudio(filename string) ([]byte, string, error)
}

// IntentService 意图识别服务接口
type IntentService interface {
	Recognize(text string) IntentResult
}

// EmbeddingService 文本向量化服务接口
type EmbeddingService interface {
	GetEmbedding(text string) ([]float32, error)
}

// VectorResult 向量搜索结果
type VectorResult struct {
	ID         string                 `json:"id"`
	Score      float32                `json:"score"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// VectorStore 向量存储接口
type VectorStore interface {
	// Add 添加向量
	Add(id string, embedding []float32, metadata map[string]interface{}) error
	// Search 搜索相似向量
	Search(embedding []float32, limit int) ([]VectorResult, error)
	// Delete 删除向量
	Delete(id string) error
}