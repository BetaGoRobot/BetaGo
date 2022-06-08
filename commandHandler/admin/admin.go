package admin

import (
	"fmt"
	"strconv"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/dbpack"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/lonelyevil/khl"
)

// ShowAdminHandler 显示管理员
//  @param targetID
//  @param quoteID
//  @return err
func ShowAdminHandler(targetID, quoteID string) (err error) {
	admins := make([]dbpack.Administrator, 0)
	dbpack.GetDbConnection().Table("betago.administrators").Find(&admins).Order("level desc")
	modules := make([]interface{}, 0)
	modules = append(modules,
		khl.CardMessageSection{
			Text: khl.CardMessageParagraph{
				Cols: 3,
				Fields: []interface{}{
					khl.CardMessageElementKMarkdown{
						Content: "**用户名**",
					},
					khl.CardMessageElementKMarkdown{
						Content: "**用户ID**",
					},
					khl.CardMessageElementKMarkdown{
						Content: "**管理等级**",
					},
				},
			},
		})
	for _, admin := range admins {
		modules = append(modules,
			khl.CardMessageSection{
				Text: khl.CardMessageParagraph{
					Cols: 3,
					Fields: []interface{}{
						khl.CardMessageElementKMarkdown{
							Content: fmt.Sprintf(`(met)%d(met)`, admin.UserID),
						},
						khl.CardMessageElementKMarkdown{
							Content: strconv.Itoa(int(admin.UserID)),
						},
						khl.CardMessageElementKMarkdown{
							Content: strconv.Itoa(int(admin.Level)),
						},
					},
				},
			},
		)
	}
	cardMessageStr, err := khl.CardMessage{
		&khl.CardMessageCard{
			Theme:   "secondary",
			Size:    "lg",
			Modules: modules,
		},
	}.BuildMessage()
	if err != nil {
		return
	}
	betagovar.GlobalSession.MessageCreate(
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

// AddAdminHandler 增加管理员
//  @param userID
//  @param userName
//  @param QuoteID
//  @param TargetID
//  @return err
func AddAdminHandler(userID, userName, QuoteID, TargetID string) (err error) {
	// 先检验是否存在
	if dbpack.GetDbConnection().Table("betago.administrators").Where("user_id = ?", utility.MustAtoI(userID)).Find(&dbpack.Administrator{}).RowsAffected != 0 {
		// 存在则不处理，返回信息
		return fmt.Errorf(fmt.Sprintf(`(met)%s(met) 已经是管理员了`, userID))
	}
	// 创建管理员
	dbRes := dbpack.GetDbConnection().Table("betago.administrators").
		Create(
			&dbpack.Administrator{
				UserID:   int64(utility.MustAtoI(userID)),
				UserName: userName,
				Level:    1,
			},
		)
	if dbRes.Error != nil {
		return dbRes.Error
	}

	cardMessageStr, err := khl.CardMessage{
		&khl.CardMessageCard{
			Theme: "secondary",
			Size:  "lg",
			Modules: []interface{}{
				khl.CardMessageHeader{
					Text: khl.CardMessageElementText{
						Content: "指令执行成功~~",
						Emoji:   false,
					},
				},
				khl.CardMessageSection{
					Text: khl.CardMessageElementKMarkdown{
						Content: fmt.Sprintf("%s 已被设置为管理员, 让我们祝贺这个B~ (met)%s(met)", userName, userID),
					},
				},
			},
		},
	}.BuildMessage()

	betagovar.GlobalSession.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeCard,
			TargetID: TargetID,
			Content:  cardMessageStr,
			Quote:    QuoteID,
		},
	})
	return
}

// RemoveAdminHandler 删除管理员
//  @param userID
//  @param targetUserID
//  @param QuoteID
//  @param TargetID
//  @return err
func RemoveAdminHandler(userID, targetUserID, QuoteID, TargetID string) (err error) {
	// 先判断目标用户是否为管理员
	if !dbpack.CheckIsAdmin(targetUserID) {
		err = fmt.Errorf("UserID=%s 不是管理员，无法删除", targetUserID)
		return
	}
	if userLevel, targetLevel := dbpack.GetAdminLevel(userID), dbpack.GetAdminLevel(targetUserID); userLevel <= targetLevel && userID != targetUserID {
		// 等级不足，无权限操作
		err = fmt.Errorf("您的等级小于或等于目标用户，无权限操作")
		return
	}
	// 删除管理员
	dbpack.GetDbConnection().Table("betago.administrators").Where("user_id = ?", utility.MustAtoI(targetUserID)).Unscoped().Delete(&dbpack.Administrator{})
	cardMessageStr, err := khl.CardMessage{
		&khl.CardMessageCard{
			Theme: "secondary",
			Size:  "lg",
			Modules: []interface{}{
				khl.CardMessageHeader{
					Text: khl.CardMessageElementText{
						Content: "指令执行成功~~",
						Emoji:   false,
					},
				},
				khl.CardMessageSection{
					Text: khl.CardMessageElementKMarkdown{
						Content: fmt.Sprintf("(met)%s(met), 这个B很不幸的被(met)%s(met)取消了管理员资格~ ", targetUserID, userID),
					},
				},
			},
		},
	}.BuildMessage()
	betagovar.GlobalSession.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeCard,
			TargetID: TargetID,
			Content:  cardMessageStr,
			Quote:    QuoteID,
		},
	})
	return
}
