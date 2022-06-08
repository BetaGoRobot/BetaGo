package roll

import (
	"fmt"
	"math/rand"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/khl"
)

// RandRollHandler 随机抽取一个数字
//  @param targetID 目标ID
//  @param quoteID 引用ID
//  @param authorID 发送者ID
//  @return err 错误信息
func RandRollHandler(targetID, quoteID, authorID string) (err error) {
	point := rand.Intn(6) + 1
	var extraStr string
	if point > 3 {
		extraStr = "运气不错呀！"
	} else if point == 1 {
		extraStr = "什么倒霉蛋！"
	} else if point == 6 {
		extraStr = "运气爆棚哇！"
	} else {
		extraStr = "运气一般般~"
	}
	cardMessageStr, err := khl.CardMessage{
		&khl.CardMessageCard{
			Theme: "info",
			Size:  "lg",
			Modules: []interface{}{
				khl.CardMessageHeader{
					Text: khl.CardMessageElementText{
						Content: string(emoji.BeerMug) + "一起掷骰子",
						Emoji:   false,
					},
				},
				khl.CardMessageSection{
					Text: khl.CardMessageElementKMarkdown{
						Content: fmt.Sprintf("(met)%s(met) %s你掷出了 **%d**\n%s", authorID, emoji.ClinkingGlasses.String(), point, extraStr),
					},
				},
			},
		},
	}.BuildMessage()

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
