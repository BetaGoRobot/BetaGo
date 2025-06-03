package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/BetaGoRobot/BetaGo/dal/aktool"
	commandBase "github.com/BetaGoRobot/BetaGo/handler/command_base"
	"github.com/BetaGoRobot/BetaGo/utility"
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

	graph := vadvisor.NewMultiSeriesLineGraph[string, float64]()

	if hours, ok := argMap["r"]; ok {
		hoursInt, e := strconv.Atoi(hours)
		if e != nil || hoursInt <= 0 {
			hoursInt = 1
		}
		var goldPrice aktool.GoldPriceDataRTList
		goldPrice, err = aktool.GetRealtimeGoldPrice(ctx)
		if err != nil {
			return
		}

		graph.
			AddPointSeries(
				func(yield func(vadvisor.XYSUnit[string, float64]) bool) {
					for _, price := range goldPrice {
						dStr := time.Now().Format(time.DateOnly) + " " + price.Time
						t, err := time.ParseInLocation(time.DateTime, dStr, utility.UTCPlus8Loc())
						if err != nil {
							return
						}
						if t.Before(time.Now().Add(-1 * time.Hour * time.Duration(hoursInt))) {
							continue
						}
						if !yield(vadvisor.XYSUnit[string, float64]{XField: t.Format(time.TimeOnly), YField: price.Price, SeriesField: price.Kind}) {
							return
						}
					}
				},
			).
			SetTitle(fmt.Sprintf("上交所黄金价格- *[T-%d]* (hour)", hoursInt))
		err = larkutils.ReplyCardTextGraph(
			ctx,
			"",
			graph,
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

		days, err := strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			days = 30
		}

		graph.
			AddPointSeries(
				func(yield func(vadvisor.XYSUnit[string, float64]) bool) {
					for _, price := range goldPrices {
						t, err := time.Parse("2006-01-02T00:00:00.000", price.Date)
						if err != nil {
							return
						}
						if t.Before(time.Now().AddDate(0, 0, -1*days)) {
							continue
						}
						d := t.Format(time.DateOnly)
						if !yield(vadvisor.XYSUnit[string, float64]{XField: d, YField: price.Close, SeriesField: "收盘价"}) ||
							!yield(vadvisor.XYSUnit[string, float64]{XField: d, YField: price.Open, SeriesField: "开盘价"}) ||
							!yield(vadvisor.XYSUnit[string, float64]{XField: d, YField: price.High, SeriesField: "最高价"}) ||
							!yield(vadvisor.XYSUnit[string, float64]{XField: d, YField: price.Low, SeriesField: "最低价"}) {
							return
						}
					}
				},
			).
			SetTitle(fmt.Sprintf("上交所黄金价格- *[T-%d]* (day)", days))
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
