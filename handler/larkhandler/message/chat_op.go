package message

import (
	"context"

	"github.com/pkg/errors"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler/lark_command/handlers"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

var _ Op = &ChatMsgOperator{}

// ChatMsgOperator  RepeatMsg Op
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:51
type ChatMsgOperator struct {
	OpBase
}

// PreRun Repeat
//
//	@receiver r *ImitateMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:35
func (r *ChatMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	// 先判断群聊的功能启用情况
	if !larkutils.CheckFunctionEnabling(*event.Event.Message.ChatId, consts.LarkFunctionRandomRepeat) {
		return errors.Wrap(consts.ErrStageSkip, "ImitateMsgOperator: Not enabled")
	}
	if larkutils.IsCommand(ctx, larkutils.PreGetTextMsg(ctx, event)) {
		return errors.Wrap(consts.ErrStageSkip, "ImitateMsgOperator: Is Command")
	}
	return
}

// Run Repeat
//
//	@receiver r *ImitateMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:41
func (r *ChatMsgOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	// 开始摇骰子, 默认概率10%
	realRate := utility.MustAtoI(utility.GetEnvWithDefault("IMITATE_DEFAULT_RATE", "10"))
	// 群聊定制化
	config, hitCache := database.FindByCacheFunc(
		database.ImitateRateCustom{
			GuildID: *event.Event.Message.ChatId,
		},
		func(d database.ImitateRateCustom) string {
			return d.GuildID
		},
	)
	span.SetAttributes(attribute.Bool("ChatRateCustom hitCache", hitCache))

	if len(config) != 0 {
		realRate = config[0].Rate
	}

	if utility.Probability(float64(realRate) / 100) {
		// sendMsg
		err := handlers.ChatHandler(ctx, event)
		if err != nil {
			return err
		}
	}
	return
}
