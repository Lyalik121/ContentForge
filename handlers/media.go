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

type GenerateRequest struct {
	Niche        string `json:"niche"`
	Audience     string `json:"audience"`
	Requirements string `json:"requirements"`
	ContentType  string `json:"content_type"`
}

func NewMediaHandler(database *sql.DB) *MediaHandler {
	return &MediaHandler{db: database}
}

func (h *MediaHandler) Generate(c *fiber.Ctx) error {
	var req GenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body.",
		})
	}

	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid user identity.",
		})
	}
	userID := int(userIDFloat)

	brief := fmt.Sprintf("Niche: %s\nAudience: %s\nRequirements: %s", req.Niche, req.Audience, req.Requirements)

	var requestID int
	insertQuery := `INSERT INTO generation_requests (user_id, prompt_modifier) OUTPUT INSERTED.id VALUES (@p1, @p2)`
	err := h.db.QueryRow(insertQuery, userID, brief).Scan(&requestID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not create request.",
		})
	}

	var result interface{}
	var contentType string

	if req.ContentType == "script" {
		script, err := processor.GenerateScriptWithRetry(brief)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Could not generate script.",
			})
		}
		result = script
		contentType = "video_script"
	} else {
		posts, err := processor.GeneratePostsWithRetry(brief)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Could not generate posts.",
			})
		}
		result = posts
		contentType = "social_posts"
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Serialization error.",
		})
	}

	genQuery := `INSERT INTO generated_content (request_id, content_type, result_text) VALUES (@p1, @p2, @p3)`
	_, err = h.db.Exec(genQuery, requestID, contentType, string(resultJSON))
	if err != nil {
		log.Printf("Помилка збереження тексту в generated_content: %v", err)
	}

	return c.JSON(fiber.Map{"status": "success", "result": result})
}

func (h *MediaHandler) Upload(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "File not found.",
		})
	}

	if fileHeader.Size > maxUploadSize {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"status":  "error",
			"message": "File is too large.",
		})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not open file.",
		})
	}
	defer src.Close()

	headBytes := make([]byte, 512)
	n, _ := src.Read(headBytes)
	contentType := http.DetectContentType(headBytes[:n])
	if !strings.HasPrefix(contentType, "audio/") && !strings.HasPrefix(contentType, "video/") {
		return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
			"status":  "error",
			"message": "Only audio and video files allowed.",
		})
	}

	src.Seek(0, io.SeekStart)

	uploadDir := "uploads"
	os.MkdirAll(uploadDir, 0755)
	ext := filepath.Ext(fileHeader.Filename)
	newName := uuid.New().String() + ext
	destPath := filepath.Join(uploadDir, newName)

	dst, err := os.Create(destPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not save file to disk.",
		})
	}
	defer dst.Close()
	io.Copy(dst, src)

	userIDFloat, _ := c.Locals("user_id").(float64)
	userID := int(userIDFloat)

	var requestID int
	reqQuery := `INSERT INTO generation_requests (user_id, prompt_modifier) OUTPUT INSERTED.id VALUES (@p1, @p2)`
	err = h.db.QueryRow(reqQuery, userID, "Media Pipeline File: "+fileHeader.Filename).Scan(&requestID)
	if err != nil {
		os.Remove(destPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not initiate pipeline request in DB.",
		})
	}

	query := `INSERT INTO media_files (user_id, file_name, file_path, status) OUTPUT INSERTED.id VALUES (@p1, @p2, @p3, @p4)`
	var mediaID int
	err = h.db.QueryRow(query, userID, fileHeader.Filename, destPath, "Uploaded").Scan(&mediaID)
	if err != nil {
		os.Remove(destPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Database error.",
		})
	}

	go h.processMedia(mediaID, requestID)

	return c.JSON(fiber.Map{
		"status":       "success",
		"media_id":     mediaID,
		"media_status": "Uploaded",
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
			"message": "Database error.",
		})
	}

	response := fiber.Map{
		"status":       "success",
		"media_id":     id,
		"media_status": mediaStatus,
	}

	if strings.ToLower(mediaStatus) == "completed" {
		var resultText string
		contentQuery := `
			SELECT TOP 1 gc.result_text 
			FROM generated_content gc
			JOIN generation_requests gr ON gc.request_id = gr.id
			JOIN media_files mf ON gr.user_id = mf.user_id AND gr.prompt_modifier = ('Media Pipeline File: ' + mf.file_name)
			WHERE mf.id = @p1
			ORDER BY gc.id DESC`

		err = h.db.QueryRow(contentQuery, id).Scan(&resultText)
		if err != nil {
			backupQuery := `SELECT TOP 1 result_text FROM generated_content WHERE content_type = 'social_posts' ORDER BY id DESC`
			err = h.db.QueryRow(backupQuery).Scan(&resultText)
		}

		if err == nil {
			var posts map[string]interface{}
			if jsonErr := json.Unmarshal([]byte(resultText), &posts); jsonErr == nil {
				if val, ok := posts["telegram"]; ok && val != nil && val != "" {
					response["telegram"] = val
				} else if val, ok := posts["twitter"]; ok && val != nil && val != "" {
					response["telegram"] = val
				} else if val, ok := posts["twitter_x"]; ok && val != nil && val != "" {
					response["telegram"] = val
				} else if val, ok := posts["x"]; ok && val != nil && val != "" {
					response["telegram"] = val
				} else {
					response["telegram"] = "Текст для Telegram не знайдено в структурі ШІ."
				}

				response["instagram"] = posts["instagram"]

				if tk, ok := posts["tiktok"]; ok && tk != nil && tk != "" {
					response["tiktok"] = tk
				} else if li, ok := posts["linkedin"]; ok {
					response["tiktok"] = li
				}

				if th, ok := posts["threads"]; ok && th != nil && th != "" {
					response["threads"] = th
				} else {
					response["threads"] = posts["instagram"]
				}
			} else {
				response["telegram"] = resultText
				response["instagram"] = resultText
				response["tiktok"] = resultText
				response["threads"] = resultText
			}
		} else {
			log.Printf("GetStatus: вміст для медіа %d не знайдено в базі: %v", id, err)
		}
	}

	return c.JSON(response)
}

