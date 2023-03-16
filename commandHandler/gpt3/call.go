package gpt3

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

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
	startTime := time.Now()
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
	endTime := time.Now()
	createTime := time.Unix(int64(resp.(map[string]interface{})["created"].(float64)), 0)
	sendingTime := createTime.Sub(startTime)
	afterwardsTime := endTime.Sub(createTime)

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

	msg = fmt.Sprintf(strings.Join(
		[]string{
			"%s",
			"---",
			"请求耗时:",
			"`sending_time_cost`: **%s**",
			"`afterwards_time_cost`: **%s**",
			"本次请求消耗: ",
			"`prompt_tokens`: **%d**=￥%f",
			"`completion_tokens`: **%d**=￥%f",
			"`total_tokens`: **%d**=￥%f",
		}, "\n"),
		strings.Trim(res.(string), "\n"),
		sendingTime.Round(time.Millisecond*100).String(),
		afterwardsTime.Round(time.Millisecond*100).String(),
		sendingTime.Round(time.Millisecond*100).String(),
		int(promptTokens.(float64)),
		promptTokens.(float64)*0.01*0.001,
		int(completionTokens.(float64)),
		completionTokens.(float64)*0.01*0.001,
		int(totalTokens.(float64)),
		totalTokens.(float64)*0.01*0.001)
	return
}
