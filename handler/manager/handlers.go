package manager

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	comcontext "github.com/BetaGoRobot/BetaGo/handler/commandHandler/context"
	errorsender "github.com/BetaGoRobot/BetaGo/handler/commandHandler/error_sender"
	"github.com/BetaGoRobot/BetaGo/handler/commandHandler/gpt3"
	"github.com/BetaGoRobot/BetaGo/handler/commandHandler/wordcontrol"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/lonelyevil/kook"
	"github.com/spyzhov/ajson"
	"go.opentelemetry.io/otel/attribute"
)

func clickEventHandler(baseCtx context.Context, ctx *kook.MessageButtonClickContext) {
	baseCtx, span := otel.BetaGoOtelTracer.Start(baseCtx, reflecting.GetCurrentFunc())
	s, _ := sonic.MarshalString(ctx)
	span.SetAttributes(attribute.Key("ClickEvent").String(s))
	defer span.End()

	if err := consts.FlowControl.Top(); err != nil {
		errorsender.SendErrorInfo(ctx.Extra.TargetID, "", "", err, context.Background())
		return
	}
	consts.FlowControl.Add()
	defer consts.FlowControl.Sub()
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
			Ctx: baseCtx,
		}
	)
	if strings.HasPrefix(command, "GPTTrace:") {
		if asyncStruct, ok := gpt3.GPTAsyncMap[command]; ok {
			if database.CheckIsAdmin(ctx.Extra.UserID) || asyncStruct.AuthorID == ctx.Extra.UserID {
				*asyncStruct.Channel <- ctx.Extra.UserInfo.Username
			} else {
				utility.SendMessageTemp(ctx.Extra.TargetID, "", commandCtx.Common.AuthorID, "这不是你创建的消息，你没有权限终止它。")
			}
		} else {
			utility.SendMessageTemp(ctx.Extra.TargetID, "", commandCtx.Common.AuthorID, "消息已经终止，请不要再尝试点击。")
		}
		return
	} else if strings.HasPrefix(command, "GPTRetry:") {
		list := strings.Split(command, ":")
		rawMsg := list[1]
		msgDetail, err := consts.GlobalSession.MessageView(ctx.Extra.MsgID)
		if err != nil {
			return
		}
		msgQuoted, err := consts.GlobalSession.MessageView(msgDetail.Quote.ID)
		if err != nil {
			return
		}
		gpt3.ClientHandlerStreamUpdate(baseCtx, ctx.Extra.GuildID, msgDetail.Quote.ID, msgQuoted.Author.ID, ctx.Extra.MsgID, rawMsg)
		return
	} else if strings.HasPrefix(command, "GPTReset:") {

		authorID := strings.Split(command, ":")[1]
		if authorID != ctx.Common.AuthorID {
			utility.SendMessageTemp(ctx.Extra.TargetID, "", ctx.Common.AuthorID, "你点击了别人的重置按钮，这将不会生效。")
		}
		gpt3.ClientHandlerStream(baseCtx, "", "", authorID, "RESET")
		utility.SendMessageTemp(ctx.Extra.TargetID, "", authorID, "你的会话已经重置成功，现在你可以点击重试来继续使用ChatGPT")
		return
	}
	if command == "Refresh" {
		m, err := consts.GlobalSession.MessageView(ctx.Extra.MsgID)
		if err != nil {
			return
		}
		root, err := ajson.Unmarshal([]byte(m.Content))
		res, err := root.JSONPath("$..title")
		if err != nil {
			return
		}
		for _, r := range res {
			str := strings.Trim(r.String(), "\"")
			sp := strings.Split(str, "- ")
			ID := sp[len(sp)-1]
			Name := strings.Join(sp[:len(sp)-1], "- ")
			r.Parent().DeleteKey("src")
			infoList, err := neteaseapi.NetEaseGCtx.GetMusicURLByID(baseCtx, []*neteaseapi.MusicIDName{{ID, Name}})
			if err != nil {
				return
			}
			for _, info := range infoList {
				r.Parent().AppendObject("src", ajson.StringNode("", info.URL))
			}
		}
		res, err = root.JSONPath("$..modules[(@.length-2)].text")
		if err != nil {
			return
		}
		res[0].DeleteKey("content")
		res[0].AppendObject("content", ajson.StringNode("", fmt.Sprintf("> 音乐无法播放？试试刷新音源\n> 当前音源版本:`%s`", time.Now().Local().Format("01-02T15:04:05"))))
		r, err := ajson.Marshal(root)
		if err != nil {
			return
		}
		fmt.Println(string(r))
		err = consts.GlobalSession.MessageUpdate(&kook.MessageUpdate{
			MessageUpdateBase: kook.MessageUpdateBase{
				MsgID:   ctx.Extra.MsgID,
				Content: string(r),
			},
		})
		if err != nil {
			log.Zlog.Error(err.Error())
			return
		}
		return
	}
	commandCtx.ContextHandler(command)
	time.Sleep(time.Second)
}

