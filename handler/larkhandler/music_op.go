package larkhandler

import (
	"context"
	"strings"

	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

var _ LarkMsgOperator = &MusicMsgOperator{}

// MusicMsgOperator Repeat
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:36:07
type MusicMsgOperator struct {
	LarkMsgOperatorBase
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
		return errors.Wrap(ErrStageSkip, "MusicMsgOperator: Not Mentioned")
	}
	if event.Event.Message.ParentId != nil {
		return errors.Wrap(ErrStageSkip, "MusicMsgOperator: Has ParentId")
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

	msg := larkutils.PreGetTextMsg(ctx, event)
	msg = larkutils.TrimAtMsg(ctx, msg)
	keywords := []string{msg}
	if keyword := strings.ToLower(strings.Join(keywords, " ")); keyword == "try panic" {
		panic("try panic!")
	}

	res, err := neteaseapi.NetEaseGCtx.SearchMusicByKeyWord(ctx, keywords...)
	if err != nil {
		return err
	}
	listMsg := larkutils.NewSearchListCard()
	for _, item := range res {
		var invalid bool
		if item.SongURL == "" { // 无效歌曲
			invalid = true
		}
		listMsg.AddColumn(ctx, item.ImageKey, item.Name, item.ArtistName, item.ID, invalid)
	}
	listMsg.AddJaegerTracer(ctx, span)
	cardStr, err := sonic.MarshalString(listMsg)
	if err != nil {
		return err
	}

	log.ZapLogger.Info("send music list", zaplog.Any("cardStr", cardStr))
	req := larkim.NewReplyMessageReqBuilder().
		Body(
			larkim.NewReplyMessageReqBodyBuilder().
				Content(cardStr).
				MsgType(larkim.MsgTypeInteractive).
				ReplyInThread(true).
				Uuid(*event.Event.Message.MessageId).
				Build(),
		).MessageId(*event.Event.Message.MessageId).
		Build()
	_, subSpan := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	resp, err := larkutils.LarkClient.Im.V1.Message.Reply(ctx, req)
	subSpan.End()

	if err != nil {
		log.ZapLogger.Error("send music list error", zaplog.Error(err))
		return err
	}
	log.ZapLogger.Info("send music list", zaplog.Any("msg", resp.CodeError.Msg))
	return
}
