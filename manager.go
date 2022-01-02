package main

import (
	"fmt"
	"strings"

	"github.com/lonelyevil/khl"
)

func replaceDirtyWords(ctx *khl.TextMessageContext) {
	message := ctx.Common.Content
	if strings.Contains(message, "傻") && strings.Contains(message, "逼") || strings.Contains(message, "傻逼") {
		// message = strings.ReplaceAll(message, "傻", "*")
		// message = strings.ReplaceAll(message, "逼", "*")
		// ctx.Session.MessageUpdate(&khl.MessageUpdate{MessageUpdateBase: khl.MessageUpdateBase{MsgID: ctx.Common.TargetID, Content: message}})
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
