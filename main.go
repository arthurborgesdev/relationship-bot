package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	app := fiber.New()

	LLMService, err := New(db)
	if err != nil {
		log.Fatalf("Failed to create LLM service: %v", err)
	}

	// MilvusService, err := vectordb.New(LLMService.llmClient)
	// if err != nil {
	// 	log.Fatalf("Failed to create Milvus service: %v", err)
	// }

	LLMService.RegisterRoutes(app)
	// MilvusService.RegisterRoutes(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	// googleCalendar()

	app.Listen(":3000")
}
