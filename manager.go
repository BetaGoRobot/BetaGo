package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/commandHandler/admin"
	"github.com/BetaGoRobot/BetaGo/commandHandler/cal"
	command_context "github.com/BetaGoRobot/BetaGo/commandHandler/context"
	errorsender "github.com/BetaGoRobot/BetaGo/commandHandler/error_sender"
	"github.com/BetaGoRobot/BetaGo/commandHandler/helper"
	"github.com/BetaGoRobot/BetaGo/commandHandler/roll"
	"github.com/BetaGoRobot/BetaGo/dbpack"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/enescakir/emoji"
	jsoniter "github.com/json-iterator/go"
	"github.com/lonelyevil/khl"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func clickEventAsyncHandler(ctx *khl.MessageButtonClickContext) {
	go clickEventHandler(ctx)
}

func clickEventHandler(ctx *khl.MessageButtonClickContext) {
	var err error
	clickValue := ctx.Extra.Value
	isAdmin := dbpack.CheckIsAdmin(ctx.Extra.UserID)
	switch clickValue {
	case "SHOWADMIN":
		if isAdmin {
			err = admin.ShowAdminHandler(ctx.Extra.TargetID, "")
		}
	case "HELP":
		if isAdmin {
			err = helper.AdminCommandHelperHandler(ctx.Extra.TargetID, "", ctx.Extra.UserID)
		} else {
			err = helper.UserCommandHelperHandler(ctx.Extra.TargetID, "", ctx.Extra.UserID)
		}
	case "ROLL":
		err = roll.RandRollHandler(ctx.Extra.TargetID, "", ctx.Extra.UserID)
	case "ONEWORD":
		err = roll.OneWordHandler(ctx.Extra.TargetID, "", ctx.Extra.UserID)
	case "PING":
		helper.PingHandler(ctx.Extra.TargetID, "", ctx.Extra.UserID)
	case "SHOWCAL":
		err = cal.ShowCalHandler(ctx.Extra.TargetID, "", ctx.Extra.UserID, ctx.Extra.GuildID)
	default:
		err = fmt.Errorf("非法操作" + emoji.Warning.String())
	}
	if err != nil {
		errorsender.SendErrorInfo(ctx.Extra.TargetID, "", ctx.Extra.UserID, err)
	}
}

func commandHandler(ctx *khl.KmarkdownMessageContext) {
	// 判断是否被at到,且消息不是引用/回复
	if !utility.IsInSlice(betagovar.RobotID, ctx.Extra.Mention) {
		return
	}
	// 示例中，由于用户发送的命令的Content格式为(met)id(met) <command> <parameters>
	// 针对解析的判断逻辑，首先判断是否为空字符串，若为空发送help信息
	// ? 解析出不包含at信息的实际内容
	trueContent := strings.TrimSpace(strings.Replace(ctx.Common.Content, "(met)"+betagovar.RobotID+"(met)", "", 1))
	if trueContent != "" {
		// 内容非空，解析命令
		var (
			command    string
			parameters []string
			slice      = strings.Split(strings.Trim(trueContent, " "), " ")
			err        error
			commandCtx *command_context.CommandContext
		)
		// 判断指令类型
		if len(slice) == 1 {
			command = slice[0]
		} else {
			command = slice[0]
			parameters = slice[1:]
		}
		command = strings.ToUpper(command)

		commandCtx = commandCtx.Init(ctx.EventHandlerCommonContext)
		commandCtx.InitExtra(ctx)
		switch command {
		case "HELP":
			err = commandCtx.HelpHandler(parameters...)
		case "ADDADMIN":
			err = commandCtx.AdminAddHandler(parameters...)
		case "REMOVEADMIN":
			err = commandCtx.AdminRemoveHandler(parameters...)
		case "SHOWADMIN":
			err = commandCtx.AdminShowHandler()
		case "ROLL":
			err = commandCtx.RollDiceHandler(parameters...)
		case "PING":
			commandCtx.PingHandler()
		case "ONEWORD":
			err = commandCtx.OneWordHandler(parameters...)
		case "SEARCHMUSIC":
			err = commandCtx.SearchMusicHandler(parameters...)
		case "GETUSER":
			err = commandCtx.GetUserInfoHandler(parameters...)
		case "SHOWCAL":
			err = commandCtx.ShowCalHandler(parameters...)
		default:
			err = fmt.Errorf(emoji.Warning.String()+"未知指令 `%s`", command)
		}
		if err != nil {
			errorsender.SendErrorInfo(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, err)
		}

	}
}

func channelJoinedAsyncHandler(ctx *khl.GuildChannelMemberAddContext) {
	go channelJoinedHandler(ctx)
}

func channelJoinedHandler(ctx *khl.GuildChannelMemberAddContext) {
	userInfo, err := utility.GetUserInfo(ctx.Extra.UserID, ctx.Common.TargetID)
	if err != nil {
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", userInfo.ID, err)
		return
	}
	channelInfo, err := utility.GetChannnelInfo(ctx.Extra.ChannelID)
	if err != nil {
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", userInfo.ID, err)
		return
	}
	// !频道日志记录
	newChanLog := &dbpack.ChannelLog{
		UserID:      userInfo.ID,
		UserName:    userInfo.Username,
		ChannelID:   channelInfo.ID,
		ChannelName: channelInfo.Name,
		JoinedTime:  ctx.Extra.JoinedAt.ToTime(),
		LeftTime:    time.Time{},
	}
	if err = newChanLog.AddJoinedRecord(); err != nil {
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", userInfo.ID, err)
	}

	cardMessageStr, err := khl.CardMessage{&khl.CardMessageCard{
		Theme: khl.CardThemeInfo,
		Size:  khl.CardSizeLg,
		Modules: []interface{}{
			khl.CardMessageSection{
				Text: khl.CardMessageElementKMarkdown{
					Content: "`" + userInfo.Nickname + "`悄悄加入了语音频道`" + channelInfo.Name + "`",
				},
			},
		},
	}}.BuildMessage()
	if err != nil {
		errorsender.SendErrorInfo(ctx.Common.TargetID, "", "", err)
		return
	}
	_, err = betagovar.GlobalSession.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeCard,
			TargetID: betagovar.NotifierChanID,
			Content:  cardMessageStr,
		},
	})
	if err != nil {
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", "", err)
	}
}

func channelLeftAsyncHandler(ctx *khl.GuildChannelMemberDeleteContext) {
	go channelLeftHandler(ctx)
}
func channelLeftHandler(ctx *khl.GuildChannelMemberDeleteContext) {
	// 离开频道时，记录频道信息
	userInfo, err := utility.GetUserInfo(ctx.Extra.UserID, ctx.Common.TargetID)
	if err != nil {
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", userInfo.ID, err)
		return
	}
	channelInfo, err := utility.GetChannnelInfo(ctx.Extra.ChannelID)
	if err != nil {
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", userInfo.ID, err)
		return
	}

	// !频道日志记录
	newChanLog := &dbpack.ChannelLog{
		UserID:      userInfo.ID,
		UserName:    userInfo.Username,
		ChannelID:   channelInfo.ID,
		ChannelName: channelInfo.Name,
		JoinedTime:  time.Time{},
		LeftTime:    ctx.Extra.ExitedAt.ToTime(),
	}
	if err = newChanLog.UpdateLeftTime(); err != nil {
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", userInfo.ID, err)
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
