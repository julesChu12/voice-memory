package server

import (
	"fmt"
	"voice-memory/internal/config"
	"voice-memory/internal/handler"
	"voice-memory/internal/router"
	"voice-memory/internal/service"

	"github.com/gin-gonic/gin"
)

// Server 服务器
type Server struct {
	config     *config.Config
	database   *service.Database
	httpServer *gin.Engine
}

// New 创建服务器
func New(cfg *config.Config) (*Server, error) {
	// 数据目录
	dataDir := "./data"

	// 创建数据库
	database, err := service.NewDatabase(dataDir)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 会话和音频目录
	sessionDir := fmt.Sprintf("%s/sessions", dataDir)
	audioDir := fmt.Sprintf("%s/audio", dataDir)

	// 创建服务
	sttService := service.NewBaiduSTT(cfg.BaiduAPIKey, cfg.BaiduSecretKey)
	ttsService := service.NewBaiduTTSWithDir(cfg.BaiduAPIKey, cfg.BaiduSecretKey, audioDir)
	glmClient := service.NewGLMClient(cfg.GLMAPIKey)

	// 创建 RAG 服务
	ragService := service.NewRAGService(cfg.GLMAPIKey)

	// 加载现有知识到向量库
	knowledges, err := database.GetAllKnowledge()
	if err == nil && len(knowledges) > 0 {
		// 限制最多加载 100 条
		maxLoad := 100
		if len(knowledges) > maxLoad {
			knowledges = knowledges[:maxLoad]
		}
		fmt.Printf("📚 正在加载 %d 条知识到向量库...\n", len(knowledges))
		for i, k := range knowledges {
			metadata := map[string]interface{}{
				"category":   k.Category,
				"tags":       k.Tags,
				"summary":    k.Summary,
				"key_points": k.KeyPoints,
				"source":     k.Source,
				"created_at": k.CreatedAt,
			}
			if err := ragService.AddKnowledge(k.ID, k.Content, metadata); err != nil {
				fmt.Printf("⚠️  知识 %s 向量化失败: %v\n", k.ID, err)
			}
			if (i+1)%20 == 0 {
				fmt.Printf("   已加载 %d/%d\n", i+1, len(knowledges))
			}
		}
		fmt.Printf("✅ 向量库初始化完成，共 %d 条知识\n", ragService.GetKnowledgeCount())
	}

	// 创建会话管理器
	sessionManager, err := service.NewSessionManagerWithStorage(sessionDir)
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("初始化会话管理器失败: %w", err)
	}

	// 创建知识组织器
	knowledgeOrganizer := service.NewKnowledgeOrganizer(glmClient)

	// 创建处理器
	sttHandler := handler.NewSTTHandler(sttService)
	chatHandler := handler.NewChatHandlerWithRAG(glmClient, sessionManager, ttsService, ragService)
	audioChatHandler := handler.NewAudioChatHandler(glmClient)
	knowledgeHandler := handler.NewKnowledgeHandler(sttService, knowledgeOrganizer, database, audioDir)
	knowledgeHandler.SetRAGService(ragService) // 设置 RAG 服务
	sessionHandler := handler.NewSessionHandler(sessionManager)
	ttsHandler := handler.NewTTSHandler(ttsService)

	// 配置路由
	httpServer := router.Setup(router.RouterConfig{
		STTHandler:       sttHandler,
		ChatHandler:      chatHandler,
		AudioChatHandler: audioChatHandler,
		KnowledgeHandler: knowledgeHandler,
		SessionHandler:   sessionHandler,
		TTSHandler:       ttsHandler,
	})

	return &Server{
		config:     cfg,
		database:   database,
		httpServer: httpServer,
	}, nil
}

// Run 启动服务器
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.config.ServerPort)

	s.printRoutes(addr)

	if err := s.httpServer.Run(addr); err != nil {
		return fmt.Errorf("启动服务器失败: %w", err)
	}

	return nil
}

// Close 关闭服务器
func (s *Server) Close() error {
	return s.database.Close()
}

// printRoutes 打印路由信息
func (s *Server) printRoutes(addr string) {
	fmt.Printf("🚀 Voice Memory Backend 启动成功\n")
	fmt.Printf("📍 服务地址: http://localhost%s\n", addr)
	fmt.Printf("🎤 STT 接口: http://localhost%s/api/stt\n", addr)
	fmt.Printf("🤖 Chat 接口: http://localhost%s/api/chat\n", addr)
	fmt.Printf("🎙️  Audio Chat: http://localhost%s/api/audio-chat\n", addr)
	fmt.Printf("🔊 TTS 接口: http://localhost%s/api/tts?text=xxx\n", addr)
	fmt.Printf("🎵 音频文件: http://localhost%s/api/audio/:filename\n", addr)
	fmt.Printf("📚 知识记录: http://localhost%s/api/knowledge/record\n", addr)
	fmt.Printf("📋 知识列表: http://localhost%s/api/knowledge/list\n", addr)
	fmt.Printf("🔍 知识搜索: http://localhost%s/api/knowledge/search\n", addr)
	fmt.Printf("💬 会话列表: http://localhost%s/api/sessions\n", addr)
	fmt.Printf("💚 健康检查: http://localhost%s/health\n", addr)
	fmt.Printf("🧠 RAG 检索: 已启用 (向量搜索)\n\n", addr)
}
