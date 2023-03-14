package gpt3

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/carlmjohnson/requests"
	"github.com/oliveagle/jsonpath"
)

var (
	proxyURL       = os.Getenv("PRIVATE_PROXY")
	ParsedProxyURL *url.URL
	ec             utility.ErrorCollector
)

func init() {
	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		panic(err)
	}
	ParsedProxyURL = parsedProxyURL

}

// Message 要发送的内容
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GPTClient GPT的请求体
type GPTClient struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// SetContent 设置内容
//
//	@receiver g
//	@param s
func (g *GPTClient) SetContent(s string) {
	g.Messages = []Message{{
		Role:    "user",
		Content: s,
	}}
}
func (g *GPTClient) GetModels() (msg string, err error) {
	var resp interface{}
	err = requests.URL("https://api.openai.com/v1/models").
		Bearer(apiKey).
		Transport(&http.Transport{
			Proxy: http.ProxyURL(ParsedProxyURL),
		}).
		ToJSON(&resp).
		Fetch(context.Background())
	print(resp)
	return
}

// Post 发送请求
//
//	@receiver g
func (g *GPTClient) Post() (msg string, err error) {
	var resp interface{}
	err = requests.
		URL("https://api.openai.com/v1/chat/completions").
		Bearer(apiKey).
		BodyJSON(&g).
		ToJSON(&resp).
		Transport(&http.Transport{
			Proxy: http.ProxyURL(ParsedProxyURL),
		}).
		Fetch(context.Background())
	if err != nil {
		return
	}
	res, err := jsonpath.JsonPathLookup(resp, "$.choices[0].message.content")
	if err != nil {
		return
	}
	promptTokens, err := jsonpath.JsonPathLookup(resp, "$.usage.prompt_tokens")
	ec.Collect(err)
	completionTokens, err := jsonpath.JsonPathLookup(resp, "$.usage.completion_tokens")
	ec.Collect(err)
	totalTokens, err := jsonpath.JsonPathLookup(resp, "$.usage.total_tokens")
	ec.Collect(err)
	err = ec.CheckError()

	msg = fmt.Sprintf("%s\n---\n本次请求消耗: \n`prompt_tokens`: **%d**\n`completion_tokens`: **%d**\n`total_tokens`: **%d**", strings.Trim(res.(string), "\n"), int(promptTokens.(float64)), int(completionTokens.(float64)), int(totalTokens.(float64)))
	return
}
