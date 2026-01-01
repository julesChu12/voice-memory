package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// SherpaSTT Sherpa Onnx 语音识别服务
type SherpaSTT struct {
	addr string
}

// NewSherpaSTT 创建 Sherpa STT 实例
func NewSherpaSTT(addr string) *SherpaSTT {
	// 确保地址不带协议前缀，或者处理 ws://
	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	addr = strings.TrimPrefix(addr, "ws://")
	addr = strings.TrimPrefix(addr, "wss://")
	return &SherpaSTT{
		addr: addr,
	}
}

// SherpaResponse Sherpa WebSocket 响应
type SherpaResponse struct {
	Text    string `json:"text"`
	Segment int    `json:"segment"`
	IsFinal bool   `json:"is_final"`
}

// Recognize 语音识别
func (s *SherpaSTT) Recognize(req *RecognizeRequest) ([]string, error) {
	u := url.URL{Scheme: "ws", Host: s.addr, Path: "/"}
	log.Printf("Connecting to Sherpa STT: %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("connect sherpa error: %w", err)
	}
	defer c.Close()

	// 转换音频数据为 float32 切片 (Assuming 16bit PCM)
	// Sherpa onnx server usually expects raw samples (float32) sent as bytes
	// But check specific server implementation. Many accept raw bytes of 16k 16bit PCM mono.
	
	// 这里假设输入已经是 16k 16bit PCM 或者 WAV
	// 如果是 WAV，需要去掉头部。
	audioData := req.AudioData
	if len(audioData) > 44 && string(audioData[:4]) == "RIFF" {
		audioData = audioData[44:]
	}

	// Converting []byte (int16) to []float32 usually required by some sherpa clients,
	// but the server might accept binary payload as raw PCM samples.
	// Let's try sending binary message directly first (most efficient).
	// Standard sherpa-onnx-online-websocket-server accepts binary messages as chunks of samples.
	// It expects 16kHz, 16-bit mono PCM if sending raw bytes? Or float32?
	
	// Reference python client sends: samples (np.float32) -> tobytes() -> send_bytes.
	// So it expects bytes representing float32 usually.
	
	samples := int16ToFloat32(audioData)
	
	// Send audio in chunks to simulate streaming (or just one big chunk)
	// Sending one big chunk might timeout or be too large.
	chunkSize := 4096 // samples
	
	// Collect results
	var results []string
	
	// Channel to receive text segments
	textChan := make(chan string)
	
	go func() {
		defer close(textChan)
		defer c.Close()
		
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				return
			}
			
			var resp SherpaResponse
			if err := json.Unmarshal(message, &resp); err != nil {
				continue
			}
			
			// If it's final, send it to channel
			if resp.IsFinal {
				textChan <- resp.Text
			} else {
				// Intermediate result
			}
		}
	}()

	// Send audio
	floatBytes := float32ToBytes(samples)
	bytesPerChunk := chunkSize * 4
	for i := 0; i < len(floatBytes); i += bytesPerChunk {
		end := i + bytesPerChunk
		if end > len(floatBytes) {
			end = len(floatBytes)
		}
		err := c.WriteMessage(websocket.BinaryMessage, floatBytes[i:end])
		if err != nil {
			return nil, fmt.Errorf("write audio error: %w", err)
		}
		time.Sleep(5 * time.Millisecond) // throttling
	}

	// Send "Done" signal
	c.WriteMessage(websocket.TextMessage, []byte("Done"))
	
	// Collect all texts
	for text := range textChan {
		if text != "" {
			results = append(results, text)
		}
	}
	
	return results, nil
}

func int16ToFloat32(data []byte) []float32 {
	samples := make([]float32, len(data)/2)
	for i := 0; i < len(data)/2; i++ {
		sample := int16(data[i*2]) | int16(data[i*2+1])<<8
		samples[i] = float32(sample) / 32768.0
	}
	return samples
}

func float32ToBytes(samples []float32) []byte {
	bytes := make([]byte, len(samples)*4)
	for i, sample := range samples {
		bits := math.Float32bits(sample)
		bytes[i*4] = byte(bits)
		bytes[i*4+1] = byte(bits >> 8)
		bytes[i*4+2] = byte(bits >> 16)
		bytes[i*4+3] = byte(bits >> 24)
	}
	return bytes
}
