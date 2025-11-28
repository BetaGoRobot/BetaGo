package message

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/dal/lark"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

var _ Op = &ReactMsgOperator{}

// ReactMsgOperator  Repeat
type ReactMsgOperator struct {
	OpBase
}

func (r *ReactMsgOperator) Name() string {
	return "ReactMsgOperator"
}

// PreRun Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *ReactMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	// // 先判断群聊的功能启用情况
	// if !larkutils.CheckFunctionEnabling(*event.Event.Message.ChatId, consts.LarkFunctionRandomReact) {
	// 	span.RecordError(err)
	// 	return errors.Wrap(consts.ErrStageSkip, "ReactMsgOperator: Not enabled")
	// }
	return
}

// Run  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *ReactMsgOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	defer func() { span.RecordError(err) }()

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
		resp, err := lark.LarkClient.Im.V1.MessageReaction.Create(ctx, req)
		if err != nil {
			logs.L().Ctx(ctx).Error("reactMessage error", zap.Error(err), zap.String("TraceID", span.SpanContext().TraceID().String()))
			return err
		}
		logs.L().Ctx(ctx).Info("reactMessage", zap.Any("resp", resp))
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
				_, err = larkutils.ReplyMsgRawContentType(
					ctx,
					*event.Event.Message.MessageId,
					larkim.MsgTypeImage,
					content,
					"_imageReact",
					false,
				)
			} else {
				content, _ := sonic.MarshalString(map[string]string{
					"file_key": target.FileID,
				})
				_, err = larkutils.ReplyMsgRawContentType(
					ctx,
					*event.Event.Message.MessageId,
					larkim.MsgTypeSticker,
					content,
					"_imageReact",
					false,
				)
			}
		}
	}

	return
}
