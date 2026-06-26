package processor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type GeneratedPosts struct {
	Instagram string `json:"instagram"`
	TikTok    string `json:"tiktok"`
	Threads   string `json:"threads"`
	Telegram  string `json:"telegram"`
}

const geminiModel = "gemini-3.1-flash-lite"
const geminiURL = "https://generativelanguage.googleapis.com/v1beta/models/" + geminiModel + ":generateContent"

const systemPrompt = `You are a social media content creator. 
Based on the transcript provided by the user, generate one ready-to-publish post for each of these platforms: Instagram, TikTok, Threads, Telegram.

Adapt the tone and length to each platform:
- instagram: engaging, friendly, with relevant emojis and 3-5 hashtags.
- tiktok: short, catchy, energetic, with a hook in the first line.
- threads: short and conversational, casual tone, max 500 characters.
- telegram: informative and clear, can be a bit longer, no hashtags needed.

Write all posts in the SAME language as the transcript.

Return your answer STRICTLY as a JSON object with EXACTLY these keys: "instagram", "tiktok", "threads", "telegram".
Each value must be a non-empty string. Do NOT add any text, explanations, or markdown outside the JSON object.`

type geminiRequest struct {
	Contents          []geminiContent   `json:"contents"`
	SystemInstruction *geminiContent    `json:"systemInstruction,omitempty"`
	GenerationConfig  *generationConfig `json:"generationConfig,omitempty"`
}
type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}
type geminiPart struct {
	Text string `json:"text"`
}

type generationConfig struct {
	ResponseMimeType string `json:"responseMimeType"`
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

func GeneratePosts(transcript string) (*GeneratedPosts, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set in .env")
	}

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: transcript}}},
		},
		GenerationConfig: &generationConfig{
			ResponseMimeType: "application/json", // force valid JSON output
		},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to build request JSON: %w", err)
	}

	req, err := http.NewRequest("POST", geminiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to Gemini failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(body))
	}

	var parsed geminiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini envelope: %w", err)
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned no content; raw body: %s", string(body))
	}
	rawPostsJSON := parsed.Candidates[0].Content.Parts[0].Text

	var posts GeneratedPosts
	if err := json.Unmarshal([]byte(rawPostsJSON), &posts); err != nil {
		return nil, fmt.Errorf("failed to parse posts JSON: %w; raw: %s", err, rawPostsJSON)
	}

	if err := validatePosts(&posts); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &posts, nil
}

const (
	instagramLimit = 2200
	tiktokLimit    = 2200
	threadsLimit   = 500
	telegramLimit  = 4096
)

func validatePosts(p *GeneratedPosts) error {

	if strings.TrimSpace(p.Instagram) == "" {
		return fmt.Errorf("instagram post is empty")
	}
	if strings.TrimSpace(p.TikTok) == "" {
		return fmt.Errorf("tiktok post is empty")
	}
	if strings.TrimSpace(p.Threads) == "" {
		return fmt.Errorf("threads post is empty")
	}
	if strings.TrimSpace(p.Telegram) == "" {
		return fmt.Errorf("telegram post is empty")
	}

	if len([]rune(p.Instagram)) > instagramLimit {
		return fmt.Errorf("instagram post too long: %d chars (max %d)", len([]rune(p.Instagram)), instagramLimit)
	}
	if len([]rune(p.TikTok)) > tiktokLimit {
		return fmt.Errorf("tiktok post too long: %d chars (max %d)", len([]rune(p.TikTok)), tiktokLimit)
	}
	if len([]rune(p.Threads)) > threadsLimit {
		return fmt.Errorf("threads post too long: %d chars (max %d)", len([]rune(p.Threads)), threadsLimit)
	}
	if len([]rune(p.Telegram)) > telegramLimit {
		return fmt.Errorf("telegram post too long: %d chars (max %d)", len([]rune(p.Telegram)), telegramLimit)
	}

	return nil
}
