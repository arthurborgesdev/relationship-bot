package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/client"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	openai "github.com/sashabaranov/go-openai"
)

var schema = &entity.Schema{
	CollectionName: "messages",
	Description:    "Messages sent in the chatbot conversation",
	Fields: []*entity.Field{
		{
			Name:       "message_id",
			DataType:   entity.FieldTypeInt64,
			AutoID:     true,
			PrimaryKey: true,
		},
		{
			Name:       "message",
			DataType:   entity.FieldTypeVarChar,
			PrimaryKey: false,
			AutoID:     false,
			TypeParams: map[string]string{
				"max_length": "65535",
			},
		},
		{
			Name:       "sender",
			DataType:   entity.FieldTypeVarChar,
			PrimaryKey: false,
			AutoID:     false,
			TypeParams: map[string]string{
				"max_length": "100",
			},
		},
		{
			Name:     "message_vector",
			DataType: entity.FieldTypeFloatVector,
			TypeParams: map[string]string{
				"dim": "128",
			},
		},
	},
	EnableDynamicField: true,
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := fiber.New()

	llmClient := openai.NewClient(os.Getenv("OPENAI_AUTH_TOKEN"))

	milvusClient, err := client.NewClient(context.Background(), client.Config{
		Address: "localhost:19530",
	})
	if err != nil {
		fmt.Printf("Can't connect to Milvus: %v\n", err)
	}
	defer milvusClient.Close()

	milvusClient.HasCollection(context.Background(), "messages")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	app.Post("/collection", func(c *fiber.Ctx) error {
		err = milvusClient.CreateCollection(
			context.Background(), // ctx
			schema,
			2, // shardNum
		)
		if err != nil {
			log.Fatal("failed to create collection:", err.Error())
			return c.SendString(err.Error())
		}

		return c.SendString("Collection Created!")
	})

	app.Post("/messages", func(c *fiber.Ctx) error {
		resp, err := llmClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: os.Getenv("OPENAI_MODEL_ID"),
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "Hello!",
					},
				},
			},
		)
		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			return c.SendString(err.Error())
		}

		fmt.Println(resp.Choices[0].Message.Content)

		return c.SendString(resp.Choices[0].Message.Content)
	})

	app.Listen(":3000")
}
