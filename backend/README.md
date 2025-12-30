# Voice Memory Backend

基于 Gin + 百度语音识别的后端服务。

## 获取百度 API Key

1. 访问 [百度 AI 开放平台](https://console.bce.baidu.com/ai/#/ai/speech/app/list)
2. 创建应用 → 选择"语音识别"
3. 获取 API Key 和 Secret Key
4. 免费额度: 5-15万次/月 (根据认证等级)

## 快速开始

### 1. 设置环境变量

```bash
export BAIDU_API_KEY=你的API_KEY
export BAIDU_SECRET_KEY=你的SECRET_KEY
```

或创建 `.env` 文件:

```bash
cp .env.example .env
# 编辑 .env 填入真实值
```

### 2. 安装依赖

```bash
cd backend
go mod download
```

### 3. 运行服务

```bash
go run cmd/main.go
```

服务将在 http://localhost:8080 启动。

## API 接口

### POST /api/stt

语音识别接口。

**请求:**
- Method: `POST`
- Content-Type: `multipart/form-data`
- Body:
  - `audio`: 音频文件 (WAV/PCM)
  - `format`: 音频格式 (默认: wav)

**响应:**
```json
{
  "success": true,
  "result": ["识别文本1", "识别文本2"]
}
```

**示例 (curl):**
```bash
curl -X POST http://localhost:8080/api/stt \
  -F "audio=@test.wav" \
  -F "format=wav"
```

### GET /health

健康检查接口。

**响应:**
```json
{
  "status": "ok",
  "service": "voice-memory-backend"
}
```

## 项目结构

```
backend/
├── cmd/                 # 主程序入口
│   └── main.go
├── internal/            # 内部包
│   ├── config/         # 配置管理
│   ├── handler/        # HTTP 处理器
│   └── service/        # 业务服务
└── go.mod
```

## 技术栈

- **Gin**: HTTP Web 框架
- **百度语音识别**: STT 服务
