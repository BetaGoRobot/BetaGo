package message

import (
	"context"

	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

var _ Op = &RepeatMsgOperator{}

// RecordMsgOperator  RepeatMsg Op
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:51
type RecordMsgOperator struct {
	OpBase
}

// PreRun Repeat
//
//	@receiver r *RepeatMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:35
func (r *RecordMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

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
func (r *RecordMsgOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	imgSeq, err := larkutils.GetAllImageFromMsgEvent(ctx, event.Event.Message)
	if err != nil {
		return
	}
	if imgSeq != nil {
		for imageKey := range imgSeq {
			err = larkutils.DownImgFromMsgAsync(
				ctx,
				*event.Event.Message.MessageId,
				larkim.MsgTypeImage,
				imageKey,
			)
			if err != nil {
				return err
			}
		}
	}
	return
}
