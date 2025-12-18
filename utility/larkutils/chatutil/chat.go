package chatutil

import (
	"context"
	"errors"

	"github.com/BetaGoRobot/BetaGo/dal/lark"
	"github.com/BetaGoRobot/BetaGo/utility/cache"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func GetChatInfo(ctx context.Context, chatID string) (chat *larkim.GetChatRespData, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	req := larkim.NewGetChatReqBuilder().ChatId(chatID).Build()
	resp, err := lark.LarkClient.Im.V1.Chat.Get(ctx, req)
	if err != nil {
		return
	}
	if !resp.Success() {
		err = errors.New(resp.Error())
		return
	}
	return resp.Data, nil
}

func GetChatInfoCache(ctx context.Context, chatID string) (chat *larkim.GetChatRespData, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	return cache.GetOrExecute(ctx, chatID, func() (chat *larkim.GetChatRespData, err error) {
		return GetChatInfo(ctx, chatID)
	})
}
