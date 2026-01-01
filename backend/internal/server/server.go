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

	// 音频目录
	audioDir := fmt.Sprintf("%s/audio", dataDir)

	// 创建基础服务
	sttService := service.NewBaiduSTT(cfg.BaiduAPIKey, cfg.BaiduSecretKey)
	ttsService := service.NewBaiduTTSWithDir(cfg.BaiduAPIKey, cfg.BaiduSecretKey, audioDir)
	glmClient := service.NewGLMClient(cfg.GLMAPIKey)
	intentRecognizer := service.NewIntentRecognizer()
	ragService := service.NewRAGService(cfg.GLMAPIKey)

	// 加载现有知识到向量库
	knowledges, err := database.GetAllKnowledge()
	if err == nil && len(knowledges) > 0 {
		fmt.Printf("📚 正在加载知识到向量库...\n")
		// ... 保持原有加载逻辑
		for _, k := range knowledges {
			metadata := map[string]interface{}{"category": k.Category, "tags": k.Tags}
			_ = ragService.AddKnowledge(k.ID, k.Content, metadata)
		}
	}

	// 创建会话管理器（带数据库）
	sessionManager := service.NewSessionManagerWithDB(database)

	// 创建知识组织器
	knowledgeOrganizer := service.NewKnowledgeOrganizer(glmClient)

	// 创建处理器 (仅保留必要的)
	sttHandler := handler.NewSTTHandler(sttService)
	knowledgeHandler := handler.NewKnowledgeHandler(sttService, knowledgeOrganizer, database, audioDir)
	knowledgeHandler.SetRAGService(ragService)
	sessionHandler := handler.NewSessionHandler(sessionManager)
	ttsHandler := handler.NewTTSHandler(ttsService)
	
	// WebSocket 处理器 (核心)
	wsHandler := handler.NewWSHandler(
		sessionManager,
		sttService,
		glmClient,
		ttsService,
		intentRecognizer,
	)

	// 配置路由
	httpServer := router.Setup(router.RouterConfig{
		STTHandler:       sttHandler,
		KnowledgeHandler: knowledgeHandler,
		SessionHandler:   sessionHandler,
		TTSHandler:       ttsHandler,
		WSHandler:        wsHandler,
	})

	return &Server{
		config:     cfg,
		database:   database,
		httpServer: httpServer,
	},
	nil
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
	fmt.Printf("🚀 Voice Memory Backend 启动成功 (Phase 2 Architecture)\n")
	fmt.Printf("📍 服务地址: http://localhost%s\n", addr)
	fmt.Printf("🔌 WebSocket: ws://localhost%s/ws\n", addr)
	fmt.Printf("📋 其他接口已清理，请优先使用 WebSocket 进行交互\n\n")
}