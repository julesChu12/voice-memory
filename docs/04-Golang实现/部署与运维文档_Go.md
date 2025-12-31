# Voice Memory - 部署与运维文档 (Golang版本)

**版本：** v1.0-Go
**日期：** 2025-12-29
**语言：** Golang

---

## 目录

1. [部署概述](#一部署概述)
2. [本地部署](#二本地部署)
3. [生产部署](#三生产部署)
4. [Docker部署](#四docker部署)
5. [监控与日志](#五监控与日志)
6. [备份与恢复](#六备份与恢复)
7. [性能优化](#七性能优化)
8. [安全配置](#八安全配置)
9. [故障排除](#九故障排除)

---

## 一、部署概述

### 1.1 部署架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Voice Memory 部署架构                     │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                      用户端                              │
│  Web Browser (PWA) / Terminal CLI / Mobile App          │
└─────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────┐
│                  Voice Memory Server                     │
│  (Golang Binary + Gin + SQLite)                         │
└─────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────┐
│                   外部服务                               │
│  • Claude API (Anthropic)                               │
│  • OpenAI API (Whisper + TTS)                           │
└─────────────────────────────────────────────────────────┘
```

### 1.2 部署模式

| 模式 | 复杂度 | 适用场景 | 扩展性 |
|------|--------|----------|--------|
| **单机部署** | 低 | 个人使用 / MVP | 无 |
| **Docker部署** | 中 | 中小团队 | 中 |
| **云托管** | 中 | 快速上线 | 高 |
| **Kubernetes** | 高 | 企业级 | 高 |

---

## 二、本地部署

### 2.1 环境准备

**系统要求**:
- Go 1.21+ (编译时)
- 100MB 磁盘空间
- 512MB RAM (最低)
- macOS / Linux / Windows

### 2.2 编译构建

```bash
# 克隆项目
git clone https://github.com/yourusername/voice-memory-go.git
cd voice-memory-go

# 安装依赖
go mod download

# 复制环境变量模板
cp .env.example .env

# 编辑.env，填入API密钥
nano .env

# 构建二进制文件
go build -o bin/voice-memory cmd/server/main.go

# 或使用Make
make build
```

### 2.3 配置文件

**.env**:
```bash
# API密钥
CLAUDE_API_KEY=your-claude-api-key-here
OPENAI_API_KEY=your-openai-api-key-here

# 服务器配置
PORT=8080
HOST=0.0.0.0

# 数据库
DATABASE_PATH=data/voice-memory.db

# 日志
LOG_LEVEL=info
LOG_FILE=logs/voice-memory.log

# 限制
MAX_UPLOAD_SIZE=10MB
RATE_LIMIT=60  # 每分钟请求数
```

### 2.4 启动服务

```bash
# 直接运行
./bin/voice-memory

# 或使用go run（开发模式）
go run cmd/server/main.go

# 后台运行
nohup ./bin/voice-memory > logs/app.log 2>&1 &

# 查看日志
tail -f logs/app.log
```

### 2.5 验证部署

```bash
# 检查服务状态
curl http://localhost:8080/health

# 测试API
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "你好"}'

# 测试WebSocket（使用wscat）
wscat -c ws://localhost:8080/ws
```

---

## 三、生产部署

### 3.1 Linux Systemd服务

**创建服务文件**:
```bash
sudo nano /etc/systemd/system/voice-memory.service
```

**内容**:
```ini
[Unit]
Description=Voice Memory Server
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=voicememory
Group=voicememory
WorkingDirectory=/opt/voice-memory

# 环境变量
Environment="CLAUDE_API_KEY=your-key"
Environment="OPENAI_API_KEY=your-key"
Environment="PORT=8080"
Environment="DATABASE_PATH=/var/lib/voice-memory/data.db"

# 执行命令
ExecStart=/opt/voice-memory/bin/voice-memory
ExecReload=/bin/kill -HUP $MAINPID

# 重启策略
Restart=always
RestartSec=5

# 安全
NoNewPrivileges=true
PrivateTmp=true

# 日志
StandardOutput=journal
StandardError=journal
SyslogIdentifier=voice-memory

[Install]
WantedBy=multi-user.target
```

**管理服务**:
```bash
# 创建用户
sudo useradd -r -s /bin/false voicememory

# 设置权限
sudo mkdir -p /opt/voice-memory
sudo mkdir -p /var/lib/voice-memory
sudo chown -R voicememory:voicememory /opt/voice-memory
sudo chown -R voicememory:voicememory /var/lib/voice-memory

# 复制文件
sudo cp bin/voice-memory /opt/voice-memory/
sudo cp -r web /opt/voice-memory/

# 重载systemd
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start voice-memory

# 开机自启
sudo systemctl enable voice-memory

# 查看状态
sudo systemctl status voice-memory

# 查看日志
sudo journalctl -u voice-memory -f
```

### 3.2 Nginx反向代理

**配置文件**:
```bash
sudo nano /etc/nginx/sites-available/voice-memory
```

**内容**:
```nginx
upstream voice_memory {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name your-domain.com;

    # 重定向到HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL证书
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    # SSL配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # 客户端上传大小限制
    client_max_body_size 10M;

    # 日志
    access_log /var/log/nginx/voice-memory-access.log;
    error_log /var/log/nginx/voice-memory-error.log;

    # HTTP API
    location /api/ {
        proxy_pass http://voice_memory;
        proxy_http_version 1.1;

        # Headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 超时
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # WebSocket
    location /ws {
        proxy_pass http://voice_memory;
        proxy_http_version 1.1;

        # WebSocket升级
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # Headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 超时
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;
    }

    # 静态文件
    location / {
        proxy_pass http://voice_memory;
    }
}
```

**启用配置**:
```bash
# 创建符号链接
sudo ln -s /etc/nginx/sites-available/voice-memory /etc/nginx/sites-enabled/

# 测试配置
sudo nginx -t

# 重载Nginx
sudo systemctl reload nginx
```

### 3.3 HTTPS配置 (Let's Encrypt)

```bash
# 安装certbot
sudo apt-get install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d your-domain.com

# 自动续期
sudo certbot renew --dry-run
```

---

## 四、Docker部署

### 4.1 Dockerfile

```dockerfile
# 多阶段构建

# 构建阶段
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o voice-memory cmd/server/main.go

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# 从构建阶段复制
COPY --from=builder /app/voice-memory .
COPY --from=builder /app/web ./web
COPY --from=builder /app/.env.example .env

# 创建数据目录
RUN mkdir -p data logs

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 运行
CMD ["./voice-memory"]
```

### 4.2 docker-compose.yml

```yaml
version: '3.8'

services:
  voice-memory:
    build: .
    container_name: voice-memory
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - CLAUDE_API_KEY=${CLAUDE_API_KEY}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - PORT=8080
      - LOG_LEVEL=info
    volumes:
      - ./data:/root/data
      - ./logs:/root/logs
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - voice-memory-network

  # 可选：Nginx反向代理
  nginx:
    image: nginx:alpine
    container_name: voice-memory-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - voice-memory
    networks:
      - voice-memory-network

networks:
  voice-memory-network:
    driver: bridge
```

### 4.3 Docker命令

```bash
# 构建镜像
docker build -t voice-memory:latest .

# 运行容器
docker run -d \
  --name voice-memory \
  -p 8080:8080 \
  -e CLAUDE_API_KEY=your-key \
  -e OPENAI_API_KEY=your-key \
  -v $(pwd)/data:/root/data \
  voice-memory:latest

# 查看日志
docker logs -f voice-memory

# 进入容器
docker exec -it voice-memory sh

# 停止容器
docker stop voice-memory

# 删除容器
docker rm voice-memory

# 使用docker-compose
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重启服务
docker-compose restart
```

### 4.4 Docker优化

**减小镜像大小**:
```dockerfile
# 使用多阶段构建（见上文）
# 使用alpine基础镜像
# 删除不必要的文件

# 构建时使用--no-cache
docker build --no-cache -t voice-memory .

# 使用.dockerignore
echo "*.md" >> .dockerignore
echo ".git" >> .dockerignore
echo "node_modules" >> .dockerignore
```

---

## 五、监控与日志

### 5.1 日志配置

**结构化日志**:
```go
// internal/logger/logger.go
package logger

import (
    "log/slog"
    "os"
)

var Logger *slog.Logger

func Init(level string) {
    var lvl slog.Level
    switch level {
    case "debug":
        lvl = slog.LevelDebug
    case "info":
        lvl = slog.LevelInfo
    case "warn":
        lvl = slog.LevelWarn
    case "error":
        lvl = slog.LevelError
    default:
        lvl = slog.LevelInfo
    }

    opts := &slog.HandlerOptions{
        Level: lvl,
    }

    // JSON格式（生产环境）
    handler := slog.NewJSONHandler(os.Stdout, opts)

    // 文本格式（开发环境）
    // handler := slog.NewTextHandler(os.Stdout, opts)

    Logger = slog.New(handler)
    slog.SetDefault(Logger)
}
```

**使用示例**:
```go
import "voice-memory/internal/logger"

logger.Logger.Info("Server started", "port", 8080)
logger.Logger.Error("Database error", "error", err)
logger.Logger.Debug("Processing request", "path", r.URL.Path)
```

### 5.2 日志轮转

**使用logrotate**:
```bash
# /etc/logrotate.d/voice-memory
/var/log/voice-memory/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 voicememory voicememory
    sharedscripts
    postrotate
        systemctl reload voice-memory > /dev/null 2>&1 || true
    endscript
}
```

### 5.3 健康检查端点

```go
// internal/api/health.go
package api

import (
    "net/http"
    "runtime"
    "time"

    "github.com/gin-gonic/gin"
)

type HealthResponse struct {
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
    Uptime    string    `json:"uptime"`
    Memory    Memory    `json:"memory"`
}

type Memory struct {
    Alloc      uint64 `json:"alloc"`
    TotalAlloc uint64 `json:"total_alloc"`
    Sys        uint64 `json:"sys"`
    NumGC      uint32 `json:"num_gc"`
}

var startTime time.Time

func init() {
    startTime = time.Now()
}

func (h *Handlers) HealthCheck(c *gin.Context) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    c.JSON(http.StatusOK, HealthResponse{
        Status:    "healthy",
        Timestamp: time.Now(),
        Uptime:    time.Since(startTime).String(),
        Memory: Memory{
            Alloc:      m.Alloc,
            TotalAlloc: m.TotalAlloc,
            Sys:        m.Sys,
            NumGC:      m.NumGC,
        },
    })
}
```

### 5.4 Prometheus监控 (可选)

```go
// internal/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    RequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "voice_memory_requests_total",
            Help: "Total number of requests",
        },
        []string{"method", "endpoint"},
    )

    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "voice_memory_request_duration_seconds",
            Help:    "Request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
)

// 中间件
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        c.Next()

        duration := time.Since(start).Seconds()
        RequestsTotal.WithLabelValues(c.Request.Method, c.FullPath()).Inc()
        RequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
    }
}
```

---

## 六、备份与恢复

### 6.1 数据库备份

**备份脚本**:
```bash
#!/bin/bash
# scripts/backup.sh

BACKUP_DIR="/backup/voice-memory"
DB_PATH="/var/lib/voice-memory/data.db"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/voice-memory_$TIMESTAMP.db"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份数据库
cp $DB_PATH $BACKUP_FILE

# 压缩
gzip $BACKUP_FILE

# 删除30天前的备份
find $BACKUP_DIR -name "voice-memory_*.db.gz" -mtime +30 -delete

echo "Backup completed: $BACKUP_FILE.gz"
```

**定时备份** (crontab):
```bash
# 每天凌晨2点备份
0 2 * * * /opt/voice-memory/scripts/backup.sh
```

### 6.2 恢复

```bash
#!/bin/bash
# scripts/restore.sh

BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: ./restore.sh <backup_file>"
    exit 1
fi

# 停止服务
sudo systemctl stop voice-memory

# 解压备份
gunzip $BACKUP_FILE

# 恢复数据库
DB_FILE="${BACKUP_FILE%.gz}"
cp $DB_FILE /var/lib/voice-memory/data.db

# 启动服务
sudo systemctl start voice-memory

echo "Restore completed"
```

---

## 七、性能优化

### 7.1 数据库优化

**PRAGMA设置**:
```go
// 优化SQLite性能
pragmaStatements := []string{
    "PRAGMA journal_mode = WAL",           // 写前日志
    "PRAGMA synchronous = NORMAL",         // 降低同步级别
    "PRAGMA cache_size = -64000",          // 64MB缓存
    "PRAGMA temp_store = MEMORY",          // 临时表在内存
    "PRAGMA mmap_size = 30000000000",      // 启用内存映射
    "PRAGMA page_size = 4096",             // 页大小
}

for _, stmt := range pragmaStatements {
    db.Exec(stmt)
}
```

### 7.2 连接池配置

```go
// 设置连接池参数
db.SetMaxOpenConns(25)       // 最大打开连接数
db.SetMaxIdleConns(25)       // 最大空闲连接数
db.SetConnMaxLifetime(5 * time.Minute)  // 连接最大生命周期
```

### 7.3 Go编译优化

```bash
# 优化编译
go build -ldflags="-s -w" -o bin/voice-memory cmd/server/main.go

# 使用-uplink标志优化
go build -ldflags="-s -w -linkmode=external" -extldflags="-static" -o bin/voice-memory cmd/server/main.go
```

---

## 八、安全配置

### 8.1 环境变量保护

```bash
# .env应该添加到.gitignore
echo ".env" >> .gitignore

# 设置文件权限
chmod 600 .env
```

### 8.2 限流配置

```go
// API限流
limiter := rate.NewLimiter(rate.Every(time.Minute), 60)  // 每分钟60个请求

func RateLimitMiddleware(l *rate.Limiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !l.Allow() {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
            })
            return
        }
        c.Next()
    }
}
```

### 8.3 CORS配置

```go
// 限制CORS来源
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")

        // 检查来源是否允许
        allowed := false
        for _, ao := range allowedOrigins {
            if origin == ao {
                allowed = true
                break
            }
        }

        if allowed {
            c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
        }

        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

