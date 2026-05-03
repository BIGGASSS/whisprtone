package main

import (
	"fmt"
	"github.com/eiannone/keyboard"
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

	// Open keyboard for keybind detection
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	stopCh := make(chan struct{})

	// Key listener goroutine: press ESC to stop recording
	go func() {
		for {
			_, key, err := keyboard.GetKey()
			if err != nil {
				log.Fatal(err)
			}
			if key == keyboard.KeyEsc {
				close(stopCh)
				return
			}
		}
	}()

	if err := utils.RecordUntil("audio.wav", stopCh); err != nil {
		panic(err)
	}

	transcript := utils.GetTranscript("audio.wav", apiKey)
	fmt.Println("Transcript: " + transcript)
	fmt.Println("Post-process: " + utils.PostProcess(transcript, apiKey))
}
