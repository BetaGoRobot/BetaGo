package context

import (
	"github.com/BetaGoRobot/BetaGo/commandHandler/admin"
	"github.com/BetaGoRobot/BetaGo/commandHandler/cal"
	errorsender "github.com/BetaGoRobot/BetaGo/commandHandler/error_sender"
	"github.com/BetaGoRobot/BetaGo/commandHandler/helper"
	"github.com/BetaGoRobot/BetaGo/commandHandler/hitokoto"
	"github.com/BetaGoRobot/BetaGo/commandHandler/music"

	"github.com/BetaGoRobot/BetaGo/commandHandler/roll"

	"github.com/BetaGoRobot/BetaGo/dbpack"
	"github.com/lonelyevil/khl"
)

// CommandContext  is a context for command.
type CommandContext struct {
	Common *CommandCommonContext
	Extra  *CommandExtraContext
}

// CommandCommonContext  is a context for command.
type CommandCommonContext struct {
	TargetID string
	AuthorID string
	MsgID    string
}

// CommandExtraContext is a context for command.
type CommandExtraContext struct {
	GuildID string
}

// IsAdmin is a function for command.
//  @receiver ctx
//  @return bool
func (ctx *CommandContext) IsAdmin() bool {
	return dbpack.CheckIsAdmin(ctx.Common.AuthorID)
}

// Init is a init function for command.
//  @receiver ctx
func (ctx *CommandContext) Init(khlCtx *khl.EventHandlerCommonContext) {
	*ctx = CommandContext{
		Common: &CommandCommonContext{
			TargetID: khlCtx.Common.TargetID,
			AuthorID: khlCtx.Common.AuthorID,
			MsgID:    khlCtx.Common.MsgID,
		},
	}
}

// InitExtra is a init function for command.
//  @receiver ctx
//  @param khlCtx
//  @return *CommandContext
func (ctx *CommandContext) InitExtra(khlCtx interface{}) {
	switch khlCtx.(type) {
	case *khl.KmarkdownMessageContext:
		khlCtx := khlCtx.(*khl.KmarkdownMessageContext)
		ctx.Extra = &CommandExtraContext{
			GuildID: khlCtx.Extra.GuildID,
		}
	case *khl.MessageButtonClickContext:
		khlCtx := khlCtx.(*khl.MessageButtonClickContext)
		ctx.Extra = &CommandExtraContext{
			GuildID: khlCtx.Extra.GuildID,
		}
	default:
		return
	}
}

// HelpHandler is a function for command.
//  @receiver ctx
//  @return error
func (ctx *CommandContext) HelpHandler(parameters ...string) {
	if ctx.IsAdmin() {
		ctx.ErrorSenderHandler(helper.AdminCommandHelperHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...))
	}
	ctx.ErrorSenderHandler(helper.UserCommandHelperHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...))
}

// AdminAddHandler is a function for command.
//  @receiver ctx
//  @param parameters
//  @return string
func (ctx *CommandContext) AdminAddHandler(parameters ...string) {
	ctx.ErrorSenderHandler(admin.AddAdminHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...))
}

// AdminRemoveHandler is a function for command.
//  @receiver ctx
//  @param parameters
//  @return error
func (ctx *CommandContext) AdminRemoveHandler(parameters ...string) {
	ctx.ErrorSenderHandler(admin.RemoveAdminHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...))
}

// AdminShowHandler is a function for command.
//  @receiver ctx
//  @return error
func (ctx *CommandContext) AdminShowHandler() {
	ctx.ErrorSenderHandler(admin.ShowAdminHandler(ctx.Common.TargetID, ctx.Common.MsgID))
}

// RollDiceHandler  is a function for command.
//  @receiver ctx
//  @return error
func (ctx *CommandContext) RollDiceHandler(parameters ...string) {
	ctx.ErrorSenderHandler(roll.RandRollHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...))
}

// PingHandler is a function for command.
//  @receiver ctx
func (ctx *CommandContext) PingHandler() {
	helper.PingHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID)
}

// SearchMusicHandler  is a function for command.
//  @receiver ctx
//  @param parameters
func (ctx *CommandContext) SearchMusicHandler(parameters ...string) {
	ctx.ErrorSenderHandler(music.SearchMusicByRobot(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...))
}

// GetUserInfoHandler is a function for command.
//  @receiver ctx
//  @param guildID
//  @param parameters
//  @return error
func (ctx *CommandContext) GetUserInfoHandler(parameters ...string) {
	ctx.ErrorSenderHandler(helper.GetUserInfoHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, ctx.Extra.GuildID, parameters...))
}

// ShowCalHandler is a function for command.
//  @receiver ctx
//  @param parameters
//  @return error
func (ctx *CommandContext) ShowCalHandler(parameters ...string) {
	ctx.ErrorSenderHandler(cal.ShowCalHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, ctx.Extra.GuildID, parameters...))
}

// GetHitokotoHandler is a function for command.
//  @receiver ctx
//  @param parameters
//  @return error
func (ctx *CommandContext) GetHitokotoHandler(parameters ...string) {
	ctx.ErrorSenderHandler(hitokoto.GetHitokotoHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...))
}

// ErrorSenderHandler is a function for command.
//  @receiver ctx
//  @param err
func (ctx *CommandContext) ErrorSenderHandler(err error) {
	if err != nil {
		errorsender.SendErrorInfo(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, err)
	}
}
