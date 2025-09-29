package larkutils

import (
	"context"
	"errors"

	"github.com/BetaGoRobot/BetaGo/dal/lark"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func GetChatName(ctx context.Context, chatID string) (chatName string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	resp, err := lark.LarkClient.Im.V1.Chat.Get(ctx, larkim.NewGetChatReqBuilder().ChatId(chatID).Build())
	if err != nil {
		return
	}
	if !resp.Success() {
		err = errors.New(resp.Error())
		return
	}
	chatName = *resp.Data.Name
	return
}

func GetChatIDFromMsgID(ctx context.Context, msgID string) (chatID string, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	resp := GetMsgFullByID(ctx, msgID)
	if !resp.Success() {
		err = errors.New(resp.Error())
		return
	}
	chatID = *resp.Data.Items[0].ChatId
	return
}
