package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "net/http/pprof"

	"github.com/lonelyevil/khl"
	"github.com/lonelyevil/khl/log_adapter/plog"
	"github.com/phuslu/log"
)

var robotName string
var robotID string
var globalSession = khl.New(os.Getenv("BOTAPI"), plog.NewLogger(&log.Logger{
	Level:  log.InfoLevel,
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
	init := NetEaseContext{}
	err := init.loginNetEase()
	if err != nil {
		log.Error().Err(err).Msg("error in init loginNetease")
	}

	globalSession.AddHandler(messageHan)
	globalSession.AddHandler(receiveDirectMessage)
}

func main() {
	// go func() {
	globalSession.Open()
	startUpMessage(globalSession)
	go dailySend()
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc
	// Cleanly close down the KHL session.
	offlineMessage(globalSession)
	globalSession.Close()
	// }()

	// http.ListenAndServe("0.0.0.0:6060", nil)
}

func messageHan(ctx *khl.TextMessageContext) {
	go func() {
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
		searchMusicByRobot(ctx)
		// scheduleEvent(ctx)
		// sendScheduledMessage(ctx)
	}()

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
