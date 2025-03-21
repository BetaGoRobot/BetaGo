package handlers

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"text/template"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/defensestation/osquery"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func ChatHandler(chatType string) func(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (err error) {
	return func(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (err error) {
		return ChatHandlerInner(ctx, event, chatType, args...)
	}
}

func ChatHandlerInner(ctx context.Context, event *larkim.P2MessageReceiveV1, chatType string, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	// sendMsg
	textMsgBuilder := larkim.NewTextMsgBuilder()
	var res string
	if chatType == "reply" {
		res, err = GenerateChatReply(ctx, event, args...)
	} else {
		res, err = GenerateChatReply(ctx, event, args...)
	}
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

func ChatHandlerWithTemplate(ctx context.Context, event *larkim.P2MessageReceiveV1, templateID int, args ...string) (err error) {
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

func GenerateChatByTemplate(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (res string, err error) {
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

	templateRows, _ := database.FindByCacheFunc(database.PromptTemplateArgs{PromptID: 1}, func(d database.PromptTemplateArgs) string {
		return strconv.Itoa(d.PromptID)
	})
	if len(templateRows) == 0 {
		return "", errors.New("prompt template not found")
	}
	promptTemplate := templateRows[0]
	promptTemplateStr := promptTemplate.TemplateStr
	tp, err := template.New("prompt").Parse(promptTemplateStr)
	if err != nil {
		return "", err
	}
	promptTemplate.UserInput = args
	promptTemplate.HistoryRecords = messageList
	b := &strings.Builder{}
	err = tp.Execute(b, promptTemplate)
	if err != nil {
		return "", err
	}

	res, err = doubao.SingleChatPrompt(ctx, b.String())
	if err != nil {
		return
	}
	span.SetAttributes(attribute.String("res", res))
	res = strings.Trim(res, "\n")
	res = strings.Trim(strings.Split(res, "\n")[0], " - ")
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

	templateRows, _ := database.FindByCacheFunc(database.PromptTemplateArgs{PromptID: 1}, func(d database.PromptTemplateArgs) string {
		return strconv.Itoa(d.PromptID)
	})
	if len(templateRows) == 0 {
		return "", errors.New("prompt template not found")
	}
	promptTemplate := templateRows[0]
	promptTemplateStr := promptTemplate.TemplateStr
	tp, err := template.New("prompt").Parse(promptTemplateStr)
	if err != nil {
		return "", err
	}
	promptTemplate.UserInput = args
	promptTemplate.HistoryRecords = messageList
	b := &strings.Builder{}
	err = tp.Execute(b, promptTemplate)
	if err != nil {
		return "", err
	}

	res, err = doubao.SingleChatPrompt(ctx, b.String())
	if err != nil {
		return
	}
	span.SetAttributes(attribute.String("res", res))
	res = strings.Trim(res, "\n")
	res = strings.Trim(strings.Split(res, "\n")[0], " - ")
	return
}

func GenerateChatReply(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (res string, err error) {
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

	templateRows, _ := database.FindByCacheFunc(
		database.PromptTemplateArgs{PromptID: 2},
		func(d database.PromptTemplateArgs) string {
			return strconv.Itoa(d.PromptID)
		},
	)
	if len(templateRows) == 0 {
		return "", errors.New("prompt template not found")
	}
	promptTemplate := templateRows[0]
	promptTemplateStr := promptTemplate.TemplateStr
	tp, err := template.New("prompt").Parse(promptTemplateStr)
	if err != nil {
		return "", err
	}
	promptTemplate.UserInput = args
	promptTemplate.HistoryRecords = messageList
	b := &strings.Builder{}
	err = tp.Execute(b, promptTemplate)
	if err != nil {
		return "", err
	}

	res, err = doubao.SingleChatPrompt(ctx, b.String())
	if err != nil {
		return
	}
	span.SetAttributes(attribute.String("res", res))
	res = strings.Trim(res, "\n")
	res = strings.Trim(strings.Split(res, "\n")[0], " - ")
	return
}
