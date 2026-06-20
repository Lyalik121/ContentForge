package handlers

import (
	"database/sql"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	return c.JSON(fiber.Map{
		"status":        "success",
		"message":       "File uploaded",
		"filename":      fileHeader.Filename,
		"saved_as":      newName,
		"size":          fileHeader.Size,
		"detected_type": contentType,
	})
}
