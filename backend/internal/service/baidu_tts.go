package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// BaiduTTS 百度语音合成服务
type BaiduTTS struct {
	apiKey     string
	secretKey  string
	token      string
	tokenExp   time.Time
	tokenFile  string
	client     *http.Client
	audioDir   string // 音频文件缓存目录
	cuid       string // 设备唯一标识
}

// NewBaiduTTS 创建百度 TTS 实例（使用用户主目录存储）
func NewBaiduTTS(apiKey, secretKey string) *BaiduTTS {
	homeDir, _ := os.UserHomeDir()
	tokenDir := filepath.Join(homeDir, ".voice-memory")
	os.MkdirAll(tokenDir, 0755)

	// 创建音频缓存目录
	audioDir := filepath.Join(homeDir, ".voice-memory", "audio")
	os.MkdirAll(audioDir, 0755)

	return newBaiduTTSService(apiKey, secretKey, tokenDir, audioDir)
}

// NewBaiduTTSWithDir 创建百度 TTS 实例（指定音频目录）
func NewBaiduTTSWithDir(apiKey, secretKey, audioDir string) *BaiduTTS {
	homeDir, _ := os.UserHomeDir()
	tokenDir := filepath.Join(homeDir, ".voice-memory")
	os.MkdirAll(tokenDir, 0755)

	// 创建指定的音频目录
	os.MkdirAll(audioDir, 0755)

	return newBaiduTTSService(apiKey, secretKey, tokenDir, audioDir)
}

