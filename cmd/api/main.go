package main

import (
	"github.com/joho/godotenv"
	"log"
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
	server.r.Run(":8000")
}
