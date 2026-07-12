package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/microsoft/go-mssqldb"
)

func ConnectDB() *sql.DB {
	err := godotenv.Load("cmd/server/.env")
	if err != nil {
		_ = godotenv.Load(".env")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "1433"
	}
	if user == "" {
		user = "sa"
	}
	if password == "" {
		password = "YourStrong@Password123"
	}
	if dbname == "" {
		dbname = "contentforge"
	}

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;encrypt=disable",
		host, user, password, port, dbname)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Помилка під час налаштування підключення:", err)
	}

	if err = db.Ping(); err != nil {
		fmt.Printf("УВАГА: База даних недоступна (%v). Але сервер продовжує запуск!\n", err)
	} else {
		fmt.Println("Успішно підключено до бази даних MS SQL!")
	}

	return db
}
