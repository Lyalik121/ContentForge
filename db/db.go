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

		createTablesQuery := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS media_files (
			id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(id) ON DELETE CASCADE,
			file_name VARCHAR(255) NOT NULL, 
			file_path VARCHAR(500) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'Uploaded', 
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT CHK_MediaStatus CHECK (status IN ('Uploaded', 'Transcribing', 'Transcribed', 'Generating', 'Completed', 'Failed'))
		);

		CREATE TABLE IF NOT EXISTS generation_requests (
			id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(id) ON DELETE CASCADE,
			media_file_id INT REFERENCES media_files(id) ON DELETE CASCADE,
			prompt_modifier TEXT, 
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS transcripts (
			id SERIAL PRIMARY KEY,
			media_file_id INT UNIQUE REFERENCES media_files(id) ON DELETE CASCADE,
			raw_text TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS generated_content (
			id SERIAL PRIMARY KEY,
			request_id INT REFERENCES generation_requests(id) ON DELETE CASCADE, 
			media_file_id INT REFERENCES media_files(id) ON DELETE CASCADE,
			content_type VARCHAR(50),
			result_text TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`

		_, err = db.Exec(createTablesQuery)
		if err != nil {
			log.Printf("Помилка автоматичного створення структур таблиць: %v", err)
		} else {
			fmt.Println("Всі таблиці бази даних успішно перевірені/створені в PostgreSQL.")
		}
	}

	return db
}
