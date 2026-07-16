# PROJECT JARVIS - TEST PLAN & VERIFICATION

This document outlines the strategy for verifying that Project Jarvis is functioning correctly across both nodes.

## 1. Component-Level Testing (Unit Tests)

### server/ollama.go
- **Test**: Verify that `Generate` sends a valid JSON request to the Ollama API and receives tokens.
- **Method**: Mock the Ollama endpoint using a simple HTTP server that returns streaming JSON fragments.
- **Success Criteria**: The `tokenChan` receives all expected fragments in order.

### server/scheduler.go
- **Test**: Verify that the 6 AM cron triggers correctly.
- **Method**: Temporarily change the trigger time to "current time + 1 minute" and monitor logs for "Executing scheduled 6 AM briefing generation...".
- **Success Criteria**: A daily briefing is generated and stored in `lastBriefing`.

### client/main.py (STT & TTS)
- **Test**: Verify local STT transcription accuracy.
- **Method**: Feed a known `.wav` file into the Whisper model instead of live microphone input.
- **Success Criteria**: Transcribed text matches the source audio with minimal error.

---

## 2. Integration Testing (The "Happy Path")

### Scenario: Standard Interaction Loop
1. **Setup**: Server is running on Ubuntu; Client is running on Linux Mint.
2. **Trigger**: User says "Jarvis".
3. **Expected behavior**:
   - Client log prints: `Wake word detected! Listening for prompt...`
   - User speaks a question (e.g., "What is the distance to Mars?").
   - Client transcribes text $\rightarrow$ sends via WebSocket to Server.
   - Server routes request $\rightarrow$ Ollama generates tokens $\rightarrow$ streams tokens back.
   - Client receives tokens $\rightarrow$ buffers sentence $\rightarrow$ speaks response via Edge-TTS.
4. **Success Criteria**: Response is heard by the user within < 2 seconds of finishing the prompt.

### Scenario: The Morning Briefing
1. **Setup**: Server has already run its 6 AM cron job (or was manually triggered).
2. **Action**: Restart the client or initiate a new connection.
3. **Expected behavior**:
   - Upon WebSocket connection, the server immediately pushes the cached briefing.
   - Client speaks: "Good morning! Your daily briefing is ready..." followed by the content.
4. **Success Criteria**: The briefing is heard without the user needing to say anything first.

---

## 3. Stability & Edge Case Testing

### Scenario: Network Interruption
1. **Action**: Disconnect the network cable or disable WiFi during a response stream.
2. **Expected behavior**: Client should log a connection error and gracefully return to Wake Word listening mode rather than crashing.
3. **Success Criteria**: System recovers automatically when network is restored.

### Scenario: High Noise Environment
1. **Action**: Play background music/noise while speaking the wake word.
2. **Expected behavior**: Verify if `openwakeword` triggers correctly or if too many false positives occur.
3. **Success Criteria**: User can tune sensitivity (if added) or confirm that Wake Word is robust enough for a home environment.

### Scenario: LLM Timeout/Crash
1. **Action**: Stop the Ollama service while a request is pending.
2. **Expected behavior**: Server should catch the HTTP error and send a `system_info` message to the client (e.g., "Error: failed to connect to Ollama").
3. **Success Criteria**: User receives an audio notification that the brain is offline instead of silence.

## 4. Performance Metrics (SLAs)
- **TTFT (Time to First Token)**: < 800ms from the moment prompt is sent over WebSocket.
- **STT Latency**: < 500ms for short prompts (< 10 words).
- **TTS Playback**: Immediate start upon receiving first sentence boundary.