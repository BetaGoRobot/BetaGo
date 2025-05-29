package handlers

import (
	"context"
	"fmt"

	"github.com/BetaGoRobot/BetaGo/dal/aktool"
	commandBase "github.com/BetaGoRobot/BetaGo/handler/command_base"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func StockHandler(stockType string) commandBase.CommandFunc[*larkim.P2MessageReceiveV1] {
	switch stockType {
	case "gold":
		return GoldHandler
	}
	return nil
}

func GoldHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	goldPrice, err := aktool.GetRealtimeGoldPrice(ctx)
	if err != nil {
		return
	}
	var latestPrice string
	if len(goldPrice) > 0 {
		latestPrice = fmt.Sprintf("*%.2f*", goldPrice[0].Price)
	}
	err = larkutils.ReplyCardText(
		ctx,
		"上交所当前金价: "+latestPrice,
		*data.Event.Message.MessageId,
		"_getID",
		false,
	)
	if err != nil {
		return
	}
	return
}
