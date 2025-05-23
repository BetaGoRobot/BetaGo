package helper

import (
	"context"
	"fmt"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/lonelyevil/kook"
	"go.opentelemetry.io/otel/attribute"
)

// GetUserInfoHandler 获取用户信息
//
//	@param userID
//	@param guildID
//	@return err
func GetUserInfoHandler(ctx context.Context, targetID, quoteID, authorID string, guildID string, args ...string) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	var userID string
	if len(args) == 1 {
		userID = args[0]
	} else {
		return fmt.Errorf("参数错误")
	}
	if userID == "" {
		return fmt.Errorf("userID is empty")
	}
	userInfo, err := utility.GetUserInfo(userID, guildID)
	if err != nil {
		return err
	}
	cardMessageModules, err := utility.BuildCardMessageCols("用户信息项", "具体信息", utility.Struct2Map(*userInfo))
	if err != nil {
		return err
	}
	cardMessageStr, err := utility.BuildCardMessage(
		string(kook.CardThemePrimary),
		string(kook.CardSizeLg),
		"",
		quoteID,
		span,
		cardMessageModules...,
	)
	if err != nil {
		return err
	}
	_, err = consts.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: targetID,
				Content:  cardMessageStr,
				Quote:    quoteID,
			},
		},
	)
	return
}
