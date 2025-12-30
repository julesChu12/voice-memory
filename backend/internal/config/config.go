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
}

// Load 从环境变量加载配置
func Load() *Config {
	return &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		BaiduAPIKey:    getEnv("BAIDU_API_KEY", ""),
		BaiduSecretKey: getEnv("BAIDU_SECRET_KEY", ""),
		GLMAPIKey:      getEnv("GLM_API_KEY", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
