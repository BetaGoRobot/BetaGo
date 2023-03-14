package gpt3

import "testing"

func TestGPTParams_Post(t *testing.T) {

	g := &GPTClient{
		Model: "gpt-3.5-turbo",
		Messages: []Message{{
			Role:    "user",
			Content: "你好啊",
		}},
	}
	g.Post()
	// g.GetModels()
}
