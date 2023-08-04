package main

import (
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

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
