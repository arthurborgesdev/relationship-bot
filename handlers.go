package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	openai "github.com/sashabaranov/go-openai"
)

func (s *LLMService) RegisterRoutes(router fiber.Router) {
	router.Post("/messages", s.chat)
	router.Get("/messagesdb", s.getMessagesRelational)
	router.Post("/messagesdb", s.insertMessageRelational)
	router.Get("/productsdb", s.getProductsRelational)
	router.Post("/productsdb", s.insertProductsRelational)

	router.Delete("/productsdb/:id", s.deleteProduct)
}

var weekday time.Weekday
var weekdayStr string
var date string

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

	weekday = time.Now().Weekday()
	weekdayStr = weekday.String()
	date = time.Now().Format("2006-01-02")

	resp, err := s.llmClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     os.Getenv("OPENAI_MODEL_ID"),
			Messages:  chatMessage,
			Functions: []openai.FunctionDefinition{getProductsAndDate},
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

	result := s.db.Where("product = ? OR flavor = ? OR quantity = ?", arguments.Item.Product, arguments.Item.Flavor, arguments.Item.Quantity).First(&product)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error,
		})
	}

	fmt.Println(product)

	return c.JSON(product)
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

func (s *LLMService) insertProductsRelational(c *fiber.Ctx) error {
	product := new(Products)

	if err := c.BodyParser(product); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	s.db.Create(&Products{Product: strings.ToLower(product.Product), Flavor: strings.ToLower(product.Flavor), Quantity: product.Quantity})

	return c.SendString("Produto salvo com sucesso!")
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
