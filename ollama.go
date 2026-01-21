package main

// Integration with the Ollama local AI API. It allows the editor to check AI
// availability and generate text completions.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nsf/termbox-go"
)

// OllamaClient handles HTTP requests to the Ollama server.
type OllamaClient struct {
	IsOnline bool   // Current availability of the local LLM.
	URL      string // Base API endpoint for status checks.
}

// GenerateRequest defines the payload for text generation.
type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// GenerateResponse defines the server's reply for text generation.
type GenerateResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// NewOllamaClient initializes the client with the configured URL.
func NewOllamaClient() *OllamaClient {
	return &OllamaClient{
		IsOnline: false,
		URL:      fmt.Sprintf("%s/api/tags", Config.OllamaURL),
	}
}

// PeriodicStatusCheck starts a background goroutine to monitor AI availability.
func (c *OllamaClient) PeriodicStatusCheck() {
	go func() {
		for {
			prevStatus := c.IsOnline
			currentStatus := c.CheckStatus()
			// If status changes, trigger a UI update to refresh the AI indicator.
			if prevStatus != currentStatus {
				termbox.Interrupt()
			}
			time.Sleep(Config.OllamaCheckInterval)
		}
	}()
}

// CheckStatus pings the Ollama server to see if it responds.
func (c *OllamaClient) CheckStatus() bool {
	client := http.Client{
		Timeout: 1 * time.Second, // Fail fast if server is down.
	}

	resp, err := client.Get(c.URL)
	if err != nil {
		c.IsOnline = false
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		c.IsOnline = true
	} else {
		c.IsOnline = false
	}

	return c.IsOnline
}

// Generate sends a prompt to the LLM and returns the generated text.
func (c *OllamaClient) Generate(prompt string) (string, error) {
	url := fmt.Sprintf("%s/api/generate", Config.OllamaURL)
	reqBody := GenerateRequest{
		Model:  Config.OllamaModel,
		Prompt: prompt,
		Stream: false, // We want the full result at once for simplified handling.
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var genResp GenerateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return "", err
	}

	return genResp.Response, nil
}
