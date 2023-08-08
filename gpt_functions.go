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
				Type: "object",
				Properties: map[string]jsonschema.Definition{
					"item": {
						Type: "string",
						Description: `O usuário informará os vapes, pods, coils e juices que ele quer comprar. 
						Ele pode informar a marca ou modelo destes. 
						Exemplos: "SMOK Nord 2, SWAG Kit, SWAG PX80" no caso de PODs, "Freebase" no caso de Juices para Vapes, etc.
						Outro exemplo: "Vou querer um pod SWAG Kit de morango." Retorne: "SWAG Kit"`,
					},
					"flavor": {
						Type: "string",
						Description: `Se o usuário informar que quer comprar um Juice como produto, ele poderá informar os sabores.
						Aqui os sabores podem ser tanto de juices quanto de nicsalts.
						Juice é para Vape e Nicsalt é para POD. 
						Exemplo: "Freebase de morango", "Nicsalt de uva". Salve apenas os sabores.
						Exemplo: "Vou querer um pode SWAG Kit de morango." Retorne: "morango"
						Se o usuário não informar item "juice" ou "nicsalt", retorne valor vazio. 
						Exemplo: "Vou querer um vape e um pod". Resposta: ""
						Exemplo: "Amanhã não é um bom dia pra mim, mas vou buscar próxima segunda-feira as 14h00". Resposta: ""`,
					},
					"quantity": {
						Type: "integer",
						Description: `O usuário informará a quantidade de itens que ele quer comprar. 
						Ele pode informar diferentes quantidades, para cada item diferente. 
						Exemplos: "2 Freebase de morango", "3 vapes de menta". Retorne apenas a quantidade.
						Exemplo: "Vou querer um juice de morango". Resposta: "1"`,
					},
					"volume": {
						Type: "integer",
						Description: `O usuário informará a quantidade volumétrica dos juices 
						e nicsalts (que variam de 15 a 100 ml) que ele quer comprar.
						Ele pode informar volumes separados pela milimetragem ou não. Salve apenas os números. 
						Exemplos: se informar "30ml" ou "60ml", salve apenas "30" ou "60".
						Se o usuário não informar valor volumétrico, retorne 0. 
						Se ele não informar nada relacionado ao volume do item, retorne valor 0.
						Exemplo: "Vou querer um pod SWAG Kit de morango." Resposta: "0",
						"Vou querer um pod SWAG Kit de morango. Resposta: "0"
						"Vou querer um juice de morango". Resposta: "0"`,
						Enum: []string{"0", "15", "30", "60", "100"},
					},
				},
				Required: []string{"product", "flavor", "quantity", "volume"},
			},
			/*
				"product": {
					Type: "string",
					Description: `O usuário informará a lista de vapes e pods que ele quer comprar. Ele pode informar a marca, o modelo,
					a quantidade, o sabor e outros dados referentes a produtos de cigarros eletônicos. Exemplos: "Freebase", "Menta".`,
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
			*/
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
		Required: []string{"product", "date", "time"},
	},
}
