package handlers

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
		"message":       "File uploaded",
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
	// Симуляція довгої обробки. Реальні Whisper/Gemini прийдуть у Пунктах 10-11.
	// time.Sleep вдає роботу, яка в житті триватиме хвилини.

	time.Sleep(3 * time.Second)
	if err := h.updateStatus(mediaID, "Transcribing"); err != nil {
		log.Printf("media %d: failed to set Transcribing: %v", mediaID, err)
		return
	}

	time.Sleep(3 * time.Second)
	if err := h.updateStatus(mediaID, "Transcribed"); err != nil {
		log.Printf("media %d: failed to set Transcribed: %v", mediaID, err)
		return
	}

	time.Sleep(3 * time.Second)
	if err := h.updateStatus(mediaID, "Generating"); err != nil {
		log.Printf("media %d: failed to set Generating: %v", mediaID, err)
		return
	}

	time.Sleep(3 * time.Second)
	if err := h.updateStatus(mediaID, "Completed"); err != nil {
		log.Printf("media %d: failed to set Completed: %v", mediaID, err)
		return
	}

	log.Printf("media %d: processing completed", mediaID)
}
