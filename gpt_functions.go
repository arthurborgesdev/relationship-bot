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
			"list": {
				Type: "array",
				Description: `O usuário informará os vapes, pods, coils e juices que ele quer comprar.
				Ele pode informar a marca ou modelo destes.
				Exemplos: "SMOK Nord 2, SWAG Kit, SWAG PX80" no caso de PODs, "Freebase" no caso de Juices para Vapes, etc.
				Para cada item informado, retorne o item, o sabor, a quantidade e o volume.
				Se algum desses campos não for informado, retorne valor vazio.
				Exemplo: "Quero um juice de morango de 30ml". Resposta: "juice", "morango", "1", "30".
				Outro exemplo: "Quero um vape". Resposta: "vape", "", "1", "0".`,
				Items: &jsonschema.Definition{
					Type: "object",
					Properties: map[string]jsonschema.Definition{
						"product": {
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
							Type: "string",
							Description: `Retorne a quantidade de ml do produto em numeral.
							"Exemplo: "Vou querer um juice de morango de 30ml". Resposta: "30".
							"Retorne "0" se o usuário não informar o volume. Exemplo: "Vou querer um juice de morango". Resposta: "0"`,
							Enum: []string{"0", "15", "30", "60", "100"},
						},
					},
					Required: []string{"product", "flavor", "quantity", "volume"},
				},
			},
			"date": {
				Type: "string",
				Description: `Hoje é ` + date + `. Então amanhã é ` + time.Now().AddDate(0, 0, 1).Format("2006-01-02") + `. Depois de amanhã é ` + time.Now().AddDate(0, 0, 2).Format("2006-01-02") + `. E assim por diante.
				Se o usuário não informar data, retorne a data de hoje. Exemplo: "Vou querer um juice de morango e um vape". Resposta: "` + date + `"`,
			},
			"time": {
				Type: "string",
				Description: `Retorne a hora informada pelo usuário no formato hh:mm.
				Use ":" para separar hora de minutos.
				Exemplo: "Vou buscar aí amanhã as 14h30", retorne: "14:30".
				Exemplo: "Vou buscar aí amanhã as 13h10", retorne: "13:10". 
				Retorne a hora nesse formato: "hh:mm" Se o usuário não infomar hora, retorne "". 
				Exemplo: "Vou querer um juice de morango e um vape". Resposta: ""`,
			},
		},
	},
}
