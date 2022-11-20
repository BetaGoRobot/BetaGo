package scheduletask

import (
	"fmt"
	"time"

	betagovar "github.com/BetaGoRobot/BetaGo/betagovar"
	command_context "github.com/BetaGoRobot/BetaGo/commandHandler/context"
	"github.com/BetaGoRobot/BetaGo/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/lonelyevil/kook"
)

// HourlyGetSen 每小时发送
func HourlyGetSen() {
	for {
		time.Sleep(time.Hour)
		commandCtx := &command_context.CommandContext{
			Common: &command_context.CommandCommonContext{
				TargetID: "3241026226723225",
			},
			Extra: &command_context.CommandExtraContext{},
		}
		commandCtx.ContextHandler("hitokoto")
	}
}

// DailyRecommand 每日发送歌曲推荐
func DailyRecommand() {
	for {
		time.Sleep(time.Hour)
		if time.Now().UTC().Format("15") == "00" {
			res, err := neteaseapi.NetEaseGCtx.GetNewRecommendMusic()
			if err != nil {
				fmt.Println("--------------", err.Error())
				return
			}

			modules := make([]interface{}, 0)
			cardMessage := make(kook.CardMessage, 0)
			var cardStr string
			var messageType kook.MessageType
			if len(res) != 0 {
				modules = append(modules, betagovar.CardMessageTextModule{
					Type: "header",
					Text: struct {
						Type    string "json:\"type\""
						Content string "json:\"content\""
					}{"plain-text", "每日8点-音乐推荐~"},
				})
				messageType = 10
				for _, song := range res {
					modules = append(modules, betagovar.CardMessageModule{
						Type:  "audio",
						Title: song.Name + " - " + song.ArtistName,
						Src:   song.SongURL,
						Cover: song.PicURL,
					})
				}
				cardMessage = append(
					cardMessage,
					&kook.CardMessageCard{
						Theme:   kook.CardThemePrimary,
						Size:    kook.CardSizeSm,
						Modules: modules,
					},
				)
				cardStr, err = cardMessage.BuildMessage()
				if err != nil {
					fmt.Println("-------------", err.Error())
					return
				}
			}
			betagovar.GlobalSession.MessageCreate(
				&kook.MessageCreate{
					MessageCreateBase: kook.MessageCreateBase{
						Type:     messageType,
						TargetID: "3241026226723225",
						Content:  cardStr,
					}})
		}
	}
}

type timeCostStru struct {
	UserID   string
	UserName string
	TimeCost time.Duration
}

// DailyRate 每日排行
func DailyRate() {
	for {
		time.Sleep(time.Hour)
		if time.Now().UTC().Format("15") == "00" {
			// 获取24小时的整体在线情况
			var timeCostList = make([]*timeCostStru, 0)
			utility.GetDbConnection().
				Debug().
				Table("betago.channel_logs").
				Select(`
					user_id,
					user_name,
					SUM(
					extract(
						'epoch'
						from
						left_time
					) - extract(
						'epoch'
						from
						joined_time
					)
					) ::bigint * 1000 * 1000 * 1000 as time_cost
				`).
				Where(`is_update = true AND left_time > NOW() - interval '24 hours'`).
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
					}{"plain-text", "语音龙王榜-测试消息"},
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
					return
				}
				betagovar.GlobalSession.MessageCreate(
					&kook.MessageCreate{
						MessageCreateBase: kook.MessageCreateBase{
							Type:     kook.MessageTypeCard,
							TargetID: "3241026226723225",
							Content:  cardMessageStr,
						},
					},
				)
			}
		}
	}

}
