// Voice Memory WebSocket Client
// 负责音频采集、WebSocket 通信和状态管理

class VoiceClient {
    constructor(url, onStateChange, onTranscript, onAudio, onSpeechStart, onAIResponse) {
        this.url = url;
        this.socket = null;
        this.audioContext = null;
        this.processor = null;
        this.mediaStream = null;
        this.isRecording = false;
        this.isSpeaking = false; // Add isSpeaking property
        
        // VAD parameters
        this.silenceThreshold = 0.03; // Increased noise threshold
        this.silenceDuration = 3000; // 3 seconds of silence
        this.silenceTimer = null;
        
        // 回调函数
        this.onStateChange = onStateChange || (() => {});
        this.onTranscript = onTranscript || (() => {});
        this.onAudio = onAudio || (() => {});
        this.onSpeechStart = onSpeechStart || (() => {});
        this.onAIResponse = onAIResponse || (() => {});
    }

    // 连接 WebSocket
    connect() {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) return;

        console.log('正在连接 WebSocket:', this.url);
        this.socket = new WebSocket(this.url);
        this.socket.binaryType = 'arraybuffer';

        this.socket.onopen = () => {
            console.log('WebSocket 已连接');
            this.onStateChange('connected');
        };

        this.socket.onclose = () => {
            console.log('WebSocket 已断开');
            this.onStateChange('disconnected');
            // 自动重连逻辑可以在这里添加
        };

        this.socket.onerror = (error) => {
            console.error('WebSocket 错误:', error);
            this.onStateChange('error');
        };

        this.socket.onmessage = (event) => {
            this.handleMessage(event);
        };
    }

    // 处理收到的消息
    handleMessage(event) {
        if (event.data instanceof ArrayBuffer) {
            // 收到音频数据 (TTS)
            console.log('<- [WS] 收到音频数据, 大小:', event.data.byteLength);
            this.onAudio(event.data);
            return;
        }

        try {
            const msg = JSON.parse(event.data);
            console.log('<- [WS] 收到 JSON:', msg);
            switch (msg.type) {
                case 'state':
                    this.onStateChange(msg.status);
                    break;
                case 'stt_intermediate':
                case 'stt_final':
                    this.onTranscript(msg.text, msg.type === 'stt_final');
                    break;
                case 'llm_reply':
                    this.onAIResponse(msg.text);
                    break;
                case 'error':
                    console.error('服务端错误:', msg.error);
                    this.onStateChange('error');
                    break;
            }
        } catch (e) {
            console.error('解析消息失败:', e);
        }
    }

    // 开始录音并发送
    async startRecording() {
        if (this.isRecording) return;
        
        if (!this.socket || this.socket.readyState === WebSocket.OPEN) {
            this.connect();
            // 等待连接建立... 简单起见这里先假设已连接或极快连接
            // 生产环境应该用 Promise 等待 onopen
        }

        try {
            this.mediaStream = await navigator.mediaDevices.getUserMedia({ audio: true });
            
            this.audioContext = new (window.AudioContext || window.webkitAudioContext)({ sampleRate: 16000 });
            const source = this.audioContext.createMediaStreamSource(this.mediaStream);
            
            // 使用 ScriptProcessor (虽然已废弃但兼容性好，Phase 3 可升级为 AudioWorklet)
            // 缓冲区大小 4096，单声道
            this.processor = this.audioContext.createScriptProcessor(4096, 1, 1);

            this.processor.onaudioprocess = (e) => {
                if (!this.isRecording) return;
                
                // 获取 PCM 数据 (Float32: -1.0 ~ 1.0)
                const inputData = e.inputBuffer.getChannelData(0);

                // --- 噪音门 (Noise Gate) & 静音检测 ---
                let sum = 0;
                for (let i = 0; i < inputData.length; i++) {
                    sum += inputData[i] * inputData[i];
                }
                const rms = Math.sqrt(sum / inputData.length);
                
                if (rms > this.silenceThreshold) { // 检测到语音
                    if (!this.isSpeaking) {
                        this.isSpeaking = true;
                        this.onSpeechStart();
                        console.log('检测到语音开始');
                    }
                    clearTimeout(this.silenceTimer); // 清除静音计时器
                    this.silenceTimer = null;
                } else { // 检测到静音
                    if (this.isSpeaking && this.silenceTimer === null) {
                        // 如果正在说话，且静音计时器未启动，则启动计时器
                        this.silenceTimer = setTimeout(() => {
                            this.handleSilence();
                        }, this.silenceDuration);
                        console.log('检测到静音，启动计时器...');
                    }
                    this.isSpeaking = false; // 不再说话
                }
                
                // 如果是说话状态，发送音频
                if (this.isSpeaking) {
                    // 转换为 Int16 PCM (百度 STT 需要)
                    const pcmData = this.floatTo16BitPCM(inputData);
                    
                                    // 发送通过 WebSocket
                                    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                                        // console.log('-> [WS] 发送音频数据块, 大小:', pcmData.byteLength); // 频率太高，默认注释掉，需要时可打开
                                        this.socket.send(pcmData);
                                    }                }
            };

            source.connect(this.processor);
            this.processor.connect(this.audioContext.destination); // 必须连接到 destination 才能工作

            this.isRecording = true;
            this.onStateChange('recording');
            
        } catch (error) {
            console.error('启动录音失败:', error);
            this.onStateChange('error');
        }
    }

    // 停止录音
    stopRecording() {
        this.isRecording = false;
        this.isSpeaking = false;
        clearTimeout(this.silenceTimer); // 清除静音计时器
        this.silenceTimer = null;
        
        if (this.processor) {
            this.processor.disconnect();
            this.processor = null;
        }
        
        if (this.mediaStream) {
            this.mediaStream.getTracks().forEach(track => track.stop());
            this.mediaStream = null;
        }
        
        if (this.audioContext) {
            this.audioContext.close();
            this.audioContext = null;
        }

        // 发送停止信号(可选，目前服务端是基于包的，这可能用于告诉服务端这波说完了)
        // this.socket.send(JSON.stringify({type: "stop"}));
        this.onStateChange('idle'); // Change state to idle after stopping
    }

    // 处理静音超时
    handleSilence() {
        console.log('静音超时，停止录音');
        if (this.isRecording) {
            this.stopRecording();
        }
    }

    // 发送纯文本
    sendText(text) {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify({ type: 'text', text: text }));
            this.onStateChange('processing');
        }
    }

    // 发送打断信号
    interrupt() {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify({ type: 'interrupt' }));
            this.onStateChange('interrupted');
        }
    }

    // 辅助：Float32 转 Int16
    floatTo16BitPCM(input) {
        const output = new Int16Array(input.length);
        for (let i = 0; i < input.length; i++) {
            const s = Math.max(-1, Math.min(1, input[i]));
            output[i] = s < 0 ? s * 0x8000 : s * 0x7FFF;
        }
        return output.buffer;
    }
}
