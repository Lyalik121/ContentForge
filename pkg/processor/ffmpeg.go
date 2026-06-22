package processor

import (
	"database/sql"
	"fmt"
	"os/exec"
	"path/filepath"
)

const ffmpegPath = "D:\\ffmpeg\\ffmpeg.exe"

func ProcessMedia(db *sql.DB, mediaID int, videoPath, outputPath, outputDir string) error {
	if err := ExtractAudio(videoPath, outputPath); err != nil {
		_, _ = db.Exec("UPDATE media_files SET status = 'Failed' WHERE id = @p1", mediaID)
		return err
	}

	_, err := db.Exec("UPDATE media_files SET status = 'Transcribing' WHERE id = @p1", mediaID)
	if err != nil {
		return fmt.Errorf("failed to update status to Transcribing: %w", err)
	}

	return nil
}

func ExtractAudio(videoPath, outputPath string) error {
	cmd := exec.Command(ffmpegPath,
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

	cmd := exec.Command(ffmpegPath,
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
