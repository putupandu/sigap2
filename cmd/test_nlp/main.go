package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sigap2/sigap2/internal/services"
)

func main() {
	godotenv.Load()
	fmt.Println("GEMINI_API_KEY set:", os.Getenv("GEMINI_API_KEY") != "")
	fmt.Println("Key prefix:", os.Getenv("GEMINI_API_KEY")[:10])
	
	desc := "terjebak longsor, saya butuh 10 makanan, 10 minuman dan 5 baju"
	needs := "-"
	
	fmt.Println("\n--- Testing AnalyzeReport ---")
	fmt.Printf("Description: %s\n", desc)
	fmt.Printf("Needs: %s\n\n", needs)
	
	urgency, items, err := services.AnalyzeReport(desc, needs)
	fmt.Println("\n--- Results ---")
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Urgency: %s\n", urgency)
	fmt.Printf("Items: %s\n", items)
}
