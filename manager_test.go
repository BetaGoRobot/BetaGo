package main

import (
	"testing"

	"github.com/BetaGoRobot/BetaGo/manager"
	"github.com/lonelyevil/kook"
)

func Test_addAdministrator(t *testing.T) {
	manager.CommandHandler(&kook.KmarkdownMessageContext{
		EventHandlerCommonContext: &kook.EventHandlerCommonContext{
			Session: &kook.Session{},
			Common: &kook.EventDataGeneral{
				ChannelType:  "",
				Type:         0,
				TargetID:     "",
				AuthorID:     "",
				Content:      "addAdmin 1234567 xxx",
				MsgID:        "",
				MsgTimestamp: 0,
				Nonce:        "",
			},
		},
		Extra: kook.EventCustomMessage{},
	})
}

func Test_clickEventHandler(t *testing.T) {
	// for a test
}
