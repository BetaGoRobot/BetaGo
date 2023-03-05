package scheduletask

import (
	"fmt"
	"time"

	betagovar "github.com/BetaGoRobot/BetaGo/betagovar"
	command_context "github.com/BetaGoRobot/BetaGo/commandHandler/context"
	"github.com/BetaGoRobot/BetaGo/commandHandler/dailyrate"
	"github.com/BetaGoRobot/BetaGo/commandHandler/news"
	"github.com/BetaGoRobot/BetaGo/neteaseapi"
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

// DailyRate 每日排行
func DailyRate() {
	for {
		time.Sleep(time.Hour)
		if time.Now().UTC().Format("15") == "00" {
			dailyrate.GetRateHandler("3241026226723225", "", "")
		}
	}
}

func DailyNews() {
	for {
		time.Sleep(time.Hour)
		if time.Now().UTC().Format("15") == "00" {
			news.Handler("3241026226723225", "", "")
		}
	}
}
