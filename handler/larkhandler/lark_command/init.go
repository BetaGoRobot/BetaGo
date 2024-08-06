package larkcommand

import (
	commandBase "github.com/BetaGoRobot/BetaGo/handler/command_base"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler/lark_command/handlers"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// LarkRootCommand lark root command node
var LarkRootCommand *commandBase.Command[*larkim.P2MessageReceiveV1]

var newCmd = commandBase.NewCommand[*larkim.P2MessageReceiveV1]

func init() {
	LarkRootCommand = commandBase.
		NewRootCommand(larkCommandNilFunc).
		AddSubCommand(
			newCmd("debug", larkCommandNilFunc).
				AddSubCommand(
					newCmd("msgid", handlers.DebugGetIDHandler),
				).
				AddSubCommand(
					newCmd("chatid", handlers.DebugGetGroupIDHandler),
				).
				AddSubCommand(
					newCmd("panic", handlers.DebugTryPanicHandler),
				).
				AddSubCommand(
					newCmd("trace", handlers.DebugTraceHandler),
				),
		).
		AddSubCommand(
			newCmd("word", larkCommandNilFunc).
				AddSubCommand(
					newCmd("add", handlers.WordAddHandler).AddArgs("word", "rate"),
				).
				AddSubCommand(
					newCmd("get", handlers.WordGetHandler),
				),
		).
		AddSubCommand(
			newCmd("reply", larkCommandNilFunc).
				AddSubCommand(
					newCmd("add", handlers.ReplyAddHandler).AddArgs("word", "reply", "type"),
				).
				AddSubCommand(
					newCmd("get", handlers.ReplyGetHandler),
				),
		).
		AddSubCommand(
			newCmd("image", larkCommandNilFunc).
				AddSubCommand(
					newCmd("add", handlers.ImageAddHandler).AddArgs("url").AddArgs("img_key"),
				).
				AddSubCommand(
					newCmd("get", handlers.ImageGetHandler),
				),
		)
	LarkRootCommand.BuildChain()
}
