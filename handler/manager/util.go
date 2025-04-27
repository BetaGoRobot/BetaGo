package manager

import (
	"context"
	"os"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/gotify"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/lonelyevil/kook"
	"github.com/lonelyevil/kook/log_adapter/plog"
	p_log "github.com/phuslu/log"
)

// SendMessageToTestChannel  is a async handler for message event
//
//	@param session
//	@param content
func SendMessageToTestChannel(session *kook.Session, content string) {
	session.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     9,
			TargetID: consts.TestChanID,
			Content:  content,
		},
	})
}

// ReconnectUsingChan pass
func ReconnectUsingChan() {
	for {
		select {
		case <-consts.ReconnectChan:
			err := Reconnect()
			if err != nil {
				gotify.SendMessage(context.Background(), "", "Reconnect failed, error is "+err.Error(), 7)
			}
		}
	}
}

// Reconnect 重建链接
func Reconnect() (err error) {
	err = consts.GlobalSession.Close()
	if err != nil {
		return
	}
	time.Sleep(time.Second)
	consts.GlobalSession = kook.New(os.Getenv("BOTAPI"), plog.NewLogger(&p_log.Logger{
		Level:  p_log.InfoLevel,
		Writer: &p_log.ConsoleWriter{},
	}))
	AddAllHandler()
	err = consts.GlobalSession.Open()
	// retryCnt := 0
	// for err != nil {
	// 	time.Sleep(100 * time.Millisecond)
	// 	betagovar.GlobalSession.Close()
	// 	betagovar.GlobalSession = kook.New(os.Getenv("BOTAPI"), plog.NewLogger(&log.Logger{
	// 		Level:  log.InfoLevel,
	// 		Writer: &log.ConsoleWriter{},
	// 	}))
	// 	err = betagovar.GlobalSession.Open()
	// 	if err != nil {
	// 		gotify.SendMessage("", "Reconnect failed, error is "+err.Error(), 7)
	// 	}
	// 	if retryCnt++; retryCnt == 5 {
	// 		return fmt.Errorf("reconnect to kook server reaches max retry cnt 5, need restart or try again" + err.Error())
	// 	}
	// }
	log.Zlog.Info("Reconnecting successfully")
	time.Sleep(time.Second * 5)
	return
}
