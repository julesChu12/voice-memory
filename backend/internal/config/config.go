package config

import "os"

// Config 应用配置
type Config struct {
	// Server 服务器配置
	ServerPort string

	// Baidu 百度语音识别配置
	BaiduAPIKey    string
	BaiduSecretKey string

	// GLM 智谱 AI 配置
	GLMAPIKey string

	// Service Providers
	STTProvider string // baidu, sherpa
	TTSProvider string // baidu, sherpa

	// Sherpa Onnx 配置
	SherpaSTTAddr string // e.g. localhost:6006
	SherpaTTSAddr string // e.g. http://localhost:19000
}

// Load 从环境变量加载配置
func Load() *Config {
	return &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		BaiduAPIKey:    getEnv("BAIDU_API_KEY", ""),
		BaiduSecretKey: getEnv("BAIDU_SECRET_KEY", ""),
		GLMAPIKey:      getEnv("GLM_API_KEY", ""),

		STTProvider:   getEnv("STT_PROVIDER", "baidu"),
		TTSProvider:   getEnv("TTS_PROVIDER", "baidu"),
		SherpaSTTAddr: getEnv("SHERPA_STT_ADDR", "localhost:6006"),
		SherpaTTSAddr: getEnv("SHERPA_TTS_ADDR", "http://localhost:19000"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
