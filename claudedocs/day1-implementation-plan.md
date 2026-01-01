# Day 1: TTS 语音播放 - 实施计划

## 现状分析

### 已有 ✅
- 百度 STT 完整实现（token 管理、文件缓存）
- GLM 文本聊天接口
- 基础 HTTP 框架和路由

### 缺失 ❌
- 任何 TTS 相关代码
- 音频返回的 API 端点
- 前端音频播放功能

---

## 技术方案决策

### 方案选择: 百度 TTS

**理由**:
1. ✅ 已有百度 API 密钥，无需额外申请
2. ✅ 可复用 STT 的 token 管理机制
3. ✅ 文档完善，接口稳定
4. ✅ 支持多种音色和语速

**备选方案** (如果百度有问题):
- 浏览器 speechSynthesis API (免费，但效果一般)

---

## 实施步骤

### 第 1 步: 创建百度 TTS 服务

**文件**: `backend/internal/service/baidu_tts.go`

**功能**:
```go
type BaiduTTS struct {
    apiKey     string
    secretKey  string
    token      string    // 复用 STT 的 token
    tokenExp   time.Time
    tokenFile  string
    client     *http.Client
}

// 核心方法
func (b *BaiduTTS) Synthesize(text string, options TTSOptions) ([]byte, error)
```

**TTS 选项**:
- `tex`: 要合成的文本
- `tok`: access token
- `ctp`: client type = 1
- `lan`: 语言 = zh
- `spd`: 语速 0-15 (默认 5)
- `pit`: 音调 0-15 (默认 5)
- `vol`: 音量 0-15 (默认 5)
- `per`: 发音人
  - 0: 女声
  - 1: 男声
  - 3: 情感合成-度逍遥
  - 4: 情感合成-度丫丫

**API 端点**:
```
https://tsn.baidu.com/text2audio
```

---

### 第 2 步: 创建 TTS Handler

**文件**: `backend/internal/handler/tts_handler.go`

**功能**:
```go
type TTSHandler struct {
    ttsService *service.BaiduTTS
}

// API 端点: GET /api/tts?text=你好&per=0
// 返回: audio/mp3 格式的音频文件
func (h *TTSHandler) HandleTTS(c *gin.Context)
```

**响应**:
- Content-Type: audio/mp3
- 直接返回音频二进制数据

---

### 第 3 步: 修改 Chat Handler

**文件**: `backend/internal/handler/chat_handler.go`

**修改内容**:
```go
type ChatResponse struct {
    Success   bool   `json:"success"`
    Reply     string `json:"reply,omitempty"`
    AudioURL  string `json:"audio_url,omitempty"`  // 新增: 音频 URL
    SessionID string `json:"session_id,omitempty"`
    Error     string `json:"error,omitempty"`
}
```

**两种模式**:
1. **仅文本模式** (当前): 只返回文本
2. **文本+音频模式** (新增): 返回文本和音频 URL

---

### 第 4 步: 添加路由

**文件**: `backend/internal/router/router.go`

**新增路由**:
```go
// TTS 路由
router.GET("/api/tts", cfg.TTSHandler.HandleTTS)
```

---

### 第 5 步: 前端音频播放

**文件**: `web/stt-demo.html`

**新增功能**:
1. **音频播放器**
```html
<audio id="audioPlayer" controls style="display:none;"></audio>
```

2. **自动播放逻辑**
```javascript
async function playAudio(audioUrl) {
    const player = document.getElementById('audioPlayer');
    player.src = audioUrl;
    player.style.display = 'block';
    await player.play();
}
```

3. **聊天集成**
```javascript
// 发送消息后，如果返回有 audio_url，自动播放
if (response.audio_url) {
    await playAudio(response.audio_url);
}
```

---

## API 设计

### 端点 1: 独立 TTS 接口
```
GET /api/tts?text=你好世界&per=0&spd=5

响应:
Content-Type: audio/mp3
Body: <音频二进制数据>
```

### 端点 2: 聊天接口增强
```
POST /api/chat
{
    "message": "你好",
    "session_id": "xxx",
    "include_audio": true    // 新增参数
}

响应:
{
    "success": true,
    "reply": "你好！有什么我可以帮你的吗？",
    "audio_url": "/api/tts?t=xxx",  // 新增
    "session_id": "xxx"
}
```

---

## 文件清单

### 新建文件
- [ ] `backend/internal/service/baidu_tts.go` - TTS 服务
- [ ] `backend/internal/handler/tts_handler.go` - TTS 处理器

### 修改文件
- [ ] `backend/internal/router/router.go` - 添加 TTS 路由
- [ ] `backend/internal/handler/chat_handler.go` - 返回 audio_url
- [ ] `backend/internal/server/server.go` - 初始化 TTS 服务
- [ ] `web/stt-demo.html` - 添加音频播放功能

---

## 测试计划

### 1. 单元测试
```bash
# 测试 TTS 服务
curl "http://localhost:8080/api/tts?text=你好" --output test.mp3
```

### 2. 集成测试
```bash
# 测试聊天带音频
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"你好","include_audio":true}'
```

### 3. 前端测试
- [ ] 打开 stt-demo.html
- [ ] 发送消息
- [ ] 验证自动播放语音

---

## 验收标准

✅ **完成标准**:
1. 调用 `/api/tts?text=xxx` 能返回 mp3 音频
2. 音频可以正常播放
3. 聊天接口可以选择返回音频 URL
4. 前端能自动播放 AI 的语音回复
5. 音质清晰，语速适中

---

## 潜在问题和解决方案

| 问题 | 解决方案 |
|------|----------|
| Token 过期 | 复用 STT 的 token 机制 |
| 文本过长 | 分段合成，然后合并 |
| 音频太大 | 使用临时文件，设置过期删除 |
| 播放失败 | 降级到文本显示 |
| CORS 问题 | 确保路由配置正确 |

---

## 时间估算

| 任务 | 预计时间 |
|------|----------|
| baidu_tts.go | 1 小时 |
| tts_handler.go | 30 分钟 |
| 路由和集成 | 30 分钟 |
| 前端播放 | 30 分钟 |
| 测试调试 | 1 小时 |
| **总计** | **3.5 小时** |

---

准备好了吗？开始实施！
