package main

import (
	"fmt"
	"log"
	"net/http"
)

// handleWebSocketWithBriefing manages WebSocket connections with briefing injection.
// It handles waking word -> STT already processed on client, then streams Ollama responses.
func (s *Server) handleWebSocketWithBriefing(w http.ResponseWriter, r *http.Request, scheduler *Scheduler) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}
	defer conn.Close()

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	// Push the cached briefing immediately on connection if it exists (6 AM cron-generated)
	briefing := scheduler.GetCachedBriefing()
	if briefing != "" {
		msg := Message{
			Type:    "response_token",
			Content: "Good morning! Your daily briefing is ready:\n\n" + briefing,
		}
		conn.WriteJSON(msg)
	}

	// Continue with standard prompt loop - streaming from Ollama via Go channels
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Read error from %s: %v", r.RemoteAddr, err)
			break
		}

		if msg.Type == "prompt" {
			tokenChan := make(chan string)   // Buffered for Ollama streaming tokens
			errChan := make(chan error, 1)    // Buffered to prevent deadlock on errors

			go s.ollamaClient.Generate(msg.Content, tokenChan, errChan) // Non-blocking streaming call

			// Collect and stream tokens to client until channel closes or error occurs
			for {
				select {
				case token, ok := <-tokenChan:
					if !ok {
						break // Channel closed by Generate - streaming complete, continue to next prompt loop iteration
					}
					respMsg := Message{Type: "response_token", Content: token}
					conn.WriteJSON(respMsg)
				case err, ok := <-errChan:
					if !ok {
						break // Error channel closed (no error occurred via normal streaming completion)
					}
					errMsg := Message{Type: "system_info", Content: fmt.Sprintf("Error: %v", err)}
					conn.WriteJSON(errMsg)
				}
			}
		}
	}

	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()
}