package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// BaiduSTT 百度语音识别服务
type BaiduSTT struct {
	apiKey     string
	secretKey  string
	token      string
	tokenExp   time.Time
	tokenFile  string
	client     *http.Client
}

// NewBaiduSTT 创建百度 STT 实例
func NewBaiduSTT(apiKey, secretKey string) *BaiduSTT {
	homeDir, _ := os.UserHomeDir()
	tokenDir := filepath.Join(homeDir, ".voice-memory")
	os.MkdirAll(tokenDir, 0755)

	return &BaiduSTT{
		apiKey:    apiKey,
		secretKey: secretKey,
		tokenFile: filepath.Join(tokenDir, "baidu_token.json"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TokenResponse 获取 token 响应
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// TokenFileData token 文件数据
type TokenFileData struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

// loadTokenFromFile 从文件加载 token
func (b *BaiduSTT) loadTokenFromFile() bool {
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
	fmt.Printf("从文件加载 token (过期时间: %v)\n", b.tokenExp)
	return true
}

// saveTokenToFile 保存 token 到文件
func (b *BaiduSTT) saveTokenToFile(expiresAt time.Time) error {
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

// getAccessToken 获取访问令牌
func (b *BaiduSTT) getAccessToken() (string, error) {
	// 如果内存中 token 为空，尝试从文件加载
	if b.token == "" {
		if b.loadTokenFromFile() {
			// 文件中的 token 有效，直接返回
			return b.token, nil
		}
	}

	// 如果 token 未过期，直接返回
	if b.token != "" && time.Now().Before(b.tokenExp) {
		fmt.Printf("使用缓存的 token (过期时间: %v)\n", b.tokenExp)
		return b.token, nil
	}

	// 请求新 token
	fmt.Printf("请求新 token... API Key: %s\n", b.apiKey[:10]+"...")
	url := fmt.Sprintf("https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=%s&client_secret=%s",
		b.apiKey, b.secretKey)

	resp, err := b.client.Post(url, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("请求 token 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	fmt.Printf("Token 响应: %s\n", string(body))

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析 token 失败: %w\n响应: %s", err, string(body))
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("token 为空\n响应: %s", string(body))
	}

	// 缓存 token (提前 5 分钟过期)
	b.token = tokenResp.AccessToken
	b.tokenExp = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	// 保存到文件
	if err := b.saveTokenToFile(b.tokenExp); err != nil {
		fmt.Printf("保存 token 到文件失败: %v\n", err)
	}

	fmt.Printf("Token 获取成功! (过期时间: %v)\n", b.tokenExp)

	return b.token, nil
}

// RecognizeRequest 识别请求
type RecognizeRequest struct {
	AudioData []byte
	Format    string // pcm/wav/amr/m4a
	Rate      int    // 采样率 16000
}

// RecognizeResponse 识别响应
type RecognizeResponse struct {
	ErrNo     int        `json:"err_no"`
	ErrMsg    string     `json:"err_msg"`
	SN        string     `json:"sn"`           // 改为 string 类型
	CorpusNo  string     `json:"corpus_no"`
	Result    []string   `json:"result"`       // 识别结果数组
}

// Recognize 语音识别
func (b *BaiduSTT) Recognize(req *RecognizeRequest) ([]string, error) {
	// 获取 access_token
	token, err := b.getAccessToken()
	if err != nil {
		return nil, err
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"format": req.Format,
		"rate":   req.Rate,
		"channel": 1,
		"cuid":   "voice-memory-client",
		"token":  token,
		"speech": base64.StdEncoding.EncodeToString(req.AudioData),
		"len":    len(req.AudioData),
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	// 发送识别请求
	url := "https://vop.baidu.com/server_api"
	httpResp, err := b.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("识别请求失败: %w", err)
	}
	defer httpResp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 调试：打印原始响应
	fmt.Printf("百度 API 原始响应: %s\n", string(respBody))

	// 解析响应
	var sttResp RecognizeResponse
	if err := json.Unmarshal(respBody, &sttResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w\n原始响应: %s", err, string(respBody))
	}

	// 检查错误
	if sttResp.ErrNo != 0 {
		return nil, fmt.Errorf("识别错误 [%d]: %s", sttResp.ErrNo, sttResp.ErrMsg)
	}

	return sttResp.Result, nil
}

// GetTokenInfo 获取当前 token 信息（供 TTS 复用）
func (b *BaiduSTT) GetTokenInfo() (string, time.Time, error) {
	if b.token == "" || time.Now().After(b.tokenExp) {
		// 尝试从文件加载
		if b.loadTokenFromFile() {
			return b.token, b.tokenExp, nil
		}
		return "", time.Time{}, fmt.Errorf("token 无效或已过期")
	}
	return b.token, b.tokenExp, nil
}
