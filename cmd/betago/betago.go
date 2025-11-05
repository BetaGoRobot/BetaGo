package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/consts/check"
	"github.com/BetaGoRobot/BetaGo/handler/commandHandler/notifier"
	server "github.com/BetaGoRobot/BetaGo/handler/webhookserver"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
)

func init() {
	go func() {
		http.ListenAndServe(":6060", nil)
	}()
	go server.Start()
}

func main() {
	check.CheckEnv()
	e := consts.GlobalSession.Open()
	if e != nil {
		logs.L.Error().Err(e).Msg("连接失败")
		panic(e)
	}
	notifier.StartAutoService()
	// go scheduletask.HourlyGetSen()
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
	// Cleanly close down the KHL session.
	notifier.OfflineMessage(consts.GlobalSession)
	consts.GlobalSession.Close()
}
