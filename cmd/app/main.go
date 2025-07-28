package main

import (
	"log"

	"github.com/nghiavan0610/btaskee-quiz-service/app"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/env"
)

func main() {
	env.Load()

	serviceApp, err := app.AppFactory()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	serviceApp.Start()
}
