package handlers

import (
	"database/sql"
	"log"
	"time"

	"contentforge/models"

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

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var dto models.RegisterDTO

	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Невалідний JSON запиту"})
	}

	if dto.Email == "" || dto.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email та пароль обов'язкові для заповнення"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Помилка шифрування пароля"})
	}

	query := `INSERT INTO users (email, password) VALUES (?, ?);`
	_, err = h.DB.ExecContext(c.Context(), query, dto.Email, string(hashedPassword))
	if err != nil {
		log.Println("Помилка при реєстрації в БД:", err)
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Користувач із таким email вже зареєстрований або виникла помилка БД"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Користувача успішно зареєстровано!"})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var dto models.LoginDTO

	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Невалідний JSON запиту"})
	}

	var user models.User

	query := `SELECT id, email, password FROM users WHERE email = ?;`
	err := h.DB.QueryRowContext(c.Context(), query, dto.Email).Scan(&user.ID, &user.Email, &user.PasswordHash)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Невірний email або пароль"})
	} else if err != nil {
		log.Println("Помилка при пошуку користувача в БД:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Помилка сервера при пошуку користувача"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Невірний email або пароль"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не вдалося створити сесію (JWT)"})
	}

	return c.JSON(fiber.Map{
		"message": "Вхід успішний!",
		"token":   tokenString,
	})
}
