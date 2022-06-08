package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/scheduletask"
	"github.com/lonelyevil/khl"
	"github.com/phuslu/log"
)

var robotName string
var robotID string

var testChannelID string

func init() {
	fmt.Println("TEst")
	if robotName = os.Getenv("ROBOT_NAME"); robotName == "" {
		robotName = "No RobotName Configured"
	}
	if testChannelID = os.Getenv("TEST_CHAN_ID"); testChannelID == "" {
	}
	if robotID = os.Getenv("ROBOT_ID"); robotID == "" {
		sendMessageToTestChannel(betagovar.GlobalSession, fmt.Sprintf("> %s 机器人未配置ID！", robotName))
	}
	init := neteaseapi.NetEaseContext{}
	err := init.LoginNetEase()
	if err != nil {
		log.Error().Err(err).Msg("error in init loginNetease")
	}

	betagovar.GlobalSession.AddHandler(messageHan)
	betagovar.GlobalSession.AddHandler(receiveDirectMessage)
}

func main() {
	betagovar.GlobalSession.Open()
	startUpMessage(betagovar.GlobalSession)
	go scheduletask.DailyRecommand()
	go scheduletask.HourlyGetSen()
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc
	// Cleanly close down the KHL session.
	offlineMessage(betagovar.GlobalSession)
	betagovar.GlobalSession.Close()
}

func messageHan(ctx *khl.KmarkdownMessageContext) {

	go func() {
		if ctx.Common.Type != khl.MessageTypeKMarkdown || ctx.Extra.Author.Bot {
			return
		}

		defer removeDirtyWords(ctx)
		commandHandler(ctx)
	}()

}
