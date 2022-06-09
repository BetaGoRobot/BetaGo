package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/commandHandler/notifier"
	"github.com/BetaGoRobot/BetaGo/commandHandler/wordcontrol"
	"github.com/BetaGoRobot/BetaGo/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/scheduletask"
	"github.com/lonelyevil/khl"
	"github.com/phuslu/log"
)

func init() {

	initNetease := neteaseapi.NetEaseContext{}
	err := initNetease.LoginNetEase()
	if err != nil {
		log.Error().Err(err).Msg("error in init loginNetease")
	}
	betagovar.GlobalSession.AddHandler(messageHan)
	betagovar.GlobalSession.AddHandler(clickEventHandler)
	betagovar.GlobalSession.AddHandler(receiveDirectMessage)
}

func test() {
	cardMessageStr, err := khl.CardMessage{&khl.CardMessageCard{
		Theme: khl.CardThemeSecondary,
		Size:  khl.CardSizeLg,
		Modules: []interface{}{
			khl.CardMessageSection{
				Mode: khl.CardMessageSectionModeRight,
				Text: khl.CardMessageElementText{
					Content: "这是一个测试",
				},
				Accessory: khl.CardMessageElementButton{
					Theme: khl.CardThemeInfo,
					Value: "clickTest",
					Click: "return-val",
					Text:  "测试按钮",
				},
			},
		},
	}}.BuildMessage()
	if err != nil {
		log.Printf("error in test: %v", err)
	}
	betagovar.GlobalSession.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeCard,
			TargetID: "7419593543056418",
			Content:  cardMessageStr,
		},
	})
}

// CheckEnv  检查环境变量
func CheckEnv() {
	if betagovar.RobotName == "" {
		sendMessageToTestChannel(betagovar.GlobalSession, ">  机器人未配置名称！")
	}
	if betagovar.RobotID == "" {
		sendMessageToTestChannel(betagovar.GlobalSession, ">  机器人未配置ID！")
		os.Exit(-1)
	}
}

func main() {
	CheckEnv()
	betagovar.GlobalSession.Open()
	test()
	notifier.StartUpMessage(betagovar.GlobalSession)
	go scheduletask.DailyRecommand()
	go scheduletask.HourlyGetSen()
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc
	// Cleanly close down the KHL session.
	notifier.OfflineMessage(betagovar.GlobalSession)
	betagovar.GlobalSession.Close()
}

func messageHan(ctx *khl.KmarkdownMessageContext) {
	go func() {
		if ctx.Common.Type != khl.MessageTypeKMarkdown || ctx.Extra.Author.Bot {
			return
		}
		defer wordcontrol.RemoveDirtyWords(ctx)
		commandHandler(ctx)
	}()

}
