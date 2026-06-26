package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"contentforge/pkg/processor"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const maxUploadSize = 500 * 1024 * 1024

type MediaHandler struct {
	db *sql.DB
}

func NewMediaHandler(database *sql.DB) *MediaHandler {
	return &MediaHandler{db: database}
}

func (h *MediaHandler) Upload(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "File not found in request. Use form field name 'file'.",
		})
	}

	if fileHeader.Size > maxUploadSize {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"status":  "error",
			"message": "File is too large. Maximum size is 500 MB.",
		})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not open uploaded file.",
		})
	}
	defer src.Close()

	headBytes := make([]byte, 512)
	n, err := src.Read(headBytes)
	if err != nil && err != io.EOF {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not read file for validation.",
		})
	}
	contentType := http.DetectContentType(headBytes[:n])
	if !strings.HasPrefix(contentType, "audio/") && !strings.HasPrefix(contentType, "video/") {
		return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
			"status":        "error",
			"message":       "Only audio and video files are allowed.",
			"detected_type": contentType,
		})
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not rewind file.",
		})
	}

	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not create upload directory.",
		})
	}

	ext := filepath.Ext(fileHeader.Filename)
	newName := uuid.New().String() + ext
	destPath := filepath.Join(uploadDir, newName)

	dst, err := os.Create(destPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not create file on disk.",
		})
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not save file.",
		})
	}

	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid user identity in token.",
		})
	}
	userID := int(userIDFloat)

	query := `INSERT INTO media_files (user_id, file_name, file_path, status)
			  OUTPUT INSERTED.id
			  VALUES (@p1, @p2, @p3, @p4)`

	var mediaID int
	err = h.db.QueryRow(query, userID, fileHeader.Filename, destPath, "Uploaded").Scan(&mediaID)
	if err != nil {
		os.Remove(destPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not save file record to database.",
		})
	}

	go h.processMedia(mediaID)

	return c.JSON(fiber.Map{
		"status":        "success",
		"message":       "File uploaded, processing started",
		"media_id":      mediaID,
		"filename":      fileHeader.Filename,
		"saved_as":      newName,
		"size":          fileHeader.Size,
		"detected_type": contentType,
	})
}

func (h *MediaHandler) GetStatus(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid media ID.",
		})
	}

	var mediaStatus string
	query := `SELECT status FROM media_files WHERE id = @p1`
	err = h.db.QueryRow(query, id).Scan(&mediaStatus)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Media file not found.",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not retrieve status.",
		})
	}

	return c.JSON(fiber.Map{
		"status":       "success",
		"media_id":     id,
		"media_status": mediaStatus,
	})
}

func (h *MediaHandler) updateStatus(mediaID int, status string) error {
	query := `UPDATE media_files SET status = @p1 WHERE id = @p2`
	_, err := h.db.Exec(query, status, mediaID)
	return err
}

