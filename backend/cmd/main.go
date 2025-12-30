package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"voice-memory/internal/config"
	"voice-memory/internal/server"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  未找到 .env 文件，使用环境变量")
	}

	// 加载配置
	cfg := config.Load()

	// 验证配置
	if err := validateConfig(cfg); err != nil {
		log.Fatal("❌", err)
	}

	// 创建并启动服务器
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatal("❌ 创建服务器失败:", err)
	}
	defer srv.Close()

	if err := srv.Run(); err != nil {
		log.Fatal("❌ 启动服务器失败:", err)
	}
}

// validateConfig 验证配置
func validateConfig(cfg *config.Config) error {
	if cfg.BaiduAPIKey == "" || cfg.BaiduSecretKey == "" {
		return fmt.Errorf("百度 API Key 或 Secret Key 未配置\n" +
			"请设置环境变量:\n" +
			"  export BAIDU_API_KEY=your_api_key\n" +
			"  export BAIDU_SECRET_KEY=your_secret_key")
	}

	if cfg.GLMAPIKey == "" {
		return fmt.Errorf("GLM API Key 未配置\n" +
			"请设置环境变量:\n" +
			"  export GLM_API_KEY=your_glm_api_key")
	}

	return nil
}
