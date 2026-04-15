package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env only when the file exists (dev). In production, env vars are
	// injected by the runtime and there is no .env file — don't fatal on that.
	if _, statErr := os.Stat(".env"); statErr == nil {
		if err := godotenv.Load(); err != nil {
			log.Fatal("Found .env file but could not parse it: ", err)
		}
	}

	server := initServer()
	server.SetupRoutes()
	// if err := SeedDBWithEvent(server, "./data/events/events.json"); err != nil {
	// 	log.Println("Failed to seed data", err.Error(), ".\nSkipping")
	// }
	// if err := SeedJournal(server, "./data/journal/journals.json"); err != nil {
	// 	log.Println("Failed to seed journal", err.Error(), ".\nSkipping")
	// }
	// if err := SeedPodcasts(server, "./data/podcasts/podcasts.json"); err != nil {
	// 	log.Println("Failed to seed podcasts", err.Error(), ".\nSkipping")
	// }
	// if err := SeedArticles(server, "./data/articles/articles.json"); err != nil {
	// 	log.Println("Failed to seed articles", err.Error(), ".\nSkipping")
	// }
	// if err := SeedCourses(server, "./data/courses.json"); err != nil {
	// 	log.Println("Failed to seed courses", err.Error(), ".\nSkipping")
	// }
	if err := SeedAdminUser(server); err != nil {
		log.Println("Failed to seed admin user", err.Error(), ".\nSkipping")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	server.r.Run(":" + port)
}
