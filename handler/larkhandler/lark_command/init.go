package larkcommand

import (
	commandBase "github.com/BetaGoRobot/BetaGo/handler/command_base"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// LarkRootCommand lark root command node
var LarkRootCommand *commandBase.Command[*larkim.P2MessageReceiveV1]

func init() {
	LarkRootCommand = commandBase.
		NewRootCommand(larkCommandNilFunc).
		AddSubCommand(
			commandBase.NewCommand("debug", larkCommandNilFunc).
				AddSubCommand(
					commandBase.NewCommand("get_id", getIDHandler),
				).
				AddSubCommand(
					commandBase.NewCommand("get_group_id", getGroupIDHandler),
				),
		)
}
