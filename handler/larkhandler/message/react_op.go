package message

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

var _ handlerbase.Operator[larkim.P2MessageReceiveV1] = &ReactMsgOperator{}

// ReactMsgOperator  Repeat
type ReactMsgOperator struct {
	handlerbase.OperatorBase[larkim.P2MessageReceiveV1]
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
	if !larkutils.CheckFunctionEnabling(*event.Event.Message.ChatId, consts.LarkFunctionRandomReact) {
		span.RecordError(err)
		return errors.Wrap(consts.ErrStageSkip, "ReactMsgOperator: Not enabled")
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
	defer span.RecordError(err)

	// React

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
							EmojiType(larkutils.GetRandomEmoji()).
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
	} else {
		if utility.Probability(float64(realRate) / 100) {
			res, hitCache := database.FindByCacheFunc(database.ReactImageMeterial{
				GuildID: *event.Event.Message.ChatId,
			}, func(d database.ReactImageMeterial) string {
				return d.GuildID
			})
			span.SetAttributes(attribute.Bool("ReactionImageMaterial hitCache", hitCache))
			if len(res) == 0 {
				return
			}
			target := utility.SampleSlice(res)
			if target.Type == consts.LarkResourceTypeImage {
				content, _ := sonic.MarshalString(map[string]string{
					"image_key": target.FileID,
				})
				err = larkutils.SendMsgRawContentType(
					ctx,
					*event.Event.Message.MessageId,
					larkim.MsgTypeImage,
					content,
					false,
				)
			} else {
				content, _ := sonic.MarshalString(map[string]string{
					"file_key": target.FileID,
				})
				err = larkutils.SendMsgRawContentType(
					ctx,
					*event.Event.Message.MessageId,
					larkim.MsgTypeSticker,
					content,
					false,
				)
			}
		}
	}

	return
}
