package handlers

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strconv"
	"strings"
	"text/template"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/redis"
	"github.com/defensestation/osquery"
	"github.com/google/uuid"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func ChatHandler(chatType string) func(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (err error) {
	return func(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (err error) {
		return ChatHandlerInner(ctx, event, chatType, args...)
	}
}

func ChatHandlerInner(ctx context.Context, event *larkim.P2MessageReceiveV1, chatType string, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	var res iter.Seq[*doubao.ModelStreamRespReasoning]
	if chatType == "reply" {
		res, err = GenerateChatReply(ctx, event, args...)
	} else {
		if ext, err := redis.GetRedisClient().
			Exists(ctx, MuteRedisKeyPrefix+*event.Event.Message.ChatId).Result(); err != nil {
			return err
		} else if ext != 0 {
			return nil // Do nothing
		}

		res, err = GenerateChatSeq(ctx, event, args...)
		if err != nil {
			return err
		}
	}

	// 先Create个卡片
	template := larkutils.GetTemplate(larkutils.StreamingReasonTemplate)
	cardContent := larkutils.NewSheetCardContent(
		ctx,
		template.TemplateID,
		template.TemplateVersion,
	).
		AddVariable("cot", "正在思考...").
		AddVariable("content", "").String()
	req := larkim.NewCreateMessageReqBuilder().
		Body(
			larkim.NewCreateMessageReqBodyBuilder().
				Content(cardContent).
				Uuid(uuid.NewString()).
				Build(),
		).
		Build()
	resp, err := larkutils.LarkClient.Im.V1.Message.Create(ctx, req)
	if err != nil {
		return err
	}
	msgID := *resp.Data.MessageId
	lastData := &doubao.ModelStreamRespReasoning{}
	writeFunc := func(data *doubao.ModelStreamRespReasoning) error {
		if data.ReasoningContent != "" {
			contentSlice := []string{}
			for _, item := range strings.Split(data.ReasoningContent, "\n") {
				contentSlice = append(contentSlice, "> "+item)
			}
			data.ReasoningContent = strings.Join(contentSlice, "\n")
		}

		cardContent := larkutils.NewSheetCardContent(
			ctx,
			template.TemplateID,
			template.TemplateVersion,
		).
			AddVariable("cot", data.ReasoningContent).
			AddVariable("content", data.Content).String()
		updateReq := larkim.NewPatchMessageReqBuilder().MessageId(msgID).
			Body(
				larkim.NewPatchMessageReqBodyBuilder().
					Content(cardContent).
					Build(),
			).
			Build()
		fmt.Println(data.ReasoningContent)
		resp, err := larkutils.LarkClient.Im.V1.Message.Patch(ctx, updateReq)
		if err != nil {
			log.ZapLogger.Error("patch message failed with error msg: " + resp.Msg)
			return err
		}
		return nil
	}
	// 更新卡片内容
	idx := 0
	for data := range res {
		idx++
		*lastData = *data

		if idx%10 == 0 {
			err = writeFunc(data)
			if err != nil {
				return err
			}
		}
	}
	err = writeFunc(lastData)
	if err != nil {
		return err
	}
	return
}

func ChatHandlerFunc(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	// 先Create个卡片
	template := larkutils.GetTemplate(larkutils.StreamingReasonTemplate)
	cardContent := larkutils.NewSheetCardContent(
		ctx,
		template.TemplateID,
		template.TemplateVersion,
	).
		AddVariable("cot", "正在思考...").
		AddVariable("content", "").String()

	req := larkim.NewCreateMessageReqBuilder().
		Body(
			larkim.NewCreateMessageReqBodyBuilder().
				Content(cardContent).
				Uuid(uuid.NewString()).
				Build(),
		).
		Build()
	resp, err := larkutils.LarkClient.Im.V1.Message.Create(ctx, req)
	if err != nil {
		return err
	}

	msgID := *resp.Data.MessageId

	res, err := GenerateChatSeq(ctx, event, args...)
	if err != nil {
		return err
	}
	lastData := &doubao.ModelStreamRespReasoning{}
	// 更新卡片内容
	idx := 0
	writeFunc := func(data *doubao.ModelStreamRespReasoning) error {
		if data.ReasoningContent != "" {
			contentSlice := []string{}
			for _, item := range strings.Split(data.ReasoningContent, "\n") {
				contentSlice = append(contentSlice, "> "+item)
			}
			data.ReasoningContent = strings.Join(contentSlice, "\n")
		}

		cardContent := larkutils.NewSheetCardContent(
			ctx,
			template.TemplateID,
			template.TemplateVersion,
		).
			AddVariable("cot", data.ReasoningContent).
			AddVariable("content", data.Content).String()
		updateReq := larkim.NewPatchMessageReqBuilder().MessageId(msgID).
			Body(
				larkim.NewPatchMessageReqBodyBuilder().
					Content(cardContent).
					Build(),
			).
			Build()
		fmt.Println(data.ReasoningContent)
		resp, err := larkutils.LarkClient.Im.V1.Message.Patch(ctx, updateReq)
		if err != nil {
			log.ZapLogger.Error("patch message failed with error msg: " + resp.Msg)
			return err
		}
		return nil
	}
	for data := range res {
		idx++
		*lastData = *data
		if idx%10 == 0 {
			err = writeFunc(data)
			if err != nil {
				return err
			}
		}
	}
	err = writeFunc(lastData)
	if err != nil {
		return err
	}

	return
}

func GenerateChatSeq(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (res iter.Seq[*doubao.ModelStreamRespReasoning], err error) {
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
		return nil, errors.New("prompt template not found")
	}
	promptTemplate := templateRows[0]
	promptTemplateStr := promptTemplate.TemplateStr
	tp, err := template.New("prompt").Parse(promptTemplateStr)
	if err != nil {
		return nil, err
	}
	promptTemplate.UserInput = args
	promptTemplate.HistoryRecords = messageList
	b := &strings.Builder{}
	err = tp.Execute(b, promptTemplate)
	if err != nil {
		return nil, err
	}

	res, err = doubao.SingleChatStreamingPrompt(ctx, b.String(), doubao.DOUBAO_THINK_EPID)
	if err != nil {
		return
	}
	// span.SetAttributes(attribute.String("res", res))
	// res = strings.Trim(res, "\n")
	// res = strings.Trim(strings.Split(res, "\n")[0], " - ")
	return
}

func GenerateChatReply(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (res iter.Seq[*doubao.ModelStreamRespReasoning], err error) {
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
		return nil, errors.New("prompt template not found")
	}
	promptTemplate := templateRows[0]
	promptTemplateStr := promptTemplate.TemplateStr
	tp, err := template.New("prompt").Parse(promptTemplateStr)
	if err != nil {
		return nil, err
	}
	promptTemplate.UserInput = args
	promptTemplate.HistoryRecords = messageList
	b := &strings.Builder{}
	err = tp.Execute(b, promptTemplate)
	if err != nil {
		return nil, err
	}

	res, err = doubao.SingleChatStreamingPrompt(ctx, b.String(), doubao.DOUBAO_THINK_EPID)
	if err != nil {
		return
	}
	// span.SetAttributes(attribute.String("res", res))
	// res = strings.Trim(res, "\n")
	// res = strings.Trim(strings.Split(res, "\n")[0], " - ")
	return
}
