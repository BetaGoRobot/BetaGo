package manager

import (
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	comcontext "github.com/BetaGoRobot/BetaGo/commandHandler/context"
	errorsender "github.com/BetaGoRobot/BetaGo/commandHandler/error_sender"
	"github.com/BetaGoRobot/BetaGo/commandHandler/wordcontrol"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/lonelyevil/kook"
)

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
	newChanLog := &utility.ChannelLogExt{
		UserID:      userInfo.ID,
		UserName:    userInfo.Username,
		ChannelID:   channelInfo.ID,
		ChannelName: channelInfo.Name,
		JoinedTime:  ctx.Extra.JoinedAt.ToTime().Format(betagovar.TimeFormat),
		LeftTime:    "",
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
					Content: "`" + userInfo.Nickname + "`加入了语音频道`" + channelInfo.Name + "`",
				},
			},
		},
	}}.BuildMessage()
	if err != nil {
		errorsender.SendErrorInfo(ctx.Common.TargetID, "", "", err)
		return
	}
	resp, err := betagovar.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: betagovar.NotifierChanID,
				Content:  cardMessageStr,
			},
		},
	)
	if err != nil {
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", "", err)
		return
	}
	newChanLog.MsgID = resp.MsgID
	// 写入数据库记录
	if err = newChanLog.AddJoinedRecord(); err != nil {
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", userInfo.ID, err)
	}
}

func guildUpdateHandler(ctx *kook.GuildUpdateContext) {
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
	newChanLog := &utility.ChannelLogExt{
		UserID:      userInfo.ID,
		UserName:    userInfo.Username,
		ChannelID:   channelInfo.ID,
		ChannelName: channelInfo.Name,
		JoinedTime:  "",
		LeftTime:    ctx.Extra.ExitedAt.ToTime().Format(betagovar.TimeFormat),
	}
	if newChanLog, err = newChanLog.UpdateLeftTime(); err != nil {
		errorsender.SendErrorInfo(betagovar.TestChanID, "", userInfo.ID, err)
		return
	}
	joinTimeT, _ := time.Parse(betagovar.TimeFormat, newChanLog.JoinedTime)
	leftTimeT, _ := time.Parse(betagovar.TimeFormat, newChanLog.LeftTime)
	cardMessageStr, err := kook.CardMessage{&kook.CardMessageCard{
		Theme: kook.CardThemeInfo,
		Size:  kook.CardSizeLg,
		Modules: []interface{}{
			kook.CardMessageSection{
				Text: kook.CardMessageElementKMarkdown{
					Content: strings.Join(
						[]string{
							"`", userInfo.Nickname, "`", "离开了频道`", channelInfo.Name, "`", "\n",
							"在线时间段：`", joinTimeT.Add(time.Hour * 8).Format("2006-01-02-15:04:05"), " - ", leftTimeT.Add(time.Hour * 8).Format("2006-01-02-15:04:05"), "`\n",
							"在线时长：**", leftTimeT.Sub(joinTimeT).String(), "**\n",
						},
						""),
				},
			},
		},
	}}.BuildMessage()
	if err != nil {
		errorsender.SendErrorInfo(ctx.Common.TargetID, "", "", err)
		return
	}
	err = betagovar.GlobalSession.MessageUpdate(
		&kook.MessageUpdate{
			MessageUpdateBase: kook.MessageUpdateBase{
				MsgID:   newChanLog.MsgID,
				Content: cardMessageStr,
			},
		},
	)
	if err != nil {
		errorsender.SendErrorInfo(betagovar.NotifierChanID, "", "", err)
		return
	}
}

func messageEventHandler(ctx *kook.KmarkdownMessageContext) {
	go func() {
		if ctx.Common.Type != kook.MessageTypeKMarkdown || ctx.Extra.Author.Bot {
			return
		}
		defer wordcontrol.RemoveDirtyWords(ctx)
		CommandHandler(ctx)
	}()
}
