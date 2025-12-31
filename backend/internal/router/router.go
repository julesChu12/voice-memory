package router

import (
	"voice-memory/internal/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RouterConfig 路由配置
type RouterConfig struct {
	STTHandler       *handler.STTHandler
	KnowledgeHandler *handler.KnowledgeHandler
	SessionHandler   *handler.SessionHandler
	TTSHandler       *handler.TTSHandler
	WSHandler        *handler.WSHandler
}

// Setup 配置路由
func Setup(cfg RouterConfig) *gin.Engine {
	router := gin.Default()

	// 配置 CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// WebSocket 路由 (Phase 2 核心)
	router.GET("/ws", cfg.WSHandler.HandleWS)

	// 以下 API 暂时保留，用于调试或特定功能
	
	// STT 路由
	router.POST("/api/stt", cfg.STTHandler.Recognize)

	// TTS 路由
	router.GET("/api/tts", cfg.TTSHandler.HandleTTS)
	router.GET("/api/audio/:filename", cfg.TTSHandler.ServeAudio)

	// 知识库路由 (未来 Phase 3 将整合进 Pipeline)
	knowledge := router.Group("/api/knowledge")
	{
		knowledge.POST("/record", cfg.KnowledgeHandler.HandleRecord)
		knowledge.GET("/list", cfg.KnowledgeHandler.HandleList)
		knowledge.POST("/search", cfg.KnowledgeHandler.HandleSearch)
	}

	// 会话历史路由 (用于前端展示归档)
	sessions := router.Group("/api/sessions")
	{
		sessions.GET("", cfg.SessionHandler.HandleListSessions)
		sessions.GET("/get", cfg.SessionHandler.HandleGetSession)
		sessions.DELETE("", cfg.SessionHandler.HandleDeleteSession)
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "voice-memory-backend",
		})
	})

	// 静态文件服务
	router.Static("/assets", "./static/assets")
	router.StaticFile("/", "./static/index.html")
	router.StaticFile("/index.html", "./static/index.html")
	router.StaticFile("/ws-client.js", "./static/ws-client.js") // 显式暴露 ws-client.js

	// SPA 路由回退
	router.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}
		c.File("./static/index.html")
	})

	return router
}
