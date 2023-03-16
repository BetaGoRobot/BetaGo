package admin

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/lonelyevil/kook"
)

// ShowAdminHandler 显示管理员
//
//	@param targetID
//	@param quoteID
//	@return err
func ShowAdminHandler(TargetID, QuoteID, authorID string, args ...string) (err error) {
	admins := make([]utility.Administrator, 0)
	utility.GetDbConnection().Table("betago.administrators").Find(&admins).Order("level DESC")
	modules := make([]interface{}, 0)
	modules = append(modules,
		kook.CardMessageSection{
			Text: kook.CardMessageParagraph{
				Cols: 3,
				Fields: []interface{}{
					kook.CardMessageElementKMarkdown{
						Content: "**用户名**",
					},
					kook.CardMessageElementKMarkdown{
						Content: "**用户ID**",
					},
					kook.CardMessageElementKMarkdown{
						Content: "**管理等级**",
					},
				},
			},
		})
	for _, admin := range admins {
		info, err := utility.GetUserInfo(strconv.Itoa(int(admin.UserID)), "")
		if err != nil {
			return err
		}
		modules = append(modules,
			kook.CardMessageSection{
				Text: kook.CardMessageParagraph{
					Cols: 3,
					Fields: []interface{}{
						kook.CardMessageElementKMarkdown{
							Content: fmt.Sprintf("`%s`", info.Nickname),
						},
						kook.CardMessageElementKMarkdown{
							Content: strconv.Itoa(int(admin.UserID)),
						},
						kook.CardMessageElementKMarkdown{
							Content: strconv.Itoa(int(admin.Level)),
						},
					},
				},
			},
		)
	}
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme:   "secondary",
			Size:    "lg",
			Modules: modules,
		},
	}.BuildMessage()
	if err != nil {
		return
	}
	betagovar.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: TargetID,
				Content:  cardMessageStr,
				Quote:    QuoteID,
			},
		},
	)
	return
}

// AddAdminHandler 增加管理员
//
//	@param userID
//	@param userName
//	@param QuoteID
//	@param TargetID
//	@return err PASS
func AddAdminHandler(TargetID, QuoteID, authorID string, args ...string) (err error) {
	var (
		succUserID []string
		ec         utility.ErrorCollector
	)
	if len(args) != 0 {
		for _, arg := range args {
			userID := strings.Trim(arg, "(met)")
			// 先检验是否存在
			if utility.GetDbConnection().Table("betago.administrators").Where("user_id = ?", utility.MustAtoI(userID)).Find(&utility.Administrator{}).RowsAffected != 0 {
				// 存在则不处理，返回信息
				return fmt.Errorf(fmt.Sprintf(`(met)%s(met) 已经是管理员了`, userID))
			}
			userInfo, err := utility.GetUserInfo(userID, "")
			if err != nil {
				return err
			}
			// 创建管理员
			dbRes := utility.GetDbConnection().Table("betago.administrators").
				Create(
					&utility.Administrator{
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
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "secondary",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: "指令执行成功~~",
						Emoji:   false,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
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
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
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
//
//	@param userID
//	@param targetUserID
//	@param QuoteID
//	@param TargetID
//	@return err PASS
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
			if !utility.CheckIsAdmin(userID) {
				// 不存在则不处理，返回信息
				return fmt.Errorf(fmt.Sprintf(`(met)%s(met) 不是管理员`, userID))
			}
			// 等级校验
			if userLevel, targetLevel := utility.GetAdminLevel(authorID), utility.GetAdminLevel(userID); userLevel <= targetLevel && userID != authorID {
				// 等级不足，无权限操作
				err = fmt.Errorf("您的等级小于或等于目标用户，无权限操作")
				return
			}
			// 删除管理员
			dbRes := utility.GetDbConnection().Table("betago.administrators").
				Where("user_id = ?", utility.MustAtoI(userID)).
				Unscoped().
				Delete(&utility.Administrator{})
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
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "secondary",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: "指令执行成功~~",
						Emoji:   false,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
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
	betagovar.GlobalSession.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     kook.MessageTypeCard,
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

// DeleteAllMessageHandler 删除频道内所有消息
//
//	@param TargetID
//	@param QuoteID
//	@param authorID
//	@param args
//	@return err
func DeleteAllMessageHandler(TargetID, QuoteID, authorID string, args ...string) (err error) {
	var (
		ec         utility.ErrorCollector
		messageNum int
	)
	defer cleaupData()
	if len(args) != 0 {
		messageNum, err = strconv.Atoi(args[0])
		if err != nil {
			return
		}
	}
	if messageNum > 50 {
		for i := 0; i < messageNum/50; i++ {
			ms, err := betagovar.GlobalSession.MessageList(TargetID, kook.MessageListWithPageSize(50))
			if err != nil {
				ec.Collect(err)
			}
			for i := len(ms) - 1; i >= 0; i-- {
				err := betagovar.GlobalSession.MessageDelete(ms[i].ID)
				backupData(ms[i].Author.Username, ms[i].Content, ms[i].ID, TargetID)
				ec.Collect(err)
			}
		}
	} else {
		ms, err := betagovar.GlobalSession.MessageList(TargetID, kook.MessageListWithPageSize(messageNum))
		if err != nil {
			ec.Collect(err)
		}
		ms = ms[:messageNum]
		if len(ms) > 50 || len(ms) > messageNum || messageNum <= 0 {
			err = fmt.Errorf("若全部删除，需要删除的消息数量>50，高危操作，请确认后`指定需要删除的消息数量`完成操作")
			return err
		}
		for i := len(ms) - 1; i >= 0; i-- {
			err := betagovar.GlobalSession.MessageDelete(ms[i].ID)
			ec.Collect(err)
			msg, err := getStringFromNode(ms[i].Content)
			ec.Collect(err)
			backupData(ms[i].Author.Username, msg, ms[i].ID, TargetID)
		}
	}
	return ec.CheckError()
}
