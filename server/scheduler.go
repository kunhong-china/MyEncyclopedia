package main

import (
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
	go client.Generate(prompt, tokenChan, errChan) // Streaming generator - channel closes when complete

loop:
	for {
		select {
		case token, ok := <-tokenChan:
			if !ok {
				break loop // Channel closed by Generate - exit outer loop
			}
			fullResponse += token
		case err := <-errChan:
			log.Printf("Scheduler error generating briefing: %v", err)
			return
		}
	}

	s.lastBriefing = fullResponse
	log.Printf("Daily briefing cached successfully.")
}

func (s *Scheduler) GetCachedBriefing() string {
	return s.lastBriefing
}
