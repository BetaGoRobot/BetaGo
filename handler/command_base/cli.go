// Package commandcli  抽象的Command执行体结构，利用泛型提供多数据类型支持的Command解析、执行流程，约定Root节点为初始节点，不会参与执行匹配
//
//	@update 2024-07-18 05:25:13
package commandBase

import (
	"context"
	"fmt"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
)

// CommandFunc Repeat
//
//	@author heyuhengmatt
//	@update 2024-07-18 04:43:42
type CommandFunc[T any] func(ctx context.Context, data T, args ...string) error

// Command Repeat
//
//	@author heyuhengmatt
//	@update 2024-07-18 04:43:37
type Command[T any] struct {
	Name        string
	SubCommands map[string]*Command[T]
	Func        CommandFunc[T]
}

// Execute 从当前节点开始，执行Command
//
//	@param c *Command[T]
//	@return Execute
//	@author heyuhengmatt
//	@update 2024-07-18 05:30:21
func (c *Command[T]) Execute(ctx context.Context, data T, args []string) error {
	if c.Func != nil { // 当前Command有执行方法，直接执行
		return c.Func(ctx, data, args...)
	}
	if len(args) == 0 { // 无执行方法且无后续参数
		return fmt.Errorf("%w: Command <b>%s</b> require sub-command", consts.ErrCommandIncomplete, c.Name)
	}
	if subcommand, ok := c.SubCommands[args[0]]; ok {
		return subcommand.Execute(ctx, data, args[1:])
	}
	availableComs := make([]string, 0, len(c.SubCommands))
	for k := range c.SubCommands {
		availableComs = append(availableComs, k)
	}

	return fmt.Errorf("%w: Command <b>%s</b> not found, available sub-commands: [%s]", consts.ErrCommandNotFound, args[0], strings.Join(availableComs, ", "))
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
