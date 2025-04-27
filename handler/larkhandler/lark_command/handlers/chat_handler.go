package handlers

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/cardutil"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/redis"
	commonutils "github.com/BetaGoRobot/go_utils/common_utils"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/defensestation/osquery"
	larkcardkit "github.com/larksuite/oapi-sdk-go/v3/service/cardkit/v1"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func ChatHandler(chatType string) func(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (err error) {
	return func(ctx context.Context, event *larkim.P2MessageReceiveV1, args ...string) (err error) {
		newChatType := chatType
		size := new(int)
		*size = 20
		argMap, input := parseArgs(args...)
		if _, ok := argMap["r"]; ok {
			newChatType = consts.MODEL_TYPE_REASON
		}
		if _, ok := argMap["c"]; ok {
			// no context
			*size = 0
		}
		return ChatHandlerInner(ctx, event, newChatType, size, input)
	}
}

func ChatHandlerInner(ctx context.Context, event *larkim.P2MessageReceiveV1, chatType string, size *int, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	var res iter.Seq[*doubao.ModelStreamRespReasoning]
	if ext, err := redis.GetRedisClient().
		Exists(ctx, MuteRedisKeyPrefix+*event.Event.Message.ChatId).Result(); err != nil {
		return err
	} else if ext != 0 {
		return nil // Do nothing
	}
	if chatType == consts.MODEL_TYPE_REASON {
		res, err = GenerateChatSeq(ctx, event, doubao.ARK_REASON_EPID, size, args...)
		// 创建卡片实体
		template := larkutils.GetTemplate(larkutils.StreamingReasonTemplate)
		cardSrc := template.TemplateSrc
		// 首先Create卡片实体
		cardEntiReq := larkcardkit.NewCreateCardReqBuilder().Body(
			larkcardkit.NewCreateCardReqBodyBuilder().
				Type(`card_json`).
				Data(cardSrc).
				Build(),
		).Build()
		createEntiResp, err := larkutils.LarkClient.Cardkit.V1.Card.Create(ctx, cardEntiReq)
		if err != nil {
			return err
		}
		cardID := *createEntiResp.Data.CardId

		// 发送卡片
		req := larkim.NewCreateMessageReqBuilder().
			ReceiveIdType(larkim.ReceiveIdTypeChatId).
			Body(
				larkim.NewCreateMessageReqBodyBuilder().
					ReceiveId(*event.Event.Message.ChatId).
					MsgType(larkim.MsgTypeInteractive).
					Content(cardutil.NewCardEntityContent(cardID).String()).
					Build(),
			).
			Build()
		resp, err := larkutils.LarkClient.Im.V1.Message.Create(ctx, req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return errors.New(resp.Error())
		}
		larkutils.RecordMessage2Opensearch(ctx, resp)
		err, lastIdx := updateCardFunc(ctx, res, cardID)
		if err != nil {
			return err
		}
		time.Sleep(time.Second * 1)
		settingUpdateReq := larkcardkit.NewSettingsCardReqBuilder().
			CardId(cardID).
			Body(larkcardkit.NewSettingsCardReqBodyBuilder().
				Settings(cardutil.DisableCardStreaming().String()).
				Sequence(lastIdx + 1).
				Build()).
			Build()
		// 发起请求
		settingUpdateResp, err := larkutils.LarkClient.Cardkit.V1.Card.
			Settings(ctx, settingUpdateReq)
		if err != nil {
			return err
		}
		if settingUpdateResp.CodeError.Err != nil {
			return errors.New(settingUpdateResp.CodeError.Error())
		}
	} else {
		res, err = GenerateChatSeq(ctx, event, doubao.ARK_NORMAL_EPID, size, args...)
		if err != nil {
			return err
		}
		lastData := &doubao.ModelStreamRespReasoning{}
		for data := range res {
			lastData = data
		}
		_, err = larkutils.ReplyMsgText(
			ctx, lastData.Content, *event.Event.Message.MessageId, "_chat_random", false,
		)
		if err != nil {
			return
		}
	}
	return
}

func updateCardFunc(ctx context.Context, res iter.Seq[*doubao.ModelStreamRespReasoning], cardID string) (err error, lastIdx int) {
	sendFunc := func(req *larkcardkit.ContentCardElementReq) {
		resp, err := larkutils.LarkClient.Cardkit.V1.CardElement.Content(ctx, req)
		if err != nil {
			log.ZapLogger.Error("patch message failed with error msg: " + resp.Msg)
			return
		}
	}
	writeFunc := func(idx int, data *doubao.ModelStreamRespReasoning) error {
		bodyBuilder := larkcardkit.
			NewContentCardElementReqBodyBuilder().
			Sequence(idx)
		updateReqBuilder := larkcardkit.
			NewContentCardElementReqBuilder().
			CardId(cardID)
		if data.ReasoningContent != "" {
			contentSlice := []string{}
			for _, item := range strings.Split(data.ReasoningContent, "\n") {
				contentSlice = append(contentSlice, "> "+item)
			}
			data.ReasoningContent = strings.Join(contentSlice, "\n")
		}

		if data.Content != "" {
			bodyBuilder.Content(data.Content)
			updateReqBuilder.ElementId("content")
			updateReqBuilder.Body(bodyBuilder.Build())
			go sendFunc(updateReqBuilder.Build())
		} else {
			bodyBuilder.Content(data.ReasoningContent)
			updateReqBuilder.ElementId("cot")
			updateReqBuilder.Body(bodyBuilder.Build())
			go sendFunc(updateReqBuilder.Build())
		}
		return nil
	}
	lastIdx = 0
	lastData := &doubao.ModelStreamRespReasoning{}
	for data := range res {
		lastIdx++
		*lastData = *data

		err = writeFunc(lastIdx, data)
		if err != nil {
			return
		}
	}
	err = writeFunc(lastIdx, lastData)
	if err != nil {
		return
	}
	return
}

func GenerateChatSeq(ctx context.Context, event *larkim.P2MessageReceiveV1, modelID string, size *int, input ...string) (res iter.Seq[*doubao.ModelStreamRespReasoning], err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	// 默认获取最近20条消息
	if size == nil {
		size = new(int)
		*size = 20
	}
	chatID := *event.Event.Message.ChatId
	query := osquery.Search().
		Query(
			osquery.Bool().Must(
				osquery.Term("chat_id", chatID),
			),
		).
		SourceIncludes("raw_message", "mentions", "create_time", "user_id", "chat_id", "user_name").
		Size(uint64(*size*3)).
		Sort("create_time", "desc")
	resp, err := opensearchdal.SearchData(
		context.Background(),
		"lark_msg_index",
		query)
	if err != nil {
		panic(err)
	}
	messageList := FilterMessage(resp.Hits.Hits, *size)

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
	member, err := larkutils.GetUserMemberFromChat(ctx, chatID, *event.Event.Sender.SenderId.OpenId)
	if err != nil {
		return
	}
	userName := ""
	if member == nil {
		userName = "NULL"
	} else {
		userName = *member.Name
	}

	promptTemplate.UserInput = commonutils.TransSliceWithSkip(input, func(s string) (string, bool) {
		if strings.TrimSpace(s) != "" {
			return fmt.Sprintf("[%s] <%s>: %s", utility.EpoMil2DateStr(*event.Event.Message.CreateTime), userName, strings.ReplaceAll(s, "\n", "\\n")), false
		}
		return "", true
	})
	promptTemplate.HistoryRecords = messageList
	b := &strings.Builder{}
	err = tp.Execute(b, promptTemplate)
	if err != nil {
		return nil, err
	}
	fmt.Println(b.String())
	res, err = doubao.SingleChatStreamingPrompt(ctx, b.String(), modelID)
	if err != nil {
		return
	}
	return
}
