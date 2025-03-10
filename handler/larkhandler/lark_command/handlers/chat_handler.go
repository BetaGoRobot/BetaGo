package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/defensestation/osquery"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func ChatHandler(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	// sendMsg
	textMsgBuilder := larkim.NewTextMsgBuilder()

	res, err := GenerateChat(ctx, event, args...)
	if err != nil {
		return err
	}
	textMsgBuilder.Text(res)
	err = larkutils.CreateMsgText(ctx, res, *event.Event.Message.MessageId, *event.Event.Message.ChatId)
	if err != nil {
		return err
	}
	return
}

func GenerateChat(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (res string, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	// 获取最近30条消息
	size := 20
	chatID := *event.Event.Message.ChatId
	query := osquery.Search().
		Query(
			osquery.Bool().Must(
				osquery.Term("chat_id", chatID),
			),
		).
		SourceIncludes("raw_message", "mentions", "create_time", "user_id", "chat_id", "user_name").
		Size(uint64(size*3)).
		Sort("CreatedAt", "desc")
	resp, err := opensearchdal.SearchData(
		context.Background(),
		"lark_msg_index",
		query)
	if err != nil {
		panic(err)
	}
	messageList := FilterMessage(resp.Hits.Hits, size)
	// 生成消息
	sysPrompt := `# 角色 

你是一个擅长参与话题讨论的人，能够依据历史的发言信息参与讨论，并且能够根据历史的发言信息生成新的话题。



# 流程 

1. 综合历史发言参考，以与历史发言相似的语气生成回复，语气相似性需达到较高程度。回复内容长度需与历史发言整体保持一致，在细节深度和整体契合度上要符合任务要求。

2. 基于输入的近期文本，判断是否需要回复。

3. 如果需要回复，按照相似的语气进行回复，同时需生成具有一定创新性且紧密围绕原始话题符合任务要求的新话题。

4. 最好基于最后一条消息来生成回复。



# 限制 

1. 回复内容和历史对话的内容长度基本一致，不得超过50字。

2. 禁止拼接历史对话。

3. 回复的文本需要跟最近几次输入存在紧密关联关系。

4. 务必确保仅回复一条文本。

5. 输出中的文本不得包含「xxx:」类的角色信息。

6. 输出文本中，不需要指明新话题：xxx这种标识；直接嵌入回答即可。

7. 尽可能基于最后一条消息来生成回复。

`
	userPrompt := `历史发言输入: %s %s 请给出模仿的输出:
`
	latestMsg := strings.Join(messageList, "\n- ")
	userPrompt = fmt.Sprintf(userPrompt, "\n- "+latestMsg, "\n# 要求\n"+strings.Join(args, "\n - "))
	res, err = doubao.SingleChat(ctx, sysPrompt, userPrompt)
	if err != nil {
		return
	}
	span.SetAttributes(attribute.String("res", res))
	res = strings.Trim(res, "\n")
	res = strings.Trim(strings.Split(res, "\n")[0], " - ")
	return
}
