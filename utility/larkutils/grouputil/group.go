package grouputil

import (
	"context"
	"errors"

	"github.com/BetaGoRobot/BetaGo/dal/lark"
	"github.com/BetaGoRobot/BetaGo/utility/cache"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func GetUserMemberFromChat(ctx context.Context, chatID, openID string) (member *larkim.ListMember, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	memberMap, err := GetUserMapFromChatIDCache(ctx, chatID)
	if err != nil {
		return
	}
	return memberMap[openID], err
}

func GetUserMapFromChatIDCache(ctx context.Context, chatID string) (memberMap map[string]*larkim.ListMember, err error) {
	return cache.GetOrExecute(chatID, func() (map[string]*larkim.ListMember, error) {
		return GetUserMapFromChatID(ctx, chatID)
	})
}

func GetUserMapFromChatID(ctx context.Context, chatID string) (memberMap map[string]*larkim.ListMember, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

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
		resp, err := lark.LarkClient.Im.ChatMembers.Get(ctx, builder.Build())
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
