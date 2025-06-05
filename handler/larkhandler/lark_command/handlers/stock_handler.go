package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/BetaGoRobot/BetaGo/dal/aktool"
	commandBase "github.com/BetaGoRobot/BetaGo/handler/command_base"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/cardutil"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
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
	case "a":
		return ZhAStockHandler
	}
	return nil
}

func GoldHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	argMap, _ := parseArgs(args...)

	var (
		cardContent *templates.TemplateCardContent
		days        int
	)
	if hours, ok := argMap["r"]; ok {
		hoursInt, e := strconv.Atoi(hours)
		if e != nil || hoursInt <= 0 {
			hoursInt = 1
		}
		cardContent, err = GetRealtimeGoldPriceGraph(ctx, hoursInt)
		if err != nil {
			return err
		}
	} else if daysStr, ok := argMap["h"]; ok {
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			days = 30
		}

		cardContent, err = GetHistoryGoldGraph(ctx, days)
		if err != nil {
			return err
		}
	} else {
		return
	}

	if metaData != nil && metaData.Refresh {
		err = larkutils.PatchCard(ctx,
			cardContent,
			*data.Event.Message.MessageId)
	} else {
		err = larkutils.ReplyCard(ctx,
			cardContent,
			*data.Event.Message.MessageId, "", false)
	}

	return
}

func ZhAStockHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	argMap, _ := parseArgs(args...)
	if stockCode, ok := argMap["code"]; !ok {
		return fmt.Errorf("stock code is required")
	} else {
		days := 1
		if daysStr, ok := argMap["days"]; ok {
			days, err = strconv.Atoi(daysStr)
			if err != nil || days <= 0 {
				days = 1
			}
		}
		graph := vadvisor.NewMultiSeriesLineGraph[string, float64]()
		stockPrice, err := aktool.GetStockPriceRT(ctx, stockCode)
		if err != nil {
			return err
		}
		stockName, err := aktool.GetStockSymbolInfo(ctx, stockCode)
		if err != nil {
			return err
		}
		graph.AddPointSeries(
			func(yield func(vadvisor.XYSUnit[string, float64]) bool) {
				for _, price := range stockPrice {
					t, err := time.ParseInLocation(time.DateTime, price.DateTime, utility.UTCPlus8Loc())
					if err != nil {
						return
					}
					if t.Before(time.Now().AddDate(0, 0, -1*days)) {
						continue
					}

					if !yield(vadvisor.XYSUnit[string, float64]{XField: t.Format(time.DateTime), YField: utility.Must2Float(price.Open), SeriesField: "开盘"}) {
						return
					}
					if !yield(vadvisor.XYSUnit[string, float64]{XField: t.Format(time.DateTime), YField: utility.Must2Float(price.Close), SeriesField: "收盘"}) {
						return
					}
					if !yield(vadvisor.XYSUnit[string, float64]{XField: t.Format(time.DateTime), YField: utility.Must2Float(price.High), SeriesField: "最高"}) {
						return
					}
					if !yield(vadvisor.XYSUnit[string, float64]{XField: t.Format(time.DateTime), YField: utility.Must2Float(price.Low), SeriesField: "最低"}) {
						return
					}
				}
			},
		)
		cardContent := cardutil.NewCardBuildGraphHelper(graph).
			SetTitle(fmt.Sprintf("沪A-[%s]%s-近<%d>天", stockCode, stockName, days)).
			Build(ctx)
		if metaData != nil && metaData.Refresh {
			err = larkutils.PatchCard(ctx,
				cardContent,
				*data.Event.Message.MessageId)
		} else {
			err = larkutils.ReplyCard(ctx,
				cardContent,
				*data.Event.Message.MessageId, "", false)
		}
	}
	return
}

func GetHistoryGoldGraph(ctx context.Context, days int) (*templates.TemplateCardContent, error) {
	graph := vadvisor.NewMultiSeriesLineGraph[string, float64]()
	goldPrices, err := aktool.GetHistoryGoldPrice(ctx)
	if err != nil {
		return nil, err
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
		)
	card := cardutil.NewCardBuildGraphHelper(graph).
		SetTitle(fmt.Sprintf("沪金所-近<%d>天", days)).
		Build(ctx)
	return card, nil
}

func GetRealtimeGoldPriceGraph(ctx context.Context, hoursInt int) (*templates.TemplateCardContent, error) {
	graph := vadvisor.NewMultiSeriesLineGraph[string, float64]()
	goldPrice, err := aktool.GetRealtimeGoldPrice(ctx)
	if err != nil {
		return nil, err
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
		)
	card := cardutil.NewCardBuildGraphHelper(graph).
		SetTitle(fmt.Sprintf("沪金所-近<%d>小时", hoursInt)).
		Build(ctx)
	return card, nil
}
