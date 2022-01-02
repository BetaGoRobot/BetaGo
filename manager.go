package main

import (
	"strings"

	"github.com/lonelyevil/khl"
)

func replaceDirtyWords(ctx *khl.TextMessageContext) {
	message := ctx.Common.Content
	if strings.Contains(message, "傻") && strings.Contains(message, "逼") || strings.Contains(message, "傻逼") {
		message = strings.ReplaceAll("傻", message, "")
		message = strings.ReplaceAll("逼", message, "")
		ctx.Session.MessageUpdate(&khl.MessageUpdate{MessageUpdateBase: khl.MessageUpdateBase{MsgID: ctx.Common.MsgID, Content: message}})
		// ctx.Session.MessageCreate(&khl.MessageCreate{
		// 	MessageCreateBase: khl.MessageCreateBase{
		// 		TargetID: ctx.Common.TargetID,
		// 		Content:  fmt.Sprintf("%s 使用了侮辱词汇，消息已被修正，我们不可以向他学习！", ctx.Extra.Author.Nickname),
		// 		Quote:    ctx.Common.MsgID,
		// 	},
		// })
		// ctx.Session.MessageDelete(ctx.Common.MsgID)
	}

}
