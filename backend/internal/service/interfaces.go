package service

// STTService 语音转文字服务接口
type STTService interface {
	Recognize(req *RecognizeRequest) ([]string, error)
}

// LLMService 大语言模型服务接口
type LLMService interface {
	// SendMessageStream 流式发送消息
	// callback: 每收到一个 chunk 就回调一次
	SendMessageStream(req ChatRequest, callback func(StreamChunk)) error
}

// TTSService 文字转语音服务接口
type TTSService interface {
	// Synthesize 合成语音
	Synthesize(options TTSOptions) ([]byte, error)
}

// IntentService 意图识别服务接口
type IntentService interface {
	Recognize(text string) IntentResult
}