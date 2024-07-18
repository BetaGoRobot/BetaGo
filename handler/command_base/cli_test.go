package commandBase

import (
	"context"
	"fmt"
	"testing"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

var larkCommandNilFunc CommandFunc[*larkim.P2MessageReceiveV1]

func bar1Handler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	_, _ = ctx, data
	fmt.Println("Executing bar1 with args:", args)
	return nil
}

func bar2Handler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	_, _ = ctx, data
	fmt.Println("Executing bar2 with args:", args)
	return nil
}

func TestCommandForLark(t *testing.T) {
	_ = t
	rootLarkCmd := NewRootCommand(larkCommandNilFunc).
		AddSubCommand(
			NewCommand("foo", larkCommandNilFunc).
				AddSubCommand(
					NewCommand("bar1", bar1Handler),
				).
				AddSubCommand(
					NewCommand("bar2", bar2Handler),
				),
		)
	fmt.Println(rootLarkCmd.Execute(context.Background(), nil, []string{"foo", "bar1", "--test"}))
	fmt.Println(rootLarkCmd.Execute(context.Background(), nil, []string{"foo", "bar2", "--test"}))
}
