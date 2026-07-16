// Wrap handleWebSocket to include briefing logic
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

	// Push the cached briefing immediately on connection if it exists
	briefing := scheduler.GetCachedBriefing()
	if briefing != "" {
		msg := Message{
			Type:    "response_token",
			Content: "Good morning! Your daily briefing is ready:\n\n" + briefing,
		}
		conn.WriteJSON(msg)
	}

	// Continue with standard prompt loop
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Read error from %s: %v", r.RemoteAddr, err)
			break
		}

		if msg.Type == "prompt" {
			tokenChan := make(chan string)
			errChan := make(chan error, 1)
			go s.ollamaClient.Generate(msg.Content, tokenChan, errChan)
			for {
				select {
				case token := <-tokenChan:
					respMsg := Message{Type: "response_token", Content: token}
					if err := conn.WriteJSON(respMsg); err != nil {
						return
					}
				case err := <-errChan:
					errMsg := Message{Type: "system_info", Content: fmt.Sprintf("Error: %v", err)}
					conn.WriteJSON(errMsg)
					break 
				}
			}
		}
	}

	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()
}