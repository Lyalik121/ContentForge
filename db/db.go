package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func ConnectDB() *sql.DB {
	_ = godotenv.Load("cmd/server/.env")
	_ = godotenv.Load(".env")

	connString := os.Getenv("DATABASE_URL")

	if connString == "" {
		host := os.Getenv("DB_HOST")
		if host == "" {
			host = "localhost"
		}

		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432"
		}

		user := os.Getenv("DB_USER")
		if user == "" {
			user = "postgres"
		}

		password := os.Getenv("DB_PASSWORD")
		if password == "" {
			password = "postgres_password"
		}

		dbname := os.Getenv("DB_NAME")
		if dbname == "" {
			dbname = "contentforge"
		}

		connString = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	}

	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal("Помилка конфігурації підключення до PostgreSQL:", err)
	}

	if err = db.Ping(); err != nil {
		fmt.Printf("УВАГА: База даних PostgreSQL недоступна (%v). Бекенд запускається без неї.\n", err)
	} else {
		fmt.Println("Успішно підключено до бази даних PostgreSQL!")
	}

	return db
}
