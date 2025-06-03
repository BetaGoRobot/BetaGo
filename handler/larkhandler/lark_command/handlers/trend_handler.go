package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility/history"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
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
func TrendHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	var (
		days     = 7
		interval = "1d"
	)

	argMap, _ := parseArgs(args...)
	if daysStr, ok := argMap["days"]; ok {
		days, err := strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			days = 30
		}
	}
	if inputInterval, ok := argMap["interval"]; ok {
		interval = inputInterval
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
		SetTitle(title).
		SetRange(float64(*min), float64(*max))
	err = larkutils.ReplyCardTextGraph(
		ctx,
		"",
		graph,
		*data.Event.Message.MessageId,
		"_getID",
		false,
	)
	return
}
