package main

import (
	"testing"

	"github.com/lonelyevil/khl"
)

func Test_addAdministrator(t *testing.T) {
	adminCommand(&khl.KmarkdownMessageContext{
		EventHandlerCommonContext: &khl.EventHandlerCommonContext{
			Session: &khl.Session{},
			Common: &khl.EventDataGeneral{
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
		Extra: khl.EventCustomMessage{},
	})
}