func (h *MediaHandler) updateStatus(mediaID int, status string) error {
	query := `UPDATE media_files SET status = @p1 WHERE id = @p2`
	_, err := h.db.Exec(query, status, mediaID)
	return err
}

func (h *MediaHandler) processMedia(mediaID int, requestID int) {
	log.Printf("media %d: початковий запуск обробки", mediaID)
	var videoPath string
	err := h.db.QueryRow("SELECT file_path FROM media_files WHERE id = @p1", mediaID).Scan(&videoPath)
	if err != nil {
		log.Printf("media %d: помилка читання шляху файлу: %v", mediaID, err)
		_ = h.updateStatus(mediaID, "Failed")
		return
	}

	audioOutputDir := "uploads"
	audioPath := filepath.Join(audioOutputDir, fmt.Sprintf("audio_%d.mp3", mediaID))

	if err = processor.ExtractAudio(videoPath, audioPath); err != nil {
		log.Printf("media %d: помилка витягування звуку: %v", mediaID, err)
		_ = h.updateStatus(mediaID, "Failed")
		return
	}
	log.Printf("media %d: звук успішно витягнуто", mediaID)

	_ = h.updateStatus(mediaID, "Transcribing")
	var transcriptText string

	if os.Getenv("MOCK_TRANSCRIPTION") == "true" {
		transcriptText = "Це автоматично згенерований контент конвеєра ContentForge."
	} else {
		openAIKey := os.Getenv("OPENAI_API_KEY")
		chunkDir := filepath.Join("uploads", fmt.Sprintf("chunks_%d", mediaID))
		transcriptText, err = processor.TranscribeLargeAudio(audioPath, chunkDir, openAIKey)
		if err != nil {
			log.Printf("media %d: помилка транскрибації: %v", mediaID, err)
			_ = h.updateStatus(mediaID, "Failed")
			return
		}
	}

	_, _ = h.db.Exec(`INSERT INTO transcripts (media_file_id, raw_text) VALUES (@p1, @p2)`, mediaID, transcriptText)

	_ = h.updateStatus(mediaID, "Generating")

	posts, err := processor.GeneratePostsWithRetry(transcriptText)
	if err != nil {
		log.Printf("media %d: помилка генерації постів через Gemini ШІ: %v", mediaID, err)
		_ = h.updateStatus(mediaID, "Failed")
		return
	}

	postsJSON, _ := json.Marshal(posts)

	finalQuery := `INSERT INTO generated_content (request_id, content_type, result_text) VALUES (@p1, @p2, @p3)`
	_, err = h.db.Exec(finalQuery, requestID, "social_posts", string(postsJSON))
	if err != nil {
		log.Printf("media %d: КРИТИЧНА ПОМИЛКА SQL ПРИ ЗБЕРЕЖЕННІ ПОСТІВ: %v. Пробую зберегти через пряму назву.", mediaID, err)

		var fallbackRequestID int
		fallbackQuery := `SELECT TOP 1 gr.id FROM generation_requests gr JOIN media_files mf ON gr.user_id = mf.user_id AND gr.prompt_modifier = ('Media Pipeline File: ' + mf.file_name) WHERE mf.id = @p1 ORDER BY gr.id DESC`
		err = h.db.QueryRow(fallbackQuery, mediaID).Scan(&fallbackRequestID)
		if err == nil {
			_, err = h.db.Exec(finalQuery, fallbackRequestID, "social_posts", string(postsJSON))
		}

		if err != nil {
			log.Printf("media %d: Помилка збереження навіть після резервного кроку: %v", mediaID, err)
			_ = h.updateStatus(mediaID, "Failed")
			return
		}
	}

	_ = h.updateStatus(mediaID, "Completed")
	log.Printf("media %d: обробка успішно завершена!", mediaID)
}
