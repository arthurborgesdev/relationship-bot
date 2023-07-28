package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	openai "github.com/sashabaranov/go-openai"
)

type Message struct {
	gorm.Model
	Content string `json:"content"`
	Role    string `json:"role"`
}

type LlmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LlmService struct {
	llmClient *openai.Client
	db        *gorm.DB
}

func getMessages(s *LlmService) ([]LlmMessage, error) {
	var messages []Message
	result := s.db.Find(&messages)

	if result.Error != nil {
		return nil, result.Error
	}

	fmt.Println(result)

	var LlmMessages []LlmMessage
	for _, message := range messages {
		LlmMessages = append(LlmMessages, LlmMessage{Role: message.Role, Content: message.Content})
	}

	return LlmMessages, nil
}

func (s *LlmService) RegisterRoutes(router fiber.Router) {
	router.Post("/messagesdb", s.insertMessageRelational)
	router.Get("/messagesdb", s.getMessagesRelational)
	router.Post("/messages", s.chat)
}

func New(db *gorm.DB) (*LlmService, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	llmClient := openai.NewClient(os.Getenv("OPENAI_AUTH_TOKEN"))

	return &LlmService{
		db:        db,
		llmClient: llmClient,
	}, nil
}

func (s *LlmService) chat(c *fiber.Ctx) error {
	message := new(Message)

	if err := c.BodyParser(message); err != nil {
		fmt.Printf("Bodyparsing error: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	previousChat, err := getMessages(s)
	if err != nil {
		fmt.Printf("Get previous chat error: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// Convert previousChat to a slice of openai.ChatCompletionMessage
	var chatHistory []openai.ChatCompletionMessage
	for _, chatMessage := range previousChat {
		chatHistory = append(chatHistory, openai.ChatCompletionMessage{
			Role:    chatMessage.Role,
			Content: chatMessage.Content,
		})
	}

	// Append the user's new message to chatHistory
	chatHistory = append(chatHistory, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message.Content,
	})

	// Include the system message at the beginning
	systemMessage := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: `Você é um chatbot que auxilia profissionais liberais a agendarem suas consultas,
		devendo responder prontamente e com polidez e ânimo a perguntas sobre o serviço prestado. Responda
		com respostas concisas, curtas e objetivas. As pessoas irão fazer perguntas sobre serviços de medicina,
		odontologia e nutrição. Responda com respostas curtas e objetivas. Para perguntas que fujam do escopo
		mencionado, apenas responda com polidez dizendo que não está apto a responder perguntas de áreas que não
		sejam medicina, odontologia e nutrição.`,
	}

	// Add system message at the beginning of the chat
	chatHistory = append([]openai.ChatCompletionMessage{systemMessage}, chatHistory...)

	resp, err := s.llmClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    os.Getenv("OPENAI_MODEL_ID"),
			Messages: chatHistory,
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return c.SendString(err.Error())
	}

	s.db.Create(&Message{Content: message.Content, Role: openai.ChatMessageRoleUser})
	s.db.Create(&Message{Content: resp.Choices[0].Message.Content, Role: openai.ChatMessageRoleAssistant})

	fmt.Println(resp.Choices[0].Message.Content)

	return c.SendString(resp.Choices[0].Message.Content)
}

func (s *LlmService) insertMessageRelational(c *fiber.Ctx) error {
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
					Content: message.Content,
				},
			},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return c.SendString(err.Error())
	}

	s.db.Create(&Message{Content: message.Content, Role: openai.ChatMessageRoleUser})
	s.db.Create(&Message{Content: resp.Choices[0].Message.Content, Role: openai.ChatMessageRoleAssistant})

	fmt.Println(resp.Choices[0].Message.Content)

	return c.SendString(resp.Choices[0].Message.Content)
}

func (s *LlmService) getMessagesRelational(c *fiber.Ctx) error {
	var messages []Message
	result := s.db.Find(&messages)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error,
		})
	}

	fmt.Println(result)

	return c.JSON(messages)
}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&Message{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	app := fiber.New()

	LlmService, err := New(db)
	if err != nil {
		log.Fatalf("Failed to create LLM service: %v", err)
	}

	// MilvusService, err := vectordb.New(LlmService.llmClient)
	// if err != nil {
	// 	log.Fatalf("Failed to create Milvus service: %v", err)
	// }

	LlmService.RegisterRoutes(app)
	// MilvusService.RegisterRoutes(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	app.Listen(":3000")
}
