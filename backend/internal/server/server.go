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

	// 创建服务
	sttService := service.NewBaiduSTT(cfg.BaiduAPIKey, cfg.BaiduSecretKey)
	glmClient := service.NewGLMClient(cfg.GLMAPIKey)

	// 会话和音频目录
	sessionDir := fmt.Sprintf("%s/sessions", dataDir)
	audioDir := fmt.Sprintf("%s/audio", dataDir)

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
	chatHandler := handler.NewChatHandlerWithSession(glmClient, sessionManager)
	audioChatHandler := handler.NewAudioChatHandler(glmClient)
	knowledgeHandler := handler.NewKnowledgeHandler(sttService, knowledgeOrganizer, database, audioDir)
	sessionHandler := handler.NewSessionHandler(sessionManager)

	// 配置路由
	httpServer := router.Setup(router.RouterConfig{
		STTHandler:       sttHandler,
		ChatHandler:      chatHandler,
		AudioChatHandler: audioChatHandler,
		KnowledgeHandler: knowledgeHandler,
		SessionHandler:   sessionHandler,
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
	fmt.Printf("📚 知识记录: http://localhost%s/api/knowledge/record\n", addr)
	fmt.Printf("📋 知识列表: http://localhost%s/api/knowledge/list\n", addr)
	fmt.Printf("🔍 知识搜索: http://localhost%s/api/knowledge/search\n", addr)
	fmt.Printf("💬 会话列表: http://localhost%s/api/sessions\n", addr)
	fmt.Printf("💚 健康检查: http://localhost%s/health\n\n", addr)
}
