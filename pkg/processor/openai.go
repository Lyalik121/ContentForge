package processor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type WhisperResponse struct {
	Text string `json:"text"`
}

func TranscribeAudioWithRetry(audioPath, apiKey string) (string, error) {
	maxRetries := 3
	backoff := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		text, err := SendToWhisper(audioPath, apiKey)
		if err == nil {
			return text, nil
		}

		// Якщо це остання спроба — повертаємо помилку
		if i == maxRetries-1 {
			return "", fmt.Errorf("whisper api failed after %d retries: %w", maxRetries, err)
		}

		fmt.Printf("Помилка Whisper API (спроба %d/%d). Наступна спроба через %v... Помилка: %v\n", i+1, maxRetries, backoff, err)
		time.Sleep(backoff)
		backoff *= 2
	}

	return "", fmt.Errorf("unexpected end of retry loop")
}

func SendToWhisper(audioPath, apiKey string) (string, error) {
	file, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err = io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	if err = writer.WriteField("model", "whisper-1"); err != nil {
		return "", fmt.Errorf("failed to write form field model: %w", err)
	}

	if err = writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error during request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var whisperResp WhisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&whisperResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return whisperResp.Text, nil
}
