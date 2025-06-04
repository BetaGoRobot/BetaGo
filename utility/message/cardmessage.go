package message

import (
	"context"
	"errors"
	"iter"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/cardutil"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	larkcardkit "github.com/larksuite/oapi-sdk-go/v3/service/cardkit/v1"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func SendAndUpdateStreamingCard(ctx context.Context, msg *larkim.EventMessage, msgSeq iter.Seq[*doubao.ModelStreamRespReasoning]) error {
	// create Card
	// 创建卡片实体
	template := templates.GetTemplate(templates.StreamingReasonTemplate)
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
				ReceiveId(*msg.ChatId).
				MsgType(larkim.MsgTypeInteractive).
				Content(cardutil.NewCardEntityContent(cardID).String()).
				Build(),
		).
		Build()
	resp, err := larkutils.LarkClient.Im.V1.Message.Create(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Success() {
		return errors.New(resp.Error())
	}

	larkutils.RecordMessage2Opensearch(ctx, resp)

	err, lastIdx := updateCardFunc(ctx, msgSeq, cardID)
	if err != nil {
		return err
	}
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
	if !settingUpdateResp.Success() {
		return errors.New(settingUpdateResp.CodeError.Error())
	}
	return nil
}

func updateCardFunc(ctx context.Context, res iter.Seq[*doubao.ModelStreamRespReasoning], cardID string) (err error, lastIdx int) {
	sendFunc := func(req *larkcardkit.ContentCardElementReq) {
		resp, err := larkutils.LarkClient.Cardkit.V1.CardElement.Content(ctx, req)
		if err != nil {
			log.Zlog.Error("patch message failed with error msg: " + resp.Msg)
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
