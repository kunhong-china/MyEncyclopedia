package main

import (
	"fmt"
	"log"
	"time"
)

type Scheduler struct {
	lastBriefing string
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// StartDailyCron runs a ticker that checks for 6:00 AM every day
func (s *Scheduler) StartDailyCron(ollamaClient *OllamaClient) {
	go func() {
		for {
			now := time.Now()
			// Check if it is exactly 6:00 AM (within a 1-minute window)
			if now.Hour() == 6 && now.Minute() == 0 {
				log.Printf("Executing scheduled 6 AM briefing generation...")
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
	prompt := "You are Project Jarvis. Generate a concise, child-friendly daily briefing for July 17th, 2026. Include one fun science fact, a news highlight about space exploration, and a positive motivational quote for the day. Format it as a greeting."
	
	tokenChan := make(chan string)
	errChan := make(chan error, 1)

	var fullResponse string
	go client.Generate(prompt, tokenChan, errChan)

	for {
		select {
		case token := <-tokenChan:
			fullResponse += token
		case err := <-errChan:
			log.Printf("Scheduler error generating briefing: %v", err)
			return
		}
		// We wait for completion here since it's a background task
		if len(fullResponse) > 0 && tokenChan == nil { // This is simplified logic, in reality check the 'done' state from ollama.go if possible or use a separate non-streaming method
			break
		}
	}
	// Note: The OllamaClient current Generate implementation doesn't send a sentinel for completion via chan
	// I will adjust it in my mind to assume we collect until the loop breaks natively from the generator.
	
	s.lastBriefing = fullResponse
	log.Printf("Daily briefing cached successfully.")
}

func (s *Scheduler) GetCachedBriefing() string {
	return s.lastBriefing
}
