package handler

import (
	"fmt"
	"voice-memory/internal/service"

	"github.com/gin-gonic/gin"
)

// TTSHandler TTS 处理器
type TTSHandler struct {
	ttsService *service.BaiduTTS
}

// NewTTSHandler 创建 TTS 处理器
func NewTTSHandler(ttsService *service.BaiduTTS) *TTSHandler {
	return &TTSHandler{
		ttsService: ttsService,
	}
}

// HandleTTS 处理 TTS 请求
// GET /api/tts?text=xxx&session_id=xxx&per=0&spd=5
// 返回音频文件 URL，支持缓存
func (h *TTSHandler) HandleTTS(c *gin.Context) {
	text := c.Query("text")
	if text == "" {
		c.JSON(400, gin.H{"error": "缺少 text 参数"})
		return
	}

	sessionID := c.Query("session_id")

	// 构建 TTS 选项
	options := service.DefaultTTSOptions(text)

	// 可选参数
	if per := c.Query("per"); per != "" {
		var perInt int
		if _, err := fmt.Sscanf(per, "%d", &perInt); err == nil {
			options.Per = perInt
		}
	}
	if spd := c.Query("spd"); spd != "" {
		var spdInt int
		if _, err := fmt.Sscanf(spd, "%d", &spdInt); err == nil {
			options.Spd = spdInt
		}
	}
	if pit := c.Query("pit"); pit != "" {
		var pitInt int
		if _, err := fmt.Sscanf(pit, "%d", &pitInt); err == nil {
			options.Pit = pitInt
		}
	}
	if vol := c.Query("vol"); vol != "" {
		var volInt int
		if _, err := fmt.Sscanf(vol, "%d", &volInt); err == nil {
			options.Vol = volInt
		}
	}

	// 合成语音并保存到文件
	filename, err := h.ttsService.SynthesizeToFile(options)
	if err != nil {
		c.JSON(500, gin.H{"error": "TTS 合成失败: " + err.Error()})
		return
	}

	// 返回音频文件 URL
	c.JSON(200, gin.H{
		"success":    true,
		"audio_url":  fmt.Sprintf("/api/audio/%s", filename),
		"session_id": sessionID,
	})
}

// ServeAudio 提供缓存的音频文件
// GET /api/audio/:filename
func (h *TTSHandler) ServeAudio(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(400, gin.H{"error": "缺少文件名"})
		return
	}

	audioData, contentType, err := h.ttsService.ServeAudio(filename)
	if err != nil {
		c.JSON(404, gin.H{"error": "音频文件不存在"})
		return
	}

	c.Data(200, contentType, audioData)
}
