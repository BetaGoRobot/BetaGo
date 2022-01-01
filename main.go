package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lonelyevil/khl"
	"github.com/lonelyevil/khl/log_adapter/plog"
	"github.com/phuslu/log"
)

func main() {
	l := log.Logger{
		Level:  log.TraceLevel,
		Writer: &log.ConsoleWriter{},
	}
	s := khl.New(os.Getenv("BOTAPI"), plog.NewLogger(&l))
	s.AddHandler(messageHan)
	s.Open()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc

	// Cleanly close down the KHL session.
	s.Close()
}

func messageHan(ctx *khl.TextMessageContext) {
	if ctx.Common.Type != khl.MessageTypeText || ctx.Extra.Author.Bot {
		return
	}
	if strings.Contains(ctx.Common.Content, "ping") {
		ctx.Session.MessageCreate(&khl.MessageCreate{
			MessageCreateBase: khl.MessageCreateBase{
				TargetID: ctx.Common.TargetID,
				Content:  "pong(此消息仅你自己可见)",
				Quote:    ctx.Common.MsgID,
			},
			TempTargetID: ctx.Common.AuthorID,
		})
	}

	// 机器人被at时返回消息
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

	if strings.Contains(ctx.Common.Content, "傻逼") && ctx.Common.AuthorID != "938697103" {
		ctx.Session.MessageCreate(&khl.MessageCreate{
			MessageCreateBase: khl.MessageCreateBase{
				TargetID: ctx.Common.TargetID,
				Content:  fmt.Sprintf("%s 使用了侮辱词汇，消息已被撤回，我们不可以向他学习！", ctx.Extra.Author.Nickname),
				Quote:    ctx.Common.MsgID,
			},
		})
		ctx.Session.MessageDelete(ctx.Common.MsgID)
	}

	if time.Now().Unix()-LastSendGeartl.Unix() > 120 {
		LastSendGeartl = time.Now()
		userChatSession, _ := ctx.Session.UserChatCreate("2423931199")
		chatCode := userChatSession.Code
		ctx.Session.DirectMessageCreate(&khl.DirectMessageCreate{MessageCreateBase: khl.MessageCreateBase{Content: "你为什么还不上线，小胖。你上线了可以在频道里多发言吗？你不说话你就是傻逼，真的。"}, ChatCode: chatCode})
	}
}

// 判断机器人是否被at到
func isInSlice(target string, slice []string) bool {
	for i := range slice {
		if slice[i] == target {
			return true
		}
	}
	return false
}
