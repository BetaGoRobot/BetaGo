// Package scheduletask 存放定时任务
package scheduletask

import (
	"context"
	"fmt"
	"sync"
	"time"

	betagovar "github.com/BetaGoRobot/BetaGo/betagovar"
	command_context "github.com/BetaGoRobot/BetaGo/commandHandler/context"
	"github.com/BetaGoRobot/BetaGo/commandHandler/dailyrate"
	"github.com/BetaGoRobot/BetaGo/commandHandler/news"
	"github.com/BetaGoRobot/BetaGo/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/gotify"
	"github.com/lonelyevil/kook"
	"github.com/patrickmn/go-cache"
)

var once = &sync.Once{}

// var SelfCheckCache = cache.New(time.Minute*30, time.Minute)

// DailyTaskCache  每日任务缓存
var DailyTaskCache = cache.New(time.Hour*3, time.Minute)

// DailyGetSen 每小时发送
func DailyGetSen() {
	for {
		time.Sleep(time.Second * 30)
		if time.Now().UTC().Format("15:04") == "00:00" {
			commandCtx := &command_context.CommandContext{
				Common: &command_context.CommandCommonContext{
					TargetID: "3241026226723225",
				},
				Extra: &command_context.CommandExtraContext{},
			}
			commandCtx.ContextHandler("hitokoto")
		}
	}
}

// DailyRecommand 每日发送歌曲推荐
func DailyRecommand() {
	for {
		time.Sleep(time.Second * 30)
		if time.Now().UTC().Format("15:04") == "00:00" {
			if getOrSetCache("DailyRecommand") {
				continue
			}
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
				messageType = kook.MessageTypeCard
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
					},
				})
		}
	}
}

// DailyRate 每日排行
func DailyRate() {
	for {
		time.Sleep(time.Second * 30)
		if time.Now().UTC().Format("15:04") == "00:00" {
			if getOrSetCache("DailyRate") {
				continue
			}
			dailyrate.GetRateHandler(context.Background(), "3241026226723225", "", "")
		}
	}
}

func getOrSetCache(key string) bool {
	if _, ok := DailyTaskCache.Get(key); ok {
		return true
	}
	DailyTaskCache.Set(key, 1, time.Hour*3)
	return false
}

// DailyNews 每日新闻
func DailyNews() {
	for {
		time.Sleep(time.Minute)
		if time.Now().UTC().Format("15") == "00" {
			if getOrSetCache("DailyNews") {
				continue
			}
			news.Handler(context.Background(), "3241026226723225", "", "")
		}
	}
}

// OnlineTest 在线测试
func OnlineTest() {
	for {
		once.Do(func() {
			time.Sleep(time.Second * 5)
		})
		selfCheckInner()
		time.Sleep(time.Second * 30)
	}
}

func selfCheckInner() {
	resp, err := betagovar.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeText,
				TargetID: betagovar.TestChanID,
				Content:  betagovar.SelfCheckMessage,
			},
		})
	if err != nil {
		fmt.Println("Cannot send message, killing...", err.Error())
		return
	}
	time.Sleep(time.Millisecond * 100)
	err = betagovar.GlobalSession.MessageDelete(resp.MsgID)
	if err != nil {
		fmt.Println("Cannot delete sent message, return...")
		return
	}
	time.Sleep(time.Second * 1)
	select {
	case <-betagovar.SelfCheckChan:
		utility.ZapLogger.Info("Self check successful")
	default:
		gotify.SendMessage("", "Self check failed, reconnecting...", 7)
		betagovar.ReconnectChan <- "reconnect"
	}
}
