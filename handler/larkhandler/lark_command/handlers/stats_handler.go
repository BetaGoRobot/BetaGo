package handlers

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func StatsGetHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	resList, hitCache := database.FindByCacheFunc(
		database.InteractionStats{}, func(item database.InteractionStats) string {
			return item.GuildID
		},
	)
	span.SetAttributes(attribute.Bool("InteractionStats hitCache", hitCache))

	for _, res := range resList {
		_ = res
	}
	return
}
