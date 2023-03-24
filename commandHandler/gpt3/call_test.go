package gpt3

import (
	"context"
	"fmt"
	"testing"
)

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

func TestGPTClient_PostWithStream(t *testing.T) {
	g := &GPTClient{
		Model: "gpt-3.5-turbo",
		Messages: []Message{{
			Role:    "user",
			Content: "请生成一段用Golang请求chatgpt的代码",
		}},
		Stream:    true,
		AsyncChan: make(chan string, 50),
	}
	go func() {
		text := ""
		for {
			select {
			case s, isClosed := <-g.AsyncChan:
				if !isClosed {
					break
				}
				text += s
				fmt.Println(text)
			}
		}
	}()
	g.PostWithStream(context.Background())
}
