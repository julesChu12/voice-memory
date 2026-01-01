# Sherpa Onnx Integration Guide

## Overview
We have integrated `sherpa-onnx` for local Speech-to-Text (STT) and Text-to-Speech (TTS) capabilities. This allows the Voice Memory system to run without relying on external cloud services (Baidu) for audio processing.

## Configuration
To use Sherpa Onnx, update your `.env` file or set the following environment variables:

```bash
# Service Providers (baidu or sherpa)
STT_PROVIDER=sherpa
TTS_PROVIDER=sherpa

# Sherpa Onnx Server Addresses
# For STT (WebSocket):
SHERPA_STT_ADDR=localhost:6006
# For TTS (HTTP):
SHERPA_TTS_ADDR=http://localhost:19000
```

## Running Sherpa Onnx Servers
Ensure your local `sherpa-onnx` servers are running.

### STT Server (Streaming)
Use `sherpa-onnx-online-websocket-server`.
Example:
```bash
sherpa-onnx-online-websocket-server \
  --port=6006 \
  --tokens=tokens.txt \
  --encoder=encoder.onnx \
  --decoder=decoder.onnx \
  --joiner=joiner.onnx
```

### TTS Server (Offline/HTTP)
Use `sherpa-onnx-offline-tts-server`.
Example:
```bash
sherpa-onnx-offline-tts-server \
  --port=19000 \
  --model=model.onnx \
  --lexicon=lexicon.txt \
  --tokens=tokens.txt \
  --espeak-data=espeak-ng-data
```

## Code Changes
1.  **Config**: Added provider and address configurations.
2.  **Service Interfaces**: Abstrated `STTService` and `TTSService` to support multiple providers.
3.  **Implementations**: Added `SherpaSTT` (WebSocket) and `SherpaTTS` (HTTP).
4.  **Backend Logic**: Updated `Server` and `Handlers` to dynamically load and use the configured provider.
