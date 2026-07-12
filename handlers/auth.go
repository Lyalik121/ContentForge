package handlers

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("super-secret-key-change-me-later")

type AuthHandler struct {
	DB *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

type RegisterDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var dto RegisterDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Невалідний формат даних"})
	}

	dto.Email = strings.TrimSpace(strings.ToLower(dto.Email))
	if dto.Email == "" || dto.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email та пароль обов'язкові"})
	}

	var exists bool
	err := h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", dto.Email).Scan(&exists)
	if err != nil {
		fmt.Println("Помилка при пошуку користувача в БД:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Помилка сервера"})
	}
	if exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Користувач з таким email вже існує"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Помилка шифрування"})
	}

	_, err = h.DB.Exec("INSERT INTO users (email, password_hash, created_at) VALUES ($1, $2, $3)", dto.Email, string(hashedPassword), time.Now())
	if err != nil {
		fmt.Println("Помилка при реєстрації в БД:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Помилка при збереженні в БД"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "message": "Користувача успішно створено"})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var dto LoginDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Невалідний формат даних"})
	}

	dto.Email = strings.TrimSpace(strings.ToLower(dto.Email))

	var dbID int
	var dbEmail string
	var dbPassword string

	err := h.DB.QueryRow("SELECT id, email, password_hash FROM users WHERE email = $1", dto.Email).Scan(&dbID, &dbEmail, &dbPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Невірний email або пароль"})
		}
		fmt.Println("Помилка при пошуку користувача в БД:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Помилка сервера"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(dto.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Невірний email або пароль"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": dbID,
		"email":   dbEmail,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не вдалося згенерувати токен"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"token":  tokenString,
		"user": fiber.Map{
			"id":    dbID,
			"email": dbEmail,
		},
	})
}
