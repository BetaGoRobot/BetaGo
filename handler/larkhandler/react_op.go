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
		return errors.New("Not enabled")
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
	resData, hitCache := database.FindByCache(&database.ReactionWhitelist{})
	span.SetAttributes(attribute.Bool("hitCache", hitCache))
	for _, data := range resData {
		if data.GuildID == *event.Event.Message.ChatId {
			chatEnabled = true
			break
		}
	}

	if chatEnabled {
	}
	if enabled, exists := reactionConfigCache.Get(*event.Event.Message.ChatId); exists {
		// 缓存中已存在，直接取值
		if !enabled.(bool) {
			return
		}
	} else {
		// 缓存中不存在，从数据库中取值
		var count int64
		database.GetDbConnection().Find(&database.ReactionWhitelist{GuildID: *event.Event.Message.ChatId}).Count(&count)
		if count == 0 {
			reactionConfigCache.Set(*event.Event.Message.ChatId, false, cache.DefaultExpiration)
			return
		}
		reactionConfigCache.Set(*event.Event.Message.ChatId, true, cache.DefaultExpiration)
	}

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
	return
}
