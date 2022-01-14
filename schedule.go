package main

import (
	"fmt"
	"time"

	"github.com/lonelyevil/khl"
)

func dailySend() {
	for {
		time.Sleep(time.Hour)
		if string(time.Now().Local().Format("15")) == "08" {
			neaseCtx := NetEaseContext{}
			res, err := neaseCtx.getNewRecommendMusic()
			if err != nil {
				fmt.Println("--------------", err.Error())
				return
			}

			modules := make([]interface{}, 0)
			cardMessage := make(khl.CardMessage, 0)
			var cardStr string
			var messageType khl.MessageType
			if len(res) != 0 {
				modules = append(modules, cardMessageTextModule{
					Type: "header",
					Text: struct {
						Type    string "json:\"type\""
						Content string "json:\"content\""
					}{"plain-text", "每日8点-音乐推荐~"},
				})
				messageType = 10
				for _, song := range res {
					modules = append(modules, cardMessageModule{
						Type:  "audio",
						Title: song.Name + " - " + song.ArtistName,
						Src:   song.SongURL,
						Cover: song.PicURL,
					})
				}
				cardMessage = append(cardMessage, &khl.CardMessageCard{Theme: khl.CardThemePrimary, Size: khl.CardSizeSm, Modules: modules})
				cardStr, err = cardMessage.BuildMessage()
				if err != nil {
					fmt.Println("-------------", err.Error())
					return
				}
			}
			globalSession.MessageCreate(
				&khl.MessageCreate{
					MessageCreateBase: khl.MessageCreateBase{
						Type:     messageType,
						TargetID: "3241026226723225",
						Content:  cardStr,
					}})
		}
	}
}
