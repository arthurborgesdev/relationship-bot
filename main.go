package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sashabaranov/go-openai/jsonschema"

	"encoding/json"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	openai "github.com/sashabaranov/go-openai"
)

type Message struct {
	gorm.Model
	Content string `json:"content"`
	Role    string `json:"role"`
}

type Products struct {
	gorm.Model
	Product  string `json:"product"`
	Flavor   string `json:"flavor"`
	Quantity int    `json:"quantity"`
}

type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMService struct {
	llmClient *openai.Client
	db        *gorm.DB
}

type Arguments struct {
	Date     string `json:"date"`
	Product  string `json:"product"`
	Flavor   string `json:"flavor"`
	Quantity string `json:"quantity"`
}

func getMessages(s *LLMService) ([]LLMMessage, error) {
	var messages []Message
	result := s.db.Find(&messages)

	if result.Error != nil {
		return nil, result.Error
	}

	fmt.Println(result)

	var LLMMessages []LLMMessage
	for _, message := range messages {
		LLMMessages = append(LLMMessages, LLMMessage{Role: message.Role, Content: message.Content})
	}

	return LLMMessages, nil
}

func (s *LLMService) RegisterRoutes(router fiber.Router) {
	router.Post("/messagesdb", s.insertMessageRelational)
	router.Post("/productsdb", s.insertProductsRelational)
	router.Get("/productsdb", s.getProductsRelational)
	router.Delete("/productsdb/:id", s.deleteProduct)
	router.Get("/messagesdb", s.getMessagesRelational)
	router.Post("/messages", s.chat)
}

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

func chatHistory(c *fiber.Ctx, s *LLMService, message *Message) ([]openai.ChatCompletionMessage, error) {
	previousChat, err := getMessages(s)
	if err != nil {
		fmt.Printf("Get previous chat error: %v\n", err)
		return nil, err
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

	return chatHistory, nil
}

func (s *LLMService) insertMessageRelational(c *fiber.Ctx) error {
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

func (s *LLMService) insertProductsRelational(c *fiber.Ctx) error {
	product := new(Products)

	if err := c.BodyParser(product); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	s.db.Create(&Products{Product: strings.ToLower(product.Product), Flavor: strings.ToLower(product.Flavor), Quantity: product.Quantity})

	return c.SendString("Produto salvo com sucesso!")
}

func (s *LLMService) getMessagesRelational(c *fiber.Ctx) error {
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

func (s *LLMService) getProductsRelational(c *fiber.Ctx) error {
	var products []Products
	result := s.db.Find(&products)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error,
		})
	}

	fmt.Println(result)

	return c.JSON(products)
}

func (s *LLMService) deleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	result := s.db.Unscoped().Delete(&Products{}, id)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error,
		})
	}

	if result.RowsAffected > 0 {
		fmt.Println("Record deleted successfully.")
		return c.SendString("Record deleted successfully.")
	} else {
		fmt.Println("No record found with the provided ID.")
		return c.SendString("No record found with the provided ID.")
	}
}

func (s *LLMService) chat(c *fiber.Ctx) error {
	message := new(Message)

	if err := c.BodyParser(message); err != nil {
		fmt.Printf("Bodyparsing error: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// chatHistory, err := chatHistory(c, s, message)
	// if err != nil {
	// 	fmt.Printf("chatHistory fetching resulted in error: %v", err)
	// 	c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	// }

	var messages []openai.ChatCompletionMessage
	chatMessage := append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message.Content,
	})

	resp, err := s.llmClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     os.Getenv("OPENAI_MODEL_ID"),
			Messages:  chatMessage,
			Functions: []openai.FunctionDefinition{getDateTime, getProductsList},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return c.SendString(err.Error())
	}

	var arguments Arguments

	incommingArguments := ""

	if resp.Choices[0].Message.Content == "" {
		incommingArguments = resp.Choices[0].Message.FunctionCall.Arguments
	} else {
		incommingArguments = resp.Choices[0].Message.Content
	}

	// Save entries in Message DB to build a history -> Useful for medical scenario (not vape)
	s.db.Create(&Message{Content: message.Content, Role: openai.ChatMessageRoleUser})
	s.db.Create(&Message{Content: incommingArguments, Role: openai.ChatMessageRoleAssistant})

	json.Unmarshal([]byte(incommingArguments), &arguments)

	var product Products
	result := s.db.Where("product = ? OR flavor = ? OR quantity = ?", arguments.Product, arguments.Flavor, arguments.Quantity).First(&product)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error,
		})
	}

	fmt.Println(result)

	return c.JSON(result)
}

var getDateTime = openai.FunctionDefinition{
	Name:        "setScheduleDate",
	Description: "Set the schedule date for the user",
	Parameters: jsonschema.Definition{
		Type: "object",
		Properties: map[string]jsonschema.Definition{
			"date": {
				Type: "string",
				Description: `O usuário poderá informar a data que ele deseja agendar a consulta. Ele pode
				informar somente o dia, ou a hora, ou o dia da semana, ou dois ou três desses dados ao mesmo tempo
				em vários formatos diferentes, por extenso ou em numeral. Exemplos: "segunda-feira", "segunda", "seg".
				Pegue esses dados e armazene inicialemtne`,
			},
		},
		Required: []string{"date"},
	},
}

var getProductsList = openai.FunctionDefinition{
	Name:        "getProductsList",
	Description: "Get products from user based on his queries",
	Parameters: jsonschema.Definition{
		Type: "object",
		Properties: map[string]jsonschema.Definition{
			"product": {
				Type: "string",
				Description: `O usuário informará a lista de vapes e pods que ele quer comprar. Ele pode informar a marca, o modelo,
				a quantidade, o sabor e outros dados referentes a produtos de cigarros eletônicos. Exemplos: "Freebase", "Menta".
				Salve esses produtos separados por vírgula`,
			},
			"flavor": {
				Type: "string",
				Description: `O usuário informará os sabores de juices que ele quer comprar. Aqui os sabores podem ser tanto de vapes quanto
				de pods. Exemplo: "Freebase de morango", "Nicsalt de uva". Salve apenas os sabores separados por vírgula`,
			},
			"quantity": {
				Type: "integer",
				Description: `O usuário informará a quantidade de vapes e pods que ele quer comprar. Ele pode informar diferentes quantidades,
				para cada item diferente. Exemplos: "2 Freebase de morango", "3 vapes de menta". Retorne apenas a quantidade separada por vírgula`,
			},
		},
		Required: []string{"product", "flavor", "quantity"},
	},
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

	err = db.AutoMigrate(&Products{})
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
