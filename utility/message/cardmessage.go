package message

import (
	"context"
	"errors"
	"iter"
	"strings"
	"sync/atomic"
	"time"

	"github.com/BetaGoRobot/BetaGo/dal/lark"
	"github.com/BetaGoRobot/BetaGo/utility/ark"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/cardutil"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	larkcardkit "github.com/larksuite/oapi-sdk-go/v3/service/cardkit/v1"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func SendAndUpdateStreamingCard(ctx context.Context, msg *larkim.EventMessage, msgSeq iter.Seq[*ark.ModelStreamRespReasoning]) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	// create Card
	// 创建卡片实体

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

func SendAndReplyStreamingCard(ctx context.Context, msg *larkim.EventMessage, msgSeq iter.Seq[*ark.ModelStreamRespReasoning], inThread bool) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

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

type KV[K comparable, V any] struct {
	Key K
	Val V
}

func updateCardFunc(ctx context.Context, res iter.Seq[*ark.ModelStreamRespReasoning], cardID string) (err error, lastIdx int) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	idx := &atomic.Int32{}
	idx.Store(0)

	defer func() {
		lastIdx = int(idx.Load())
	}()
	sendFunc := func(key, content string) {
		ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
		defer span.End()
		defer func() { span.RecordError(err) }()
		body := larkcardkit.NewContentCardElementReqBodyBuilder().Content(content).Sequence(int(idx.Add(1))).Build()
		req := larkcardkit.NewContentCardElementReqBuilder().CardId(cardID).ElementId(key).Body(body).Build()
		resp, err := lark.LarkClient.Cardkit.V1.CardElement.Content(ctx, req)
		if err != nil {
			logs.L().Ctx(ctx).Error("patch message failed with error msg", zap.Error(err))
			return
		}
		if !resp.Success() {
			logs.L().Ctx(ctx).Error("patch message failed with error msg", zap.String("CodeError.Error", resp.CodeError.Error()))
			return
		}
	}
	var (
		msgChan = make(chan KV[string, string], 10)
		ticker  = time.NewTicker(time.Millisecond * 20)
	)
	defer ticker.Stop()
	eg := errgroup.Group{}

	eg.Go(func() error {
		defer close(msgChan)

		writeFunc := func(data ark.ModelStreamRespReasoning) error {
			_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
			defer span.End()
			defer func() { span.RecordError(err) }()

			if data.ReasoningContent != "" {
				contentSlice := []string{}
				for _, item := range strings.Split(data.ReasoningContent, "\n") {
					contentSlice = append(contentSlice, "> "+item)
				}
				data.ReasoningContent = strings.Join(contentSlice, "\n")
			}

			if data.Content != "" {
				msgChan <- KV[string, string]{"content", data.Content}
			}

			if data.ReasoningContent != "" {
				msgChan <- KV[string, string]{"cot", data.ReasoningContent}
			}
			return nil
		}

		for data := range res {
			err = writeFunc(*data)
			if err != nil {
				return err
			}
		}
		return nil
	})

	chunkQueue := make(map[string]string)
	clearQueue := func() {
		if len(chunkQueue) > 0 {
			for key, content := range chunkQueue {
				if key == "content" {
					// 尝试修复 JSON 字符串
					contentStruct := &ark.ContentStruct{}
					if !strings.HasSuffix(content, "}") {
						content += "}"
					}
					if sonic.UnmarshalString(content+"}", &contentStruct); contentStruct != nil {
						content = contentStruct.BuildOutput()
					}
				}
				sendFunc(key, content)
			}
			chunkQueue = map[string]string{}
		}
	}
updateChunkLoop:
	for {
		select {
		case chunk, ok := <-msgChan:
			if !ok {
				break updateChunkLoop
			}
			chunkQueue[chunk.Key] = chunk.Val
		case <-ticker.C:
			clearQueue()
		}
	}
	clearQueue()
	return
}
