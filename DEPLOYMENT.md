# ==============================================================================
# PROJECT JARVIS - DEPLOYMENT & SETUP GUIDE
# ==============================================================================

This guide explains how to deploy the Project Jarvis AI Knowledge Agent across 
your network.

## 1. Compute Node Setup (Ubuntu Server + RTX 4090)

### A. Prerequisites
- **Network**: Ensure the Ubuntu server has a **static IP address**. This is critical as the client relies on it for WebSocket communication.
- **Firewall**: Open port `8080` to allow incoming traffic from the local network:
  `sudo ufw allow 8080/tcp`
- **GPU Drivers**: Install NVIDIA Drivers and CUDA Toolkit.
- **Ollama**: Install [Ollama](https://ollama.ai/):
  `curl -fsSL https://ollama.com/install.sh | sh`

### B. Model Installation
Pull the Gemma 2 27B model:
  `ollama pull gemma2:27b`

### C. Server Deployment
1. Clone this repository to the server.
2. Install Go (1.21+).
3. Initialize and run:
   ```bash
   cd server
   go mod tidy
   go run .
   ```

---

## 2. Thin Client Setup (Linux Mint)

### A. System Dependencies
Install audio headers, FFmpeg, and the MP3 player needed for TTS:
  `sudo apt-get update && sudo apt-get install -y portaudio19-dev libasound2-dev ffmpeg mpg123 python3-pip`

### B. Python Environment
It is recommended to use a virtual environment:
  ```bash
  cd client
  python3 -m venv venv
  source venv/bin/activate
  pip install -r requirements.txt
  ```

### C. Configuration
Edit `client/main.py` and replace `<UBUNTU_SERVER_IP>` with the actual IP address of your Ubuntu server.

### D. Run Manually
  `python3 main.py`

---

## 3. Persistent Mode (Auto-Start on Mint)

To have Jarvis start automatically on boot as a background daemon:

1. Move the project folder to `/home/jarvisuser/MyEncyclopedia`.
2. Copy the service file:
   `sudo cp client/jarvis.service /etc/systemd/system/jarvis.service`
3. Enable and start the service:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable jarvis.service
   sudo systemctl start jarvis.service
   ```

## 4. Architecture Summary
- **Client (Mint):** Wake Word $\rightarrow$ Fast-Whisper STT $\rightarrow$ WebSocket $\rightarrow$ Edge-TTS Audio.
- **Server (Ubuntu):** Go Router $\rightarrow$ Ollama (`gemma2:27b`) $\rightarrow$ Token Streaming.
- **Automation:** Server generates a daily briefing at 6 AM; pushed to client upon first connection.
