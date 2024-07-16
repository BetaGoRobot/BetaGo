package larkcards

import (
	"context"
	"fmt"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func PreGetTextMsg(ctx context.Context, event *larkim.P2MessageReceiveV1) string {
	msgMap := make(map[string]interface{})
	msg := *event.Event.Message.Content
	err := sonic.UnmarshalString(msg, &msgMap)
	if err != nil {
		log.ZapLogger.Error("repeatMessage", zaplog.Error(err))
		return ""
	}
	if text, ok := msgMap["text"]; ok {
		msg = text.(string)
	}
	return msg
}

func QuoteMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	msg := PreGetTextMsg(ctx, event)
	if strings.Contains(msg, "下班") {
		req := larkim.NewReplyMessageReqBuilder().
			Body(
				larkim.NewReplyMessageReqBodyBuilder().
					Content(larkim.NewTextMsgBuilder().Text("这么早你就惦记着下班了?").Build()).
					MsgType(larkim.MsgTypeText).
					ReplyInThread(true).
					Uuid(*event.Event.Message.MessageId).
					Build(),
			).MessageId(*event.Event.Message.MessageId).
			Build()
		_, subSpan := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
		resp, err := larkutils.LarkClient.Im.V1.Message.Reply(ctx, req)
		subSpan.End()
		if err != nil {
			fmt.Println(resp)
			return err
		}
		fmt.Println(resp.CodeError.Msg)
	}
	return
}
