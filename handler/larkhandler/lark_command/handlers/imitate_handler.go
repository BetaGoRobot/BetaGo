package handlers

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/grouputil"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkmsgutils"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/defensestation/osquery"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"go.opentelemetry.io/otel/attribute"
)

func ImitateHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	defer func() { span.RecordError(err) }()

	quoteList := data.Event.Message.Mentions
	if len(quoteList) == 1 {
		userMsgs := SearchByUserID(
			*quoteList[0].Id.OpenId, 100, 50)
		if len(userMsgs) == 0 {
			return nil
		}
		historyMsg := "- " + strings.Join(userMsgs, "\n- ")
		latestMsg := "- " + strings.Join(SearchExcludeUserID(
			*quoteList[0].Id.OpenId, *data.Event.Message.ChatId, 100, 15,
		), "\n- ")
		sysPrompt := `# 角色
你是一个擅长模仿别人说话语气的人，务必要保持语言风格和用词方式

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
		if err != nil {
			return err
		}
		span.SetAttributes(attribute.String("res", res))
		res = strings.Trim(res, "\n")
		userName := ""
		name, _ := grouputil.GetUserMemberFromChat(ctx, *data.Event.Message.ChatId, *data.Event.Message.Mentions[0].Id.OpenId)
		if name != nil {
			userName = *name.Name
		}
		if userName != "" {
			res = userName + ": " + res
		}
		_, err = larkutils.ReplyMsgText(ctx, res, *data.Event.Message.MessageId, "__imitate", false)
		if err != nil {
			return err
		}
	}

	return
}

type MessageDoc struct {
	UserID     string `json:"user_id"`
	ChatID     string `json:"chat_id"`
	UserName   string `json:"user_name"`
	Mentions   string `json:"mentions"`
	RawMessage string `json:"raw_message"`
	CreateTime string `json:"create_time"`
}

func SearchByUserID(UserID string, batch, size uint64) (messageList []string) {
	query := osquery.Search().
		Query(
			osquery.Bool().Must(
				osquery.Term("user_id", UserID),
			).MustNot(
				osquery.MatchPhrase(
					"raw_message_seg", "file _ key",
				),
			),
		).
		SourceIncludes("raw_message", "mentions", "create_time").
		Size(batch).
		Sort("create_time", "desc")
	resp, err := opensearchdal.SearchData(
		context.Background(),
		consts.LarkMsgIndex,
		query)
	if err != nil {
		panic(err)
	}
	messageList = FilterMessage(resp.Hits.Hits, int(size))
	return
}

func SearchExcludeUserID(UserID, chatID string, batch, size uint64) (messageList []string) {
	query := osquery.Search().
		Query(
			osquery.Bool().Must(
				osquery.Term("chat_id", chatID),
			).MustNot(
				osquery.MatchPhrase(
					"raw_message_seg", "file _ key",
				),
			),
		).
		SourceIncludes("raw_message", "mentions", "create_time").
		Size(batch).
		Sort("create_time", "desc")
	resp, err := opensearchdal.SearchData(
		context.Background(),
		consts.LarkMsgIndex,
		query,
	)
	if err != nil {
		panic(err)
	}
	messageList = FilterMessage(resp.Hits.Hits, int(size))
	return
}

func FilterMessage(hits []opensearchapi.SearchHit, size int) (msgList []string) {
	msgList = make([]string, 0)
	for _, hit := range hits {
		res := &handlertypes.MessageIndex{}
		b, _ := hit.Source.MarshalJSON()
		err := sonic.ConfigStd.Unmarshal(b, res)
		if err != nil {
			continue
		}
		mentions := make([]*larkmsgutils.Mention, 0)

		if res.Mentions != "null" {
			err = sonic.UnmarshalString(res.Mentions, &mentions)
			if err != nil {
				continue
			}
		}

		r := larkmsgutils.ReplaceMentionToName(res.RawMessage, mentions)

		r = strings.ReplaceAll(r, "\n", "\\n")
		r = fmt.Sprintf("[%s](%s) <%s>: %s", res.CreateTime, res.UserID, res.UserName, r)
		if r != "" {
			msgList = append(msgList, r)
		}
	}
	if len(msgList) > int(size) {
		msgList = msgList[:size]
	}
	slices.Reverse(msgList)
	return msgList
}
