package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/commandHandler/admin"
	"github.com/BetaGoRobot/BetaGo/commandHandler/roll"
	"github.com/enescakir/emoji"

	errorsender "github.com/BetaGoRobot/BetaGo/commandHandler/error_sender"
	"github.com/BetaGoRobot/BetaGo/commandHandler/helper"
	"github.com/BetaGoRobot/BetaGo/dbpack"
	"github.com/BetaGoRobot/BetaGo/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/qqmusicapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	goaway "github.com/TwiN/go-away"
	"github.com/lonelyevil/khl"
)

var reMatch = regexp.MustCompile(`^(?P<command>\w+)\s+(?P<parameter>.*)$`)

func commandHandler(ctx *khl.KmarkdownMessageContext) {
	// 判断是否被at到,且消息不是引用/回复
	if !utility.IsInSlice(robotID, ctx.Extra.Mention) {
		return
	}
	// 示例中，由于用户发送的命令的Content格式为(met)id(met) <command> <parameters>
	// 针对解析的判断逻辑，首先判断是否为空字符串，若为空发送help信息
	// ? 解析出不包含at信息的实际内容
	trueContent := strings.Trim(ctx.Common.Content[strings.LastIndex(ctx.Common.Content, `)`)+1:], " ")
	if trueContent != "" {
		// 内容非空，解析命令
		var (
			command, parameter string
			adminNotSolved     bool
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
		var err error
		// 进入指令执行逻辑,首先判断是否为Admin
		if dbpack.CheckIsAdmin(ctx.Common.AuthorID) {
			// 是Admin，判断指令
			switch command {
			case "help":
				err = helper.AdminCommandHelperHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID)
				adminNotSolved = true
			case "addAdmin":
				userID, userName, found := strings.Cut(parameter, " ")
				if found {
					err = admin.AddAdminHandler(userID, userName, ctx.Common.MsgID, ctx.Common.TargetID)
				} else {
					err = fmt.Errorf("请输入正确的用户ID和用户名格式")
				}
				adminNotSolved = true
			case "removeAdmin":
				targetUserID := parameter
				err = admin.RemoveAdminHandler(ctx.Common.AuthorID, targetUserID, ctx.Common.MsgID, ctx.Common.TargetID)
				adminNotSolved = true
			case "showAdmin":
				err = admin.ShowAdminHandler(ctx.Common.TargetID, ctx.Common.MsgID)
				adminNotSolved = true
			default:
				adminNotSolved = false
			}
			if err != nil {
				errorsender.SendErrorInfo(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, err)
			}
		}
		// 非管理员命令
		if !adminNotSolved {
			switch command {
			case "help":
				err = helper.UserCommandHelperHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID)
			case "roll":
				err = roll.RandRollHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID)
			case "ping":
				helper.PingHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID)
			case "oneword":
				err = roll.OneWordHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID)
			default:
				err = fmt.Errorf("未知的指令`%s`, 尝试使用command `help`来获取可用的命令列表", command)
			}
			if err != nil {
				errorsender.SendErrorInfo(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, err)
			}
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

		var (
			modulesNetese = make([]interface{}, 0)
			modulesQQ     = make([]interface{}, 0)
		)
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
				modulesNetese = append(modulesNetese, khl.CardMessageFile{
					Type:  khl.CardMessageFileTypeAudio,
					Src:   song.SongURL,
					Title: song.Name + " - " + song.ArtistName,
					Cover: song.PicURL,
				})
				tempMap[song.Name+" - "+song.ArtistName] = 0
			}
			modulesNetese = append([]interface{}{
				khl.CardMessageHeader{
					Text: khl.CardMessageElementText{
						Content: emoji.Headphone.String() + "网易云音乐-搜索结果" + emoji.MagnifyingGlassTiltedLeft.String(),
						Emoji:   false,
					},
				},
			}, modulesNetese...)
			tempMap = make(map[string]byte)
			// 添加QQ音乐搜索的结果
			for _, song := range resQQmusic {
				if _, ok := tempMap[song.Name+" - "+song.ArtistName]; ok {
					continue
				}
				modulesQQ = append(modulesQQ, khl.CardMessageFile{
					Type:  khl.CardMessageFileTypeAudio,
					Src:   song.SongURL,
					Title: song.Name + " - " + song.ArtistName,
					Cover: song.PicURL,
				})
				tempMap[song.Name+" - "+song.ArtistName] = 0
			}
			modulesQQ = append([]interface{}{
				khl.CardMessageHeader{
					Text: khl.CardMessageElementText{
						Content: emoji.MusicalNote.String() + "QQ音乐-搜索结果" + emoji.MagnifyingGlassTiltedLeft.String(),
						Emoji:   false,
					},
				},
			}, modulesQQ...)

			cardMessage = append(cardMessage,
				&khl.CardMessageCard{
					Theme:   khl.CardThemePrimary,
					Size:    khl.CardSizeSm,
					Modules: modulesNetese,
				},
				&khl.CardMessageCard{
					Theme:   khl.CardThemePrimary,
					Size:    khl.CardSizeSm,
					Modules: modulesQQ,
				},
			)
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

func startUpMessage(session *khl.Session) (err error) {
	currentIP, err := utility.GetOutBoundIP()
	if err != nil {
		return
	}
	cardMessage, err := khl.CardMessage{
		&khl.CardMessageCard{
			Theme: "info",
			Size:  "lg",
			Modules: []interface{}{
				khl.CardMessageHeader{
					Text: khl.CardMessageElementText{
						Content: emoji.DesertIsland.String() + "Online Notifacation" + emoji.Information.String(),
						Emoji:   false,
					},
				},
				khl.CardMessageSection{
					Text: khl.CardMessageElementKMarkdown{
						Content: "Name: \t**" + robotName + "**\n" + "CurrentTime: \t**" + time.Now().Format("2006-01-02 15:04:05") + "**\n" + "IP: \t**" + currentIP + "**",
					},
				},
			},
		},
	}.BuildMessage()
	session.MessageCreate(
		&khl.MessageCreate{
			MessageCreateBase: khl.MessageCreateBase{
				Type:     khl.MessageTypeCard,
				TargetID: testChannelID,
				Content:  cardMessage,
			},
		},
	)
	return
}

func offlineMessage(session *khl.Session) (err error) {
	currentIP, err := utility.GetOutBoundIP()
	if err != nil {
		return
	}
	cardMessage, err := khl.CardMessage{
		&khl.CardMessageCard{
			Theme: "info",
			Size:  "lg",
			Modules: []interface{}{
				khl.CardMessageHeader{
					Text: khl.CardMessageElementText{
						Content: emoji.DesertIsland.String() + "Offline Notifacation" + emoji.Information.String(),
						Emoji:   false,
					},
				},
				khl.CardMessageSection{
					Text: khl.CardMessageElementKMarkdown{
						Content: "Name: \t**" + robotName + "**\n" + "CurrentTime: \t**" + time.Now().Format("2006-01-02 15:04:05") + "**\n" + "IP: \t**" + currentIP + "**",
					},
				},
			},
		},
	}.BuildMessage()
	session.MessageCreate(
		&khl.MessageCreate{
			MessageCreateBase: khl.MessageCreateBase{
				Type:     khl.MessageTypeCard,
				TargetID: testChannelID,
				Content:  cardMessage,
			},
		},
	)
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
