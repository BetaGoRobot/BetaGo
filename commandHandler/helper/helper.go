package helper

import (
	"fmt"
	"strings"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/dbpack"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/khl"
)

// AdminCommandHelperHandler 查看帮助
//  @param targetID
//  @param quoteID
//  @param authorID
//  @return err
func AdminCommandHelperHandler(targetID, quoteID, authorID string) (err error) {
	// title := "嗨，你可以使用的指令如下:"
	// !对无参数指令，使用Button展示
	var commandInfoList []*dbpack.CommandInfo
	if dbpack.GetDbConnection().Table("betago.command_infos").Where("command_param_len=0").Order("command_name desc").Find(&commandInfoList).RowsAffected == 0 {
		err = fmt.Errorf("no command info found")
		return
	}

	var (
		modules = []interface{}{
			khl.CardMessageHeader{
				Text: khl.CardMessageElementText{
					Content: string(emoji.ComputerMouse) + "无参数指令:",
					Emoji:   true,
				},
			},
			khl.CardMessageActionGroup{},
		}
	)
	count := 0
	for _, commandInfo := range commandInfoList {
		count++
		if count%4 == 0 {
			modules = append(modules, khl.CardMessageActionGroup{})
		}
		modules[count/4+1] = append(modules[count/4+1].(khl.CardMessageActionGroup),
			khl.CardMessageElementButton{
				Theme: khl.CardThemeSuccess,
				Value: strings.ToUpper(strings.Trim(commandInfo.CommandName, "`")),
				Click: string(khl.CardMessageElementButtonClickReturnVal),
				Text:  strings.ToUpper(strings.Trim(commandInfo.CommandName, "`")),
			},
		)
	}

	// !对有参指令，使用文本展示
	commandInfoList = make([]*dbpack.CommandInfo, 0)
	if dbpack.GetDbConnection().Table("betago.command_infos").Where("command_param_len!=0").Order("command_name desc").Find(&commandInfoList).RowsAffected == 0 {
		err = fmt.Errorf("no command info found")
		return
	}
	modules = append(modules,
		khl.CardMessageHeader{
			Text: khl.CardMessageElementText{
				Content: string(emoji.ComputerMouse) + "含参数指令:",
				Emoji:   true,
			},
		},
		khl.CardMessageSection{
			Text: khl.CardMessageParagraph{
				Cols: 2,
				Fields: []interface{}{
					khl.CardMessageElementKMarkdown{
						Content: "**指令名称**",
					},
					khl.CardMessageElementKMarkdown{
						Content: "**指令功能**",
					},
				},
			},
		},
	)
	for _, commandInfo := range commandInfoList {
		modules = append(modules, khl.CardMessageSection{
			Text: khl.CardMessageParagraph{
				Cols: 2,
				Fields: []interface{}{
					khl.CardMessageElementKMarkdown{
						Content: commandInfo.CommandName,
					},
					khl.CardMessageElementKMarkdown{
						Content: commandInfo.CommandDesc,
					},
				},
			},
		})
	}

	cardMessageStr, err := khl.CardMessage{
		&khl.CardMessageCard{
			Theme:   "secondary",
			Size:    "lg",
			Modules: modules,
		},
	}.BuildMessage()

	if err != nil {
		err = fmt.Errorf("building cardMessage error %s", err.Error())
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
			TempTargetID: authorID,
		},
	)
	return
}

// UserCommandHelperHandler 查看帮助
//  @param targetID
//  @param quoteID
//  @param authorID
//  @return err
func UserCommandHelperHandler(targetID, quoteID, authorID string) (err error) {
	// 帮助信息

	// !对无参数指令，使用Button展示
	var commandInfoList []*dbpack.CommandInfo
	if dbpack.GetDbConnection().Table("betago.command_infos").Where("command_param_len=0 and command_type='user'").Order("command_name desc").Find(&commandInfoList).RowsAffected == 0 {
		err = fmt.Errorf("no command info found")
		return
	}

	var (
		modules = []interface{}{
			khl.CardMessageHeader{
				Text: khl.CardMessageElementText{
					Content: string(emoji.ComputerMouse) + "无参数指令:",
					Emoji:   true,
				},
			},
			khl.CardMessageActionGroup{},
		}
	)
	count := 0
	for _, commandInfo := range commandInfoList {
		count++
		if count%4 == 0 {
			modules = append(modules, khl.CardMessageActionGroup{})
		}
		modules[count/4+1] = append(modules[count/4+1].(khl.CardMessageActionGroup),
			khl.CardMessageElementButton{
				Theme: khl.CardThemeSuccess,
				Value: strings.ToUpper(strings.Trim(commandInfo.CommandName, "`")),
				Click: string(khl.CardMessageElementButtonClickReturnVal),
				Text:  strings.ToUpper(strings.Trim(commandInfo.CommandName, "`")),
			},
		)
	}

	// !对有参指令，使用文本展示
	commandInfoList = make([]*dbpack.CommandInfo, 0)
	if dbpack.GetDbConnection().Table("betago.command_infos").Where("command_param_len!=0 and command_type='user'").Order("command_name desc").Find(&commandInfoList).RowsAffected == 0 {
		err = fmt.Errorf("no command info found")
		return
	}
	modules = append(modules,
		khl.CardMessageHeader{
			Text: khl.CardMessageElementText{
				Content: string(emoji.ComputerMouse) + "含参数指令:",
				Emoji:   true,
			},
		},
		khl.CardMessageSection{
			Text: khl.CardMessageParagraph{
				Cols: 2,
				Fields: []interface{}{
					khl.CardMessageElementKMarkdown{
						Content: "**指令名称**",
					},
					khl.CardMessageElementKMarkdown{
						Content: "**指令功能**",
					},
				},
			},
		},
	)
	for _, commandInfo := range commandInfoList {
		modules = append(modules, khl.CardMessageSection{
			Text: khl.CardMessageParagraph{
				Cols: 2,
				Fields: []interface{}{
					khl.CardMessageElementKMarkdown{
						Content: commandInfo.CommandName,
					},
					khl.CardMessageElementKMarkdown{
						Content: commandInfo.CommandDesc,
					},
				},
			},
		})
	}

	cardMessageStr, err := khl.CardMessage{
		&khl.CardMessageCard{
			Theme:   "secondary",
			Size:    "lg",
			Modules: modules,
		},
	}.BuildMessage()

	if err != nil {
		err = fmt.Errorf("building cardMessage error %s", err.Error())
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
			TempTargetID: authorID,
		},
	)
	return
}
