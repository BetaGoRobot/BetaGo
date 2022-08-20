package wordcontrol

import (
	"fmt"
	"strings"

	goaway "github.com/TwiN/go-away"
	"github.com/lonelyevil/khl"
)

// RemoveDirtyWords 删除脏词
//
//	@param ctx
func RemoveDirtyWords(ctx *khl.KmarkdownMessageContext) {
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
