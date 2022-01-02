package main

import (
	"fmt"
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

// 机器人被at时返回消息
func replyToMention(ctx *khl.TextMessageContext) {
	if isInSlice("3508390651", ctx.Extra.Mention) {
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
	currentIp, err := GetOutBoundIP()
	if err != nil {
		return
	}
	session.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     9,
			TargetID: "7419593543056418",
			Content:  fmt.Sprintf("---------\n> Robot `BetaGo-Nightly` is \n`online`\n IP:\t%s\n Time:\t%s", currentIp, time.Now().String()),
		}})
	return
}

func offlineMessage(session *khl.Session) (err error) {
	currentIp, err := GetOutBoundIP()
	if err != nil {
		return
	}
	session.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     9,
			TargetID: "7419593543056418",
			Content:  fmt.Sprintf("---------\n> Robot `BetaGo-Nightly` is \n`offline`\n IP:\t%s\n Time:\t%s", currentIp, time.Now().String()),
		}})
	return
}