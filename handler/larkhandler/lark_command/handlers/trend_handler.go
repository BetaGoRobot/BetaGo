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
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/vadvisor"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
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
	defer func() { span.RecordError(err) }()

	var (
		days     = 7
		interval = "1d"
		st, et   time.Time
	)

	argMap, _ := parseArgs(args...)
	if inputInterval, ok := argMap["interval"]; ok {
		interval = inputInterval
	}
	if daysStr, ok := argMap["days"]; ok {
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			days = 30
		}
	}

	st, et = GetBackDays(days)
	// 如果有st，et的配置，用st，et的配置来覆盖
	if stStr, ok := argMap["st"]; ok {
		if etStr, ok := argMap["et"]; ok {
			st, err = time.Parse(time.RFC3339, stStr)
			if err != nil {
				return err
			}
			et, err = time.Parse(time.RFC3339, etStr)
			if err != nil {
				return err
			}
		}
	}
	helper := &trendInternalHelper{
		days:     days,
		st:       st,
		et:       et,
		msgID:    *data.Event.Message.MessageId,
		chatID:   *data.Event.Message.ChatId,
		interval: interval,
	}

	trend, err := helper.TrendByUser(ctx)
	if err != nil {
		return err
	}

	if playType, ok := argMap["play"]; ok {
		switch playType {
		case "bar":
			err = helper.DrawTrendBar(ctx, trend, !metaData.Refresh)
		default:
			err = helper.DrawTrendPie(ctx, trend, !metaData.Refresh)
		}
	} else {
		graph := vadvisor.NewMultiSeriesLineGraph[string, int64]()
		graph.AddPointSeries(
			func(yield func(vadvisor.XYSUnit[string, int64]) bool) {
				for _, item := range trend {
					if item.Key == "你" {
						item.Key = "机器人"
					}
					if !yield(vadvisor.XYSUnit[string, int64]{
						X: item.Time,
						Y: item.Value,
						S: item.Key,
					}) {
						return
					}
				}
			},
		)
		title := fmt.Sprintf("[%s]水群频率表-%ddays", larkutils.GetChatName(ctx, *data.Event.Message.ChatId), days)
		cardContent := cardutil.NewCardBuildGraphHelper(graph).
			SetTitle(title).Build(ctx)
		if metaData.Refresh {
			err = larkutils.PatchCard(ctx, cardContent, *data.Event.Message.MessageId)
		} else {
			err = larkutils.ReplyCard(ctx, cardContent, *data.Event.Message.MessageId, "", false)
		}
	}

	return
}

type trendInternalHelper struct {
	days          int
	st, et        time.Time
	msgID, chatID string
	interval      string
}

func (h *trendInternalHelper) DrawTrendPie(ctx context.Context, trend history.TrendSeries, reply bool) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

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
	title := fmt.Sprintf("[%s]水群频率表-%ddays", larkutils.GetChatName(ctx, h.chatID), h.days)
	cardContent := cardutil.NewCardBuildGraphHelper(graph).
		SetStartTime(h.st).
		SetEndTime(h.et).
		SetTitle(title).Build(ctx)
	if reply {
		return larkutils.ReplyCard(ctx, cardContent, h.msgID, "", false)
	}
	return larkutils.PatchCard(ctx, cardContent, h.msgID)
}

func (h *trendInternalHelper) DrawTrendBar(ctx context.Context, trend history.TrendSeries, reply bool) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

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
	title := fmt.Sprintf("[%s]水群频率表-%ddays", larkutils.GetChatName(ctx, h.chatID), h.days)
	cardContent := cardutil.NewCardBuildGraphHelper(graph).
		SetStartTime(h.st).
		SetEndTime(h.et).
		SetTitle(title).Build(ctx)
	if reply {
		return larkutils.ReplyCard(ctx, cardContent, h.msgID, "", false)
	}
	return larkutils.PatchCard(ctx, cardContent, h.msgID)
}

func (h *trendInternalHelper) TrendByUser(ctx context.Context) (trend history.TrendSeries, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	trend, err = history.New(ctx).
		Query(
			osquery.Bool().
				Must(
					osquery.Term("chat_id", h.chatID),
					osquery.Range("create_time_v2").
						Gte(h.st.Format(time.RFC3339)).
						Lte(h.et.Format(time.RFC3339)),
				),
		).
		GetTrend(
			h.interval,
			"user_name",
		)
	return
}

func (h *trendInternalHelper) TrendRate(ctx context.Context, indexName, field string, size uint64) (singleDimAggs *history.SingleDimAggregate, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	singleDimAggs = &history.SingleDimAggregate{}
	// 通过Opensearch统计发言数量
	req := osquery.Search().Query(
		osquery.Bool().
			Must(
				osquery.Term("chat_id", h.chatID),
				osquery.Range("create_time_v2").
					Gte(h.st.Format(time.RFC3339)).
					Lte(h.et.Format(time.RFC3339)),
			),
	).Size(0).Aggs(osquery.TermsAgg("dimension", field).Size(size))

	resp, err := opensearchdal.
		SearchData(
			ctx,
			indexName,
			req,
		)

	err = sonic.Unmarshal(resp.Aggregations, singleDimAggs)
	return
}

func GetBackDays(days int) (st, et time.Time) {
	st, et = time.Now().AddDate(0, 0, -1*days), time.Now()
	return
}
