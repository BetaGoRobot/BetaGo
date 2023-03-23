package gpt3

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/oliveagle/jsonpath"
)

var apiKey = os.Getenv("GPT_TOKEN")

// CreateChatCompletion
//
//	@param msg
//	@return message
//	@return err
func CreateChatCompletion(ctx context.Context, msg, authorID string) (message string, err error) {
	// if sessionID := GetExsistConversation(ctx, authorID); sessionID != "" {
	// } else {
	// 	// create a new conversation
	// }
	moderationRes := ModerationCheck(ctx, msg)
	if len(moderationRes) > 0 {
		return fmt.Sprintf("很遗憾, 您的对话无法被发送到OpenAI服务器, \n文本违反了以下条例: \n`%s`", strings.Join(moderationRes, "\t")), nil
	}
	gptClient := &GPTClient{Model: "gpt-3.5-turbo"}
	gptClient.SetContent(msg)
	message, err = gptClient.Post()
	return
}

func ModerationCheck(ctx context.Context, content string) (res []string) {
	var resp interface{}
	g := &struct {
		Input string `json:"input"`
	}{
		Input: content,
	}
	_ = requests.
		URL("https://api.openai.com/v1/moderations").
		Bearer(apiKey).
		BodyJSON(&g).
		ToJSON(&resp).
		Transport(&http.Transport{
			Proxy: http.ProxyURL(ParsedProxyURL),
		}).
		Fetch(ctx)
	isFlagged, _ := jsonpath.JsonPathLookup(resp, "$.results[0].flagged")
	if isFlagged.(bool) {
		categories, _ := jsonpath.JsonPathLookup(resp, "$.results[0].categories")
		for k, v := range categories.(map[string]interface{}) {
			if v.(bool) {
				res = append(res, k)
			}
		}
	}
	return
}

// func GetExsistConversation(ctx context.Context, authorID string) (sessionID string) {
// 	r := &utility.ChatContextRecord{}
// 	utility.GetDbConnection().
// 		Table("betago.chat_context_records").
// 		Where("user_id = %s", authorID).
// 		Find(r).Debug()
// 	return r.SessionID
// }
