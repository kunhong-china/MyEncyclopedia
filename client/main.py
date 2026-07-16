import asyncio
import json
import sys
import os
from websockets import connect
from openwakeword.model import Model
import whisper
import edge_tts
import pyaudio
import wave
import numpy as np

# Configuration
SERVER_URL = "ws://<UBUNTU_SERVER_IP>:8080/ws"
WAKE_WORD = "jarvis"
STT_MODEL_SIZE = "base" # Options: tiny, base, small...

class JarvisClient:
    def __init__(self):
        # Initialize Wake Word Model
        self.oww_model = Model(wakeword_model=WAKE_WORD)
        
        # Initialize Whisper STT
        self.stt_model = whisper.load_model(STT_MODEL_SIZE)
        
        # Audio Setup
        self.p = pyaudio.PyAudio()
        self.stream = self.p.open(format=pyaudio.paInt16,
                                channels=1,
                                rate=16000,
                                input=True,
                                frames_per_buffer=1280)

    async def text_to_speech(self, text):
        """Convert text to audio using edge-tts and play it immediately."""
        communicate = edge_tts.Communicate(text, "en-US-GuyNeural")
        await communicate.save("output.mp3")
        # On Linux Mint, we can use a simple system call to play the mp3
        os.system("mpg123 -q output.mp3")

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
                # Read audio chunk from microphone
                data = self.stream.read(1280, exception_on_overflow=False)
                audio_data = np.frombuffer(data, dtype=np.int16).astype(np.float32) / 32768.0
                
                # Feed into OpenWakeWord
                prediction = self.oww_model.predict(audio_data)
                
                if prediction > 0.5 and not listening_for_prompt:
                    print("Wake word detected! Listening for prompt...")
                    listening_for_prompt = True
                    audio_buffer = [] # Reset buffer for STT session
                
                if listening_for_prompt:
                    audio_buffer.extend(audio_data)
                    
                    # Simple VAD (Voice Activity Detection) via silence gap or time limit
                    # For this implementation, we record for 4 seconds or until a pause is detected
                    if len(audio_buffer) > 16000 * 5: 
                        listening_for_prompt = False
                        await self.process_stt_and_send(audio_buffer, websocket)

    async def process_stt_and_send(self, audio_buffer, websocket):
        print("Processing speech...")
        # Save buffer to temporary file for Whisper
        with wave.open("temp_prompt.wav", "wb") as f:
            f.setnchannels(1)
            f.setsampwidth(2)
            f.setframerate(16000)
            # Convert float32 back to int16 for wav file
            f.writeframes((np.array(audio_buffer) * 32767).astype(np.int16).tobytes())

        # Transcribe using Fast-Whisper / Whisper
        result = self.stt_model.transcribe("temp_prompt.wav")
        text = result["text"].strip()
        print(f"You said: {text}")

        if text:
            payload = {"type": "prompt", "content": text}
            await websocket.send(json.dumps(payload))
            
            # Handle streaming response from server
            full_response = ""
            async for message in websocket:
                data = json.loads(message)
                if data['type'] == 'response_token':
                    token = data['content']
                    print(token, end="", flush=True)
                    full_response += token
                elif data['type'] == 'system_info':
                    print(f"\nSystem: {data['content']}")
                    break
                
                # In a real production scenario, you'd want to stream audio tokens.
                # For this version, we buffer the sentence and then speak it.
                if "." in token or "?" in token or "!" in token:
                    await self.text_to_speech(full_response)
                    full_response = ""
            print("\nDone.")

async def main():
    client = JarvisClient()
    await client.listen_and_process()

if __name__ == "__main__":
    asyncio.run(main())
