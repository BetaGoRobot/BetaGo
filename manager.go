package main

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/dbpack"
	"github.com/BetaGoRobot/BetaGo/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/qqmusicapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	goaway "github.com/TwiN/go-away"
	"github.com/lonelyevil/khl"
)

var reMatch = regexp.MustCompile(`^(?P<command>\w+)\s+(?P<parameter>.*)$`)

func debugInterfaceHandler(ctx *khl.KmarkdownMessageContext) {
	message := ctx.Common.Content

	if strings.Contains(message, "debug") && ctx.Common.AuthorID == "938697103" {
		command := strings.Index(message, "debug")
		switch command {

		}
	}
}

func adminCommand(ctx *khl.KmarkdownMessageContext) {
	if !utility.IsInSlice(robotID, ctx.Extra.Mention) {
		return
	}
	var (
		command, parameter string
	)
	content := strings.Trim(ctx.Common.Content[strings.LastIndex(ctx.Common.Content, `)`)+1:], " ")
	reRes := reMatch.FindAllStringSubmatch(content, -1)
	if len(reRes) == 0 {
		command = content
	} else if len(reRes[0]) >= 3 {
		command = reRes[0][1]
		parameter = reRes[0][2]
	} else if content == "" {
		return
	} else {
		return
	}
	var successFlag bool
	if dbpack.CheckIsAdmin(ctx.Common.AuthorID) {
		// 确认为管理员再执行
		var commandResStr string
		switch command {
		case "addAdmin":
			successFlag = true
			userID, userName, _ := strings.Cut(parameter, " ")
			userIDInt, _ := strconv.Atoi(userID)
			// 先检验是否存在
			if dbpack.GetDbConnection().Table("betago.administrators").Where("user_id = ?", userIDInt).Find(&dbpack.Administrator{}).RowsAffected != 0 {
				// 存在则不处理，返回信息
				ctx.Session.MessageCreate(&khl.MessageCreate{
					MessageCreateBase: khl.MessageCreateBase{
						Type:     9,
						TargetID: ctx.Common.TargetID,
						Content:  fmt.Sprintf("%s 已经为管理员, 无需处理~", userName),
						Quote:    ctx.Common.MsgID,
					},
				})
			}

			dbRes := dbpack.GetDbConnection().Table("betago.administrators").Create(&dbpack.Administrator{
				UserID:   int64(userIDInt),
				UserName: userName,
				Level:    1,
			})
			if dbRes.Error != nil {
				return
			}
			commandResStr = fmt.Sprintf("%s 已被设置为管理员, 让我们祝贺这个B~ (met)%s(met)", userName, userID)
		case "showAdmin":
			successFlag = true
			admins := make([]*dbpack.Administrator, 0)
			dbpack.GetDbConnection().Table("betago.administrators").Find(&admins)
			for _, admin := range admins {
				commandResStr += fmt.Sprintf("%s\n", admin.UserName)
			}
		default:
			successFlag = false
			ctx.Session.MessageCreate(&khl.MessageCreate{
				MessageCreateBase: khl.MessageCreateBase{
					Type:     9,
					TargetID: ctx.Common.TargetID,
					Content:  "该指令暂不支持",
					Quote:    ctx.Common.MsgID,
				},
			})
		}
		if successFlag {
			// 在频道中发送成功消息
			ctx.Session.MessageCreate(&khl.MessageCreate{
				MessageCreateBase: khl.MessageCreateBase{
					Type:     khl.MessageTypeKMarkdown,
					TargetID: ctx.Common.TargetID,
					Content:  `指令执行成功---->` + "\n" + commandResStr,
					Quote:    ctx.Common.MsgID,
				},
			})
		}
	}
}

func removeDirtyWords(ctx *khl.KmarkdownMessageContext) {
	message := ctx.Common.Content
	if strings.Contains(message, "傻") && strings.Contains(message, "逼") || strings.Contains(message, "傻逼") || goaway.IsProfane(message) {
		ctx.Session.MessageCreate(&khl.MessageCreate{
			MessageCreateBase: khl.MessageCreateBase{
				TargetID: ctx.Common.TargetID,
				Content:  fmt.Sprintf("%s 使用了侮辱发言%s, 消息已被移除, 不可以向他学习哦", ctx.Extra.Author.Nickname, goaway.Censor(message)),
				Quote:    ctx.Common.MsgID,
			},
		})
		ctx.Session.MessageDelete(ctx.Common.MsgID)
	}

}

var (
	reg = regexp.MustCompile(`(?i)(\.search)\ (.*)`)
)

