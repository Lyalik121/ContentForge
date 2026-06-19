package handlers

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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

	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not open uploaded file.",
		})
	}
	defer src.Close()

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
		"status":   "success",
		"message":  "File uploaded",
		"filename": fileHeader.Filename,
		"saved_as": newName,
		"size":     fileHeader.Size,
	})
}
