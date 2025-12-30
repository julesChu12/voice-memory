package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewBaiduTTS 测试创建 TTS 实例
func TestNewBaiduTTS(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 保存原始 HOME，设置临时目录
	origHome := os.Getenv("HOME")
	defer func() {
		if origHome != "" {
			os.Setenv("HOME", origHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	os.Setenv("HOME", tempDir)

	tts := NewBaiduTTS("test_api_key", "test_secret_key")

	if tts == nil {
		t.Fatal("NewBaiduTTS 返回 nil")
	}

	if tts.apiKey != "test_api_key" {
		t.Errorf("期望 apiKey 'test_api_key', 得到 '%s'", tts.apiKey)
	}

	if tts.secretKey != "test_secret_key" {
		t.Errorf("期望 secretKey 'test_secret_key', 得到 '%s'", tts.secretKey)
	}

	// 检查目录是否创建
	tokenDir := filepath.Join(tempDir, ".voice-memory")
	if _, err := os.Stat(tokenDir); os.IsNotExist(err) {
		t.Errorf("token 目录未创建: %s", tokenDir)
	}

	audioDir := filepath.Join(tempDir, ".voice-memory", "audio")
	if _, err := os.Stat(audioDir); os.IsNotExist(err) {
		t.Errorf("audio 目录未创建: %s", audioDir)
	}
}

// TestDefaultTTSOptions 测试默认 TTS 选项
func TestDefaultTTSOptions(t *testing.T) {
	text := "你好，世界"
	opts := DefaultTTSOptions(text)

	if opts.Text != text {
		t.Errorf("期望 Text '%s', 得到 '%s'", text, opts.Text)
	}

	if opts.Per != 4195 {
		t.Errorf("期望 Per 4195 (精品女声), 得到 %d", opts.Per)
	}

	if opts.Spd != 6 {
		t.Errorf("期望 Spd 6 (稍快语速), 得到 %d", opts.Spd)
	}

	if opts.Pit != 6 {
		t.Errorf("期望 Pit 6 (中等音调), 得到 %d", opts.Pit)
	}

	if opts.Vol != 8 {
		t.Errorf("期望 Vol 8 (较高音量), 得到 %d", opts.Vol)
	}
}

// TestTTSOptions 测试自定义 TTS 选项
func TestTTSOptions(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		per      int
		spd      int
		pit      int
		vol      int
	}{
		{
			name: "默认选项",
			text: "测试",
			per:  0,
			spd:  5,
			pit:  5,
			vol:  5,
		},
		{
			name: "男声",
			text: "测试",
			per:  1,
			spd:  5,
			pit:  5,
			vol:  5,
		},
		{
			name: "快速语速",
			text: "测试",
			per:  0,
			spd:  10,
			pit:  5,
			vol:  5,
		},
		{
			name: "高音调",
			text: "测试",
			per:  0,
			spd:  5,
			pit:  10,
			vol:  5,
		},
		{
			name: "最大音量",
			text: "测试",
			per:  0,
			spd:  5,
			pit:  5,
			vol:  15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := TTSOptions{
				Text: tt.text,
				Per:  tt.per,
				Spd:  tt.spd,
				Pit:  tt.pit,
				Vol:  tt.vol,
			}

			if opts.Text != tt.text {
				t.Errorf("Text 不匹配: 期望 %s, 得到 %s", tt.text, opts.Text)
			}
			if opts.Per != tt.per {
				t.Errorf("Per 不匹配: 期望 %d, 得到 %d", tt.per, opts.Per)
			}
			if opts.Spd != tt.spd {
				t.Errorf("Spd 不匹配: 期望 %d, 得到 %d", tt.spd, opts.Spd)
			}
			if opts.Pit != tt.pit {
				t.Errorf("Pit 不匹配: 期望 %d, 得到 %d", tt.pit, opts.Pit)
			}
			if opts.Vol != tt.vol {
				t.Errorf("Vol 不匹配: 期望 %d, 得到 %d", tt.vol, opts.Vol)
			}
		})
	}
}

// TestSimpleHash 测试简单哈希函数
func TestSimpleHash(t *testing.T) {
	// 测试相同输入产生相同输出
	h1 := simpleHash("test")
	h2 := simpleHash("test")
	if h1 != h2 {
		t.Errorf("相同输入应产生相同哈希: %s != %s", h1, h2)
	}

	// 测试不同输入产生不同输出
	h3 := simpleHash("test1")
	h4 := simpleHash("test2")
	if h3 == h4 {
		t.Errorf("不同输入应产生不同哈希: %s == %s", h3, h4)
	}

	// 测试空字符串
	h5 := simpleHash("")
	if h5 != "0" {
		t.Errorf("空字符串哈希应为 '0', 得到 '%s'", h5)
	}

	// 测试哈希值不为空且是有效的十六进制
	h6 := simpleHash("hello")
	if h6 == "" {
		t.Error("哈希值不应为空")
	}
	// 验证是有效的十六进制字符串（只包含 0-9 和 a-f）
	for _, c := range h6 {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("哈希值应只包含十六进制字符, 得到 '%s'", h6)
		}
	}
}

