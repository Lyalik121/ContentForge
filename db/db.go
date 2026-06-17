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

	if port == "" {
		port = "1433"
		host = "localhost"
		user = "sa"
		password = "YourStrong@Password123"
		dbname = "contentforge"
	}

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;encrypt=disable",
		host, user, password, port, dbname)

	db, err := sql.Open("mssql", connString)
	if err != nil {
		log.Fatal("Помилка під час налаштування підключення:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("База даних недоступна (Ping failed):", err)
	}

	fmt.Println("Успішно підключено до бази даних MS SQL!")
	return db
}
