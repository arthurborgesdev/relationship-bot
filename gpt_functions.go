package main

import (
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var getProductsAndDate = openai.FunctionDefinition{
	Name:        "getProductsAndDate",
	Description: "Get products from user based on his queries and date of delivery",
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
			"date": {
				Type: "string",
				Description: `Hoje é ` + date + `. Então amanhã é ` + time.Now().AddDate(0, 0, 1).Format("2006-01-02") + `. Depois de amanhã é ` + time.Now().AddDate(0, 0, 2).Format("2006-01-02") + `. E assim por diante.
				Se o usuário não informar data, retorne a data de hoje. Exemplo: "Vou querer um juice de morango e um vape". Resposta: "` + date + `"`,
			},
			"time": {
				Type: "string",
				Description: `Retorne a hora informada pelo usuário no formato hh:mm. Exemplo: "10:00", "14:30". Retorne a hora nesse formato: "hh:mm" Se
				o usuário não infomar data, retorne "00:00". Exemplo: "Vou querer um juice de morango e um vape". Resposta: "00:00"`,
			},
		},
		Required: []string{"product", "flavor", "quantity", "date", "time"},
	},
}
