package gpt3

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/oliveagle/jsonpath"
)

var (
	proxyURL       = os.Getenv("PRIVATE_PROXY")
	ParsedProxyURL *url.URL
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
	msg = strings.Trim(res.(string), "\n")
	return
}