func channelJoinedHandler(baseCtx context.Context, ctx *kook.GuildChannelMemberAddContext) {
	defer utility.CollectPanic(baseCtx, ctx.Common, ctx.Common.TargetID, ctx.Common.MsgID, "")
	userInfo, err := utility.GetUserInfo(ctx.Extra.UserID, ctx.Common.TargetID)
	if err != nil {
		errorsender.SendErrorInfo(consts.NotifierChanID, "", userInfo.ID, err, baseCtx)
		return
	}
	channelInfo, err := utility.GetChannnelInfo(ctx.Extra.ChannelID)
	if err != nil {
		errorsender.SendErrorInfo(consts.NotifierChanID, "", userInfo.ID, err, baseCtx)
		return
	}
	// !频道日志记录
	newChanLog := &database.ChannelLogExt{
		UserID:      userInfo.ID,
		UserName:    userInfo.Username,
		ChannelID:   channelInfo.ID,
		ChannelName: channelInfo.Name,
		JoinedTime:  ctx.Extra.JoinedAt.ToTime().Format(consts.TimeFormat),
		LeftTime:    "",
		GuildID:     utility.GetGuildIDFromChannelID(channelInfo.ID),
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
		errorsender.SendErrorInfo(ctx.Common.TargetID, "", "", err, baseCtx)
		return
	}
	resp, err := consts.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: consts.NotifierChanID,
				Content:  cardMessageStr,
			},
		},
	)
	if err != nil {
		errorsender.SendErrorInfo(consts.NotifierChanID, "", "", err, baseCtx)
		return
	}
	newChanLog.MsgID = resp.MsgID
	// 写入数据库记录
	if err = newChanLog.AddJoinedRecord(); err != nil {
		errorsender.SendErrorInfo(consts.NotifierChanID, "", userInfo.ID, err, baseCtx)
	}
}

func guildUpdateHandler(kookCtx *kook.GuildUpdateContext) {
}

func channelLeftHandler(baseCtx context.Context, kookCtx *kook.GuildChannelMemberDeleteContext) {
	defer utility.CollectPanic(baseCtx, kookCtx.Extra, kookCtx.Common.TargetID, "", kookCtx.Extra.UserID)
	// 离开频道时，记录频道信息
	userInfo, err := utility.GetUserInfo(kookCtx.Extra.UserID, kookCtx.Common.TargetID)
	if err != nil {
		errorsender.SendErrorInfo(consts.TestChanID, "", userInfo.ID, err, baseCtx)
		return
	}
	channelInfo, err := utility.GetChannnelInfo(kookCtx.Extra.ChannelID)
	if err != nil {
		errorsender.SendErrorInfo(consts.TestChanID, "", userInfo.ID, err, baseCtx)
		return
	}

	// !频道日志记录
	newChanLog := &database.ChannelLogExt{
		UserID:      userInfo.ID,
		UserName:    userInfo.Username,
		ChannelID:   channelInfo.ID,
		ChannelName: channelInfo.Name,
		JoinedTime:  "",
		LeftTime:    kookCtx.Extra.ExitedAt.ToTime().Format(consts.TimeFormat),
		GuildID:     "",
	}
	if newChanLog, err = newChanLog.UpdateLeftTime(); err != nil {
		errorsender.SendErrorInfo(consts.TestChanID, "", userInfo.ID, err, baseCtx)
		return
	}
	joinTimeT, _ := time.Parse(consts.TimeFormat, newChanLog.JoinedTime)
	leftTimeT, _ := time.Parse(consts.TimeFormat, newChanLog.LeftTime)
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
		errorsender.SendErrorInfo(kookCtx.Common.TargetID, "", "", err, baseCtx)
		return
	}
	err = consts.GlobalSession.MessageUpdate(
		&kook.MessageUpdate{
			MessageUpdateBase: kook.MessageUpdateBase{
				MsgID:   newChanLog.MsgID,
				Content: cardMessageStr,
			},
		},
	)
	if err != nil {
		errorsender.SendErrorInfo(consts.NotifierChanID, "", "", err, baseCtx)
		return
	}
}

func messageEventHandler(baseCtx context.Context, kookCtx *kook.KmarkdownMessageContext) {
	baseCtx, span := otel.BetaGoOtelTracer.Start(baseCtx, reflecting.GetCurrentFunc())
	rawRecord, _ := json.Marshal(&kookCtx.Extra)
	span.SetAttributes(attribute.Key("Record").String(string(rawRecord)))
	defer span.End()
	if kookCtx.Common.Type != kook.MessageTypeKMarkdown {
		return
	}
	// 配合每分钟自我健康检查，接收到指定消息写入chan
	if kookCtx.Common.Content == consts.SelfCheckMessage && kookCtx.Extra.Author.Bot {
		consts.SelfCheckChan <- "ok"
	}
	if kookCtx.Extra.Author.Bot {
		return
	}
	defer wordcontrol.RemoveDirtyWords(kookCtx)
	CommandHandler(baseCtx, kookCtx)
}