// newBaiduTTSService 内部构造函数
func newBaiduTTSService(apiKey, secretKey, tokenDir, audioDir string) *BaiduTTS {
	// 生成设备唯一标识 (cuid)
	b := make([]byte, 16)
	rand.Read(b)
	cuid := hex.EncodeToString(b)

	return &BaiduTTS{
		apiKey:    apiKey,
		secretKey: secretKey,
		tokenFile: filepath.Join(tokenDir, "baidu_token.json"),
		audioDir:  audioDir,
		cuid:      cuid,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TTSOptions TTS 选项
type TTSOptions struct {
	Text  string // 要合成的文本
	Per   int    // 发音人 0:女声 1:男声 3:度逍遥 4:度丫丫
	Spd   int    // 语速 0-15, 默认5
	Pit   int    // 音调 0-15, 默认5
	Vol   int    // 音量 0-15, 默认5
}

// DefaultTTSOptions 默认 TTS 选项
func DefaultTTSOptions(text string) TTSOptions {
	return TTSOptions{
		Text: text,
		Per:  4195,  // 精品发音人 - 情感女声
		Spd:  6,     // 语速 (0-15)
		Pit:  6,     // 音调 (0-15)
		Vol:  8,     // 音量 (0-15)
	}
}

// loadTokenFromFile 从文件加载 token
func (b *BaiduTTS) loadTokenFromFile() bool {
	data, err := os.ReadFile(b.tokenFile)
	if err != nil {
		return false
	}

	var tokenData TokenFileData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return false
	}

	// 检查是否过期
	if time.Now().Unix() > tokenData.ExpiresAt {
		return false
	}

	b.token = tokenData.AccessToken
	b.tokenExp = time.Unix(tokenData.ExpiresAt, 0)
	return true
}

// saveTokenToFile 保存 token 到文件
func (b *BaiduTTS) saveTokenToFile(expiresAt time.Time) error {
	tokenData := TokenFileData{
		AccessToken: b.token,
		ExpiresAt:   expiresAt.Unix(),
	}

	data, err := json.MarshalIndent(tokenData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(b.tokenFile, data, 0644)
}

// getAccessToken 获取访问令牌（复用 STT 的 token）
func (b *BaiduTTS) getAccessToken() (string, error) {
	// 如果内存中 token 为空，尝试从文件加载
	if b.token == "" {
		if b.loadTokenFromFile() {
			return b.token, nil
		}
	}

	// 如果 token 未过期，直接返回
	if b.token != "" && time.Now().Before(b.tokenExp) {
		return b.token, nil
	}

	// 请求新 token
	reqURL := fmt.Sprintf("https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=%s&client_secret=%s",
		b.apiKey, b.secretKey)

	resp, err := b.client.Post(reqURL, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("请求 token 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析 token 失败: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("token 为空")
	}

	// 缓存 token (提前 5 分钟过期)
	b.token = tokenResp.AccessToken
	b.tokenExp = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	// 保存到文件
	if err := b.saveTokenToFile(b.tokenExp); err != nil {
		fmt.Printf("保存 token 到文件失败: %v\n", err)
	}

	return b.token, nil
}

// Synthesize 合成语音，返回音频数据
func (b *BaiduTTS) Synthesize(options TTSOptions) ([]byte, error) {
	// 获取 access_token
	token, err := b.getAccessToken()
	if err != nil {
		return nil, err
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("tex", options.Text)
	params.Set("tok", token)
	params.Set("cuid", b.cuid)
	params.Set("ctp", "1")
	params.Set("lan", "zh")
	params.Set("aue", "3") // 3=mp3格式

	if options.Per > 0 {
		params.Set("per", fmt.Sprintf("%d", options.Per))
	}
	if options.Spd > 0 {
		params.Set("spd", fmt.Sprintf("%d", options.Spd))
	}
	if options.Pit > 0 {
		params.Set("pit", fmt.Sprintf("%d", options.Pit))
	}
	if options.Vol > 0 {
		params.Set("vol", fmt.Sprintf("%d", options.Vol))
	}

	// 日志输出 TTS 参数
	fmt.Printf("TTS 请求参数: per=%d, spd=%d, pit=%d, vol=%d, text_len=%d\n",
		options.Per, options.Spd, options.Pit, options.Vol, len(options.Text))

	// 发送 TTS 请求
	ttsURL := "https://tsn.baidu.com/text2audio"
	resp, err := http.PostForm(ttsURL, params)
	if err != nil {
		return nil, fmt.Errorf("TTS 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取音频失败: %w", err)
	}

	// 检查是否是错误响应（返回 JSON 而非音频）
	if len(audioData) > 0 && audioData[0] == '{' {
		var errResp struct {
			ErrNo   int    `json:"err_no"`
			ErrMsg  string `json:"err_msg"`
		}
		if err := json.Unmarshal(audioData, &errResp); err == nil && errResp.ErrNo != 0 {
			return nil, fmt.Errorf("TTS 错误 [%d]: %s", errResp.ErrNo, errResp.ErrMsg)
		}
	}

	return audioData, nil
}

// SynthesizeToFile 合成语音并保存到文件
func (b *BaiduTTS) SynthesizeToFile(options TTSOptions) (string, error) {
	audioData, err := b.Synthesize(options)
	if err != nil {
		return "", err
	}

	// 生成文件名（使用时间戳和文本哈希）
	filename := fmt.Sprintf("tts_%d_%s.mp3", time.Now().Unix(), simpleHash(options.Text))
	filepath := filepath.Join(b.audioDir, filename)

	fmt.Printf("SynthesizeToFile: audioDir=%s, filename=%s, fullpath=%s\n", b.audioDir, filename, filepath)

	if err := os.WriteFile(filepath, audioData, 0644); err != nil {
		return "", fmt.Errorf("保存音频文件失败: %w", err)
	}

	return filename, nil
}

// simpleHash 简单哈希函数
func simpleHash(s string) string {
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	return fmt.Sprintf("%x", hash&0xFFFFFFFF)
}

// GetAudioDir 获取音频文件目录
func (b *BaiduTTS) GetAudioDir() string {
	return b.audioDir
}

// ServeAudio 通过 HTTP 提供音频文件
func (b *BaiduTTS) ServeAudio(filename string) ([]byte, string, error) {
	filepath := filepath.Join(b.audioDir, filename)
	fmt.Printf("ServeAudio: audioDir=%s, filename=%s, fullpath=%s\n", b.audioDir, filename, filepath)
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, "", err
	}
	return data, "audio/mpeg", nil
}

// SynthesizeBase64 合成语音并返回 Base64 编码
func (b *BaiduTTS) SynthesizeBase64(options TTSOptions) (string, error) {
	audioData, err := b.Synthesize(options)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(audioData), nil
}

// CleanupOldFiles 清理超过指定时间的音频文件
func (b *BaiduTTS) CleanupOldFiles(maxAge time.Duration) error {
	entries, err := os.ReadDir(b.audioDir)
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-maxAge)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filepath := filepath.Join(b.audioDir, entry.Name())
			os.Remove(filepath)
			fmt.Printf("清理旧音频文件: %s\n", entry.Name())
		}
	}

	return nil
}

// GetAccessTokenFromSTT 从 STT 服务复制 token（避免重复获取）
func (b *BaiduTTS) GetAccessTokenFromSTT(stt *BaiduSTT) error {
	token, exp, err := stt.GetTokenInfo()
	if err != nil {
		return err
	}
	b.token = token
	b.tokenExp = exp
	return nil
}

// GetTokenInfo 获取当前 token 信息
func (b *BaiduTTS) GetTokenInfo() (string, time.Time, error) {
	if b.token == "" || time.Now().After(b.tokenExp) {
		return "", time.Time{}, fmt.Errorf("token 无效或已过期")
	}
	return b.token, b.tokenExp, nil
}
