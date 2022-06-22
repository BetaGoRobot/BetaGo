package main

import (
	"fmt"
	"log"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	comcontext "github.com/BetaGoRobot/BetaGo/commandHandler/context"
	errorsender "github.com/BetaGoRobot/BetaGo/commandHandler/error_sender"
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
	var (
		command    = ctx.Extra.Value
		commandCtx = &comcontext.CommandContext{
			Common: &comcontext.CommandCommonContext{
				TargetID: ctx.Extra.TargetID,
				AuthorID: ctx.Extra.UserID,
				MsgID:    "",
			},
			Extra: &comcontext.CommandExtraContext{
				GuildID: ctx.Extra.GuildID,
			},
		}
	)
	switch command {
	case "SHOWADMIN":
		commandCtx.AdminShowHandler()
	case "HELP":
		commandCtx.HelpHandler()
	case "ROLL":
		commandCtx.RollDiceHandler()
	case "ONEWORD":
		commandCtx.GetHitokotoHandler()
	case "PING":
		commandCtx.PingHandler()
	case "SHOWCAL":
		commandCtx.ShowCalHandler()
	default:
		commandCtx.ErrorSenderHandler(fmt.Errorf("非法操作" + emoji.Warning.String()))
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
	command, parameters := utility.GetCommandWithParameters(ctx.Common.Content)
	commandCtx := comcontext.GetNewCommandCtx().Init(ctx.EventHandlerCommonContext).InitExtra(ctx)
	if command != "" {
		switch command {
		case "HELP":
			commandCtx.HelpHandler(parameters...)
		case "ADDADMIN":
			commandCtx.AdminAddHandler(parameters...)
		case "REMOVEADMIN":
			commandCtx.AdminRemoveHandler(parameters...)
		case "SHOWADMIN":
			commandCtx.AdminShowHandler()
		case "ROLL":
			commandCtx.RollDiceHandler(parameters...)
		case "PING":
			commandCtx.PingHandler()
		case "ONEWORD":
			commandCtx.GetHitokotoHandler(parameters...)
		case "SEARCHMUSIC":
			commandCtx.SearchMusicHandler(parameters...)
		case "GETUSER":
			commandCtx.GetUserInfoHandler(parameters...)
		case "SHOWCAL":
			commandCtx.ShowCalHandler(parameters...)
		default:
			commandCtx.ErrorSenderHandler(fmt.Errorf(emoji.QuestionMark.String()+"未知指令 `%s`", command))
		}
	} else {
		// 内容为空，发送help信息
		commandCtx.HelpHandler()
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
