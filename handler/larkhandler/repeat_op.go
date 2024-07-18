package larkhandler

import (
	"context"

	"github.com/pkg/errors"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

var _ LarkMsgOperator = &RepeatMsgOperator{}

// RepeatMsgOperator  RepeatMsg Op
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:51
type RepeatMsgOperator struct {
	LarkMsgOperatorBase
}

// PreRun Repeat
//
//	@receiver r *RepeatMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:35
func (r *RepeatMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	// 先判断群聊的功能启用情况
	if !checkFunctionEnabling(*event.Event.Message.ChatId, consts.LarkFunctionRandomRepeat) {
		return errors.Wrap(ErrStageSkip, "RepeatMsgOperator: Not enabled")
	}
	if larkutils.IsCommand(ctx, larkutils.PreGetTextMsg(ctx, event)) {
		return errors.Wrap(ErrStageSkip, "RepeatMsgOperator: Is Command")
	}
	return
}

// Run Repeat
//
//	@receiver r *RepeatMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:41
func (r *RepeatMsgOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	// Repeat
	msg := larkutils.PreGetTextMsg(ctx, event)

	// 开始摇骰子, 默认概率10%
	realRate := utility.MustAtoI(utility.GetEnvWithDefault("REPEAT_DEFAULT_RATE", "10"))

	config, hitCache := database.FindByCacheFunc(
		database.RepeatWordsRate{
			Word: msg,
		},
		func(d database.RepeatWordsRate) string {
			return d.Word
		},
	)
	span.SetAttributes(attribute.Bool("RepeatWordsRate hitCache", hitCache))
	for _, data := range config {
		if data.Word == msg {
			realRate = data.Rate
			break
		}
	}

	if utility.Probability(float64(realRate) / 100) {
		// sendMsg
		textMsg := larkim.NewTextMsgBuilder().Text(msg).Build()
		req := larkim.NewCreateMessageReqBuilder().ReceiveIdType(larkim.ReceiveIdTypeChatId).Body(
			larkim.NewCreateMessageReqBodyBuilder().
				ReceiveId(*event.Event.Message.ChatId).
				Content(textMsg).
				MsgType(larkim.MsgTypeText).
				Uuid(*event.Event.Message.MessageId + "repeat").
				Build(),
		).Build()
		resp, err := larkutils.LarkClient.Im.V1.Message.Create(ctx, req)
		if err != nil {
			log.ZapLogger.Error("repeatMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
		}
		log.ZapLogger.Info("repeatMessage", zaplog.Any("resp", resp), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
	}
	return
}
