package gpt3

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/spyzhov/ajson"
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
	a := &GPTClient{
		Model: "gpt-3.5-turbo",
		Messages: []Message{{
			Role:    "user",
			Content: "你好",
		}},
		Stream:    true,
		AsyncChan: make(chan string, 50),
	}
	r, err := json.Marshal(&a)
	if err != nil {
		log.Println(err.Error())
	}
	fmt.Println(string(r))
	resp, err := betagovar.HttpClientWithProxy.R().
		SetAuthScheme("Bearer").
		SetHeader("Content-Type", "application/json").
		SetAuthToken(apiKey).
		SetBody(r).
		SetDoNotParseResponse(true).
		Post("https://api.openai.com/v1/chat/completions")
	if err != nil {
		log.Println(err.Error())
	}
	reader := bufio.NewReader(resp.RawBody())
	var s string
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		lineJSON := bytes.TrimLeft(line, "data: ")
		res, err := ajson.JSONPath(lineJSON, "$..content")
		if err != nil {
			continue
		}
		if len(res) > 0 {
			r, err := res[0].GetString()
			if err != nil {
				log.Println(err.Error())
			}
			s += r
		}
	}
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
