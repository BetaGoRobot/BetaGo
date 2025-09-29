package message

import (
	"context"
	"errors"
	"iter"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/dal/lark"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/cardutil"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkcardkit "github.com/larksuite/oapi-sdk-go/v3/service/cardkit/v1"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"golang.org/x/sync/errgroup"
)

func SendAndUpdateStreamingCard(ctx context.Context, msg *larkim.EventMessage, msgSeq iter.Seq[*doubao.ModelStreamRespReasoning]) error {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	// create Card
	// 创建卡片实体
	// template := templates.GetTemplate(templates.StreamingReasonTemplate)
	// cardSrc := template.TemplateSrc
	cardContent := templates.NewCardContent(ctx, templates.NormalCardReplyTemplate)
	// 首先Create卡片实体
	cardEntiReq := larkcardkit.NewCreateCardReqBuilder().Body(
		larkcardkit.NewCreateCardReqBodyBuilder().
			// Type(`card_json`).
			Type(`template`).
			Data(cardContent.DataString()).
			Build(),
	).Build()
	createEntiResp, err := lark.LarkClient.Cardkit.V1.Card.Create(ctx, cardEntiReq)
	if err != nil {
		return err
	}
	if !createEntiResp.Success() {
		return errors.New(createEntiResp.CodeError.Error())
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
	resp, err := lark.LarkClient.Im.V1.Message.Create(ctx, req)
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
	settingUpdateResp, err := lark.LarkClient.Cardkit.V1.Card.
		Settings(ctx, settingUpdateReq)
	if err != nil {
		return err
	}
	if !settingUpdateResp.Success() {
		return errors.New(settingUpdateResp.CodeError.Error())
	}
	return nil
}

func SendAndReplyStreamingCard(ctx context.Context, msg *larkim.EventMessage, msgSeq iter.Seq[*doubao.ModelStreamRespReasoning], inThread bool) error {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	// create Card
	// 创建卡片实体
	// template := templates.GetTemplate(templates.StreamingReasonTemplate)
	// cardSrc := template.TemplateSrc
	cardContent := templates.NewCardContent(ctx, templates.NormalCardReplyTemplate)
	// 首先Create卡片实体
	cardEntiReq := larkcardkit.NewCreateCardReqBuilder().Body(
		larkcardkit.NewCreateCardReqBodyBuilder().
			// Type(`card_json`).
			Type(`template`).
			Data(cardContent.DataString()).
			Build(),
	).Build()
	createEntiResp, err := lark.LarkClient.Cardkit.V1.Card.Create(ctx, cardEntiReq)
	if err != nil {
		return err
	}
	if !createEntiResp.Success() {
		return errors.New(createEntiResp.CodeError.Error())
	}
	cardID := *createEntiResp.Data.CardId

	// 发送卡片
	req := larkim.NewReplyMessageReqBuilder().
		MessageId(*msg.MessageId).
		Body(
			larkim.NewReplyMessageReqBodyBuilder().ReplyInThread(inThread).
				MsgType(larkim.MsgTypeInteractive).
				Content(cardutil.NewCardEntityContent(cardID).String()).
				Build(),
		).
		Build()
	resp, err := lark.LarkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Success() {
		return errors.New(resp.Error())
	}

	larkutils.RecordReplyMessage2Opensearch(ctx, resp)

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
	settingUpdateResp, err := lark.LarkClient.Cardkit.V1.Card.
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
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	sendFunc := func(req *larkcardkit.ContentCardElementReq) {
		ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
		defer span.End()

		resp, err := lark.LarkClient.Cardkit.V1.CardElement.Content(ctx, req)
		if err != nil {
			log.Zlog.Error("patch message failed with error msg: " + resp.Msg)
			return
		}
	}
	var (
		msgChan = make(chan *larkcardkit.ContentCardElementReq, 10)
		ticker  = time.NewTicker(time.Millisecond * 20)
	)
	defer ticker.Stop()

	eg := errgroup.Group{}
	eg.Go(func() error {
		defer close(msgChan)
		writeFunc := func(idx int, data *doubao.ModelStreamRespReasoning) error {
			_, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
			defer span.End()

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
			} else {
				bodyBuilder.Content(data.ReasoningContent)
				updateReqBuilder.ElementId("cot")
			}
			updateReqBuilder.Body(bodyBuilder.Build())
			msgChan <- updateReqBuilder.Build()
			return nil
		}
		lastIdx = 0
		lastData := &doubao.ModelStreamRespReasoning{}
		for data := range res {
			lastIdx++
			*lastData = *data

			err = writeFunc(lastIdx, data)
			if err != nil {
				return err
			}
		}
		err = writeFunc(lastIdx, lastData)
		if err != nil {
			return err
		}
		return nil
	})

	var lastChunk *larkcardkit.ContentCardElementReq
updateChunkLoop:
	for {
		select {
		case chunk, ok := <-msgChan:
			if !ok {
				break updateChunkLoop
			}
			lastChunk = chunk
		case <-ticker.C:
			if lastChunk != nil {
				sendFunc(lastChunk)
				lastChunk = nil
			}
		}
	}
	if lastChunk != nil {
		sendFunc(lastChunk)
	}
	return
}
