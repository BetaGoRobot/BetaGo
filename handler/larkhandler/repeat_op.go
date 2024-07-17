package larkhandler

import (
	"context"
	"errors"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/patrickmn/go-cache"
)

var _ LarkMsgOperator = &RepeatMsgOperator{}

type RepeatMsgOperator struct{}

func (r *RepeatMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	// 先判断群聊的功能启用情况
	if !checkFunctionEnabling(*event.Event.Message.ChatId, consts.LarkFunctionRandomRepeat) {
		return errors.New("Not enabled")
	}
	return
}

func (r *RepeatMsgOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	// Repeat
	msg := larkutils.PreGetTextMsg(ctx, event)

	// 开始摇骰子, 默认概率10%
	realRate := utility.MustAtoI(utility.GetEnvWithDefault("REPEAT_DEFAULT_RATE", "10"))
	if rate, exists := repeatWordRateCache.Get(msg); exists {
		if r := rate.(int); r != -1 {
			repeatWordRateCache.Set(msg, rate.(int), cache.DefaultExpiration)
			realRate = r
		}
	} else {
		wordRate := database.RepeatWordsRate{
			Word: msg,
		}
		database.GetDbConnection().Find(&wordRate)
		if wordRate.Rate != 0 && wordRate.Rate != -1 {
			realRate = wordRate.Rate
		}
		repeatWordRateCache.Set(msg, realRate, cache.DefaultExpiration)
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

func (r *RepeatMsgOperator) PostRun(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	return
}