func searchMusicByRobot(ctx *khl.KmarkdownMessageContext) {
	// ctx.Session.AssetCreate()
	message := ctx.Common.Content
	if res := reg.FindStringSubmatch(message); res != nil && len(res) > 2 {
		// 使用网易云搜索
		neaseCtx := neteaseapi.NetEaseContext{}
		resNetease, err := neaseCtx.SearchMusicByKeyWord(strings.Split(res[2], " "))
		if err != nil {
			log.Println("--------------", err.Error())
			return
		}

		// 使用QQ音乐搜索
		qqmusicCtx := qqmusicapi.QQmusicContext{}
		resQQmusic, err := qqmusicCtx.SearchMusic(strings.Split(res[2], " "))
		if err != nil {
			log.Println("--------------", err.Error())
			return
		}

		modules := make([]interface{}, 0)
		cardMessage := make(khl.CardMessage, 0)
		var cardStr string
		var messageType khl.MessageType
		if len(resNetease) != 0 || len(resQQmusic) != 0 {
			tempMap := make(map[string]byte, 0)
			messageType = 10
			// 添加网易云搜索的结果
			for _, song := range resNetease {
				if _, ok := tempMap[song.Name+" - "+song.ArtistName]; ok {
					continue
				}
				modules = append(modules, betagovar.CardMessageModule{
					Type:  "audio",
					Title: "网易 - " + song.Name + " - " + song.ArtistName,
					Src:   song.SongURL,
					Cover: song.PicURL,
				})
				tempMap[song.Name+" - "+song.ArtistName] = 0
			}

			// 添加QQ音乐搜索的结果
			for _, song := range resQQmusic {
				if _, ok := tempMap[song.Name+" - "+song.ArtistName]; ok {
					continue
				}
				modules = append(modules, betagovar.CardMessageModule{
					Type:  "audio",
					Title: "QQ - " + song.Name + " - " + song.ArtistName,
					Src:   song.SongURL,
					Cover: song.PicURL,
				})
				tempMap[song.Name+" - "+song.ArtistName] = 0
			}

			cardMessage = append(cardMessage, &khl.CardMessageCard{Theme: khl.CardThemePrimary, Size: khl.CardSizeSm, Modules: modules})
			cardStr, err = cardMessage.BuildMessage()
			if err != nil {
				log.Println("-------------", err.Error())
				return
			}
		} else {
			messageType = 9
			cardStr = "--------\n> (ins)没有找到你要搜索的歌曲哦, 换一个关键词试试~(ins)\n\n--------------"
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

// 机器人被at时返回消息
func replyToMention(ctx *khl.KmarkdownMessageContext) {
	if utility.IsInSlice(robotID, ctx.Extra.Mention) {
		//! 被At到
		content := ctx.Common.Content
		NowTime := time.Now().Unix()
		if NowTime-LastMentionedTime.Unix() > 1 {
			LastMentionedTime = time.Now()
			if strings.Contains(content, "roll") {
				point := rand.Intn(6) + 1
				msg := &khl.MessageCreate{
					MessageCreateBase: khl.MessageCreateBase{
						TargetID: ctx.Common.TargetID,
						Content:  "你的点数是" + strconv.Itoa(point),
						Quote:    ctx.Common.MsgID,
					},
					// TempTargetID: ctx.Common.AuthorID,
				}
				if point > 3 {
					msg.Content += "，运气不错呀！"
				} else if point == 1 {
					msg.Content += "，什么倒霉孩子"
				} else if point == 6 {
					msg.Content += "，运气爆棚哇！"
				} else {
					msg.Content += "，运气一般般啦~"
				}
				ctx.Session.MessageCreate(msg)
			} else {
				ctx.Session.MessageCreate(
					&khl.MessageCreate{
						MessageCreateBase: khl.MessageCreateBase{
							TargetID: ctx.Common.TargetID,
							Content:  "@我干什么? 没事干了吗! (此消息仅你可见)",
							Quote:    ctx.Common.MsgID,
						},
						TempTargetID: ctx.Common.AuthorID,
					})
			}
		}

	}
}

func startUpMessage(session *khl.Session) (err error) {
	currentIP, err := utility.GetOutBoundIP()
	if err != nil {
		return
	}
	session.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     9,
			TargetID: testChannelID,
			Content:  fmt.Sprintf("---------\n> Robot `%s` is \n`online`\n IP:\t%s\n Time:\t%s\n---------", robotName, currentIP, utility.GetCurrentTime()),
		}})
	return
}

func offlineMessage(session *khl.Session) (err error) {
	currentIP, err := utility.GetOutBoundIP()
	if err != nil {
		return
	}
	session.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     9,
			TargetID: testChannelID,
			Content:  fmt.Sprintf("---------\n> Robot `%s` is \n`offline`\n IP:\t%s\n Time:\t%s\n---------", robotName, currentIP, utility.GetCurrentTime()),
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

func replayToDirectMessage(ctx *khl.KmarkdownMessageContext) {
	log.Println("-----------Test")
}
