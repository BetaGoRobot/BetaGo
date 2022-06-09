package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/commandHandler/admin"
	errorsender "github.com/BetaGoRobot/BetaGo/commandHandler/error_sender"
	"github.com/BetaGoRobot/BetaGo/commandHandler/helper"
	"github.com/BetaGoRobot/BetaGo/commandHandler/music"
	"github.com/BetaGoRobot/BetaGo/commandHandler/roll"
	"github.com/BetaGoRobot/BetaGo/dbpack"
	"github.com/BetaGoRobot/BetaGo/utility"
	jsoniter "github.com/json-iterator/go"
	"github.com/lonelyevil/khl"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func commandHandler(ctx *khl.KmarkdownMessageContext) {
	// 判断是否被at到,且消息不是引用/回复
	if !utility.IsInSlice(betagovar.RobotID, ctx.Extra.Mention) {
		return
	}
	// 示例中，由于用户发送的命令的Content格式为(met)id(met) <command> <parameters>
	// 针对解析的判断逻辑，首先判断是否为空字符串，若为空发送help信息
	// ? 解析出不包含at信息的实际内容
	trueContent := strings.Trim(ctx.Common.Content[strings.LastIndex(ctx.Common.Content, `)`)+1:], " ")
	if trueContent != "" {
		// 内容非空，解析命令
		var (
			command        string
			parameters     []string
			adminNotSolved bool
		)
		// 首先执行正则解析
		slice := strings.Split(strings.Trim(trueContent, " "), " ")
		// 判断指令类型
		if len(slice) == 1 {
			command = slice[0]
		} else {
			command = slice[0]
			parameters = slice[1:]
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
				if len(parameters) == 2 {
					userID, userName := parameters[0], parameters[1]
					err = admin.AddAdminHandler(userID, userName, ctx.Common.MsgID, ctx.Common.TargetID)
				} else {
					err = fmt.Errorf("请输入正确的用户ID和用户名格式")
				}
				adminNotSolved = true
			case "removeAdmin":
				if len(parameters) == 1 {
					targetUserID := parameters[0]
					err = admin.RemoveAdminHandler(ctx.Common.AuthorID, targetUserID, ctx.Common.MsgID, ctx.Common.TargetID)
				} else {
					err = fmt.Errorf("请输入正确的用户ID和用户名格式")
				}
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
			case "searchMusic":
				err = music.SearchMusicByRobot(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters)
			case "getuser":
				err = helper.GetUserInfoHandler(parameters[0], ctx.Extra.GuildID, ctx.Common.TargetID, ctx.Common.MsgID)
			default:
				err = fmt.Errorf("未知的指令`%s`, 尝试使用command `help`来获取可用的命令列表", command)
			}
			if err != nil {
				errorsender.SendErrorInfo(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, err)
			}
		}
	}

}

func sendMessageToTestChannel(session *khl.Session, content string) {

	session.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     9,
			TargetID: betagovar.TestChanID,
			Content:  content,
		}})
}

func receiveDirectMessage(ctx *khl.DirectMessageReactionAddContext) {
	log.Println("-----------Test")
}
