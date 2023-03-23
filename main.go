package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/check"
	"github.com/BetaGoRobot/BetaGo/commandHandler/notifier"
	"github.com/BetaGoRobot/BetaGo/manager"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/kevinmatthe/zaplog"
)

var (
	zapLogger   = utility.ZapLogger
	sugerLogger = utility.SugerLogger
)

func init() {
	utility.InitLogger()
	go func() {
		http.ListenAndServe(":6060", nil)
	}()
}

func main() {
	check.CheckEnv()
	go manager.ReconnectUsingChan()
	e := betagovar.GlobalSession.Open()
	if e != nil {
		utility.ZapLogger.Error("连接失败", zaplog.Error(e))
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
	notifier.OfflineMessage(betagovar.GlobalSession)
	betagovar.GlobalSession.Close()
}
