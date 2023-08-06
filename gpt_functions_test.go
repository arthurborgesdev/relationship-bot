package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"encoding/json"

	"time"

	"github.com/sashabaranov/go-openai"
)

func gptCall(message string) (Arguments, error) {
	client := openai.NewClient(os.Getenv("OPENAI_AUTH_TOKEN"))

	var messages []openai.ChatCompletionMessage
	chatMessage := append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	weekday = time.Now().Weekday()
	weekdayStr = weekday.String()
	date = time.Now().Format("2006-01-02")

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     os.Getenv("OPENAI_MODEL_ID"),
			Messages:  chatMessage,
			Functions: []openai.FunctionDefinition{getProductsAndDate},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
	}

	var arguments Arguments

	incommingArguments := ""

	if resp.Choices[0].Message.Content == "" {
		incommingArguments = resp.Choices[0].Message.FunctionCall.Arguments
	} else {
		incommingArguments = resp.Choices[0].Message.Content
	}

	json.Unmarshal([]byte(incommingArguments), &arguments)

	fmt.Printf("ChatCompletion response: %v\n", arguments) // Parsear e mostrar em formato JSON
	fmt.Printf("Date: %v\n", arguments.Date)
	fmt.Printf("Day: %v\n", arguments.Day)
	fmt.Printf("Time: %v\n", arguments.Time)
	fmt.Printf("Product: %s\n", arguments.Product)
	fmt.Printf("Flavor: %s\n", arguments.Flavor)
	fmt.Printf("Quantity: %d\n", arguments.Quantity)

	argumentsJSON, err := json.Marshal(arguments)
	if err != nil {
		fmt.Printf("Error marshaling arguments: %v\n", err)
		return arguments, err
	}

	// Print the JSON string
	fmt.Printf("ChatCompletion response in JSON format: %s\n", argumentsJSON)
	return arguments, nil
}

func TestGPTFunction(t *testing.T) {
	message := "Vou querer um juice de morango e um vape. Vou buscar aí amanhã as 14h00"

	arguments, err := gptCall(message)
	if err != nil {
		fmt.Printf("Error calling GPT: %v\n", err)
		return
	}

	if arguments.Product != "vape" {
		t.Errorf("Product is not correct: %s", arguments.Product)
	}

	if arguments.Flavor != "morango" {
		t.Errorf("Flavor is not correct: %s", arguments.Flavor)
	}

	if arguments.Quantity != 1 {
		t.Errorf("Quantity is not correct: %d", arguments.Quantity)
	}

	if arguments.Day != 1 {
		t.Errorf("Day is not correct: %d", arguments.Day)
	}

	if arguments.Time != "14:00" {
		t.Errorf("Time is not correct: %s", arguments.Time)
	}
}

func TestGPTFunctionDates(t *testing.T) {
	message := "Vou buscar aí amanhã as 14h00"

	arguments, err := gptCall(message)
	if err != nil {
		fmt.Printf("Error calling GPT: %v\n", err)
		return
	}

	if arguments.Date != "2023-08-07" {
		t.Errorf("Date is not correct: %s", arguments.Date)
	}
}
