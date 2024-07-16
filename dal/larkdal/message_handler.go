package larkdal

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func MessageV2Handler(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer larkutils.RecoverMsg(ctx, *event.Event.Message.MessageId)
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()

	stamp, err := strconv.ParseInt(*event.Event.Message.CreateTime, 10, 64)
	if err != nil {
		panic(err)
	}
	if time.Now().Sub(time.Unix(stamp/1000, 0)) > time.Second*10 {
		return nil
	}
	if !IsMentioned(event.Event.Message.Mentions) || *event.Event.Sender.SenderId.OpenId == BotOpenID {
		larkhandler.MainLarkHandler.RunParallelStages(ctx, event)
		return nil
	}
	msgMap := make(map[string]interface{})
	msg := *event.Event.Message.Content
	err = sonic.UnmarshalString(msg, &msgMap)
	if err != nil {
		return err
	}
	if text, ok := msgMap["text"]; ok {
		msg = text.(string)
	}
	go getMusicAndSend(ctx, event, msg)
	fmt.Println(larkcore.Prettify(event))
	return nil
}

func getMusicAndSend(ctx context.Context, event *larkim.P2MessageReceiveV1, msg string) (err error) {
	defer larkutils.RecoverMsg(ctx, *event.Event.Message.MessageId)
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()

	keywords := strings.Split(msg, " ")[1:]
	if keyword := strings.ToLower(strings.Join(keywords, " ")); keyword == "try panic" {
		panic("try panic!")
	}
	res, err := neteaseapi.NetEaseGCtx.SearchMusicByKeyWord(ctx, keywords...)
	if err != nil {
		return err
	}
	listMsg := NewSearchListCard()
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
