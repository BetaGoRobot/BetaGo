package helper

import (
	"fmt"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/lonelyevil/khl"
)

// GetUserInfoHandler 获取用户信息
//  @param userID
//  @param guildID
//  @return err
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
	cardMessageStr, err := khl.CardMessage{&khl.CardMessageCard{
		Theme:   khl.CardThemePrimary,
		Size:    khl.CardSizeLg,
		Modules: cardMessageModules,
	}}.BuildMessage()
	if err != nil {
		return err
	}
	_, err = betagovar.GlobalSession.MessageCreate(
		&khl.MessageCreate{
			MessageCreateBase: khl.MessageCreateBase{
				Type:     khl.MessageTypeCard,
				TargetID: targetID,
				Content:  cardMessageStr,
				Quote:    quoteID,
			},
		},
	)
	return

}
