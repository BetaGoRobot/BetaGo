package doubao

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
)

func GenerateChat(ctx context.Context, historyMsgList ...string) (answer string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	// historyMsg := "- " + strings.Join(historyMsgList, "\n- ")

	return
}
