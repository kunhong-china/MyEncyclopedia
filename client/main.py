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
        
        # Audio Setup
        self.p = pyaudio.PyAudio()
        self.stream = None
        self.init_audio_stream()
        
    def init_audio_stream(self):
        """Initialize or reinitialize audio stream."""
        try:
            if self.stream:
                self.stream.close()
            self.stream = self.p.open(
                format=pyaudio.paInt16,
                channels=1,
                rate=16000,
                input=True,
                frames_per_buffer=1280
            )
        except Exception as e:
            print(f"Error initializing audio: {e}")
            sys.exit(1)

    async def text_to_speech(self, text):
        """Convert text to audio using edge-tts and play it in background."""
        if not text.strip():
            return
            
        try:
            communicate = edge_tts.Communicate(text, "en-US-GuyNeural")
            await communicate.save("output.mp3")
            
            # Use subprocess.Popen for non-blocking playback
            subprocess.Popen(
                ["mpg123", "-q", "output.mp3"],
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL
            )
        except Exception as e:
            print(f"\nTTS Error: {e}")

    async def listen_and_process(self):
        """Main loop with automatic reconnection."""
        while True:
            try:
                await self._connect_and_listen()
            except KeyboardInterrupt:
                print("\nShutting down...")
                break
            except Exception as e:
                print(f"\nConnection error: {e}")
                print(f"Reconnecting in {RECONNECT_DELAY} seconds...")
                await asyncio.sleep(RECONNECT_DELAY)

    async def _connect_and_listen(self):
        """Connect to server and handle wake word detection."""
        print(f"Connecting to {SERVER_URL}...")
        
        async with connect(SERVER_URL) as websocket:
            print(f"✓ Connected! Listening for wake word '{WAKE_WORD}'...")
            
            # Handle initial briefing if sent by server
            await self._check_for_briefing(websocket)
            
            # Main wake word detection loop
            audio_buffer = []
            listening_for_prompt = False
            
            while True:
                try:
                    # Read audio chunk from microphone
                    data = self.stream.read(1280, exception_on_overflow=False)
                    audio_data = np.frombuffer(data, dtype=np.int16).astype(np.float32) / 32768.0
                    
                    # Feed into OpenWakeWord
                    prediction = self.oww_model.predict(audio_data)
                    
                    # OpenWakeWord returns dict: {model_name: score}
                    wake_score = prediction.get(WAKE_WORD, 0.0)
                    
                    if wake_score > 0.5 and not listening_for_prompt:
                        print("\n🎙️  Wake word detected! Listening for your question...")
                        listening_for_prompt = True
                        audio_buffer = []
                    
                    if listening_for_prompt:
                        audio_buffer.extend(audio_data)
                        
                        # Record for specified duration
                        if len(audio_buffer) >= 16000 * RECORDING_DURATION:
                            listening_for_prompt = False
                            await self.process_stt_and_send(audio_buffer, websocket)
                            audio_buffer = []
                            print(f"\nListening for wake word '{WAKE_WORD}'...")
                            
                except Exception as e:
                    print(f"\nAudio processing error: {e}")
                    self.init_audio_stream()

    async def _check_for_briefing(self, websocket):
        """Check if server sends an initial briefing."""
        try:
            greeting = await asyncio.wait_for(websocket.recv(), timeout=2.0)
            data = json.loads(greeting)
            if data['type'] == 'response_token':
                print(f"\n📢 {data['content']}\n")
                await self.text_to_speech(data['content'])
        except asyncio.TimeoutError:
            pass
        except Exception as e:
            print(f"Briefing error: {e}")

    async def process_stt_and_send(self, audio_buffer, websocket):
        """Transcribe audio and send to server, then handle response."""
        print("⏳ Processing speech...")
        
        try:
            # Save buffer to temporary file for Whisper
            wav_file = "temp_prompt.wav"
            with wave.open(wav_file, "wb") as f:
                f.setnchannels(1)
                f.setsampwidth(2)
                f.setframerate(16000)
                f.writeframes((np.array(audio_buffer) * 32767).astype(np.int16).tobytes())

            # Transcribe using Whisper
            result = self.stt_model.transcribe(wav_file)
            text = result["text"].strip()
            
            if not text:
                print("❌ No speech detected. Please try again.")
                return
                
            print(f"💭 You said: {text}")

            # Send prompt to server
            payload = {"type": "prompt", "content": text}
            await websocket.send(json.dumps(payload))
            
            # Handle streaming response
            await self._handle_response_stream(websocket)
            
        except Exception as e:
            print(f"STT/Send error: {e}")
    
    async def _handle_response_stream(self, websocket):
        """Handle streaming token response from server."""
        print("🤖 Jarvis: ", end="", flush=True)
        
        sentence_buffer = ""
        response_complete = False
        
        while not response_complete:
            try:
                message = await asyncio.wait_for(websocket.recv(), timeout=30.0)
                data = json.loads(message)
                
                if data['type'] == 'response_token':
                    token = data['content']
                    print(token, end="", flush=True)
                    sentence_buffer += token
                    
                    # Speak complete sentences for better flow
                    if any(punct in token for punct in ['. ', '! ', '? ', '\n']):
                        await self.text_to_speech(sentence_buffer)
                        sentence_buffer = ""
                        
                elif data['type'] == 'system_info':
                    print(f"\n⚠️  System: {data['content']}")
                    response_complete = True
                    
            except asyncio.TimeoutError:
                print("\n⏱️  Response timeout")
                response_complete = True
            except WebSocketException:
                print("\n❌ Connection lost during response")
                raise
                
        # Speak any remaining text
        if sentence_buffer.strip():
            await self.text_to_speech(sentence_buffer)
            
        print("\n")

    def cleanup(self):
        """Clean up resources."""
        if self.stream:
            self.stream.stop_stream()
            self.stream.close()
        self.p.terminate()

async def main():
    client = JarvisClient()
    try:
        await client.listen_and_process()
    finally:
        client.cleanup()

if __name__ == "__main__":
    asyncio.run(main())
