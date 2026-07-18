# Quick Reference Guide

## 🚀 Getting Started

### First Time Setup
```bash
# From project root
./setup.sh

# Configure server
nano server/.env

# Configure client
nano client/.env
```

### Running (Development)
```bash
# Terminal 1 - Server
cd server && ./run.sh

# Terminal 2 - Client  
cd client && ./run.sh
```

---

## 🔧 Configuration

### Server (.env)
```bash
JARVIS_PORT=:8080
OLLAMA_ENDPOINT=http://localhost:11434
OLLAMA_MODEL=gemma2:27b
```

### Client (.env)
```bash
JARVIS_SERVER_URL=ws://192.168.1.100:8080/ws
WAKE_WORD=jarvis
STT_MODEL_SIZE=base
RECORDING_DURATION=5
RECONNECT_DELAY=5
```

---

## 🩺 Health Checks

### Check Server
```bash
curl http://localhost:8080/health
```

### Check Ollama
```bash
ollama list
curl http://localhost:11434/api/tags
```

### View Logs
```bash
# Server
sudo journalctl -u jarvis-server.service -f

# Client
sudo journalctl -u jarvis.service -f
```

---

## 🐛 Troubleshooting

### Client Won't Connect
```bash
# Check server is running
curl http://<SERVER_IP>:8080/health

# Test network
telnet <SERVER_IP> 8080

# Check client config
cat client/.env | grep JARVIS_SERVER_URL
```

### Wake Word Not Working
```bash
# List audio devices
arecord -l

# Test microphone
arecord -d 5 test.wav && aplay test.wav

# Check permissions
groups | grep audio
```

### Ollama Not Responding
```bash
# Check status
systemctl status ollama

# Check GPU
nvidia-smi

# Test Ollama
ollama run gemma2:27b "Hello"
```

### Audio Not Playing
```bash
# Test speaker
speaker-test -c 2 -t wav

# Test mpg123
mpg123 --test

# Check volume
alsamixer
```

---

## 📁 Project Structure

```
MyEncyclopedia/
├── client/              # Python voice client
│   ├── main.py         # Main client logic
│   ├── .env            # Client config
│   └── run.sh          # Start script
├── server/              # Go compute server
│   ├── main.go         # Server entry point
│   ├── ollama.go       # Ollama integration
│   ├── scheduler.go    # Daily briefing
│   ├── .env            # Server config
│   └── run.sh          # Start script
└── setup.sh            # Setup automation
```

---

## 🎤 Using Jarvis

1. Say "**Jarvis**" (wake word)
2. Wait for "🎙️ Wake word detected!"
3. Ask your question (5 seconds to speak)
4. Listen to response
5. Repeat!

---

## 🔄 Common Commands

### Update Code
```bash
git pull
./setup.sh
```

### Restart Services
```bash
# Development
# Just Ctrl+C and restart ./run.sh

# Production
sudo systemctl restart jarvis-server.service
sudo systemctl restart jarvis.service
```

### Clean Up
```bash
# Remove temp files
rm -f client/*.wav client/*.mp3

# Clear Python cache
rm -rf client/__pycache__
```

---

## 📊 Monitoring

### Active Connections
Look for: `📊 Active clients: N` in server logs

### Response Times
Look for: `✅ [IP] Response complete` in server logs

### Errors
Look for: `❌` emoji in logs

---

## 🆘 Emergency

### Stop Everything
```bash
# Kill development processes
pkill -f "python main.py"
pkill -f "go run"

# Stop services
sudo systemctl stop jarvis.service
sudo systemctl stop jarvis-server.service
```

### Reset to Clean State
```bash
# Stop services
sudo systemctl stop jarvis.service jarvis-server.service

# Clean install
cd MyEncyclopedia
git reset --hard
git pull
./setup.sh

# Reconfigure and restart
```

---

## 📚 More Information

- Full setup: `DEPLOYMENT.md`
- Architecture: `README.md`
- Testing: `TEST_PLAN.md`
- Changes: `CHANGELOG.md`
- Improvements: `IMPROVEMENTS.md`
