package handler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"voice-memory/internal/service"

	"github.com/gin-gonic/gin"
)

// KnowledgeHandler 知识库处理器
type KnowledgeHandler struct {
	sttService   service.STTService
	organizer    *service.KnowledgeOrganizer
	database     *service.Database
	audioDir     string
	ragService   *service.RAGService
}

// NewKnowledgeHandler 创建知识库处理器
func NewKnowledgeHandler(
	sttService service.STTService,
	organizer *service.KnowledgeOrganizer,
	database *service.Database,
	audioDir string,
) *KnowledgeHandler {
	return &KnowledgeHandler{
		sttService: sttService,
		organizer:  organizer,
		database:   database,
		audioDir:   audioDir,
		ragService: nil,
	}
}

// SetRAGService 设置 RAG 服务
func (h *KnowledgeHandler) SetRAGService(ragService *service.RAGService) {
	h.ragService = ragService
}

// RecordRequest 语音记录请求
type RecordRequest struct {
	AutoOrganize bool   `form:"auto_organize"` // 是否自动整理（默认 true）
	Text         string `form:"text"`          // 直接传入的文本（可选，如果提供则跳过 STT）
	SessionID    string `form:"session_id"`    // 关联的会话ID（可选）
}

// RecordResponse 语音记录响应
type RecordResponse struct {
	Success   bool                   `json:"success"`
	Knowledge *service.Knowledge     `json:"knowledge,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// HandleRecord 处理语音记录
func (h *KnowledgeHandler) HandleRecord(c *gin.Context) {
	// 解析参数
	var req RecordRequest
	req.AutoOrganize = c.DefaultPostForm("auto_organize", "true") == "true"
	req.Text = c.PostForm("text")
	req.SessionID = c.PostForm("session_id") // 解析会话ID

	// 生成知识 ID
	knowledgeID := fmt.Sprintf("kb_%d", generateIntID())

	var text string
	var audioPath string
	var source string

	// 优先使用直接传入的文本
	if req.Text != "" {
		text = req.Text
		source = "manual"
		// 如果同时有音频文件，也保存
		fileHeader, err := c.FormFile("audio")
		if err == nil {
			audioPath = filepath.Join(h.audioDir, knowledgeID+".wav")
			if err := c.SaveUploadedFile(fileHeader, audioPath); err != nil {
				c.JSON(500, RecordResponse{
					Success: false,
					Error:   "保存音频失败: " + err.Error(),
				})
				return
			}
		}
	} else {
		// 获取上传的音频文件
		fileHeader, err := c.FormFile("audio")
		if err != nil {
			c.JSON(400, RecordResponse{
				Success: false,
				Error:   "音频文件上传失败: " + err.Error(),
			})
			return
		}

		// 保存音频文件
		audioPath = filepath.Join(h.audioDir, knowledgeID+".wav")
		if err := c.SaveUploadedFile(fileHeader, audioPath); err != nil {
			c.JSON(500, RecordResponse{
				Success: false,
				Error:   "保存音频失败: " + err.Error(),
			})
			return
		}

		// 打开音频文件进行 STT
		file, err := os.Open(audioPath)
		if err != nil {
			c.JSON(500, RecordResponse{
				Success: false,
				Error:   "打开音频文件失败: " + err.Error(),
			})
			return
		}
		defer file.Close()

		// 读取音频数据
		audioData, err := io.ReadAll(file)
		if err != nil {
			c.JSON(500, RecordResponse{
				Success: false,
				Error:   "读取音频数据失败: " + err.Error(),
			})
			return
		}

		// 调用 STT
		results, err := h.sttService.Recognize(&service.RecognizeRequest{
			AudioData: audioData,
			Format:    "wav",
			Rate:      16000,
		})
		if err != nil {
			c.JSON(500, RecordResponse{
				Success: false,
				Error:   "语音识别失败: " + err.Error(),
			})
			return
		}

		// 取第一个识别结果
		if len(results) > 0 {
			text = results[0]
		}
		source = "voice"
	}

	// 构建知识条目
	knowledge := &service.Knowledge{
		ID:        knowledgeID,
		Content:   text,
		Source:    source,
		AudioURL:  audioPath,
		SessionID: req.SessionID, // 关联会话ID，保存完整对话历史
		Metadata:  make(map[string]string),
	}

	// 如果提供了 session_id，生成会话标题和摘要
	var sessionForSummary *service.Session = nil
	if req.SessionID != "" {
		session, err := h.database.GetSession(req.SessionID)
		if err != nil {
			fmt.Printf("获取会话失败: %v\n", err)
		} else if session != nil && len(session.Messages) > 0 {
			sessionForSummary = session
			// 使用 AI 生成标题
			title, err := h.organizer.GenerateTitleFromSession(session)
			if err != nil {
				fmt.Printf("AI 生成标题失败: %v\n", err)
				// 失败时使用默认标题
				knowledge.Title = "会话记录 - " + time.Now().Format("2006-01-02 15:04")
			} else {
				knowledge.Title = title
			}
		}
	}

	// AI 自动整理（如果有 session 则使用完整对话，否则只用输入文本）
	if req.AutoOrganize {
		var organizeText string
		if sessionForSummary != nil {
			// 构建完整对话内容用于摘要生成
			var conversationText string
			for i, msg := range sessionForSummary.Messages {
				role := "用户"
				if msg.Role == "assistant" {
					role = "AI助手"
				}
				conversationText += fmt.Sprintf("%d. %s: %s\n", i+1, role, msg.Content)
			}
			organizeText = conversationText
		} else {
			organizeText = text
		}

		organizeResult, err := h.organizer.Organize(organizeText)
		if err != nil {
			// 整理失败不影响存储，只记录日志
			fmt.Printf("AI 整理警告: %v\n", err)
		} else {
			knowledge.Summary = organizeResult.Summary
			knowledge.KeyPoints = organizeResult.KeyPoints
			knowledge.Entities = organizeResult.Entities
			knowledge.Category = organizeResult.Category
			knowledge.Tags = organizeResult.Tags
			knowledge.Importance = organizeResult.Importance
			knowledge.Sentiment = organizeResult.Sentiment
		}
	}

	// 保存到数据库
	if err := h.database.SaveKnowledge(knowledge); err != nil {
		c.JSON(500, RecordResponse{
			Success: false,
			Error:   "保存知识库失败: " + err.Error(),
		})
		return
	}

	// 同步到 RAG 向量库
	if h.ragService != nil && knowledge.Content != "" {
		metadata := map[string]interface{}{
			"category":   knowledge.Category,
			"tags":       knowledge.Tags,
			"summary":    knowledge.Summary,
			"key_points": knowledge.KeyPoints,
			"source":     knowledge.Source,
			"created_at": knowledge.CreatedAt,
		}
		if err := h.ragService.AddKnowledge(knowledge.ID, knowledge.Content, metadata); err != nil {
			// 向量化失败不影响主流程，只记录日志
			fmt.Printf("⚠️  RAG 向量化失败: %v\n", err)
		} else {
			fmt.Printf("✅ 知识 %s 已添加到向量库\n", knowledge.ID)
		}
	}

	c.JSON(200, RecordResponse{
		Success:   true,
		Knowledge: knowledge,
	})
}

// ListKnowledgeRequest 列表请求
type ListKnowledgeRequest struct {
	Category string `form:"category"` // 按分类筛选
}

// ListKnowledgeResponse 列表响应
type ListKnowledgeResponse struct {
	Success    bool                   `json:"success"`
	Knowledges []service.Knowledge    `json:"knowledges,omitempty"`
	Stats      map[string]int         `json:"stats,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// HandleList 处理知识列表
func (h *KnowledgeHandler) HandleList(c *gin.Context) {
	var req ListKnowledgeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Category = ""
	}

	var knowledges []service.Knowledge
	var err error

	if req.Category != "" {
		knowledges, err = h.database.GetKnowledgeByCategory(req.Category)
	} else {
		knowledges, err = h.database.GetAllKnowledge()
	}

	if err != nil {
		c.JSON(500, ListKnowledgeResponse{
			Success: false,
			Error:   "获取知识列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, ListKnowledgeResponse{
		Success:    true,
		Knowledges: knowledges,
	})
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query string `json:"query" binding:"required"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Success    bool                 `json:"success"`
	Knowledges []service.Knowledge  `json:"knowledges,omitempty"`
	Error      string               `json:"error,omitempty"`
}

// HandleSearch 处理搜索
func (h *KnowledgeHandler) HandleSearch(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, SearchResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	results, err := h.database.SearchKnowledge(req.Query)
	if err != nil {
		c.JSON(500, SearchResponse{
			Success: false,
			Error:   "搜索失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, SearchResponse{
		Success:    true,
		Knowledges: results,
	})
}

func generateIntID() int64 {
	return time.Now().UnixNano()
}
