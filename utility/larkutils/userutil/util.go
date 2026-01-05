package userutil

import (
	"context"
	"errors"

	"github.com/BetaGoRobot/BetaGo/dal/lark"
	"github.com/BetaGoRobot/BetaGo/utility/cache"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/grouputil"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	"go.uber.org/zap"
)

func GetUserInfo(ctx context.Context, userID string) (user *larkcontact.User, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	resp, err := lark.LarkClient.Contact.V3.User.Get(ctx, larkcontact.NewGetUserReqBuilder().UserId(userID).Build())
	if err != nil {
		return
	}
	if !resp.Success() {
		err = errors.New(resp.Error())
		return
	}
	return resp.Data.User, nil
}

func GetUserInfoCache(ctx context.Context, chatID, userID string) (user *larkcontact.User, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	res, err := cache.GetOrExecute(ctx, userID, func() (*larkcontact.User, error) {
		return GetUserInfo(ctx, userID)
	})
	logs.L().Ctx(ctx).Info("GetUserInfoCache", zap.Any("user", res))
	// userInfo失败了，走群聊试试
	groupMember, err := grouputil.GetUserMemberFromChat(ctx, chatID, userID)
	if err != nil {
		logs.L().Ctx(ctx).Error("GetUserMemberFromChat", zap.Any("user", groupMember))
		return
	}
	if groupMember == nil {
		err = errors.New("user not found in chat")
		return
	}
	res = &larkcontact.User{
		UserId: groupMember.MemberId,
		OpenId: &userID,
		Name:   groupMember.Name,
	}
	return res, err
}
