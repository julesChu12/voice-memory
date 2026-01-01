package service

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SherpaTTS Sherpa Onnx 语音合成服务
type SherpaTTS struct {
	addr     string
	audioDir string
}

// NewSherpaTTS 创建 Sherpa TTS 实例
func NewSherpaTTS(addr string) *SherpaTTS {
	homeDir, _ := os.UserHomeDir()
	audioDir := filepath.Join(homeDir, ".voice-memory", "audio")
	return NewSherpaTTSWithDir(addr, audioDir)
}

// NewSherpaTTSWithDir 创建 Sherpa TTS 实例（指定音频目录）
func NewSherpaTTSWithDir(addr, audioDir string) *SherpaTTS {
	// 确保地址以 http:// 开头
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}
	// 去掉结尾的斜杠
	addr = strings.TrimSuffix(addr, "/")

	os.MkdirAll(audioDir, 0755)

	return &SherpaTTS{
		addr:     addr,
		audioDir: audioDir,
	}
}

// Synthesize 合成语音
func (s *SherpaTTS) Synthesize(options TTSOptions) ([]byte, error) {
	// Construct URL: /generate
	reqURL := s.addr + "/generate"
	
	params := url.Values{}
	params.Set("text", options.Text)
	params.Set("sid", fmt.Sprintf("%d", options.Per)) // Speaker ID
	params.Set("speed", fmt.Sprintf("%f", float32(options.Spd)/5.0)) // Normalize speed? 
	// Baidu Spd is 0-15 (default 5). Sherpa is usually 1.0 default.
	// Let's assume options.Spd=5 maps to 1.0.
	
	// Send request
	resp, err := http.PostForm(reqURL, params)
	if err != nil {
		return nil, fmt.Errorf("sherpa tts request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sherpa tts error %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read audio failed: %w", err)
	}

	return data, nil
}

// SynthesizeToFile 合成语音并保存到文件
func (s *SherpaTTS) SynthesizeToFile(options TTSOptions) (string, error) {
	audioData, err := s.Synthesize(options)
	if err != nil {
		return "", err
	}

	// Generate filename
	filename := fmt.Sprintf("sherpa_tts_%d_%s.wav", time.Now().Unix(), simpleHash(options.Text))
	filepath := filepath.Join(s.audioDir, filename)

	if err := os.WriteFile(filepath, audioData, 0644); err != nil {
		return "", fmt.Errorf("save audio file failed: %w", err)
	}

	return filename, nil
}

// SynthesizeBase64 合成语音并返回 Base64
func (s *SherpaTTS) SynthesizeBase64(options TTSOptions) (string, error) {
	audioData, err := s.Synthesize(options)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(audioData), nil
}

// GetAudioDir 获取音频目录
func (s *SherpaTTS) GetAudioDir() string {
	return s.audioDir
}

// ServeAudio 提供音频文件
func (s *SherpaTTS) ServeAudio(filename string) ([]byte, string, error) {
	filepath := filepath.Join(s.audioDir, filename)
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, "", err
	}
	return data, "audio/wav", nil
}
