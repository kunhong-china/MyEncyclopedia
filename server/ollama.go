package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type OllamaClient struct {
	Endpoint string
	Model    string
}

func NewOllamaClient(endpoint, model string) *OllamaClient {
	return &OllamaClient{
		Endpoint: endpoint,
		Model:    model,
	}
}

// Generate implements streaming response from Ollama with proper channel lifecycle management.
// The output channel is always closed on function exit to signal completion to receivers.
func (c *OllamaClient) Generate(prompt string, tokenChan chan<- string, errChan chan<- error) {
	defer close(tokenChan) // Ensure channel is closed when generator completes

	reqBody := OllamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: true,
	}

	jsonBody, _ := json.Marshal(reqBody)
	resp, err := http.Post(c.Endpoint+"/api/generate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		errChan <- fmt.Errorf("failed to connect to Ollama: %v", err)
		return // Channel still closes via defer
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	for {
		var response OllamaResponse
		if err := decoder.Decode(&response); err != nil {
			if err == io.EOF {
				break // Normal stream end, channel will close via defer
			}
			errChan <- fmt.Errorf("error decoding Ollama stream: %v", err)
			return // Channel still closes via defer
		}

		tokenChan <- response.Response
		if response.Done {
			break // Ollama signals stream complete, defer closes channel
		}
	}
}