package manager

import (
	"strings"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	comcontext "github.com/BetaGoRobot/BetaGo/commandHandler/context"
	"github.com/BetaGoRobot/BetaGo/utility"
	jsoniter "github.com/json-iterator/go"
	"github.com/lonelyevil/kook"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// CommandHandler  is a async handler for command event
//
//	@param ctx
func CommandHandler(ctx *kook.KmarkdownMessageContext) {
	// 判断是否被at到,且消息不是引用/回复
	if !utility.IsInSlice(betagovar.RobotID, ctx.Extra.Mention) &&
		!strings.Contains(ctx.Common.Content, betagovar.RobotID) &&
		!strings.HasPrefix(ctx.Common.Content, ".") {
		return
	}
	// 示例中，由于用户发送的命令的Content格式为(met)id(met) <command> <parameters>
	// 针对解析的判断逻辑，首先判断是否为空字符串，若为空发送help信息
	// ! 解析出不包含at信息的实际内容
	command, parameters := utility.GetCommandWithParameters(ctx.Common.Content)
	commandCtx := comcontext.GetNewCommandCtx().Init(ctx.EventHandlerCommonContext).InitExtra(ctx)
	if command != "" {
		commandCtx.ContextHandler(command, parameters...)
	} else {
		// 内容为空，发送help信息
		commandCtx.ContextHandler(comcontext.CommandContextTypeHelper)
	}
}

// ChannelJoinedAsyncHandler is a async handler for channel joined event
//
//	@param ctx
func ChannelJoinedAsyncHandler(ctx *kook.GuildChannelMemberAddContext) {
	go channelJoinedHandler(ctx)
}

// guildUpdateAsyncHandler  is a async handler for guild update event
//
//	@param ctx
func guildUpdateAsyncHandler(ctx *kook.GuildUpdateContext) {
	go guildUpdateHandler(ctx)
}

// ChannelLeftAsyncHandler  is a async handler for channel left event
//
//	@param ctx
func ChannelLeftAsyncHandler(ctx *kook.GuildChannelMemberDeleteContext) {
	go channelLeftHandler(ctx)
}

// ClickEventAsyncHandler  is a async handler for click event
//
//	@param ctx
func ClickEventAsyncHandler(ctx *kook.MessageButtonClickContext) {
	go clickEventHandler(ctx)
}

// MessageEventAsyncHandler  is a async handler for message event
//
//	@param ctx
func MessageEventAsyncHandler(ctx *kook.KmarkdownMessageContext) {
	go messageEventHandler(ctx)
}
