package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/lonelyevil/khl"
	"github.com/lonelyevil/khl/log_adapter/plog"
	"github.com/phuslu/log"
)

var robotName string
var robotID string
var globalSession = khl.New(os.Getenv("BOTAPI"), plog.NewLogger(&log.Logger{
	Level:  log.TraceLevel,
	Writer: &log.ConsoleWriter{},
}))

var testChannelID string

func init() {
	if robotName = os.Getenv("ROBOT_NAME"); robotName == "" {
		robotName = "No RobotName Configured"
	}
	if testChannelID = os.Getenv("TEST_CHAN_ID"); testChannelID == "" {
	}
	if robotID = os.Getenv("ROBOT_ID"); robotID == "" {
		sendMessageToTestChannel(globalSession, fmt.Sprintf("> %s 机器人未配置ID！", robotName))
	}
	globalSession.AddHandler(messageHan)
}

func main() {
	globalSession.Open()
	startUpMessage(globalSession)
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc
	// Cleanly close down the KHL session.
	offlineMessage(globalSession)
	globalSession.Close()
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

	replyToMention(ctx)
	replaceDirtyWords(ctx)
	// sendScheduledMessage(ctx)
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
