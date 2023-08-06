package main

import (
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
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
	Quantity int    `json:"quantity"`
	Day      int    `json:"day"`
	Time     string `json:"time"`
}
