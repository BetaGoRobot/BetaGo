package roll

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/khl"
)

// RandRollHandler 随机抽取一个数字
//  @param targetID 目标ID
//  @param quoteID 引用ID
//  @param authorID 发送者ID
//  @return err 错误信息
func RandRollHandler(targetID, quoteID, authorID string, args ...string) (err error) {
	var (
		min, max int
	)
	if len(args) == 0 {
		// 如果没有参数，使用默认range[1,7)
		min = 1
		max = 6

	} else if len(args) == 2 {
		// 如果有参数，使用自定义range
		min, err = strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		max, err = strconv.Atoi(args[1])
		if err != nil {
			return err
		}
		if max <= min {
			return fmt.Errorf("参数错误，max必须大于min")
		}
	} else {
		return fmt.Errorf("参数错误~")
	}
	point := rand.Intn(max-min+1) + min
	var extraStr string
	if point == max {
		extraStr = "最佳运气！" + emoji.BeerMug.String()
	} else if point > (max-min)*2/3+min {
		extraStr = "运气爆棚哇！"
	} else if point > (max-min)/3+min {
		extraStr = "运气不错呀！"
	} else if point > min {
		extraStr = "运气一般般~"
	} else if point == min {
		extraStr = "什么倒霉蛋！"
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
						Content: fmt.Sprintf("范围 `[%d,%d]` (met)%s(met) %s你掷出了 **%d**\n%s", min, max, authorID, emoji.ClinkingGlasses.String(), point, extraStr),
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
