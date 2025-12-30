package router

import (
	"voice-memory/internal/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RouterConfig 路由配置
type RouterConfig struct {
	STTHandler       *handler.STTHandler
	ChatHandler      *handler.ChatHandler
	AudioChatHandler *handler.AudioChatHandler
	KnowledgeHandler *handler.KnowledgeHandler
	SessionHandler   *handler.SessionHandler
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

	// STT 路由
	router.POST("/api/stt", cfg.STTHandler.Recognize)

	// Chat 路由
	router.POST("/api/chat", cfg.ChatHandler.HandleChat)
	router.POST("/api/audio-chat", cfg.AudioChatHandler.HandleAudioChat)

	// 知识库路由
	knowledge := router.Group("/api/knowledge")
	{
		knowledge.POST("/record", cfg.KnowledgeHandler.HandleRecord)
		knowledge.GET("/list", cfg.KnowledgeHandler.HandleList)
		knowledge.POST("/search", cfg.KnowledgeHandler.HandleSearch)
	}

	// 会话路由
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

	return router
}
