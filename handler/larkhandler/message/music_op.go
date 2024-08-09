package message

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/cardutil"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

var _ handlerbase.Operator[larkim.P2MessageReceiveV1] = &MusicMsgOperator{}

// MusicMsgOperator Repeat
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:36:07
type MusicMsgOperator struct {
	handlerbase.OperatorBase[larkim.P2MessageReceiveV1]
}

// PreRun Music
//
//	@receiver r *MusicMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:34:09
func (r *MusicMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	if !larkutils.IsMentioned(event.Event.Message.Mentions) {
		return errors.Wrap(consts.ErrStageSkip, "MusicMsgOperator: Not Mentioned")
	}
	if event.Event.Message.ParentId != nil {
		return errors.Wrap(consts.ErrStageSkip, "MusicMsgOperator: Has ParentId")
	}
	if larkutils.IsCommand(ctx, larkutils.PreGetTextMsg(ctx, event)) {
		return errors.Wrap(consts.ErrStageSkip, "MusicMsgOperator: Is Command")
	}
	return
}

// Run  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *MusicMsgOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()
	defer span.RecordError(err)

	msg := larkutils.PreGetTextMsg(ctx, event)
	msg = larkutils.TrimAtMsg(ctx, msg)
	keywords := []string{msg}

	res, err := neteaseapi.NetEaseGCtx.SearchMusicByKeyWord(ctx, keywords...)
	if err != nil {
		return err
	}
	cardContent, err := cardutil.SendMusicListCard(ctx, res, cardutil.MusicItemNoTrans, neteaseapi.CommentTypeSong)
	if err != nil {
		return
	}
	err = larkutils.ReplyMsgRawContentType(ctx, *event.Event.Message.MessageId, larkim.MsgTypeInteractive, cardContent, "_RunMusicOp", false)
	if err != nil {
		log.ZapLogger.Error("send music list error", zaplog.Error(err))
		return err
	}
	return
}
