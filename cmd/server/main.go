package main

import (
	"fmt"
	"log"

	"contentforge/db"
	"contentforge/handlers"
	"contentforge/middleware"

	"github.com/gofiber/fiber/v2"
)

func main() {
	database := db.ConnectDB()
	defer database.Close()

	app := fiber.New(fiber.Config{
		BodyLimit: 500 * 1024 * 1024, // 500 MB
	})

	authHandler := handlers.NewAuthHandler(database)
	mediaHandler := handlers.NewMediaHandler(database)

	app.Get("/", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")

		return c.SendString(`
			<h1>🚀 Сервер ContentForge успішно працює!</h1>
			<p>Доступні роути для перевірки:</p>
			<ul>
				<li><a href="/health">Перевірити статус сервера (/health)</a></li>
			</ul>
		`)
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Post("/echo", func(c *fiber.Ctx) error {
		return c.Send(c.Body())
	})

	auth := app.Group("/api/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	media := app.Group("/api/media")
	media.Post("/upload", middleware.Protected(), mediaHandler.Upload)

	api := app.Group("/api")

	api.Get("/profile", middleware.Protected(), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		email := c.Locals("email")

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Ви успішно увійшли в захищену зону!",
			"user_id": userID,
			"email":   email,
		})
	})

	fmt.Println("Сервер Fiber запускається на порту :3000...")
	log.Fatal(app.Listen(":3000"))
}
