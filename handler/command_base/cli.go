// Package commandBase  抽象的Command执行体结构，利用泛型提供多数据类型支持的Command解析、执行流程，约定Root节点为初始节点，不会参与执行匹配
//
//	@update 2024-07-18 05:25:13
package commandBase

import (
	"context"
	"fmt"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
)

// CommandFunc Repeat
//
//	@author heyuhengmatt
//	@update 2024-07-18 04:43:42
type CommandFunc[T any] func(ctx context.Context, data T, metaData *handlerbase.BaseMetaData, args ...string) error

// Command Repeat
//
//	@author heyuhengmatt
//	@update 2024-07-18 04:43:37
type Command[T any] struct {
	Name        string
	SubCommands map[string]*Command[T]
	Func        CommandFunc[T]
	Usage       string
	SupportArgs map[string]struct{}
	curComChain []string
}

// Execute 从当前节点开始，执行Command
//
//	@param c *Command[T]
//	@return Execute
//	@author heyuhengmatt
//	@update 2024-07-18 05:30:21
func (c *Command[T]) Execute(ctx context.Context, data T, metaData *handlerbase.BaseMetaData, args []string) error {
	if c.Func != nil { // 当前Command有执行方法，直接执行
		return c.Func(ctx, data, metaData, args...)
	}
	if len(args) == 0 { // 无执行方法且无后续参数
		return fmt.Errorf("%w: %s", consts.ErrCommandIncomplete, c.FormatUsage())
	}

	if subcommand, ok := c.SubCommands[args[0]]; ok {
		if usage, ok := subcommand.CheckUsage(args[1:]...); ok {
			return fmt.Errorf("%w: %s", consts.ErrCheckUsage, usage)
		}
		err := subcommand.Execute(ctx, data, metaData, args[1:])
		if err != nil && err == consts.ErrArgsIncompelete {
			return fmt.Errorf("%w: %s", consts.ErrArgsIncompelete, subcommand.FormatUsage())
		}
		return err
	}

	return fmt.Errorf(
		"%w: Command <b>%s</b> not found, available sub-commands: %s",
		consts.ErrCommandNotFound,
		args[0],
		fmt.Sprintf(" [%s]", strings.Join(c.GetSubCommands(), ", ")),
	)
}

// BuildChain 从当前节点开始，执行Command
//
//	@param c *Command[T]
//	@return Execute
//	@author heyuhengmatt
//	@update 2024-07-18 05:30:21
func (c *Command[T]) BuildChain() {
	for _, subcommand := range c.SubCommands {
		subcommand.curComChain = append(c.curComChain, subcommand.Name)
		subcommand.BuildChain()
	}
}

// FormatUsage 获取当前节点的所有SubCommands
//
//	@param c
//	@return GetSubCommands
func (c *Command[T]) FormatUsage() string {
	if c.Usage == "" {
		baseUsage := fmt.Sprintf("Usage: %s", "/"+strings.Join(c.curComChain, " "))
		if len(c.SupportArgs) != 0 {
			baseUsage += fmt.Sprintf(" <%s>", strings.Join(c.GetSupportArgs(), ", "))
		}
		if len(c.SubCommands) != 0 {
			baseUsage += fmt.Sprintf(" [%s]", strings.Join(c.GetSubCommands(), ", "))
		}

		return baseUsage
	}
	return c.Usage
}

// CheckUsage 获取当前节点的所有SubCommands
//
//	@param c
//	@return GetSubCommands
func (c *Command[T]) CheckUsage(args ...string) (usage string, isHelp bool) {
	if len(args) == 1 {
		if args[0] == "--help" {
			return c.FormatUsage(), true
		}
	}
	for index, arg := range args {
		if _, ok := c.SupportArgs[arg]; ok {
			continue
		}
		if subcommand, ok := c.SubCommands[arg]; ok {
			return subcommand.CheckUsage(args[index+1:]...)
		}
	}
	return "", false
}

// GetSubCommands 获取当前节点的所有SubCommands
//
//	@param c
//	@return GetSubCommands
func (c *Command[T]) GetSubCommands() []string {
	availableComs := make([]string, 0, len(c.SubCommands))
	for k := range c.SubCommands {
		availableComs = append(availableComs, k)
	}
	return availableComs
}

// GetSupportArgs 获取当前节点的所有SubCommands
//
//	@param c
//	@return GetSubCommands
func (c *Command[T]) GetSupportArgs() []string {
	supportArgs := make([]string, 0, len(c.SupportArgs))
	for k := range c.SupportArgs {
		supportArgs = append(supportArgs, k)
	}
	return supportArgs
}

// Validate 从当前节点开始，执行Command
//
//	@param c *Command[T]
//	@return Execute
//	@author heyuhengmatt
//	@update 2024-07-18 05:30:21
func (c *Command[T]) Validate(ctx context.Context, data T, args []string) bool {
	if c.Func != nil { // 当前Command有执行方法，直接执行
		return true
	}
	if len(args) == 0 { // 无执行方法且无后续参数
		return false
	}
	if subcommand, ok := c.SubCommands[args[0]]; ok {
		return subcommand.Validate(ctx, data, args[1:])
	}
	return true
}

// AddSubCommand 添加一个SubCommand
//
//	@param c *Command[T]
//	@return AddSubCommand
//	@author heyuhengmatt
//	@update 2024-07-18 05:30:07
func (c *Command[T]) AddSubCommand(subCommand *Command[T]) *Command[T] {
	c.SubCommands[subCommand.Name] = subCommand
	return c
}

// AddUsage 添加一个SubCommand
//
//	@param c *Command[T]
//	@return AddUsage
func (c *Command[T]) AddUsage(usage string) *Command[T] {
	c.Usage = usage
	return c
}

// AddArgs 添加一个SubCommand
//
//	@param c *Command[T]
//	@return AddUsage
func (c *Command[T]) AddArgs(args ...string) *Command[T] {
	for _, arg := range args {
		c.SupportArgs[arg] = struct{}{}
	}
	return c
}

// NewCommand 创建一个新的Command结构
//
//	@param name string
//	@param fn CommandFunc[T]
//	@return *Command
//	@author heyuhengmatt
//	@update 2024-07-18 05:29:58
func NewCommand[T any](name string, fn CommandFunc[T]) *Command[T] {
	return &Command[T]{
		Name:        name,
		SubCommands: make(map[string]*Command[T]),
		Func:        fn,
		SupportArgs: make(map[string]struct{}),
	}
}

// NewCommand 创建一个新的Command结构
//
//	@param name string
//	@param fn CommandFunc[T]
//	@return *Command
//	@author heyuhengmatt
//	@update 2024-07-18 05:29:58
func NewRootCommand[T any](fn CommandFunc[T]) *Command[T] {
	return &Command[T]{
		Name:        "root",
		SubCommands: make(map[string]*Command[T]),
		Func:        fn,
	}
}
