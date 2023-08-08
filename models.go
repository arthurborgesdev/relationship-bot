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
	Products []struct {
		Item     string `json:"item"`
		Flavor   string `json:"flavor"`
		Quantity int    `json:"quantity"`
		Volume   string `json:"volume"`
	} `json:"products"`

	Date string `json:"date"`
	Time string `json:"time"`
}
