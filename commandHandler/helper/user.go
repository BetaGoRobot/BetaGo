package helper

import (
	"context"
	"fmt"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/lonelyevil/kook"
	"go.opentelemetry.io/otel/attribute"
)

// GetUserInfoHandler 获取用户信息
//
//	@param userID
//	@param guildID
//	@return err
func GetUserInfoHandler(ctx context.Context, targetID, quoteID, authorID string, guildID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
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
	cardMessageStr, err := kook.CardMessage{&kook.CardMessageCard{
		Theme: kook.CardThemePrimary,
		Size:  kook.CardSizeLg,
		Modules: append(
			cardMessageModules,
			&kook.CardMessageDivider{},
			&kook.CardMessageSection{
				Mode: kook.CardMessageSectionModeRight,
				Text: &kook.CardMessageElementKMarkdown{
					Content: "TraceID: `" + span.SpanContext().TraceID().String() + "`",
				},
				Accessory: kook.CardMessageElementButton{
					Theme: kook.CardThemeSuccess,
					Value: "https://jaeger.kevinmatt.top/trace/" + span.SpanContext().TraceID().String(),
					Click: "link",
					Text:  "链路追踪",
				},
			},
		),
	}}.BuildMessage()
	if err != nil {
		return err
	}
	_, err = betagovar.GlobalSession.MessageCreate(
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
