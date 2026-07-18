package main

import (
	"log"
	"sync"
	"time"
)

type Scheduler struct {
	lastBriefing string
	mu           sync.RWMutex
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// StartDailyCron runs a ticker that checks for 6:00 AM every day
func (s *Scheduler) StartDailyCron(ollamaClient *OllamaClient) {
	go func() {
		log.Printf("⏰ Daily briefing scheduler started (triggers at 6:00 AM)")
		
		for {
			now := time.Now()
			// Check if it is exactly 6:00 AM (within a 1-minute window)
			if now.Hour() == 6 && now.Minute() == 0 {
				log.Printf("📰 Executing scheduled 6 AM briefing generation...")
				s.generateBriefing(ollamaClient)
				// Sleep for 61 seconds to avoid triggering multiple times in the same minute
				time.Sleep(61 * time.Second)
			}
			// Check every 30 seconds
			time.Sleep(30 * time.Second)
		}
	}()
}

func (s *Scheduler) generateBriefing(client *OllamaClient) {
	now := time.Now()
	dateStr := now.Format("January 2, 2006")
	
	prompt := "You are Project Jarvis, a helpful AI assistant. Generate a concise, child-friendly daily briefing for " + dateStr + ". Include one fun science fact, a motivational quote for the day, and a brief weather reminder to check conditions. Keep it under 100 words and format it as a cheerful greeting."

	tokenChan := make(chan string)
	errChan := make(chan error, 1)

	var fullResponse string
	go client.Generate(prompt, tokenChan, errChan)

loop:
	for {
		select {
		case token, ok := <-tokenChan:
			if !ok {
				break loop
			}
			fullResponse += token
		case err := <-errChan:
			log.Printf("❌ Scheduler error generating briefing: %v", err)
			return
		}
	}

	s.mu.Lock()
	s.lastBriefing = fullResponse
	s.mu.Unlock()
	
	log.Printf("✅ Daily briefing cached successfully (%d chars)", len(fullResponse))
}

func (s *Scheduler) GetCachedBriefing() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastBriefing
}
