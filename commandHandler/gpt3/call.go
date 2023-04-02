package gpt3

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/carlmjohnson/requests"
	"github.com/oliveagle/jsonpath"
	"github.com/spyzhov/ajson"
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
	Model      string      `json:"model"`
	Messages   []Message   `json:"messages"`
	Stream     bool        `json:"stream,omitempty"`
	AsyncChan  chan string `json:"-"`
	StopChan   chan string `json:"-"`
	StopAuthor string      `json:"-"`
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
	headers := make(map[string][]string)
	startTime := time.Now()
	err = requests.
		URL("https://api.openai.com/v1/chat/completions").
		Bearer(apiKey).
		BodyJSON(&g).
		ToJSON(&resp).
		CopyHeaders(headers).
		Transport(&http.Transport{
			Proxy: http.ProxyURL(ParsedProxyURL),
		}).
		Fetch(context.Background())
	endTime := time.Now()
	if err != nil {
		return
	}
	createTime := time.Unix(int64(resp.(map[string]interface{})["created"].(float64)), 0)
	sendingTime := createTime.Sub(startTime).Abs()
	processTime, _ := time.ParseDuration(headers["Openai-Processing-Ms"][0] + "ms")
	returnTime := endTime.Sub(createTime) - processTime

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
			"`processing_time_cost`: **%s**",
			"`return_time_cost`: **%s**",
			"本次请求消耗: ",
			"`prompt_tokens`: **%d**=￥%f",
			"`completion_tokens`: **%d**=￥%f",
			"`total_tokens`: **%d**=￥%f",
		}, "\n"),
		strings.Trim(res.(string), "\n"),
		sendingTime.Round(time.Millisecond*100).String(),
		processTime.Round(time.Millisecond*100).String(),
		returnTime.Round(time.Millisecond*100).String(),
		int(promptTokens.(float64)),
		promptTokens.(float64)*0.01*0.001,
		int(completionTokens.(float64)),
		completionTokens.(float64)*0.01*0.001,
		int(totalTokens.(float64)),
		totalTokens.(float64)*0.01*0.001)
	return
}

// PostWithStream  流式请求
//
//	@receiver g
//	@param ctx
//	@return executeMsg
//	@return err
func (g *GPTClient) PostWithStream(ctx context.Context) (err error) {
	jsonBody, err := json.Marshal(&g)
	if err != nil {
		return
	}
	resp, err := betagovar.HttpClientWithProxy.R().
		SetAuthScheme("Bearer").
		SetHeader("Content-Type", "application/json").
		SetAuthToken(apiKey).
		SetBody(jsonBody).
		SetDoNotParseResponse(true).
		Post("https://api.openai.com/v1/chat/completions")
	if err != nil {
		return
	}
	reader := bufio.NewReader(resp.RawBody())
	for {
		var line []byte
		line, err = reader.ReadBytes('\n')
		if err == io.EOF {
			err = nil
			break
		}
		lineJSON := bytes.TrimLeft(line, "data: ")
		if bytes.Contains(lineJSON, []byte("maximum")) {
			err = fmt.Errorf(
				`		你的会话已经达到了GPT-3.5-turbo的4096 token上限, 请使用指令` + "`" + ".gpt RESET" + "`" + `来重置会话。
		请注意: 为了保护你的隐私,
***目前***我们不会保存你与ChatGPT交流的历史会话。在后续BetaGo上线OAuth鉴权后, 你可能需要同意[EULA协议]来允许我们保存你的历史会话。
`,
			)
			break
		}
		res, err := ajson.JSONPath(lineJSON, "$..content")
		if err != nil {
			continue
		}
		if len(res) > 0 {
			r, err := res[0].GetString()
			if err != nil {
				log.Println(err.Error())
			}
			select {
			case g.StopAuthor = <-g.StopChan:
				close(g.AsyncChan)
				close(g.StopChan)
				return nil
			default:
				g.AsyncChan <- r
			}
		}
	}
	close(g.AsyncChan)
	close(g.StopChan)
	return
}
