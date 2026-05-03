package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"whisprtone/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")

	if err := utils.ConstructAudio("audio.wav", 5); err != nil {
		panic(err)
	}

	transcript := utils.GetTranscript("audio.wav", apiKey)
	fmt.Println("Transcript: " + transcript)
	fmt.Println("Post-process: " + utils.PostProcess(transcript, apiKey))
}
