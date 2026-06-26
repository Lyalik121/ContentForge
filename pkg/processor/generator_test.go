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
