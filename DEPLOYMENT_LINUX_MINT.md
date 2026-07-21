# Jarvis Deployment Guide for Linux Mint

This guide provides step-by-step instructions to deploy and run the Jarvis client application on Linux Mint hardware.

## Prerequisites

Before installing Jarvis, ensure your system meets these requirements:
- Linux Mint 20 or later
- Python 3.8 or higher
- Internet connection for package installation
- Audio input/output capabilities

## System Dependencies Installation

Install required system packages using the following commands:

```bash
# Update package list
sudo apt update

# Install audio development libraries and audio players
sudo apt install portaudio19-dev libasound-dev mpg123

# Install additional useful tools 
sudo apt install git python3-pip python3-venv
```

## Python Environment Setup

Create a virtual environment to isolate the project dependencies:

```bash
# Create virtual environment
python3 -m venv jarvis_env

# Activate the virtual environment
source jarvis_env/bin/activate

# Install Python requirements
pip install -r client/requirements.txt
```

## Hardware Setup

### Audio Configuration

The application automatically detects audio input devices. However, you can manually check your audio devices:

```bash
# List available audio devices
arecord -l
aplay -l

# Check current audio input/output
pactl list sources short
pactl list sinks short
```

## Environment Configuration

Create a `.env` file in the client directory to configure application settings:

```bash
cd client/
cp .env.example .env
```

Edit the `.env` file to set your configuration:
```
JARVIS_SERVER_URL=ws://<UBUNTU_SERVER_IP>:8080/ws
WAKE_WORD=jarvis
STT_MODEL_SIZE=base
RECORDING_DURATION=5
RECONNECT_DELAY=5
```

## Running Jarvis

### Start the Application

```bash
# Activate virtual environment
source jarvis_env/bin/activate

# Navigate to client directory and run
cd client/
python3 main.py
```

### Using systemd service (Optional)

To run Jarvis automatically on system boot:

1. Create a systemd service file:
```bash
sudo nano /etc/systemd/system/jarvis.service
```

2. Add the following content:
```ini
[Unit]
Description=Jarvis Client
After=network.target

[Service]
Type=simple
User=your-username
WorkingDirectory=/path/to/MyEncyclopedia/client
Environment=PYTHONPATH=/path/to/MyEncyclopedia/client
ExecStart=/path/to/MyEncyclopedia/client/jarvis_env/bin/python3 main.py
Restart=always

[Install]
WantedBy=multi-user.target
```

3. Start and enable the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable jarvis.service
sudo systemctl start jarvis.service
```

## Troubleshooting

### Audio Issues
If you encounter audio problems, try:

1. **Check audio permissions:**
```bash
# Add user to audio group
sudo usermod -a -G audio $USER
```

2. **Verify audio device detection:**
```bash
# Check if microphone is working
arecord -d 5 test.wav
```

### Common Error Messages

1. **"Could not play audio"**: Install additional audio players:
```bash
sudo apt install alsa-utils pulseaudio-utils libav-tools
```

2. **Python import errors**: Reinstall requirements in virtual environment:
```bash
pip install -r requirements.txt
```

## Linux Mint-Specific Notes

### Version Compatibility
This implementation has been tested with:
- Linux Mint 20.x (Ubuntu 20.04 base)
- Linux Mint 21.x (Ubuntu 22.04 base)

### Audio Driver Considerations
Linux Mint may use different audio drivers that require additional configuration:

```bash
# Check audio subsystem status
pulseaudio --version
alsamixer

# If needed, configure default audio device
echo 'export AUDIODEV=hw:0,0' >> ~/.bashrc
```

## Testing Your Installation

After setup, test the application:
1. Run `python3 main.py`
2. Say "jarvis" to activate the wake word detection
3. The system should start listening and process your voice commands
4. Verify text-to-speech works with sample responses

## Support

For additional support or troubleshooting assistance, contact:
- Your development team or documentation maintainer
- Linux Mint community forums