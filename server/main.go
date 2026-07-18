package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Type    string `json:"type"` // "prompt", "response_token", "system_info"
	Content string `json:"content"`
}

type Server struct {
	ollamaClient *OllamaClient
	clients      map[*websocket.Conn]bool
	mu           sync.Mutex
}

func NewServer(ollamaEndpoint, ollamaModel string) *Server {
	return &Server{
		ollamaClient: NewOllamaClient(ollamaEndpoint, ollamaModel),
		clients:      make(map[*websocket.Conn]bool),
	}
}

func main() {
	// Configuration constants
	const (
		Port           = ":8080"
		OllamaEndpoint = "http://localhost:11434"
		OllamaModel    = "gemma2:27b"
	)

	server := NewServer(OllamaEndpoint, OllamaModel)
	scheduler := NewScheduler()
	
	// Start the background cron job
	scheduler.StartDailyCron(server.ollamaClient)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// We wrap the handler to inject the scheduler's briefing if needed
		server.handleWebSocketWithBriefing(w, r, scheduler)
	})

	log.Printf("Jarvis Compute Node starting on %s...", Port)
	if err := http.ListenAndServe(Port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
