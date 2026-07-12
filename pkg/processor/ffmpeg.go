package processor

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getFFmpegPath() string {
	path := "ffmpeg"

	if _, err := os.Stat(`D:\ffmpeg\ffmpeg.exe`); err == nil {
		path = `D:\ffmpeg\ffmpeg.exe`
	}
	return path
}

func ProcessMedia(db *sql.DB, mediaID int, videoPath, outputPath, outputDir string) error {
	if err := ExtractAudio(videoPath, outputPath); err != nil {
		_, _ = db.Exec("UPDATE media_files SET status = 'Failed' WHERE id = $1", mediaID)
		return err
	}
	_, err := db.Exec("UPDATE media_files SET status = 'Transcribing' WHERE id = $1", mediaID)
	if err != nil {
		return fmt.Errorf("failed to update status to Transcribing: %w", err)
	}

	return nil
}

func ExtractAudio(videoPath, outputPath string) error {
	cmd := exec.Command(getFFmpegPath(),
		"-i", videoPath,
		"-vn",
		"-acodec", "libmp3lame",
		"-ab", "128k",
		outputPath,
		"-y",
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg extract error: %w", err)
	}
	return nil
}

func SplitAudio(audioPath, outputDir string) ([]string, error) {
	outputMask := filepath.Join(outputDir, "chunk_%03d.mp3")

	cmd := exec.Command(getFFmpegPath(),
		"-i", audioPath,
		"-f", "segment",
		"-segment_time", "1200",
		"-c", "copy",
		outputMask,
		"-y",
	)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg split error: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(outputDir, "chunk_*.mp3"))
	if err != nil {
		return nil, err
	}

	return files, nil
}

const maxWhisperFileSize = 20 * 1024 * 1024

func TranscribeLargeAudio(audioPath, chunkDir, apiKey string) (string, error) {
	info, err := os.Stat(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat audio file: %w", err)
	}

	if info.Size() <= maxWhisperFileSize {
		return TranscribeAudioWithRetry(audioPath, apiKey)
	}

	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create chunk dir: %w", err)
	}

	chunks, err := SplitAudio(audioPath, chunkDir)
	if err != nil {
		return "", fmt.Errorf("failed to split audio: %w", err)
	}
	if len(chunks) == 0 {
		return "", fmt.Errorf("split produced no chunks")
	}

	var parts []string
	for i, chunk := range chunks {
		log.Printf("transcribing chunk %d/%d: %s", i+1, len(chunks), chunk)
		text, err := TranscribeAudioWithRetry(chunk, apiKey)
		if err != nil {
			return "", fmt.Errorf("chunk %d/%d failed: %w", i+1, len(chunks), err)
		}
		parts = append(parts, text)
	}

	if err := os.RemoveAll(chunkDir); err != nil {
		log.Printf("warning: failed to clean up chunk dir %s: %v", chunkDir, err)
	}

	return strings.Join(parts, " "), nil
}
