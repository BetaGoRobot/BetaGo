package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/BetaGoRobot/BetaGo/dal/aktool"
	commandBase "github.com/BetaGoRobot/BetaGo/handler/command_base"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/vadvisor"
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

	argMap, _ := parseArgs(args...)
	if _, ok := argMap["r"]; ok {
		var goldPrice aktool.GoldPriceDataRTList
		goldPrice, err = aktool.GetRealtimeGoldPrice(ctx)
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
	} else if daysStr, ok := argMap["h"]; ok {
		goldPrices, err := aktool.GetHistoryGoldPrice(ctx)
		if err != nil {
			return err
		}
		graph := vadvisor.NewMultiSeriesLineGraph[string, float64]()
		days, err := strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			days = 30
		}
		var min, max *float64
		for _, price := range goldPrices {
			t, err := time.Parse("2006-01-02T00:00:00.000", price.Date)
			if err != nil {
				return err
			}
			if t.Before(time.Now().AddDate(0, 0, -1*days)) {
				continue
			}
			d := t.Format(time.DateOnly)
			graph.AddData(d, price.Close, "收盘价")
			graph.AddData(d, price.Open, "开盘价")
			graph.AddData(d, price.High, "最高价")
			graph.AddData(d, price.Low, "最低价")
			if min == nil || max == nil {
				min, max = new(float64), new(float64)
				*min, *max = price.Low, price.High
			}
			if price.Low < *min {
				*min = price.Low
			}
			if price.High > *max {
				*max = price.High
			}
		}

		graph.SetRange((1-0.1)**min, (1+0.1)**max)
		graph.SetTitle(fmt.Sprintf("上交所黄金价格-%ddays", days))
		err = larkutils.ReplyCardTextGraph(
			ctx,
			"",
			graph,
			*data.Event.Message.MessageId,
			"_getID",
			false,
		)
	}

	return
}
