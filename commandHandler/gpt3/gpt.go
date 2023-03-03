package gpt3

import (
	"time"

	"github.com/chatgp/gpt3"
)

var apiKey = "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"

var cli = getClient()

func getClient() (cli *gpt3.Client) {
	// new gpt-3 client
	cli, _ = gpt3.NewClient(&gpt3.Options{
		ApiKey:  apiKey,
		Timeout: 30 * time.Second,
		Debug:   true,
	})
	return
}

func CreateChatCompletion(msg string) (message string) {
	uri := "/v1/chat/completions"
	params := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]interface{}{
			{"role": "user", "content": msg},
		},
	}

	res, err := cli.Post(uri, params)
	if err != nil {
		panic(err)
	}
	message = res.Get("choices.0.message.content").String()

	return
}
