package helper

import (
	"fmt"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/lonelyevil/khl"
)

var commandHelper = map[string]string{
	"`help`":        "查看帮助 \n`@BetaGo` `help`",
	"`ping`":        "检查机器人是否运行正常 \n`@BetaGo` `ping`",
	"`roll`":        "掷骰子 \n`@BetaGo` `roll`",
	"`addAdmin`":    "添加管理员 \n`@BetaGo` `addAdmin <userID> <userName>`",
	"`removeAdmin`": "移除管理员 \n`@BetaGo` `removeAdmin <userID>`",
	"`showAdmin`":   "显示所有管理员 \n`@BetaGo` `showAdmin`",
}

// CommandHelperHandler 查看帮助
//  @param targetID
//  @param quoteID
//  @param authorID
//  @return err
func CommandHelperHandler(targetID, quoteID, authorID string) (err error) {
	// 帮助信息
	var modules []interface{}
	modules = append(modules,
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
	for command, helper := range commandHelper {
		modules = append(modules,
			khl.CardMessageSection{
				Text: khl.CardMessageParagraph{
					Cols: 2,
					Fields: []interface{}{
						khl.CardMessageElementKMarkdown{
							Content: command,
						},
						khl.CardMessageElementKMarkdown{
							Content: helper,
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
