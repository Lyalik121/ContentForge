package main

import (
	"fmt"
	"log"

	"contentforge/db"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Підключаємося до бази даних MS SQL
	database := db.ConnectDB()
	defer database.Close()

	// 2. Ініціалізуємо веб-сервер Fiber
	app := fiber.New()

	// 3. Головна сторінка з примусовим кодуванням UTF-8
	app.Get("/", func(c *fiber.Ctx) error {
		// Кажемо браузеру читати текст як UTF-8, щоб прибрати "кракозябри"
		c.Set("Content-Type", "text/html; charset=utf-8")

		return c.SendString(`
			<h1>🚀 Сервер ContentForge успішно працює!</h1>
			<p>Доступні роути для перевірки:</p>
			<ul>
				<li><a href="/health">Перевірити статус сервера (/health)</a></li>
			</ul>
		`)
	})

	// 4. Роут перевірки здоров'я додатка
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// 5. Тестовий ехо-роут
	app.Post("/echo", func(c *fiber.Ctx) error {
		return c.Send(c.Body())
	})

	// 6. Старт сервера
	fmt.Println("Сервер Fiber запускається на порту :3000...")
	log.Fatal(app.Listen(":3000"))
}
