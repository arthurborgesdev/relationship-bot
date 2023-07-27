package vectordb

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	openai "github.com/sashabaranov/go-openai"
)

type Message struct {
	MessageText string `json:"message"`
}

type MilvusService struct {
	llmClient    *openai.Client
	milvusClient client.Client
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

func insertVector(llmClient *openai.Client, milvusClient client.Client, c *fiber.Ctx, message, user string) error {
	vector, err := embed(message, llmClient)
	if err != nil {
		fmt.Printf("Embedding error: %v\n", err)
		return err
	}

	messageColumn := entity.NewColumnVarChar("message", []string{message})
	senderColumn := entity.NewColumnVarChar("sender", []string{user})
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
		return err
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
		return err
	}

	err = milvusClient.LoadCollection(
		context.Background(),
		"messages",
		false,
	)
	if err != nil {
		log.Println("Failed to load collection:", err.Error())
		return err
	}

	return nil
}

func New(llmClient *openai.Client) (*MilvusService, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	milvusClient, err := client.NewClient(context.Background(), client.Config{
		Address: "localhost:19530",
	})
	if err != nil {
		fmt.Printf("Can't connect to Milvus: %v\n", err)
		return nil, err
	}

	return &MilvusService{
		llmClient:    llmClient,
		milvusClient: milvusClient,
	}, nil
}

func (s *MilvusService) createCollection(c *fiber.Ctx) error {
	err := s.milvusClient.CreateCollection(
		context.Background(), // ctx
		schema,
		2, // shardNum
	)
	if err != nil {
		log.Println("failed to create collection:", err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendString("Collection Created!")
}

func (s *MilvusService) RegisterRoutes(router fiber.Router) {
	router.Post("/messages", s.insertMessage)
	router.Get("/messages", s.vectorSearch)
	router.Post("/vectors", s.insertVector)
	router.Post("/embeddings", s.createEmbeddings)
	router.Delete("/collections/:name", s.deleteCollection)
	router.Get("/collections", s.listCollections)
	router.Get("/collections/:name", s.showCollection)
	router.Post("/collections", s.createCollection)
}

func (s *MilvusService) showCollection(c *fiber.Ctx) error {
	collDesc, err := s.milvusClient.DescribeCollection( // Return the name and schema of the collection.
		context.Background(),
		c.Params("name"),
	)
	if err != nil {
		log.Println("failed to check collection schema:", err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	log.Printf("%v\n", collDesc)

	return c.JSON(collDesc)
}

func (s *MilvusService) listCollections(c *fiber.Ctx) error {
	listColl, err := s.milvusClient.ListCollections(
		context.Background(), // ctx
	)
	if err != nil {
		log.Println("failed to list all collections", err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	log.Println(listColl)

	return c.JSON(listColl)
}

func (s *MilvusService) deleteCollection(c *fiber.Ctx) error {
	err := s.milvusClient.DropCollection(
		context.Background(), // ctx
		c.Params("name"),
	)
	if err != nil {
		log.Println("failed to delete collection", err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendString("Collection Deleted!")
}

func (s *MilvusService) createEmbeddings(c *fiber.Ctx) error {
	message := new(Message)

	if err := c.BodyParser(message); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	embeddingReq := openai.EmbeddingRequest{
		Input: message.MessageText,
		Model: openai.AdaEmbeddingV2,
	}

	resp, err := s.llmClient.CreateEmbeddings(context.Background(), embeddingReq)
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
}

func (s *MilvusService) insertVector(c *fiber.Ctx) error {
	message := new(Message)

	if err := c.BodyParser(message); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	vector, err := embed(message.MessageText, s.llmClient)
	if err != nil {
		fmt.Printf("Embedding error: %v\n", err)
	}

	messageColumn := entity.NewColumnVarChar("message", []string{message.MessageText})
	senderColumn := entity.NewColumnVarChar("sender", []string{"user"})
	messageVectorColumn := entity.NewColumnFloatVector("message_vector", 1536, [][]float32{vector})

	_, err = s.milvusClient.Insert(
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
	err = s.milvusClient.CreateIndex(
		context.Background(),
		"messages",
		"message_vector",
		idx,
		false,
	)
	if err != nil {
		log.Fatal("fail to create index:", err.Error())
	}

	err = s.milvusClient.LoadCollection(
		context.Background(),
		"messages",
		false,
	)
	if err != nil {
		log.Println("Failed to load collection:", err.Error())
		return c.SendString(err.Error())
	}

	return c.SendString("Data inserted and collection loaded in memory!")
}

func (s *MilvusService) insertMessage(c *fiber.Ctx) error {
	message := new(Message)

	if err := c.BodyParser(message); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	resp, err := s.llmClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: os.Getenv("OPENAI_MODEL_ID"),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: message.MessageText,
				},
			},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return c.SendString(err.Error())
	}

	err = insertVector(s.llmClient, s.milvusClient, c, message.MessageText, "user")
	if err != nil {
		fmt.Printf("Insert user message error in Vector DB: %v\n", err)
		return c.SendString(err.Error())
	}

	err = insertVector(s.llmClient, s.milvusClient, c, resp.Choices[0].Message.Content, "llm")
	if err != nil {
		fmt.Printf("Insert llm message error in Vector DB: %v\n", err)
		return c.SendString(err.Error())
	}

	fmt.Println(resp.Choices[0].Message.Content)

	return c.SendString(resp.Choices[0].Message.Content)
}

func (s *MilvusService) vectorSearch(c *fiber.Ctx) error {
	message := new(Message)
	if err := c.BodyParser(message); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	err := s.milvusClient.LoadCollection(
		context.Background(),
		"messages",
		false,
	)
	if err != nil {
		log.Println("Failed to load collection:", err.Error())
		return c.SendString(err.Error())
	}

	sp, _ := entity.NewIndexIvfFlatSearchParam( // NewIndex*SearchParam func
		10, // searchParam
	)

	opt := client.SearchQueryOptionFunc(func(option *client.SearchQueryOption) {
		option.Limit = 3
		option.Offset = 0
		option.ConsistencyLevel = entity.ClStrong
		option.IgnoreGrowing = false
	})

	vector, err := embed(message.MessageText, s.llmClient)
	if err != nil {
		fmt.Printf("Embedding error: %v\n", err)
		return err
	}

	searchResult, err := s.milvusClient.Search(
		context.Background(), // ctx
		"messages",           // CollectionName
		[]string{},           // partitionNames
		"",                   // expr
		[]string{"message"},  // outputFields
		[]entity.Vector{entity.FloatVector(vector)}, // vectors
		"message_vector", // vectorField
		entity.L2,        // metricType
		10,               // topK
		sp,               // sp
		opt,
	)
	if err != nil {
		log.Fatal("fail to search collection:", err.Error())
	}

	fmt.Printf("%#v\n", searchResult)

	for _, sr := range searchResult {
		fmt.Println(sr.IDs)
		fmt.Println(sr.Scores)
	}

	err = s.milvusClient.ReleaseCollection(
		context.Background(), // ctx
		"messages",           // CollectionName
	)
	if err != nil {
		log.Fatal("failed to release collection:", err.Error())
	}

	return c.SendString("Collection loaded in memory!")
}
