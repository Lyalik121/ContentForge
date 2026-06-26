package processor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type GeneratedPosts struct {
	Instagram string `json:"instagram"`
	TikTok    string `json:"tiktok"`
	Threads   string `json:"threads"`
	Telegram  string `json:"telegram"`
}

const geminiModel = "gemini-3.1-flash-lite"
const geminiURL = "https://generativelanguage.googleapis.com/v1beta/models/" + geminiModel + ":generateContent"

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}
type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}
type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func callGemini(prompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is not set in .env")
	}

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to build request JSON: %w", err)
	}

	req, err := http.NewRequest("POST", geminiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to Gemini failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(body))
	}

	var parsed geminiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no text; raw body: %s", string(body))
	}

	return parsed.Candidates[0].Content.Parts[0].Text, nil
}
