package processor

import (
	"fmt"
	"testing"

	"github.com/joho/godotenv"
)

func TestCallGemini(t *testing.T) {

	_ = godotenv.Load("../../cmd/server/.env")

	answer, err := callGemini("Привітайся українською, одне коротке речення.")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Gemini answered:", answer)
}

func TestGeneratePosts(t *testing.T) {
	_ = godotenv.Load("../../cmd/server/.env")

	transcript := `Сьогодні я хочу розповісти про те, як штучний інтелект змінює 
	нашу роботу. ШІ допомагає автоматизувати рутинні задачі, економить час 
	і дозволяє зосередитись на творчості. Але важливо памʼятати, що це лише 
	інструмент — головне залишається за людиною.`

	posts, err := GeneratePosts(transcript)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("=== INSTAGRAM ===\n", posts.Instagram)
	fmt.Println("\n=== TIKTOK ===\n", posts.TikTok)
	fmt.Println("\n=== THREADS ===\n", posts.Threads)
	fmt.Println("\n=== TELEGRAM ===\n", posts.Telegram)
}

func TestGeneratePostsWithRetry(t *testing.T) {
	_ = godotenv.Load("../../cmd/server/.env")

	transcript := "Сьогодні поговоримо про користь ранкових пробіжок для здоровʼя та настрою."

	posts, err := GeneratePostsWithRetry(transcript)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("=== THREADS ===\n", posts.Threads)
}
