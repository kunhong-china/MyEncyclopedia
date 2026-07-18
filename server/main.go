package main

import (
	"log"
	"net/http"
	"os"
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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Configuration with environment variable support
	port := getEnv("JARVIS_PORT", ":8080")
	ollamaEndpoint := getEnv("OLLAMA_ENDPOINT", "http://localhost:11434")
	ollamaModel := getEnv("OLLAMA_MODEL", "gemma2:27b")

	server := NewServer(ollamaEndpoint, ollamaModel)
	scheduler := NewScheduler()
	
	// Start the background cron job
	scheduler.StartDailyCron(server.ollamaClient)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.handleWebSocketWithBriefing(w, r, scheduler)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("🚀 Jarvis Compute Node starting on %s...", port)
	log.Printf("📡 Ollama endpoint: %s (model: %s)", ollamaEndpoint, ollamaModel)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
