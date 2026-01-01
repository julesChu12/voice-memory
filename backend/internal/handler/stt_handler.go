package handler

import (
	"io"
	"voice-memory/internal/service"

	"github.com/gin-gonic/gin"
)

// STTHandler STT 处理器
type STTHandler struct {
	sttService service.STTService
}

// NewSTTHandler 创建 STT 处理器
func NewSTTHandler(sttService service.STTService) *STTHandler {
	return &STTHandler{
		sttService: sttService,
	}
}

// RecognizeRequest 识别请求
type RecognizeRequest struct {
	Format string `json:"format" binding:"required"` // wav/pcm
	Rate   int    `json:"rate" binding:"required"`   // 采样率
}

// RecognizeResponse 识别响应
type RecognizeResponse struct {
	Success bool     `json:"success"`
	Result  []string `json:"result,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// Recognize 语音识别接口 (前端直接上传 WAV)
func (h *STTHandler) Recognize(c *gin.Context) {
	// 读取上传的音频文件
	fileHeader, err := c.FormFile("audio")
	if err != nil {
		c.JSON(400, RecognizeResponse{
			Success: false,
			Error:   "音频文件上传失败: " + err.Error(),
		})
		return
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, RecognizeResponse{
			Success: false,
			Error:   "打开音频文件失败: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// 读取音频数据 (前端已经是 WAV 格式)
	audioData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(500, RecognizeResponse{
			Success: false,
			Error:   "读取音频数据失败: " + err.Error(),
		})
		return
	}

	// 直接调用百度 STT (无需转换)
	results, err := h.sttService.Recognize(&service.RecognizeRequest{
		AudioData: audioData,
		Format:    "wav",
		Rate:      16000,
	})

	if err != nil {
		c.JSON(500, RecognizeResponse{
			Success: false,
			Error:   "语音识别失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, RecognizeResponse{
		Success: true,
		Result:  results,
	})
}