// TestGetAudioDir 测试获取音频目录
func TestGetAudioDir(t *testing.T) {
	tempDir := t.TempDir()

	origHome := os.Getenv("HOME")
	defer func() {
		if origHome != "" {
			os.Setenv("HOME", origHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()
	os.Setenv("HOME", tempDir)

	tts := NewBaiduTTS("test_key", "test_secret")
	audioDir := tts.GetAudioDir()

	expectedDir := filepath.Join(tempDir, ".voice-memory", "audio")
	if audioDir != expectedDir {
		t.Errorf("期望音频目录 %s, 得到 %s", expectedDir, audioDir)
	}
}

// TestSynthesizeBase64 测试 Base64 编码
func TestSynthesizeBase64(t *testing.T) {
	// 这个测试需要真实的 API，先跳过
	t.Skip("需要真实 API key")

	// TODO: 当有 API key 时，可以测试：
	// 1. 调用 SynthesizeBase64
	// 2. 验证返回的是有效的 Base64 字符串
	// 3. 验证可以解码回音频数据
}

// TestCleanupOldFiles 测试清理旧文件
func TestCleanupOldFiles(t *testing.T) {
	tempDir := t.TempDir()

	origHome := os.Getenv("HOME")
	defer func() {
		if origHome != "" {
			os.Setenv("HOME", origHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()
	os.Setenv("HOME", tempDir)

	tts := NewBaiduTTS("test_key", "test_secret")
	audioDir := tts.GetAudioDir()

	// 创建一些测试文件
	testFiles := []string{"old1.mp3", "old2.mp3", "recent.mp3"}
	for _, fname := range testFiles {
		fpath := filepath.Join(audioDir, fname)
		if err := os.WriteFile(fpath, []byte("test data"), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}
	}

	// 修改一个文件的时间为 2 小时前（超过 maxAge）
	oldFile := filepath.Join(audioDir, "old1.mp3")
	twoHoursAgo := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldFile, twoHoursAgo, twoHoursAgo); err != nil {
		t.Fatalf("修改文件时间失败: %v", err)
	}

	// 清理 1 小时前的文件
	maxAge := 1 * time.Hour
	if err := tts.CleanupOldFiles(maxAge); err != nil {
		t.Fatalf("CleanupOldFiles 失败: %v", err)
	}

	// 验证旧文件被删除
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("old1.mp3 应该被删除，但仍然存在")
	}

	// 验证新文件仍然存在
	recentFile := filepath.Join(audioDir, "recent.mp3")
	if _, err := os.Stat(recentFile); os.IsNotExist(err) {
		t.Error("recent.mp3 不应该被删除")
	}
}

// TestGetTokenInfo 测试获取 token 信息
func TestGetTokenInfo(t *testing.T) {
	tts := NewBaiduTTS("test_key", "test_secret")

	// 没有 token 时应该返回错误
	_, exp, err := tts.GetTokenInfo()
	if err == nil {
		t.Error("期望返回错误，但得到了成功")
	}
	if !exp.IsZero() {
		t.Error("期望过期时间为零值")
	}
}

// TestGetAccessTokenFromSTT 测试从 STT 复用 token
func TestGetAccessTokenFromSTT(t *testing.T) {
	// 创建 STT 和 TTS 实例
	stt := NewBaiduSTT("test_key", "test_secret")
	tts := NewBaiduTTS("test_key", "test_secret")

	// 手动设置一个测试 token
	testToken := "test_token_123"
	testExp := time.Now().Add(1 * time.Hour)

	stt.token = testToken
	stt.tokenExp = testExp

	// 从 STT 复制 token
	if err := tts.GetAccessTokenFromSTT(stt); err != nil {
		t.Fatalf("GetAccessTokenFromSTT 失败: %v", err)
	}

	// 验证 token 被正确复制
	token, exp, err := tts.GetTokenInfo()
	if err != nil {
		t.Fatalf("获取 token 信息失败: %v", err)
	}

	if token != testToken {
		t.Errorf("期望 token '%s', 得到 '%s'", testToken, token)
	}

	// 检查过期时间在 1 分钟内
	if exp.Sub(testExp) > time.Minute {
		t.Errorf("过期时间不匹配，差异: %v", exp.Sub(testExp))
	}
}

// BenchmarkSimpleHash 性能测试
func BenchmarkSimpleHash(b *testing.B) {
	str := "这是一个测试字符串，用于哈希性能测试"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		simpleHash(str)
	}
}

// BenchmarkDefaultTTSOptions 性能测试
func BenchmarkDefaultTTSOptions(b *testing.B) {
	text := "这是一个测试文本"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DefaultTTSOptions(text)
	}
}