func (h *MediaHandler) processMedia(mediaID int) {
	log.Printf("media %d: початковий запуск обробки", mediaID)

	var videoPath string
	err := h.db.QueryRow("SELECT file_path FROM media_files WHERE id = @p1", mediaID).Scan(&videoPath)
	if err != nil {
		log.Printf("media %d: не вдалося знайти шлях до файлу в БД: %v", mediaID, err)
		_ = h.updateStatus(mediaID, "Failed")
		return
	}

	audioOutputDir := "uploads"
	audioPath := filepath.Join(audioOutputDir, fmt.Sprintf("audio_%d.mp3", mediaID))

	err = processor.ExtractAudio(videoPath, audioPath)
	if err != nil {
		log.Printf("media %d: помилка FFmpeg під час витягування звуку: %v", mediaID, err)
		_ = h.updateStatus(mediaID, "Failed")
		return
	}

	log.Printf("media %d: звук успішно витягнуто в %s", mediaID, audioPath)

	fileInfo, err := os.Stat(audioPath)
	if err != nil {
		log.Printf("media %d: не вдалося прочитати розмір аудіофайлу: %v", mediaID, err)
		_ = h.updateStatus(mediaID, "Failed")
		return
	}

	const maxAudioSize = 20 * 1024 * 1024

	if fileInfo.Size() > maxAudioSize {
		log.Printf("media %d: аудіофайл більше 20МБ (%d байт). Запускаємо чанкінг...", mediaID, fileInfo.Size())

		chunksDir := filepath.Join(audioOutputDir, fmt.Sprintf("chunks_%d", mediaID))
		_ = os.MkdirAll(chunksDir, 0755)

		chunks, err := processor.SplitAudio(audioPath, chunksDir)
		if err != nil {
			log.Printf("media %d: помилка під час нарізки файлу: %v", mediaID, err)
			_ = h.updateStatus(mediaID, "Failed")
			return
		}

		log.Printf("media %d: успішно нарізано на %d шматків: %v", mediaID, len(chunks), chunks)
	}

	if err := h.updateStatus(mediaID, "Transcribing"); err != nil {
		log.Printf("media %d: не вдалося встановити статус Transcribing: %v", mediaID, err)
		return
	}

	var transcriptText string

	if os.Getenv("MOCK_TRANSCRIPTION") == "true" {
		log.Printf("media %d: MOCK режим увімкнено — реальний виклик Whisper пропущено", mediaID)
		transcriptText = "Це тестова транскрипція (mock). Виклик OpenAI пропущено."
	} else {
		openAIKey := os.Getenv("OPENAI_API_KEY")
		if openAIKey == "" {
			log.Printf("media %d: помилка, OPENAI_API_KEY не вказано в .env", mediaID)
			_ = h.updateStatus(mediaID, "Failed")
			return
		}

		log.Printf("media %d: відправка аудіо до OpenAI Whisper API...", mediaID)

		transcriptText, err = processor.TranscribeAudioWithRetry(audioPath, openAIKey)
		if err != nil {
			log.Printf("media %d: помилка транскрибації після всіх спроб: %v", mediaID, err)
			_ = h.updateStatus(mediaID, "Failed")
			return
		}
	}

	txQuery := `INSERT INTO transcripts (media_file_id, raw_text) VALUES (@p1, @p2)`
	_, err = h.db.Exec(txQuery, mediaID, transcriptText)
	if err != nil {
		log.Printf("media %d: помилка збереження транскрипту в БД: %v", mediaID, err)
		_ = h.updateStatus(mediaID, "Failed")
		return
	}

	_ = h.updateStatus(mediaID, "Transcribed")
	log.Printf("media %d: транскрипт збережено, статус: Transcribed", mediaID)

	if err := h.updateStatus(mediaID, "Generating"); err != nil {
		log.Printf("media %d: не вдалося встановити статус Generating: %v", mediaID, err)
		return
	}
	log.Printf("media %d: запуск генерації постів через Gemini...", mediaID)

	posts, err := processor.GeneratePostsWithRetry(transcriptText)
	if err != nil {
		log.Printf("media %d: генерація постів провалилась: %v", mediaID, err)
		_ = h.updateStatus(mediaID, "Failed")
		return
	}

	postsJSON, err := json.Marshal(posts)
	if err != nil {
		log.Printf("media %d: не вдалося серіалізувати пости в JSON: %v", mediaID, err)
		_ = h.updateStatus(mediaID, "Failed")
		return
	}

	genQuery := `INSERT INTO generated_content (media_file_id, content_type, result_text)
				 VALUES (@p1, @p2, @p3)`
	_, err = h.db.Exec(genQuery, mediaID, "social_posts", string(postsJSON))
	if err != nil {
		log.Printf("media %d: помилка збереження постів у БД: %v", mediaID, err)
		_ = h.updateStatus(mediaID, "Failed")
		return
	}

	_ = h.updateStatus(mediaID, "Completed")
	log.Printf("media %d: обробка повністю завершена! Статус: Completed", mediaID)
}
