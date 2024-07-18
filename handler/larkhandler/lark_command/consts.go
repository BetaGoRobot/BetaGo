package larkcommand

import (
	commandBase "github.com/BetaGoRobot/BetaGo/handler/command_base"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

var larkCommandNilFunc commandBase.CommandFunc[*larkim.P2MessageReceiveV1]
