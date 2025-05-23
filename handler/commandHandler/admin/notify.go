package admin

import (
	"context"
	"fmt"

	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"go.opentelemetry.io/otel/attribute"
)

func Handler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	if !database.CheckIsAdmin(authorID) {
		return fmt.Errorf("你不是管理员，无法执行此操作")
	}

	return
}
