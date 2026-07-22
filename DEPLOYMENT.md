# ==============================================================================
# PROJECT JARVIS - DEPLOYMENT & SETUP GUIDE
# ==============================================================================

This guide explains how to deploy the Project Jarvis AI Knowledge Agent across 
your network.

## Quick Setup (Recommended)

From the project root directory:

```bash
./setup.sh
```

This will configure both the client and server automatically. Then:

1. Edit `server/.env` with your configuration
2. Edit `client/.env` with your server IP address
3. Follow the manual deployment steps below

---

## 1. Compute Node Setup (Ubuntu Server + RTX 4090)

### A. Prerequisites
- **Network**: Ensure the Ubuntu server has a **static IP address**. This is critical as the client relies on it for WebSocket communication.
- **Firewall**: Open port `8080` to allow incoming traffic from the local network:
  ```bash
  sudo ufw allow 8080/tcp
  ```
- **GPU Drivers**: Install NVIDIA Drivers and CUDA Toolkit.
- **Ollama**: Install [Ollama](https://ollama.ai/):
  ```bash
  curl -fsSL https://ollama.com/install.sh | sh
  ```

### B. Model Installation
Pull the Gemma 2 27B model:
```bash
ollama pull gemma4:31b
```

### C. Server Configuration
1. Copy the environment template:
   ```bash
   cd server
   cp .env.example .env
   ```

2. Edit `.env` to configure your setup:
   ```bash
   JARVIS_PORT=:8080
   OLLAMA_ENDPOINT=http://localhost:11434
   OLLAMA_MODEL=gemma4:31b
   ```

### D. Server Deployment

**Development Mode:**
```bash
cd server
./run.sh
```

**Production Mode (systemd service):**
```bash
sudo cp server/jarvis-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable jarvis-server.service
sudo systemctl start jarvis-server.service
```

---

## 2. Thin Client Setup (Linux Mint)

### A. System Dependencies
Install audio headers, FFmpeg, and the MP3 player needed for TTS:
```bash
sudo apt-get update && sudo apt-get install -y \
    portaudio19-dev \
    libasound2-dev \
    ffmpeg \
    mpg123 \
    python3-pip \
    python3-venv
```

### B. Python Environment
The setup script creates this automatically, or manually:
```bash
cd client
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### C. Configuration
1. Copy the environment template:
   ```bash
   cd client
   cp .env.example .env
   ```

2. Edit `.env` with your server IP:
   ```bash
   JARVIS_SERVER_URL=ws://192.168.1.100:8080/ws
   WAKE_WORD=jarvis
   STT_MODEL_SIZE=base
   RECORDING_DURATION=5
   RECONNECT_DELAY=5
   ```

### D. Run Development Mode
```bash
cd client
./run.sh
```

---

## 3. Persistent Mode (Auto-Start on Mint)

To have Jarvis start automatically on boot as a background daemon:

1. Move the project folder to `/home/jarvisuser/MyEncyclopedia`.
2. Update the `.env` file with production settings.
3. Copy the service file:
   ```bash
   sudo cp client/jarvis.service /etc/systemd/system/jarvis.service
   ```
4. Enable and start the service:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable jarvis.service
   sudo systemctl start jarvis.service
   ```

5. Check status:
   ```bash
   sudo systemctl status jarvis.service
   ```

---

## 4. Architecture Summary

- **Client (Mint):** Wake Word → Fast-Whisper STT → WebSocket → Edge-TTS Audio
- **Server (Ubuntu):** Go Router → Ollama (`gemma4:31b`) → Token Streaming
- **Automation:** Server generates a daily briefing at 6 AM; pushed to client upon first connection
- **Error Handling:** Automatic reconnection on network failures
- **Configuration:** Environment-based config for easy deployment

---

## 5. Health Checks

### Server Health
```bash
curl http://<SERVER_IP>:8080/health
```

### Check Logs
**Server:**
```bash
sudo journalctl -u jarvis-server.service -f
```

**Client:**
```bash
sudo journalctl -u jarvis.service -f
```

---

## 6. Troubleshooting

### Client can't connect to server
- Verify server IP in `client/.env`
- Check firewall: `sudo ufw status`
- Test connectivity: `telnet <SERVER_IP> 8080`

### Wake word not detected
- Check microphone permissions
- Adjust wake word sensitivity in code if needed
- Verify microphone input: `arecord -l`

### Ollama errors
- Ensure Ollama is running: `systemctl status ollama`
- Check model is loaded: `ollama list`
- Verify GPU memory: `nvidia-smi`

### Audio playback issues
- Test mpg123: `mpg123 --test`
- Check audio output: `speaker-test -c 2`

