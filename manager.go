package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/lonelyevil/khl"
)

func replaceDirtyWords(ctx *khl.TextMessageContext) {
	message := ctx.Common.Content
	if strings.Contains(message, "傻") && strings.Contains(message, "逼") || strings.Contains(message, "傻逼") {
		ctx.Session.MessageCreate(&khl.MessageCreate{
			MessageCreateBase: khl.MessageCreateBase{
				TargetID: ctx.Common.TargetID,
				Content:  fmt.Sprintf("%s 使用了侮辱词汇，消息已被移除，不可以向他学习哦", ctx.Extra.Author.Nickname),
				Quote:    ctx.Common.MsgID,
			},
		})
		ctx.Session.MessageDelete(ctx.Common.MsgID)
	}

}

var (
	reg = regexp.MustCompile(`(?i)(\.search)\ (.*)`)
)

func searchMusicByRobot(ctx *khl.TextMessageContext) {
	message := ctx.Common.Content
	if res := reg.FindStringSubmatch(message); res != nil && len(res) > 2 {
		neaseCtx := NetEaseContext{}
		res, err := neaseCtx.searchMusicByKeyWord(strings.Split(res[2], " "))
		if err != nil {
			log.Println("--------------", err.Error())
			return
		}

		modules := make([]interface{}, 0)
		cardMessage := make(khl.CardMessage, 0)
		var cardStr string
		var messageType khl.MessageType
		if len(res) != 0 {
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
				log.Println("-------------", err.Error())
				return
			}
		} else {
			messageType = 9
			cardStr = "--------\n> (ins)没有找到你要搜索的歌曲哦，换一个关键词试试~(ins)\n\n--------------"
		}
		ctx.Session.MessageCreate(
			&khl.MessageCreate{
				MessageCreateBase: khl.MessageCreateBase{
					Type:     messageType,
					TargetID: ctx.Common.TargetID,
					Content:  cardStr,
					Quote:    ctx.Common.MsgID,
				}})
	}
	return
}

// func scheduleEvent(ctx *khl.TextMessageContext) {
// 	if string(time.Now().Local().Format("15")) == "05" {
// 		neaseCtx := NetEaseContext{}
// 		res, err := neaseCtx.getNewRecommendMusic()
// 		if err != nil {
// 			log.Println("--------------", err.Error())
// 			return
// 		}

// 		modules := make([]interface{}, 0)
// 		cardMessage := make(khl.CardMessage, 0)
// 		var cardStr string
// 		var messageType khl.MessageType
// 		if len(res) != 0 {
// 			modules = append(modules, cardMessageTextModule{
// 				Type: "header",
// 				Text: struct {
// 					Type    string "json:\"type\""
// 					Content string "json:\"content\""
// 				}{"plain-text", "每日8点-音乐推荐~"},
// 			})
// 			messageType = 10
// 			for _, song := range res {
// 				modules = append(modules, cardMessageModule{
// 					Type:  "audio",
// 					Title: song.Name + " - " + song.ArtistName,
// 					Src:   song.SongURL,
// 					Cover: song.PicURL,
// 				})
// 			}
// 			cardMessage = append(cardMessage, &khl.CardMessageCard{Theme: khl.CardThemePrimary, Size: khl.CardSizeSm, Modules: modules})
// 			cardStr, err = cardMessage.BuildMessage()
// 			if err != nil {
// 				log.Println("-------------", err.Error())
// 				return
// 			}
// 		} else {
// 			messageType = 9
// 			cardStr = "--------\n> (ins)没有找到你要搜索的歌曲哦，换一个关键词试试~(ins)\n\n--------------"
// 		}
// 		ctx.Session.MessageCreate(
// 			&khl.MessageCreate{
// 				MessageCreateBase: khl.MessageCreateBase{
// 					Type:     messageType,
// 					TargetID: ctx.Common.TargetID,
// 					Content:  cardStr,
// 					Quote:    ctx.Common.MsgID,
// 				}})
// 	}
// }

// 机器人被at时返回消息
func replyToMention(ctx *khl.TextMessageContext) {
	if isInSlice(robotID, ctx.Extra.Mention) {
		NowTime := time.Now().Unix()
		if NowTime-LastMentionedTime.Unix() > 1 {
			LastMentionedTime = time.Now()
			ctx.Session.MessageCreate(&khl.MessageCreate{
				MessageCreateBase: khl.MessageCreateBase{
					TargetID: ctx.Common.TargetID,
					Content:  "@我干什么？没事干了吗! (此消息仅你可见)",
					Quote:    ctx.Common.MsgID,
				},
				TempTargetID: ctx.Common.AuthorID,
			})
		}
	}
}

func startUpMessage(session *khl.Session) (err error) {
	currentIP, err := GetOutBoundIP()
	if err != nil {
		return
	}
	session.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     9,
			TargetID: testChannelID,
			Content:  fmt.Sprintf("---------\n> Robot `%s` is \n`online`\n IP:\t%s\n Time:\t%s\n---------", robotName, currentIP, GetCurrentTime()),
		}})
	return
}

func offlineMessage(session *khl.Session) (err error) {
	currentIP, err := GetOutBoundIP()
	if err != nil {
		return
	}
	session.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     9,
			TargetID: testChannelID,
			Content:  fmt.Sprintf("---------\n> Robot `%s` is \n`offline`\n IP:\t%s\n Time:\t%s\n---------", robotName, currentIP, GetCurrentTime()),
		}})
	return
}

func sendMessageToTestChannel(session *khl.Session, content string) {

	session.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     9,
			TargetID: testChannelID,
			Content:  content,
		}})
}

func receiveDirectMessage(ctx *khl.DirectMessageReactionAddContext) {
	log.Println("-----------Test")
}
