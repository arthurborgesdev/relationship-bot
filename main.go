package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/client"

	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	openai "github.com/sashabaranov/go-openai"
)

type Message struct {
	MessageText string `json:"message"`
}

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
				"dim": "1536",
			},
		},
	},
	EnableDynamicField: true,
}

func embed(message string, llmClient *openai.Client) ([]float32, error) {
	embeddingReq := openai.EmbeddingRequest{
		Input: message,
		Model: openai.AdaEmbeddingV2,
	}

	resp, err := llmClient.CreateEmbeddings(context.Background(), embeddingReq)
	if err != nil {
		fmt.Printf("Embedding error: %v\n", err)
		return nil, err
	}

	fmt.Println(resp.Data[0].Embedding)

	return resp.Data[0].Embedding, nil
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

	app.Post("/collections", func(c *fiber.Ctx) error {
		err = milvusClient.CreateCollection(
			context.Background(), // ctx
			schema,
			2, // shardNum
		)
		if err != nil {
			log.Println("failed to create collection:", err.Error())
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return c.SendString("Collection Created!")
	})

	app.Get("/collections/:name", func(c *fiber.Ctx) error {
		collDesc, err := milvusClient.DescribeCollection( // Return the name and schema of the collection.
			context.Background(),
			c.Params("name"),
		)
		if err != nil {
			log.Println("failed to check collection schema:", err.Error())
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		log.Printf("%v\n", collDesc)

		return c.JSON(collDesc)
	})

	app.Get("/collections", func(c *fiber.Ctx) error {
		listColl, err := milvusClient.ListCollections(
			context.Background(), // ctx
		)
		if err != nil {
			log.Println("failed to list all collections", err.Error())
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		log.Println(listColl)

		return c.JSON(listColl)
	})

	app.Delete("/collections/:name", func(c *fiber.Ctx) error {
		err := milvusClient.DropCollection(
			context.Background(), // ctx
			c.Params("name"),
		)
		if err != nil {
			log.Println("failed to delete collection", err.Error())
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return c.SendString("Collection Deleted!")
	})

	app.Post("/embeddings", func(c *fiber.Ctx) error {

		message := new(Message)

		if err := c.BodyParser(message); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		embeddingReq := openai.EmbeddingRequest{
			Input: message.MessageText,
			Model: openai.AdaEmbeddingV2,
		}

		resp, err := llmClient.CreateEmbeddings(context.Background(), embeddingReq)
		if err != nil {
			fmt.Printf("Embedding error: %v\n", err)
			return c.SendString(err.Error())
		}

		fmt.Println(resp.Data[0].Embedding)

		strData := make([]string, len(resp.Data[0].Embedding))

		for i, v := range resp.Data[0].Embedding {
			strData[i] = strconv.FormatFloat(float64(v), 'f', 6, 64)
		}

		result := strings.Join(strData, " ")

		return c.SendString(string(result))
	})

	app.Post("/insert", func(c *fiber.Ctx) error {
		message := new(Message)

		if err := c.BodyParser(message); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		vector, err := embed(message.MessageText, llmClient)
		if err != nil {
			fmt.Printf("Embedding error: %v\n", err)
		}

		messageColumn := entity.NewColumnVarChar("message", []string{message.MessageText})
		senderColumn := entity.NewColumnVarChar("sender", []string{"user"})
		messageVectorColumn := entity.NewColumnFloatVector("message_vector", 1536, [][]float32{vector})

		_, err = milvusClient.Insert(
			context.Background(), // ctx
			"messages",           // CollectionName
			"",                   // partitionName
			messageColumn,        // columnarData
			senderColumn,         // columnarData
			messageVectorColumn,  // columnarData
		)
		if err != nil {
			log.Println("failed to insert data:", err.Error())
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		idx, err := entity.NewIndexIvfFlat(
			entity.L2,
			1024,
		)
		if err != nil {
			log.Fatal("fail to create ivf flat index parameter:", err.Error())
		}
		err = milvusClient.CreateIndex(
			context.Background(),
			"messages",
			"message_vector",
			idx,
			false,
		)
		if err != nil {
			log.Fatal("fail to create index:", err.Error())
		}

		err = milvusClient.LoadCollection(
			context.Background(),
			"messages",
			false,
		)
		if err != nil {
			log.Println("Failed to load collection:", err.Error())
			return c.SendString(err.Error())
		}

		return c.SendString("Data inserted and collection loaded in memory!")
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
