package main

import (
	"log"
	"os"

	"github.com/arthurborgesdev/relationship-bot/vectordb"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := fiber.New()

	llmClient := openai.NewClient(os.Getenv("OPENAI_AUTH_TOKEN"))

	MilvusService, err := vectordb.New(llmClient)
	if err != nil {
		log.Fatalf("Failed to create Milvus service: %v", err)
	}

	MilvusService.RegisterRoutes(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	app.Listen(":3000")
}