---

## 九、故障排除

### 9.1 常见问题

**问题1: 服务无法启动**
```bash
# 检查端口占用
sudo lsof -i :8080

# 检查日志
sudo journalctl -u voice-memory -n 50

# 检查配置文件
cat /opt/voice-memory/.env
```

**问题2: 数据库锁定**
```bash
# 检查SQLite锁
sqlite3 /var/lib/voice-memory/data.db "PRAGMA database_list"

# 重启服务释放锁
sudo systemctl restart voice-memory
```

**问题3: API调用失败**
```bash
# 检查API密钥
echo $CLAUDE_API_KEY

# 测试网络连接
curl -v https://api.anthropic.com

# 检查日志
tail -f /var/log/voice-memory/app.log
```

### 9.2 日志级别调试

```bash
# 临时启用调试级别
sudo systemctl edit voice-memory

# 添加：
[Service]
Environment="LOG_LEVEL=debug"

# 重启服务
sudo systemctl restart voice-memory
```

### 9.3 性能分析

```bash
# 启用pprof（需要代码支持）
go tool pprof http://localhost:8080/debug/pprof/profile

# 查看堆内存
go tool pprof http://localhost:8080/debug/pprof/heap
```

---

## 十、更新与维护

### 10.1 更新流程

```bash
# 1. 备份数据
./scripts/backup.sh

# 2. 停止服务
sudo systemctl stop voice-memory

# 3. 下载新版本
git pull origin main

# 4. 构建
go build -o bin/voice-memory cmd/server/main.go

# 5. 复制到部署目录
sudo cp bin/voice-memory /opt/voice-memory/

# 6. 运行迁移（如果有）
./bin/voice-memory --migrate

# 7. 启动服务
sudo systemctl start voice-memory

# 8. 验证
curl http://localhost:8080/health
```

### 10.2 回滚

```bash
# 如果更新失败，回滚
sudo systemctl stop voice-memory
sudo cp /opt/voice-memory/bin/voice-memory.backup /opt/voice-memory/voice-memory
sudo systemctl start voice-memory
```

---

**文档版本历史**:
- v1.0-Go (2025-12-29): 创建Golang版本部署与运维文档
