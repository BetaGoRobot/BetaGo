package doubao

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
)

func GenerateChat(ctx context.Context, historyMsgList ...string) (answer string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	// historyMsg := "- " + strings.Join(historyMsgList, "\n- ")

	return
}
