package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

func New(db *gorm.DB) (*LLMService, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	llmClient := openai.NewClient(os.Getenv("OPENAI_AUTH_TOKEN"))

	return &LLMService{
		db:        db,
		llmClient: llmClient,
	}, nil
}
