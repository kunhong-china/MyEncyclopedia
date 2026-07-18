package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// handleWebSocketWithBriefing manages WebSocket connections with briefing injection.
func (s *Server) handleWebSocketWithBriefing(w http.ResponseWriter, r *http.Request, scheduler *Scheduler) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Upgrade error: %v", err)
		return
	}
	defer conn.Close()

	clientAddr := r.RemoteAddr
	log.Printf("✅ Client connected: %s", clientAddr)

	s.mu.Lock()
	s.clients[conn] = true
	clientCount := len(s.clients)
	s.mu.Unlock()
	
	log.Printf("📊 Active clients: %d", clientCount)

	// Set read/write deadlines to prevent hung connections
	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
	conn.SetWriteDeadline(time.Now().Add(30 * time.Second))

	// Push cached briefing immediately on connection if it exists
	briefing := scheduler.GetCachedBriefing()
	if briefing != "" {
		msg := Message{
			Type:    "response_token",
			Content: "Good morning! Your daily briefing is ready:\n\n" + briefing,
		}
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("❌ Error sending briefing to %s: %v", clientAddr, err)
		}
	}

	// Main prompt handling loop
	for {
		// Reset read deadline for each message
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
		
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("📴 Client %s disconnected: %v", clientAddr, err)
			break
		}

		if msg.Type == "prompt" {
			log.Printf("💬 [%s] Prompt: %s", clientAddr, msg.Content)
			
			tokenChan := make(chan string)
			errChan := make(chan error, 1)

			go s.ollamaClient.Generate(msg.Content, tokenChan, errChan)

		streamLoop:
			for {
				select {
				case token, ok := <-tokenChan:
					if !ok {
						log.Printf("✅ [%s] Response complete", clientAddr)
						break streamLoop
					}
					
					respMsg := Message{Type: "response_token", Content: token}
					conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
					
					if err := conn.WriteJSON(respMsg); err != nil {
						log.Printf("❌ Write error to %s: %v", clientAddr, err)
						s.cleanupClient(conn)
						return
					}
					
				case err := <-errChan:
					log.Printf("⚠️  [%s] Ollama error: %v", clientAddr, err)
					errMsg := Message{
						Type:    "system_info",
						Content: fmt.Sprintf("I encountered an error: %v. Please try again.", err),
					}
					conn.WriteJSON(errMsg)
					break streamLoop
				}
			}
		}
	}

	s.cleanupClient(conn)
}

func (s *Server) cleanupClient(conn *websocket.Conn) {
	s.mu.Lock()
	delete(s.clients, conn)
	clientCount := len(s.clients)
	s.mu.Unlock()
	
	log.Printf("📊 Active clients: %d", clientCount)
}
