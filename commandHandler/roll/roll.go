package roll

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/kook"
	"go.opentelemetry.io/otel/attribute"
)

// RandRollHandler 随机抽取一个数字
//
//	@param targetID 目标ID
//	@param quoteID 引用ID
//	@param authorID 发送者ID
//	@return err 错误信息
func RandRollHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.End()

	var min, max int
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
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "info",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: string(emoji.BeerMug) + "一起掷骰子",
						Emoji:   false,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: fmt.Sprintf("范围 `[%d,%d]` (met)%s(met) %s你掷出了 **%d**\n%s", min, max, authorID, emoji.ClinkingGlasses.String(), point, extraStr),
					},
				},
			},
		},
	}.BuildMessage()

	betagovar.GlobalSession.MessageCreate(
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
