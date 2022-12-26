package dailyrate

import (
	"fmt"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/kook"
)

type timeCostStru struct {
	UserID   string
	UserName string
	TimeCost time.Duration
}

func GetRateHandler(targetID, msgID, authorID string, args ...string) (err error) {
	// 获取24小时的整体在线情况
	var timeCostList = make([]*timeCostStru, 0)
	utility.GetDbConnection().
		Debug().
		Table("betago.channel_log_exts").
		Select(`
			(
				SUM(
					EXTRACT(
						epoch
						FROM
						to_timestamp(left_time, 'YYYY\-MM\-DD\THH24\:MI\.MS') - to_timestamp(joined_time, 'YYYY-MM-DD\THH24\:MI\.MS')
					)
				)* 1000 * 1000 * 1000
			):: bigint as time_cost
		`).
		Where(`is_update = true AND to_timestamp(left_time,'YYYY-MM-DDTHH24:MI:SS.MS') > NOW() - interval '24 hours'`).
		Group("user_id, user_name").
		Order("time_cost DESC").
		Limit(3).
		Find(&timeCostList)
	if len(timeCostList) != 0 {
		modules := make([]interface{}, 0)
		modules = append(modules, betagovar.CardMessageTextModule{
			Type: "header",
			Text: struct {
				Type    string "json:\"type\""
				Content string "json:\"content\""
			}{"plain-text", "勤恳在线排行榜" + emoji.DesktopComputer.String()},
		})
		modules = append(modules,
			kook.CardMessageSection{
				Text: kook.CardMessageParagraph{
					Cols: 3,
					Fields: []interface{}{
						kook.CardMessageElementKMarkdown{
							Content: "**用户ID**",
						},
						kook.CardMessageElementKMarkdown{
							Content: "**用户名**",
						},
						kook.CardMessageElementKMarkdown{
							Content: "**语音时长**",
						},
					},
				},
			},
		)
		for _, user := range timeCostList {
			modules = append(modules,
				kook.CardMessageSection{
					Text: kook.CardMessageParagraph{
						Cols: 3,
						Fields: []interface{}{
							kook.CardMessageElementKMarkdown{
								Content: user.UserID,
							},
							kook.CardMessageElementKMarkdown{
								Content: fmt.Sprintf("`%s`", user.UserName),
							},
							kook.CardMessageElementKMarkdown{
								Content: user.TimeCost.String(),
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
			return err
		}
		betagovar.GlobalSession.MessageCreate(
			&kook.MessageCreate{
				MessageCreateBase: kook.MessageCreateBase{
					Type:     kook.MessageTypeCard,
					TargetID: betagovar.NotifierChanID,
					Content:  cardMessageStr,
					Quote:    msgID,
				},
			},
		)
	}
	return
}
