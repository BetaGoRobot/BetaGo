package larkutils

import (
	"context"
	"errors"

	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

var LarkClient *lark.Client = lark.NewClient(env.LarkAppID, env.LarkAppSecret)

func GetUserMapFromChatID(ctx context.Context, chatID string) (memberMap map[string]*larkim.ListMember, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	memberMap = make(map[string]*larkim.ListMember)
	hasMore := true
	pageToken := ""
	for hasMore {
		builder := larkim.
			NewGetChatMembersReqBuilder().
			MemberIdType(`open_id`).
			ChatId(chatID).
			PageSize(100)
		if pageToken != "" {
			builder.PageToken(pageToken)
		}
		resp, err := LarkClient.Im.ChatMembers.Get(ctx, builder.Build())
		if err != nil {
			return memberMap, err
		}
		if !resp.Success() {
			err = errors.New(resp.Error())
			return memberMap, err
		}
		for _, item := range resp.Data.Items {
			memberMap[*item.MemberId] = item
		}
		hasMore = *resp.Data.HasMore
		pageToken = *resp.Data.PageToken
	}
	return
}

func GetUserMemberFromChat(ctx context.Context, chatID, openID string) (member *larkim.ListMember, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	memberMap, err := GetUserMapFromChatID(ctx, chatID)
	if err != nil {
		return
	}
	return memberMap[openID], err
}

func GetChatName(ctx context.Context, chatID string) (chatName string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	resp, err := LarkClient.Im.V1.Chat.Get(ctx, larkim.NewGetChatReqBuilder().ChatId(chatID).Build())
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
