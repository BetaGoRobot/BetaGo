package scheduletask

import (
	"fmt"
	"time"

	betagovar "github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/yiyan"
	"github.com/lonelyevil/khl"
)

// HourlyGetSen 每小时发送
func HourlyGetSen() {
	for {
		time.Sleep(time.Hour)
		betagovar.GlobalSession.MessageCreate(&khl.MessageCreate{
			MessageCreateBase: khl.MessageCreateBase{
				Type:     9,
				TargetID: "3241026226723225",
				Content:  "来自一言的句子: `" + yiyan.GetSen() + "`",
			},
		})
	}
}

// DailyRecommand 每日发送歌曲推荐
func DailyRecommand() {
	for {
		time.Sleep(time.Minute * 5)
		if time.Now().Local().Format("15") == "08" {
			neaseCtx := neteaseapi.NetEaseContext{}
			res, err := neaseCtx.GetNewRecommendMusic()
			if err != nil {
				fmt.Println("--------------", err.Error())
				return
			}

			modules := make([]interface{}, 0)
			cardMessage := make(khl.CardMessage, 0)
			var cardStr string
			var messageType khl.MessageType
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
					&khl.CardMessageCard{
						Theme:   khl.CardThemePrimary,
						Size:    khl.CardSizeSm,
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
				&khl.MessageCreate{
					MessageCreateBase: khl.MessageCreateBase{
						Type:     messageType,
						TargetID: "3241026226723225",
						Content:  cardStr,
					}})
		}
	}
}
