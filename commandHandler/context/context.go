package context

import (
	"fmt"
	"time"

	errorsender "github.com/BetaGoRobot/BetaGo/commandHandler/error_sender"
	"github.com/enescakir/emoji"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/lonelyevil/kook"
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
//
//	@receiver ctx
//	@return bool
func (ctx *CommandContext) IsAdmin() bool {
	return utility.CheckIsAdmin(ctx.Common.AuthorID)
}

// CommandContextFunc is a function for command.
//
//	@param TargetID
//	@param MsgID
//	@param AuthorID
//	@param parameters
//	@return error
type CommandContextFunc func(TargetID, MsgID, AuthorID string, parameters ...string) error

// CommandContextWithGuildIDFunc is a function for command.
//
//	@param targetID
//	@param quoteID
//	@param authorID
//	@param guildID
//	@param args
//	@return error
type CommandContextWithGuildIDFunc func(targetID, quoteID, authorID string, guildID string, args ...string) error

// GetNewCommandCtx  is a function for command.
//
//	@return *CommandContext
func GetNewCommandCtx() *CommandContext {
	return &CommandContext{
		Common: &CommandCommonContext{},
		Extra:  &CommandExtraContext{},
	}
}

// Init is a init function for command.
//
//	@receiver ctx
func (ctx *CommandContext) Init(khlCtx *kook.EventHandlerCommonContext) *CommandContext {
	*ctx = CommandContext{
		Common: &CommandCommonContext{
			TargetID: khlCtx.Common.TargetID,
			AuthorID: khlCtx.Common.AuthorID,
			MsgID:    khlCtx.Common.MsgID,
		},
	}
	return ctx
}

// InitExtra is a init function for command.
//
//	@receiver ctx
//	@param khlCtx
//	@return *CommandContext
func (ctx *CommandContext) InitExtra(khlCtx interface{}) *CommandContext {
	switch khlCtx.(type) {
	case *kook.KmarkdownMessageContext:
		khlCtx := khlCtx.(*kook.KmarkdownMessageContext)
		ctx.Extra = &CommandExtraContext{
			GuildID: khlCtx.Extra.GuildID,
		}
	case *kook.MessageButtonClickContext:
		khlCtx := khlCtx.(*kook.MessageButtonClickContext)
		ctx.Extra = &CommandExtraContext{
			GuildID: khlCtx.Extra.GuildID,
		}
	}
	return ctx
}

// ErrorSenderHandler is a function for command.
//
//	@receiver ctx
//	@param err
func (ctx *CommandContext) ErrorSenderHandler(err error) {
	if err != nil {
		errorsender.SendErrorInfo(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, err)
	}
}

// ErrorSenderHandlerNew  is a function for command.
//
//	@receiver ctx
//	@param err
func (ctx *CommandContext) ErrorSenderHandlerNew(ctxFunc interface{}, parameters ...string) {
	var err error
	if realFunc, ok := ctxFunc.(CommandContextFunc); ok {
		err = realFunc(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, parameters...)
	}
	if realFunc, ok := ctxFunc.(CommandContextWithGuildIDFunc); ok {
		err = realFunc(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, ctx.Extra.GuildID, parameters...)
	}
	if err != nil {
		errorsender.SendErrorInfo(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, err)
	}
}

// ContextHandler is a function for command.
//
//	@receiver ctx
//	@param Command
//	@param parameters
func (ctx *CommandContext) ContextHandler(Command string, parameters ...string) {
	defer utility.CollectPanic(ctx, ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID)
	var ctxFunc CommandContextFunc
	var ctxGuildFunc CommandContextWithGuildIDFunc
	ctxFunc = commandMapping[Command]
	ctxGuildFunc = commandMappingWithGuildID[Command]
	if ctxFunc == nil && ctxGuildFunc == nil {
		ctx.ErrorSenderHandler(fmt.Errorf(emoji.QuestionMark.String()+"未知指令 `%s`", Command))
		return
	}
	if ctxFunc != nil {
		defer utility.GetTimeCost(time.Now(), Command)
		ctx.ErrorSenderHandlerNew(ctxFunc, parameters...)
	}
	if ctxGuildFunc != nil {
		defer utility.GetTimeCost(time.Now(), Command)
		ctx.ErrorSenderHandlerNew(ctxGuildFunc, parameters...)
	}
}
