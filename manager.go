package main

import (
	"log"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	comcontext "github.com/BetaGoRobot/BetaGo/commandHandler/context"
	errorsender "github.com/BetaGoRobot/BetaGo/commandHandler/error_sender"
	"github.com/BetaGoRobot/BetaGo/utility"
	jsoniter "github.com/json-iterator/go"
	"github.com/lonelyevil/kook"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func clickEventAsyncHandler(ctx *kook.MessageButtonClickContext) {
	go clickEventHandler(ctx)
}

func clickEventHandler(ctx *kook.MessageButtonClickContext) {
	if err := betagovar.FlowControl.Top(); err != nil {
		errorsender.SendErrorInfo(ctx.Extra.TargetID, "", "", err)
		return
	}
	betagovar.FlowControl.Add()
	defer betagovar.FlowControl.Sub()
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
	commandCtx.ContextHandler(command)
	time.Sleep(time.Second)
}

func commandHandler(ctx *kook.KmarkdownMessageContext) {
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
		commandCtx.ContextHandler(command, parameters...)
	} else {
		// 内容为空，发送help信息
		commandCtx.ContextHandler(comcontext.CommandContextTypeHelper)
	}
}

func channelJoinedAsyncHandler(ctx *kook.GuildChannelMemberAddContext) {
	go channelJoinedHandler(ctx)
}

func channelJoinedHandler(ctx *kook.GuildChannelMemberAddContext) {
	defer utility.CollectPanic(ctx.Common, ctx.Common.TargetID, ctx.Common.MsgID, "")
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
	newChanLog := &utility.ChannelLog{
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
	if strings.Contains(channelInfo.Name, "躲避女人") {
		return
	}
	cardMessageStr, err := kook.CardMessage{&kook.CardMessageCard{
		Theme: kook.CardThemeInfo,
		Size:  kook.CardSizeLg,
		Modules: []interface{}{
			kook.CardMessageSection{
				Text: kook.CardMessageElementKMarkdown{
					Content: "`" + userInfo.Nickname + "`悄悄加入了语音频道`" + channelInfo.Name + "`",
				},
			},
		},
	}}.BuildMessage()
	if err != nil {
		errorsender.SendErrorInfo(ctx.Common.TargetID, "", "", err)
		return
	}
	_, err = betagovar.GlobalSession.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     kook.MessageTypeCard,
			TargetID: betagovar.NotifierChanID,
			Content:  cardMessageStr,
		},
	})
	if err != nil {
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", "", err)
	}
}

func channelLeftAsyncHandler(ctx *kook.GuildChannelMemberDeleteContext) {
	go channelLeftHandler(ctx)
}
func channelLeftHandler(ctx *kook.GuildChannelMemberDeleteContext) {
	defer utility.CollectPanic(ctx.Extra, ctx.Common.TargetID, "", ctx.Extra.UserID)
	// 离开频道时，记录频道信息
	userInfo, err := utility.GetUserInfo(ctx.Extra.UserID, ctx.Common.TargetID)
	if err != nil {
		errorsender.SendErrorInfo(betagovar.TestChanID, "", userInfo.ID, err)
		return
	}
	channelInfo, err := utility.GetChannnelInfo(ctx.Extra.ChannelID)
	if err != nil {
		errorsender.SendErrorInfo(betagovar.TestChanID, "", userInfo.ID, err)
		return
	}

	// !频道日志记录
	newChanLog := &utility.ChannelLog{
		UserID:      userInfo.ID,
		UserName:    userInfo.Username,
		ChannelID:   channelInfo.ID,
		ChannelName: channelInfo.Name,
		JoinedTime:  time.Time{},
		LeftTime:    ctx.Extra.ExitedAt.ToTime(),
	}
	if err = newChanLog.UpdateLeftTime(); err != nil {
		errorsender.SendErrorInfo(betagovar.TestChanID, "", userInfo.ID, err)
	}
}

func sendMessageToTestChannel(session *kook.Session, content string) {

	session.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     9,
			TargetID: betagovar.TestChanID,
			Content:  content,
		}})
}

func receiveDirectMessage(ctx *kook.DirectMessageReactionAddContext) {
	log.Println("-----------Test")
}
