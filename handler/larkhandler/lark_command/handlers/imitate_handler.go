package handlers

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	"github.com/defensestation/osquery"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func ImitateHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	quoteList := data.Event.Message.Mentions
	if len(quoteList) == 1 {
		historyMsg := "- " + strings.Join(SearchByUserID(
			*quoteList[0].Id.OpenId, 50), "\n- ")
		latestMsg := "- " + strings.Join(SearchExcludeUserID(
			*quoteList[0].Id.OpenId, *data.Event.Message.ChatId, 15,
		), "\n- ")
		sysPrompt := `# 角色
你是一个擅长模仿别人说话语气的人，你言简意赅，不会有很多的修饰语

# 历史发言参考：
%s

# 流程
1. 综合上面的历史发言参考，模仿他的语气
2. 基于输入的近期文本，判断是否需要回复
3. 如果需要回复，按照他的语气进行回复

# 限制
1. 回复内容和历史对话的平均长度基本一致
2. 禁止拼接历史对话
3. 回复的文本需要跟最近一次输入存在关联关系
4. 仅回复一条
`
		userPrompt := `输入:
%s
请给出模仿的输出:`
		sysPrompt = fmt.Sprintf(sysPrompt, historyMsg)
		userPrompt = fmt.Sprintf(userPrompt, latestMsg)
		span.SetAttributes(attribute.String("sys_prompt", sysPrompt))
		span.SetAttributes(attribute.String("user_prompt", userPrompt))
		res, err := doubao.SingleChat(ctx, sysPrompt, userPrompt)
		span.SetAttributes(attribute.String("res", res))
		fmt.Println(sysPrompt, userPrompt)
		if err != nil {
			return err
		}
		res = strings.Trim(res, "\n")
		err = larkutils.CreateMsgText(ctx, res, *data.Event.Message.MessageId, *data.Event.Message.ChatId)
		if err != nil {
			return err
		}
	}

	return
}

type Mention struct {
	Key string `json:"key"`
	ID  struct {
		UserID  string `json:"user_id"`
		OpenID  string `json:"open_id"`
		UnionID string `json:"union_id"`
	} `json:"id"`
	Name      string `json:"name"`
	TenantKey string `json:"tenant_key"`
}
type MessageDoc struct {
	Mentions   string `json:"mentions"`
	RawMessage string `json:"raw_message"`
	CreateTime string `json:"create_time"`
}

func SearchByUserID(UserID string, size uint64) (messageList []string) {
	query := osquery.Search().
		Query(
			osquery.Bool().Must(
				osquery.Term("user_id", UserID),
				osquery.Bool().MustNot(
					osquery.Prefix("message_str", "/"),
				),
			),
		).SourceIncludes("raw_message", "mentions", "create_time").Size(size).Sort("create_time", "desc")
	resp, err := opensearchdal.SearchData(context.Background(), "lark_msg_index", query)
	if err != nil {
		panic(err)
	}
	messageList = []string{}
	for _, hit := range resp.Hits.Hits {
		res := &MessageDoc{}
		b, _ := hit.Source.MarshalJSON()
		err = sonic.ConfigStd.Unmarshal(b, res)
		if err != nil {
			continue
		}
		mentions := make([]Mention, 0)

		if res.Mentions != "null" {
			err = sonic.UnmarshalString(res.Mentions, &mentions)
			if err != nil {
				continue
			}
		}

		r := replaceMention(res.RawMessage, mentions)
		if r != "" {
			messageList = append(messageList, r)
		}
	}
	slices.Reverse(messageList)
	return
}

func SearchExcludeUserID(UserID, chatID string, size uint64) (messageList []string) {
	query := osquery.Search().
		Query(
			osquery.Bool().Must(
				osquery.Bool().MustNot(
					osquery.Term("user_id", UserID),
					osquery.Prefix("message_str", "/"),
				),
				osquery.Term("chat_id", chatID),
			),
		).SourceIncludes(
		"raw_message", "mentions", "create_time").Size(size).Sort("create_time", "desc")
	resp, err := opensearchdal.SearchData(
		context.Background(),
		"lark_msg_index",
		query,
	)
	if err != nil {
		panic(err)
	}
	messageList = []string{}
	for _, hit := range resp.Hits.Hits {
		res := &MessageDoc{}
		b, _ := hit.Source.MarshalJSON()
		err = sonic.ConfigStd.Unmarshal(b, res)
		if err != nil {
			continue
		}
		mentions := make([]Mention, 0)

		if res.Mentions != "null" {
			err = sonic.UnmarshalString(res.Mentions, &mentions)
			if err != nil {
				continue
			}
		}
		r := replaceMention(res.RawMessage, mentions)
		if r != "" {
			messageList = append(messageList, r)
		}
	}
	slices.Reverse(messageList)
	return
}

func replaceMention(input string, mentions []Mention) string {
	if mentions != nil {
		for _, mention := range mentions {
			// input = strings.ReplaceAll(input, mention.Key, fmt.Sprintf("<at user_id=\\\"%s\\\">%s</at>", mention.ID.UserID, mention.Name))
			input = strings.ReplaceAll(input, mention.Key, "")
			if len(input) > 0 && string(input[0]) == "/" {
				if inputs := strings.Split(input, " "); len(inputs) > 0 {
					input = strings.Join(inputs[1:], " ")
				}
			}

		}
	}
	return input
}
