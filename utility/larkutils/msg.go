package larkutils

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// PreGetTextMsg 获取消息内容
//
//	@param ctx
//	@param event
//	@return string
func PreGetTextMsg(ctx context.Context, event *larkim.P2MessageReceiveV1) string {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
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
