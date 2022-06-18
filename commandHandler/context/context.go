package context

import (
	"github.com/BetaGoRobot/BetaGo/commandHandler/admin"
	"github.com/BetaGoRobot/BetaGo/commandHandler/cal"
	"github.com/BetaGoRobot/BetaGo/commandHandler/helper"
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

// // Init is a init function for command.
// //  @receiver ctx
// func (ctx *CommandContext) Init(khlCtx *khl.EventHandlerCommonContext) *CommandContext {
// 	return &CommandContext{
// 		Common: &CommandCommonContext{
// 			TargetID: khlCtx.Common.TargetID,
// 			AuthorID: khlCtx.Common.AuthorID,
// 			MsgID:    khlCtx.Common.MsgID,
// 		},
// 	}
// }

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
	default:
		return
	}
}

// HelpHandler is a function for command.
//  @receiver ctx
//  @return error
func (ctx *CommandContext) HelpHandler(parameters ...string) error {
	if ctx.IsAdmin() {
		return helper.AdminCommandHelperHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...)
	}
	return helper.UserCommandHelperHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...)
}

// AdminAddHandler is a function for command.
//  @receiver ctx
//  @param parameters
//  @return string
func (ctx *CommandContext) AdminAddHandler(parameters ...string) error {
	return admin.AddAdminHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...)
}

// AdminRemoveHandler is a function for command.
//  @receiver ctx
//  @param parameters
//  @return error
func (ctx *CommandContext) AdminRemoveHandler(parameters ...string) error {
	return admin.RemoveAdminHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...)
}

// AdminShowHandler is a function for command.
//  @receiver ctx
//  @return error
func (ctx *CommandContext) AdminShowHandler() error {
	return admin.ShowAdminHandler(ctx.Common.TargetID, ctx.Common.MsgID)
}

// RollDiceHandler  is a function for command.
//  @receiver ctx
//  @return error
func (ctx *CommandContext) RollDiceHandler(parameters ...string) error {
	return roll.RandRollHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...)
}

// PingHandler is a function for command.
//  @receiver ctx
func (ctx *CommandContext) PingHandler() {
	helper.PingHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID)
}

// OneWordHandler is a function for command.
//  @receiver ctx
//  @param parameters
//  @return error
func (ctx *CommandContext) OneWordHandler(parameters ...string) error {
	return roll.OneWordHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...)
}

// SearchMusicHandler  is a function for command.
//  @receiver ctx
//  @param parameters
func (ctx *CommandContext) SearchMusicHandler(parameters ...string) error {
	return music.SearchMusicByRobot(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...)
}

// GetUserInfoHandler is a function for command.
//  @receiver ctx
//  @param guildID
//  @param parameters
//  @return error
func (ctx *CommandContext) GetUserInfoHandler(parameters ...string) error {
	return helper.GetUserInfoHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, ctx.Extra.GuildID, parameters...)
}

// ShowCalHandler is a function for command.
//  @receiver ctx
//  @param parameters
//  @return error
func (ctx *CommandContext) ShowCalHandler(parameters ...string) error {
	return cal.ShowCalHandler(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, ctx.Extra.GuildID, parameters...)
}
