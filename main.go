package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/commandHandler/notifier"
	"github.com/BetaGoRobot/BetaGo/commandHandler/wordcontrol"
	"github.com/BetaGoRobot/BetaGo/scheduletask"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/heyuhengmatt/zaplog"
	"github.com/lonelyevil/khl"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	zapLogger   = utility.ZapLogger
	sugerLogger = utility.SugerLogger
)

func init() {
	go func() {
		port := ":2112"
		if os.Getenv("IS_TEST") == "true" {
			port = ":2113"
		}
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(port, nil)
	}()
	utility.InitLogger()
	betagovar.GlobalSession.AddHandler(messageHan)
	betagovar.GlobalSession.AddHandler(clickEventAsyncHandler)
	betagovar.GlobalSession.AddHandler(receiveDirectMessage)
	betagovar.GlobalSession.AddHandler(channelJoinedAsyncHandler)
	betagovar.GlobalSession.AddHandler(channelLeftAsyncHandler)
	go func() {
		// pprof监控
		http.ListenAndServe(":6060", nil)
	}()
}

// CheckEnv  检查环境变量
func CheckEnv() {
	if betagovar.RobotName == "" {
		sendMessageToTestChannel(betagovar.GlobalSession, ">  机器人未配置名称！")
	}
	if betagovar.RobotID == "" {
		zaplog.Logger.Fatal("机器人未配置ID！")
	}
}

func main() {
	CheckEnv()
	e := betagovar.GlobalSession.Open()
	if e != nil {
		zaplog.Logger.Error("连接失败", zaplog.Error(e))
		panic(e)
	}
	notifier.StartUpMessage(betagovar.GlobalSession)
	go scheduletask.DailyRecommand()
	// go scheduletask.HourlyGetSen()
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
