# Changelog - Project Jarvis Improvements

## 2026-07-18 - Major Bug Fixes and Improvements

### 🔴 Critical Bug Fixes

#### Server (Go)
- **Fixed infinite loop in `server_handlers.go`**: Added labeled break (`streamLoop`) to properly exit token streaming loop when channel closes
- **Fixed race condition in scheduler**: Added mutex locks to `lastBriefing` access
- **Added connection timeout handling**: WebSocket connections now have read/write deadlines to prevent hung connections

#### Client (Python)
- **Fixed OpenWakeWord integration**: Changed from accessing float value to dict lookup (`prediction.get(WAKE_WORD, 0.0)`)
- **Fixed event loop hang**: Client now properly returns to wake word listening after processing each prompt
- **Fixed WebSocket message consumption**: Response streaming now completes before returning to prompt loop
- **Fixed blocking TTS**: Changed from `os.system()` to `subprocess.Popen()` for non-blocking audio playback

### ✨ New Features

#### Configuration Management
- **Environment-based configuration**: Both client and server now use `.env` files
- **`.env.example` templates**: Easy configuration with sensible defaults
- **Runtime configuration**: Server IP, wake word, model size all configurable without code changes

#### Error Handling & Resilience
- **Automatic reconnection**: Client automatically reconnects on WebSocket failures
- **Graceful error recovery**: Client returns to wake word listening on errors instead of crashing
- **Connection health monitoring**: Server tracks active client count and logs connection lifecycle
- **Improved logging**: Emoji-prefixed logs for better readability (🚀 ✅ ❌ 💬 etc.)

#### Developer Experience
- **Quick setup script**: `./setup.sh` automates entire setup process
- **Run scripts**: Simple `./run.sh` scripts for both client and server
- **Health check endpoint**: `/health` endpoint for monitoring server status
- **Better service files**: Improved systemd service configurations

### 🔧 Improvements

#### Server
- Added connection timeouts to prevent resource leaks
- Better error messages to clients
- Proper client cleanup on disconnect
- Thread-safe scheduler with read/write mutex
- Dynamic date in daily briefing (no longer hardcoded)
- Configuration via environment variables

#### Client
- Sentence-based TTS for better flow
- Non-blocking audio playback
- Timeout handling for server responses
- Improved user feedback with status messages
- Audio stream recovery on errors
- Resource cleanup on shutdown

#### Documentation
- Updated README with quick start guide
- Comprehensive DEPLOYMENT.md with troubleshooting
- Added dependency installation instructions
- Included systemd service templates

### 🐛 Minor Fixes
- Fixed typo in `jarvis.service` WorkingDirectory path
- Added `python-dotenv` to requirements.txt
- Proper virtual environment detection in run scripts

### 📝 Code Quality
- Removed commented-out code
- Improved error messages
- Better code organization
- Consistent logging format

---

## Migration Guide

### For Existing Installations

1. **Backup your current setup**
2. **Pull latest changes**: `git pull`
3. **Run setup**: `./setup.sh`
4. **Create `.env` files**:
   ```bash
   # Server
   cd server
   cp .env.example .env
   # Edit .env with your settings
   
   # Client
   cd ../client
   cp .env.example .env
   # Edit .env with your server IP
   ```
5. **Test the system**:
   ```bash
   # Terminal 1
   cd server && ./run.sh
   
   # Terminal 2
   cd client && ./run.sh
   ```

### Breaking Changes
- Client no longer requires editing `main.py` for server IP (use `.env` instead)
- OpenWakeWord model initialization uses array syntax: `wakeword_models=[WAKE_WORD]`

---

## Known Limitations
- Single client support (multiple clients connect but may conflict)
- Fixed 5-second recording duration (no dynamic VAD)
- No LangChain tools integrated yet (Wikipedia, Math, News)
- No vision/webcam processing yet

## Future Improvements
- Add proper Voice Activity Detection (VAD)
- Implement multi-client support with session management
- Add LangChain tool integration
- Add vision model support
- Create web-based monitoring dashboard
- Add unit and integration tests
