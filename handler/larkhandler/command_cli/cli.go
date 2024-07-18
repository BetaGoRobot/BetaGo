// Package commandcli a package
package commandcli

import (
	"context"
	"strings"

	"github.com/dlclark/regexp2"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// var _ IMCommand = &IMCommandHandler[any]{}

var Re = regexp2.MustCompile(`\/(\w+)(\s+|$)`, 0)

type CommandContext struct {
	Name string
	Args []string
}

func (c *CommandContext) ParseCommand(input string) string {
	for _, in := range strings.Fields(input) {
		if c.Name != "" && strings.HasPrefix(in, "/") {
			c.Name = strings.TrimLeft(in, "/")
			continue
		}
		if c.Name != "" {
			if strings.HasPrefix(in, "--") {
				c.Args = append(c.Args, strings.TrimLeft(in, "--"))
			}
		}

	}
	return c.Name
}

// IMCommand IMCommand
type IMCommand interface {
	Init()
	Run(ctx context.Context, command string) error
}

// IMCommandHandler test
type IMCommandHandler[T any] struct {
	Commands []*IMCommandOperatorBase[T]
}

// Init test
//
//	@param *IMCommandHandler[T]
//	@return Init
func (*IMCommandHandler[T]) Init() {
}

// Run  Run
//
//	@param *IMCommandHandler[T]
//	@return Run
func (h *IMCommandHandler[T]) Run(ctx context.Context, data T, command string) error {
	if command == "" {
		return nil
	}
	if !strings.HasPrefix(command, "/") {
		return nil
	}
	command = strings.TrimLeft(command, "/")
	sList := strings.Fields(command)
	command = sList[0]
	args := strings.Join(sList[1:], " ")
	for _, c := range h.Commands {
		if c.Name == command {
			return c.RunSubCommands(ctx, data, args)
		}
	}
	return nil
}

// RunSubCommands test
//
//	@param op
//	@return RunSubCommands
func (op *IMCommandOperatorBase[T]) RunSubCommands(ctx context.Context, data T, command string) error {
	if command == "" {
		return nil
	}
	if op.Func != nil {
		return op.Func(ctx, data, command)
	}
	if len(op.SubCommands) > 0 {
		for _, c := range op.SubCommands {
			if c.Name == command {
				return c.RunSubCommands(ctx, data, command)
			}
		}
	}
	return nil
}

// IMCommandOperatorBase IMCommandOperatorBase
type IMCommandOperatorBase[T any] struct {
	Name        string
	SubCommands []*IMCommandOperatorBase[T]
	Func        func(ctx context.Context, data T, args ...string) error
}

func init() {
	larkCommand := &IMCommandOperatorBase[*larkim.P2MessageReceiveV1]{
		Name: "debug",
		SubCommands: []*IMCommandOperatorBase[*larkim.P2MessageReceiveV1]{
			{
				Name:        "get_id",
				SubCommands: []*IMCommandOperatorBase[*larkim.P2MessageReceiveV1]{},
				Func:        GetIDHandler,
			},
		},
	}
	handler := &IMCommandHandler[*larkim.P2MessageReceiveV1]{
		Commands: []*IMCommandOperatorBase[*larkim.P2MessageReceiveV1]{larkCommand},
	}
	handler.Run(context.Background(), &larkim.P2MessageReceiveV1{}, "/debug get_id")
	_ = larkCommand
}
