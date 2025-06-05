package handlers

import (
	"cmp"
	"context"
	"fmt"
	"strconv"
	"time"

	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/history"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/cardutil"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/vadvisor"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/defensestation/osquery"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

// TrendHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return err error
//	@author kevinmatthe
//	@update 2025-05-30 15:19:56
func TrendHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	var (
		days     = 7
		interval = "1d"
	)

	argMap, _ := parseArgs(args...)
	if daysStr, ok := argMap["days"]; ok {
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			days = 30
		}
	}
	trend, err := history.New(ctx).
		Query(
			osquery.Bool().
				Must(
					osquery.Term("chat_id", *data.Event.Message.ChatId),
					osquery.Range("create_time").
						Gte(time.Now().AddDate(0, 0, -1*days).Format(time.DateTime)).
						Lte(time.Now().Format(time.DateTime)),
				),
		).
		GetTrend(
			interval,
			"user_name",
		)
	if err != nil {
		return err
	}

	if playType, ok := argMap["play"]; !ok {
		if inputInterval, ok := argMap["interval"]; ok {
			interval = inputInterval
		}
		graph := vadvisor.NewMultiSeriesLineGraph[string, int64]()
		var min, max *int64
		for _, item := range trend {
			if item.Key == "你" {
				item.Key = "机器人"
			}
			graph.AddData(item.Time, item.Value, item.Key)

			if min == nil || max == nil {
				min, max = new(int64), new(int64)
				*min, *max = item.Value, item.Value
			}

			if item.Value < *min {
				*min = item.Value
			}
			if item.Value > *max {
				*max = item.Value
			}
		}
		title := fmt.Sprintf("[%s]水群频率表-%ddays", larkutils.GetChatName(ctx, *data.Event.Message.ChatId), days)
		graph.
			SetRange(float64(*min), float64(*max))
		cardContent := cardutil.NewCardBuildGraphHelper(graph).
			SetTitle(title).Build(ctx)
		if metaData.Refresh {
			err = larkutils.PatchCard(ctx, cardContent, *data.Event.Message.MessageId)
		} else {
			err = larkutils.ReplyCard(ctx, cardContent, *data.Event.Message.MessageId, "", false)
		}
	} else {
		switch playType {
		case "pie":
			err = DrawTrendPie(ctx, trend, data, days, !metaData.Refresh)
		case "bar":
			err = DrawTrendBar(ctx, trend, data, days, !metaData.Refresh)
		default:
			err = DrawTrendPie(ctx, trend, data, days, !metaData.Refresh)
		}
	}

	return
}

func DrawTrendPie(ctx context.Context, trend history.TrendSeries, data *larkim.P2MessageReceiveV1, days int, reply bool) (err error) {
	graph := vadvisor.NewPieChartsGraphWithPlayer[string, int64]()
	for _, item := range trend {
		t, err := time.ParseInLocation(time.DateTime, item.Time, utility.UTCPlus8Loc())
		if err != nil {
			return err
		}
		if item.Key == "你" {
			item.Key = "机器人"
		}
		graph.AddData(
			t.Format(time.DateOnly),
			&vadvisor.ValueUnit[string, int64]{
				XField:      t.Format(time.DateOnly),
				SeriesField: item.Key,
				YField:      item.Value,
			})

	}
	graph.BuildPlayer(ctx)
	title := fmt.Sprintf("[%s]水群频率表-%ddays", larkutils.GetChatName(ctx, *data.Event.Message.ChatId), days)
	cardContent := cardutil.NewCardBuildGraphHelper(graph).
		SetTitle(title).Build(ctx)
	if reply {
		return larkutils.ReplyCard(ctx, cardContent, *data.Event.Message.MessageId, "", false)
	}
	return larkutils.PatchCard(ctx, cardContent, *data.Event.Message.MessageId)
}

func DrawTrendBar(ctx context.Context, trend history.TrendSeries, data *larkim.P2MessageReceiveV1, days int, reply bool) (err error) {
	graph := vadvisor.NewBarChartsGraphWithPlayer[string, int64]()
	for _, item := range trend {
		t, err := time.ParseInLocation(time.DateTime, item.Time, utility.UTCPlus8Loc())
		if err != nil {
			return err
		}
		if item.Key == "你" {
			item.Key = "机器人"
		}
		graph.AddData(
			t.Format(time.DateOnly),
			&vadvisor.ValueUnit[string, int64]{
				XField:      item.Key,
				SeriesField: item.Key,
				YField:      item.Value,
			},
		)

	}
	graph.SetDirection("horizontal").ReverseAxis()
	graph.SetSortFunc(func(a, b *vadvisor.ValueUnit[string, int64]) int {
		return cmp.Compare(b.YField, a.YField)
	})
	graph.BuildPlayer(ctx)
	title := fmt.Sprintf("[%s]水群频率表-%ddays", larkutils.GetChatName(ctx, *data.Event.Message.ChatId), days)
	cardContent := cardutil.NewCardBuildGraphHelper(graph).
		SetTitle(title).Build(ctx)
	if reply {
		return larkutils.ReplyCard(ctx, cardContent, *data.Event.Message.MessageId, "", false)
	}
	return larkutils.PatchCard(ctx, cardContent, *data.Event.Message.MessageId)
}
