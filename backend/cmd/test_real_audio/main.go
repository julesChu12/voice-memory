package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	serverAddr := flag.String("addr", "localhost:8080", "http service address")
	audioFile := flag.String("file", "../web/16k.wav", "path to audio file")
	flag.Parse()

	// 1. 读取音频文件
	// 尝试相对于 backend 目录或当前目录寻找文件
	pathsToCheck := []string{
		*audioFile,
		filepath.Join("..", *audioFile),
		filepath.Join("web", "16k.wav"),
		filepath.Join("..", "web", "16k.wav"),
		"/Users/yt/Documents/project/voice-memory/web/16k.wav",
	}

	var fileData []byte
	var err error
	var finalPath string

	for _, p := range pathsToCheck {
		fileData, err = os.ReadFile(p)
		if err == nil {
			finalPath = p
			break
		}
	}

	if err != nil {
		log.Fatal("❌ 无法找到音频文件:", err)
	}
	log.Printf("📂 加载音频文件: %s (大小: %d bytes)", finalPath, len(fileData))

	// 跳过 WAV 头 (44 bytes)
	if len(fileData) > 44 {
		fileData = fileData[44:]
	}

	// 2. 连接 WebSocket
	u := url.URL{Scheme: "ws", Host: *serverAddr, Path: "/ws"}
	log.Printf("🔗 连接服务器: %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("❌ 连接失败 (请确保 'air' 或 'go run' 正在运行):", err)
	}
	defer c.Close()

	done := make(chan struct{})

	// 3. 接收消息协程
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("❌ 读取结束:", err)
				return
			}
			log.Printf("📩 收到消息: %s", message)
		}
	}()

	// 4. 发送音频流
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	log.Println("🎙️ 开始发送音频流...")
	
	chunkSize := 3200 // 100ms at 16k sample rate (2 bytes per sample) -> 16000 * 2 * 0.1 = 3200 bytes
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	offset := 0
	
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if offset >= len(fileData) {
				log.Println("✅ 音频发送完毕")
				// 保持连接一会等待回复
				select {
				case <-done:
				case <-time.After(5 * time.Second): // 等待 5 秒
					log.Println("⏳ 测试超时或结束")
				}
				c.Close()
				return
			}

			end := offset + chunkSize
			if end > len(fileData) {
				end = len(fileData)
			}

			err := c.WriteMessage(websocket.BinaryMessage, fileData[offset:end])
			if err != nil {
				log.Println("❌ 发送错误:", err)
				return
			}
			// log.Printf("-> 发送块 %d-%d", offset, end)
			offset = end
		case <-interrupt:
			log.Println("🛑 用户打断")
			
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
