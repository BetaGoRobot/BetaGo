package commandcli

import (
	"context"
	"errors"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func GetIDHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	req := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().
			Content(larkim.NewTextMsgBuilder().Text(*data.Event.Message.ParentId).Build()).
			MsgType(larkim.MsgTypeText).
			ReplyInThread(true).
			Uuid(*data.Event.Message.MessageId + "reply").
			Build(),
	).MessageId(*data.Event.Message.MessageId).Build()

	resp, err := larkutils.LarkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		return
	}
	if !resp.Success() {
		return errors.New(resp.Error())
	}
	return
}
