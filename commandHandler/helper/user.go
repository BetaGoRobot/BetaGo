package helper

import (
	"fmt"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/lonelyevil/kook"
)

// GetUserInfoHandler 获取用户信息
//
//	@param userID
//	@param guildID
//	@return err
func GetUserInfoHandler(targetID, quoteID, authorID string, guildID string, args ...string) (err error) {
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
		Theme:   kook.CardThemePrimary,
		Size:    kook.CardSizeLg,
		Modules: cardMessageModules,
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
