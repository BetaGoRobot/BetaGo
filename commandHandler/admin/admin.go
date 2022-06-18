package admin

import (
	"fmt"
	"strconv"
	"strings"

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
	dbpack.GetDbConnection().Table("betago.administrators").Find(&admins).Order("level DESC")
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
//  @return err PASS
func AddAdminHandler(TargetID, QuoteID, authorID string, args ...string) (err error) {
	var (
		succUserID []string
		ec         utility.ErrorCollector
	)
	if len(args) != 0 {
		for _, arg := range args {
			userID := strings.Trim(arg, "(met)")
			// 先检验是否存在
			if dbpack.GetDbConnection().Table("betago.administrators").Where("user_id = ?", utility.MustAtoI(userID)).Find(&dbpack.Administrator{}).RowsAffected != 0 {
				// 存在则不处理，返回信息
				return fmt.Errorf(fmt.Sprintf(`(met)%s(met) 已经是管理员了`, userID))
			}
			userInfo, err := utility.GetUserInfo(userID, "")
			if err != nil {
				return err
			}
			// 创建管理员
			dbRes := dbpack.GetDbConnection().Table("betago.administrators").
				Create(
					&dbpack.Administrator{
						UserID:   int64(utility.MustAtoI(userID)),
						UserName: userInfo.Nickname,
						Level:    1,
					},
				)
			if dbRes.Error != nil {
				ec.Collect(dbRes.Error)
				continue
			}
			succUserID = append(succUserID, userID)
		}
		if !ec.NoError() {
			return ec.GenErr()
		}
	} else {
		return fmt.Errorf("请输入用户ID")
	}
	var succStr string
	for _, userID := range succUserID {
		succStr += fmt.Sprintf("(met)%s(met) 已被设置为管理员\n", userID)
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
						Content: succStr,
					},
				},
			},
		},
	}.BuildMessage()
	if err != nil {
		ec.Collect(err)
		return ec.GenErr()
	}
	betagovar.GlobalSession.MessageCreate(
		&khl.MessageCreate{
			MessageCreateBase: khl.MessageCreateBase{
				Type:     khl.MessageTypeCard,
				TargetID: TargetID,
				Content:  cardMessageStr,
				Quote:    QuoteID,
			},
		},
	)
	if !ec.NoError() {
		err = ec.GenErr()
	}
	return
}

// RemoveAdminHandler 删除管理员
//  @param userID
//  @param targetUserID
//  @param QuoteID
//  @param TargetID
//  @return err PASS
func RemoveAdminHandler(TargetID, QuoteID, authorID string, args ...string) (err error) {
	var (
		ec         utility.ErrorCollector
		succUserID []string
	)
	if len(args) != 0 {
		// 参数有效
		for _, arg := range args {
			userID := strings.Trim(arg, "(met)")
			// 先检验是否存在
			if !dbpack.CheckIsAdmin(userID) {
				// 不存在则不处理，返回信息
				return fmt.Errorf(fmt.Sprintf(`(met)%s(met) 不是管理员`, userID))
			}
			// 等级校验
			if userLevel, targetLevel := dbpack.GetAdminLevel(authorID), dbpack.GetAdminLevel(userID); userLevel <= targetLevel && userID != authorID {
				// 等级不足，无权限操作
				err = fmt.Errorf("您的等级小于或等于目标用户，无权限操作")
				return
			}
			// 删除管理员
			dbRes := dbpack.GetDbConnection().Table("betago.administrators").
				Where("user_id = ?", utility.MustAtoI(userID)).
				Unscoped().
				Delete(&dbpack.Administrator{})
			if dbRes.Error != nil {
				ec.Collect(dbRes.Error)
				continue
			}
			succUserID = append(succUserID, userID)
		}
	} else {
		return fmt.Errorf("参数不足")
	}

	var succStr string
	for _, userID := range succUserID {
		succStr += fmt.Sprintf("(met)%s(met) 管理员已被移除\n", userID)
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
						Content: succStr,
					},
				},
			},
		},
	}.BuildMessage()
	if err != nil {
		ec.Collect(err)
		return ec.GenErr()
	}
	betagovar.GlobalSession.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeCard,
			TargetID: TargetID,
			Content:  cardMessageStr,
			Quote:    QuoteID,
		},
	})
	if !ec.NoError() {
		err = ec.GenErr()
	}
	return
}
