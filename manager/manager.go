package manager

import (
	"context"
	"strings"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	comcontext "github.com/BetaGoRobot/BetaGo/commandHandler/context"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	jsoniter "github.com/json-iterator/go"
	"github.com/lonelyevil/kook"
	"go.opentelemetry.io/otel/attribute"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// CommandHandler  is a async handler for command event
//
//	@param ctx
func CommandHandler(baseCtx context.Context, kookCtx *kook.KmarkdownMessageContext) {
	baseCtx, span := jaeger_client.BetaGoCommandTracer.Start(baseCtx, utility.GetCurrentFunc())
	rawRecord, _ := json.Marshal(&kookCtx.Extra)
	span.SetAttributes(attribute.Key("Record").String(string(rawRecord)))
	defer span.End()

	if kookCtx.Extra.Author.Bot {
		return
	}
	// 判断是否被at到,且消息不是引用/回复
	if !utility.IsInSlice(betagovar.RobotID, kookCtx.Extra.Mention) &&
		!strings.Contains(kookCtx.Common.Content, betagovar.RobotID) &&
		!strings.HasPrefix(kookCtx.Common.Content, ".") {
		return
	}
	// 示例中，由于用户发送的命令的Content格式为(met)id(met) <command> <parameters>
	// 针对解析的判断逻辑，首先判断是否为空字符串，若为空发送help信息
	// ! 解析出不包含at信息的实际内容
	command, parameters := utility.GetCommandWithParameters(kookCtx.Common.Content)
	commandCtx := comcontext.GetNewCommandCtx().Init(kookCtx.EventHandlerCommonContext).InitContext(baseCtx).InitExtra(kookCtx)
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
func ChannelJoinedAsyncHandler(kookCtx *kook.GuildChannelMemberAddContext) {
	newCtx := context.Background()
	newCtx, span := jaeger_client.BetaGoCommandTracer.Start(newCtx, utility.GetCurrentFunc())
	rawRecord, _ := json.Marshal(&kookCtx.Extra)
	span.SetAttributes(attribute.Key("Record").String(string(rawRecord)))
	defer span.End()
	go channelJoinedHandler(newCtx, kookCtx)
}

// guildUpdateAsyncHandler  is a async handler for guild update event
//
//	@param ctx
func guildUpdateAsyncHandler(kookCtx *kook.GuildUpdateContext) {
	go guildUpdateHandler(kookCtx)
}

// ChannelLeftAsyncHandler  is a async handler for channel left event
//
//	@param ctx
func ChannelLeftAsyncHandler(kookCtx *kook.GuildChannelMemberDeleteContext) {
	newCtx := context.Background()
	newCtx, span := jaeger_client.BetaGoCommandTracer.Start(newCtx, utility.GetCurrentFunc())
	rawRecord, _ := json.Marshal(&kookCtx.Extra)
	span.SetAttributes(attribute.Key("Record").String(string(rawRecord)))
	defer span.End()
	go channelLeftHandler(newCtx, kookCtx)
}

// ClickEventAsyncHandler  is a async handler for click event
//
//	@param ctx
func ClickEventAsyncHandler(kookCtx *kook.MessageButtonClickContext) {
	newCtx := context.Background()
	newCtx, span := jaeger_client.BetaGoCommandTracer.Start(newCtx, utility.GetCurrentFunc())
	rawRecord, _ := json.Marshal(&kookCtx.Extra)
	span.SetAttributes(attribute.Key("Record").String(string(rawRecord)))
	defer span.End()
	go clickEventHandler(newCtx, kookCtx)
}

// MessageEventAsyncHandler  is a async handler for message event
//
//	@param ctx
func MessageEventAsyncHandler(kookCtx *kook.KmarkdownMessageContext) {
	newCtx := context.Background()
	newCtx, span := jaeger_client.BetaGoCommandTracer.Start(newCtx, utility.GetCurrentFunc())
	rawRecord, _ := json.Marshal(&kookCtx.Extra)
	span.SetAttributes(attribute.Key("Record").String(string(rawRecord)))
	defer span.End()
	go messageEventHandler(newCtx, kookCtx)
}
