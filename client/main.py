import asyncio
import json
import sys
import os
import subprocess
from websockets import connect
from websockets.exceptions import WebSocketException
from openwakeword.model import Model
import whisper
import edge_tts
import pyaudio
import wave
import numpy as np
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Configuration
SERVER_URL = os.getenv("JARVIS_SERVER_URL", "ws://localhost:8080/ws")
WAKE_WORD = os.getenv("WAKE_WORD", "jarvis")
STT_MODEL_SIZE = os.getenv("STT_MODEL_SIZE", "base")
RECORDING_DURATION = int(os.getenv("RECORDING_DURATION", "5"))
RECONNECT_DELAY = int(os.getenv("RECONNECT_DELAY", "5"))

class JarvisClient:
    def __init__(self):
        print("Initializing Jarvis Client...")
        
        # Initialize Wake Word Model
        self.oww_model = Model(wakeword_models=[WAKE_WORD])
        
        # Initialize Whisper STT
        print(f"Loading Whisper {STT_MODEL_SIZE} model...")
        self.stt_model = whisper.load_model(STT_MODEL_SIZE)
        
        # Audio Setup - Improved for Linux Mint compatibility
        try:
            self.p = pyaudio.PyAudio()
            # Try to find a working audio input device
            input_device_index = None
            for i in range(self.p.get_device_count()):
                dev_info = self.p.get_device_info_by_index(i)
                if (dev_info['maxInputChannels'] > 0 and 
                    'audio' in dev_info['name'].lower() or 
                    'input' in dev_info['name'].lower()):
                    input_device_index = i
                    break
            
            # If no specific input device found, use default (index 0)
            if input_device_index is None:
                input_device_index = 0
                
            self.stream = self.p.open(format=pyaudio.paInt16,
                                    channels=1,
                                    rate=16000,
                                    input=True,
                                    frames_per_buffer=1280,
                                    input_device_index=input_device_index)
            print(f"Using audio input device: {self.p.get_device_info_by_index(input_device_index)['name']}")
            
        except Exception as e:
            print(f"Error setting up audio: {e}")
            raise

    async def text_to_speech(self, text):
        """Convert text to audio using edge-tts and play it in background."""
        if not text.strip():
            return
            
        try:
            communicate = edge_tts.Communicate(text, "en-US-GuyNeural")
            await communicate.save("output.mp3")
            
            # On Linux Mint, we can use different audio players with fallbacks
            player_commands = [
                "mpg123 -q output.mp3",
                "aplay output.mp3",  # Alternative for WAV files, if needed
                "paplay output.mp3 --device=alsa_output.pci-0000_00_1f.3.analog-stereo"  # PulseAudio with device spec
            ]
            
            # Try each player until one works
            played = False
            for cmd in player_commands:
                try:
                    result = os.system(f"{cmd} 2>/dev/null")
                    if result == 0:
                        played = True
                        break
                except Exception as e:
                    print(f"Failed to play with {cmd}: {e}")
                    continue
            
            if not played:
                print("Warning: Could not play audio. Please ensure mpg123 or another audio player is installed.")
                
        except Exception as e:
            print(f"Error in text-to-speech conversion: {e}")

    async def listen_and_process(self):
        print(f"Jarvis is listening... (Wake word: {WAKE_WORD})")
        
        # Buffer for audio data to pass to Whisper
        audio_buffer = []
        listening_for_prompt = False
        
        async with connect(SERVER_URL) as websocket:
            # Handle the initial greeting/briefing from server if any
            try:
                greeting = await asyncio.wait_for(websocket.recv(), timeout=2.0)
                data = json.loads(greeting)
                if data['type'] == 'response_token':
                    print(f"Server: {data['content']}")
                    await self.text_to_speech(data['content'])
            except asyncio.TimeoutError:
                pass

            while True:
                try:
                    # Listen for wake word
                    wake_word_detected = False
                    
                    # Get audio data from microphone
                    audio_data = self.stream.read(1024)
                    audio_buffer.append(audio_data)
                    
                    # Convert to numpy array (necessary for openwakeword or speech processing)
                    audio_np = np.frombuffer(audio_data, dtype=np.int16)
                    
                    # Wake word detection
                    if not listening_for_prompt:
                        wake_word_detected = self.oww_model.predict(audio_np)
                        if wake_word_detected:
                            print("Wake word detected!")
                            listening_for_prompt = True
                            audio_buffer = []  # Reset audio buffer for prompt recording
                            continue
                    
                    # Record prompt when wake word is detected
                    if listening_for_prompt:
                        # Here we would normally process the recorded audio
                        # For simplicity, showing the connection workflow but actual prompt processing logic would go here
                        
                        # Send the audio (as text string for demo purposes)
                        try:
                            # Simulate sending a message to server
                            await websocket.send(json.dumps({
                                "type": "prompt",
                                "content": "Processing your request..."
                            }))
                            
                            # Receive response from server
                            while True:
                                try:
                                    response = await asyncio.wait_for(websocket.recv(), timeout=10.0)
                                    data = json.loads(response)
                                    if data['type'] == 'response':
                                        print(f"Server: {data['content']}")
                                        await self.text_to_speech(data['content'])
                                        # Stop listening after response
                                        listening_for_prompt = False
                                        break
                                    elif data['type'] == 'response_token':
                                        print(f"System: {data['content']}")
                                        break
                                except asyncio.TimeoutError:
                                    break
                        except Exception as e:
                            print(f"Error in communication with server: {e}")
                            listening_for_prompt = False
                            
                        # Reset audio buffer for next prompt
                        audio_buffer = []
                        
                except KeyboardInterrupt:
                    print("\nJarvis shutdown requested...")
                    break
                except WebSocketException as e:
                    print(f"WebSocket error: {e}. Reconnecting in {RECONNECT_DELAY}s...")
                    await asyncio.sleep(RECONNECT_DELAY)
                    # Note: This would need proper reconnect logic in a real implementation
                    
            print("\nDone.")

async def main():
    try:
        client = JarvisClient()
        await client.listen_and_process()
    except KeyboardInterrupt:
        print("\nJarvis shutdown requested...")
    except Exception as e:
        print(f"Fatal error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    asyncio.run(main())