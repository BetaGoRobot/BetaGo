package larkhandler

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

var _ LarkMsgOperator = &ReactMsgOperator{}

// ReactMsgOperator  Repeat
type ReactMsgOperator struct {
	LarkMsgOperatorBase
}

// PreRun Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *ReactMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	// 先判断群聊的功能启用情况
	if !checkFunctionEnabling(*event.Event.Message.ChatId, consts.LarkFunctionRandomReact) {
		return errors.Wrap(ErrStageSkip, "RepeatMsgOperator: Not enabled")
	}
	return
}

// Run  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *ReactMsgOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	// React
	// 先判断群聊的功能启用情况
	chatEnabled := false
	resData, hitCache := database.FindByCacheFunc(
		database.ReactionWhitelist{GuildID: *event.Event.Message.ChatId},
		func(d database.ReactionWhitelist) string {
			return d.GuildID
		},
	)
	span.SetAttributes(attribute.Bool("ReactionWhitelist hitCache", hitCache))
	for _, data := range resData {
		if data.GuildID == *event.Event.Message.ChatId {
			chatEnabled = true
			break
		}
	}

	if chatEnabled {
		// 开始摇骰子, 默认概率10%
		realRate := utility.MustAtoI(utility.GetEnvWithDefault("REACTION_DEFAULT_RATE", "10"))
		if utility.Probability(float64(realRate) / 100) {
			// sendMsg
			req := larkim.NewCreateMessageReactionReqBuilder().
				MessageId(*event.Event.Message.MessageId).
				Body(
					larkim.NewCreateMessageReactionReqBodyBuilder().
						ReactionType(
							larkim.NewEmojiBuilder().
								EmojiType(getRandomEmoji()).
								Build(),
						).
						Build(),
				).
				Build()
			resp, err := larkutils.LarkClient.Im.V1.MessageReaction.Create(ctx, req)
			if err != nil {
				log.ZapLogger.Error("reactMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
				return err
			}
			log.ZapLogger.Info("reactMessage", zaplog.Any("resp", resp), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
		}
	}

	return
}
