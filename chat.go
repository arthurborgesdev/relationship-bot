package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	openai "github.com/sashabaranov/go-openai"
)

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
