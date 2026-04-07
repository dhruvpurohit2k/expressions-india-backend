package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Could not load ENV", err)
	}

	server := initServer()
	server.SetupRoutes()
	if err := SeedDBWithEvent(server, "./data/events/events.json"); err != nil {
		log.Println("Failed to seed data", err.Error(), ".\nSkipping")
	}
	if err := SeedJournal(server, "./data/journal/journals.json"); err != nil {
		log.Println("Failed to seed journal", err.Error(), ".\nSkipping")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	server.r.Run(":" + port)
}
