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
	"github.com/BetaGoRobot/BetaGo/commandHandler/admin"
	"github.com/BetaGoRobot/BetaGo/commandHandler/helper"
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
	// 判断是否被at到
	if !utility.IsInSlice(robotID, ctx.Extra.Mention) {
		return
	}
	// 示例中，由于用户发送的命令的RawContent格式为(met)id(met) <command> <parameters>
	// 针对解析的判断逻辑，首先判断是否为空字符串，若为空发送help信息
	// ? 解析出不包含at信息的实际内容
	trueContent := strings.Trim(ctx.Common.Content[strings.LastIndex(ctx.Common.Content, `)`)+1:], " ")
	if trueContent != "" {
		// 内容非空，解析命令
		var (
			command, parameter string
		)
		// 首先执行正则解析
		res := reMatch.FindAllStringSubmatch(trueContent, -1)
		// 判断指令类型
		switch len(res) {
		case 0:
			// 单指令
			command = trueContent
		case 1:
			// 含参指令
			command = res[0][1]
			parameter = res[0][2]
		default:
			// 异常指令
			return
		}
		// 进入指令执行逻辑,首先判断是否为Admin
		if dbpack.CheckIsAdmin(ctx.Common.AuthorID) {
			// 是Admin，判断指令
			var err error
			switch command {
			case "help":
				err = helper.CommandHelperHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID)
			case "addAdmin":
				userID, userName, found := strings.Cut(parameter, " ")
				if found {
					err = admin.AddAdminHandler(userID, userName, ctx.Common.MsgID, ctx.Common.TargetID)
				}
				err = fmt.Errorf("请输入正确的用户ID和用户名格式")
			case "removeAdmin":
				targetUserID := parameter
				err = admin.RemoveAdminHandler(ctx.Common.AuthorID, targetUserID, ctx.Common.MsgID, ctx.Common.TargetID)
			case "showAdmin":
				err = admin.ShowAdminHandler(ctx.Common.TargetID, ctx.Common.MsgID)
			}
			if err != nil {
				sendErrorInfo(err)
			}
		}
	}

}

func sendErrorInfo(err error) {

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
