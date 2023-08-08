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

	fmt.Printf("ChatCompletion response: %v\n", arguments)
	fmt.Printf("Date: %v\n", arguments.Date)
	fmt.Printf("Time: %v\n", arguments.Time)

	for i, product := range arguments.Products {
		fmt.Printf("Product %d:\n", i+1)
		fmt.Printf("\tItem: %s\n", product.Item)
		fmt.Printf("\tFlavor: %s\n", product.Flavor)
		fmt.Printf("\tQuantity: %d\n", product.Quantity)
		fmt.Printf("\tVolume: %s\n", product.Volume)
	}

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
	fmt.Printf("TestMessage: %s\n", message)

	arguments, err := gptCall(message)
	if err != nil {
		fmt.Printf("Error calling GPT: %v\n", err)
		return
	}

	foundVape := false
	foundJuice := false
	for _, product := range arguments.Products {
		if product.Item == "vape" {
			foundVape = true
			if product.Quantity != 1 {
				t.Errorf("Quantity is not correct for vape: %d", product.Quantity)
			}
		}
		if product.Item == "juice" {
			foundJuice = true
			if product.Flavor != "morango" {
				t.Errorf("Flavor is not correct for juice: %s", product.Flavor)
			}
		}
	}

	if !foundVape {
		t.Errorf("Vape product not found")
	}
	if !foundJuice {
		t.Errorf("Juice product not found")
	}

	if arguments.Date != time.Now().Add(time.Hour*24).Format("2006-01-02") {
		t.Errorf("Date is not correct: %s", arguments.Date)
	}

	if arguments.Time != "14:00" {
		t.Errorf("Time is not correct: %s", arguments.Time)
	}
}
func TestGPTFunctionDates(t *testing.T) {
	message := "Vou buscar aí amanhã às 13h10"
	fmt.Printf("TestMessage: %s\n", message)

	arguments, err := gptCall(message)
	if err != nil {
		fmt.Printf("Error calling GPT: %v\n", err)
		return
	}

	if arguments.Date != time.Now().Add(time.Hour*24).Format("2006-01-02") {
		t.Errorf("Date is not correct: %s", arguments.Date)
	}

	if arguments.Time != "13:10" {
		t.Errorf("Time is not correct: %s", arguments.Time)
	}
}

func TestGPTFunctionWeekdays(t *testing.T) {
	message := "Amanhã não é um bom dia pra mim, mas vou buscar próxima segunda-feira às 14h25"
	fmt.Printf("TestMessage: %s\n", message)

	arguments, err := gptCall(message)
	if err != nil {
		fmt.Printf("Error calling GPT: %v\n", err)
		return
	}

	if arguments.Date != time.Now().Add(time.Hour*24*6).Format("2006-01-02") {
		t.Errorf("Date is not correct: %s", arguments.Date)
	}

	if arguments.Time != "14:25" {
		t.Errorf("Time is not correct: %s", arguments.Time)
	}
}

func TestGPTFunctionVolume(t *testing.T) {
	message := "Vou querer um juice de morango de 40ml"
	fmt.Printf("TestMessage: %s\n", message)

	arguments, err := gptCall(message)
	if err != nil {
		fmt.Printf("Error calling GPT: %v\n", err)
		return
	}

	product := arguments.Products[0]

	if product.Item != "juice" {
		t.Errorf("Item is not correct: %s", product.Item)
	}

	if product.Flavor != "morango" {
		t.Errorf("Flavor is not correct: %s", product.Flavor)
	}

	if product.Volume != "40" {
		t.Errorf("Volume is not correct: %s", product.Volume)
	}
}
