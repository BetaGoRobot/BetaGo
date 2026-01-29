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
	"github.com/BetaGoRobot/BetaGo/utility/ark/tools"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/cardutil"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/vadvisor"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/gg/goption"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
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
	defer func() { span.RecordError(err) }()

	argMap, _ := parseArgs(args...)

	var (
		cardContent *templates.TemplateCardContent
		days        int
		hoursInt    int
		defaultDays = 30
		st, et      time.Time
	)
	defer func() {
		if err != nil {
			metaData.SetExtra("gold_result", "执行失败，错误原因"+err.Error())
		} else {
			metaData.SetExtra("gold_result", "执行成功")
		}
	}()
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

	if hours, ok := argMap["h"]; ok {
		if st.IsZero() || et.IsZero() {
			hoursInt, err = strconv.Atoi(hours)
			if err != nil || hoursInt <= 0 {
				hoursInt = 1
			}
			st = time.Now().Add(time.Duration(-1*hoursInt) * time.Hour)
			et = time.Now()
		}

		cardContent, err = GetRealtimeGoldPriceGraph(ctx, st, et)
		fmt.Println(cardContent.String())
		if err != nil {
			return err
		}
	} else if daysStr, ok := argMap["d"]; ok {
		if st.IsZero() || et.IsZero() {
			days, err = strconv.Atoi(daysStr)
			if err != nil || days <= 0 {
				days = defaultDays
			}
			st = time.Now().AddDate(0, 0, -1*days)
			et = time.Now()
		}

		cardContent, err = GetHistoryGoldGraph(ctx, st, et)
		if err != nil {
			return err
		}
	} else {
		cardContent, err = GetHistoryGoldGraph(ctx, st, et)
		if err != nil {
			return err
		}
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
	defer func() { span.RecordError(err) }()

	var (
		days                  = 1
		defaultDays           = 1
		st, et      time.Time = time.Now().AddDate(0, 0, -1*defaultDays), time.Now()
		stockCode   string
	)

	argMap, _ := parseArgs(args...)

	stockCode, ok := argMap["code"]
	if !ok {
		return fmt.Errorf("stock code is required")
	}

	if daysStr, ok := argMap["days"]; ok {
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			days = defaultDays
		}
		st, et = time.Now().AddDate(0, 0, -1*days), time.Now()
	}

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
				if t.Before(st) || t.After(et) {
					continue
				}

				if !yield(vadvisor.XYSUnit[string, float64]{X: t.In(utility.UTCPlus8Loc()).Format(time.DateTime), Y: utility.Must2Float(price.Open), S: "开盘"}) {
					return
				}
				if !yield(vadvisor.XYSUnit[string, float64]{X: t.In(utility.UTCPlus8Loc()).Format(time.DateTime), Y: utility.Must2Float(price.Close), S: "收盘"}) {
					return
				}
				if !yield(vadvisor.XYSUnit[string, float64]{X: t.In(utility.UTCPlus8Loc()).Format(time.DateTime), Y: utility.Must2Float(price.High), S: "最高"}) {
					return
				}
				if !yield(vadvisor.XYSUnit[string, float64]{X: t.In(utility.UTCPlus8Loc()).Format(time.DateTime), Y: utility.Must2Float(price.Low), S: "最低"}) {
					return
				}
			}
		},
	)
	cardContent := cardutil.NewCardBuildGraphHelper(graph).
		SetTitle(fmt.Sprintf("沪A-[%s]%s-近<%d>天", stockCode, stockName, days)).
		SetStartTime(st).
		SetEndTime(et).
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
	return
}

func GetHistoryGoldGraph(ctx context.Context, st, et time.Time) (*templates.TemplateCardContent, error) {
	logs.L().Ctx(ctx).Info("GetHistoryGoldGraph", zap.String("st", st.Format(time.RFC3339)), zap.String("et", et.Format(time.RFC3339)))
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
					if t.Before(st) || t.After(et) {
						continue
					}
					d := t.Format(time.DateOnly)
					if !yield(vadvisor.XYSUnit[string, float64]{X: d, Y: price.Close, S: "收盘价"}) ||
						!yield(vadvisor.XYSUnit[string, float64]{X: d, Y: price.Open, S: "开盘价"}) ||
						!yield(vadvisor.XYSUnit[string, float64]{X: d, Y: price.High, S: "最高价"}) ||
						!yield(vadvisor.XYSUnit[string, float64]{X: d, Y: price.Low, S: "最低价"}) {
						return
					}
				}
			},
		)
	card := cardutil.NewCardBuildGraphHelper(graph).
		SetTitle("沪金所价格数据").
		SetStartTime(st).
		SetEndTime(et).
		Build(ctx)
	return card, nil
}

func GetRealtimeGoldPriceGraph(ctx context.Context, st, et time.Time) (*templates.TemplateCardContent, error) {
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
					if t.Before(st) || t.After(et) {
						continue
					}
					if !yield(vadvisor.XYSUnit[string, float64]{X: t.Format(time.TimeOnly), Y: price.Price, S: price.Kind}) {
						return
					}
				}
			},
		)
	card := cardutil.NewCardBuildGraphHelper(graph).
		SetTitle("沪金所价格数据").
		SetStartTime(st).
		SetEndTime(et).
		Build(ctx)
	return card, nil
}

func init() {
	params := tools.NewParameters("object").
		AddProperty("start_time", &tools.Property{
			Type:        "string",
			Description: "开始时间，默认可以不穿，格式为YYYY-MM-DD HH:MM:SS",
		}).
		AddProperty("end_time", &tools.Property{
			Type:        "string",
			Description: "结束时间，默认可以不传，格式为YYYY-MM-DD HH:MM:SS",
		}).
		AddProperty("hours", &tools.Property{
			Type:        "number",
			Description: "查询的小时数，默认1小时",
		}).
		AddProperty("days", &tools.Property{
			Type:        "number",
			Description: "查询的天数，默认30天",
		})
	fcu := tools.NewFunctionCallUnit().
		Name("gold_price_get").Desc("搜索指定时间范围内的金价变化情况，可选相对时间天或小时，也可以指定时间范围").Params(params).Func(goldWrap)
	tools.M().Add(fcu)
}

func goldWrap(ctx context.Context, meta *tools.FunctionCallMeta, args string) (any, error) {
	s := struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Days      *int   `json:"days"`
		Hours     *int   `json:"hours"`
	}{}
	err := utility.UnmarshallStringPre(args, &s)
	if err != nil {
		return nil, err
	}
	argsSlice := make([]string, 0)
	if s.Days != nil && *s.Days > 0 {
		argsSlice = append(argsSlice, "--d", strconv.Itoa(*s.Days))
	}
	if s.Hours != nil && *s.Hours > 0 {
		argsSlice = append(argsSlice, "--h", strconv.Itoa(*s.Hours))
	}
	if s.StartTime != "" {
		argsSlice = append(argsSlice, "--st="+s.StartTime)
	}
	if s.EndTime != "" {
		argsSlice = append(argsSlice, "--et="+s.EndTime)
	}
	metaData := handlerbase.NewBaseMetaDataWithChatIDUID(ctx, meta.ChatID, meta.UserID)
	if err := GoldHandler(ctx, meta.LarkData, metaData, argsSlice...); err != nil {
		return nil, err
	}
	return goption.Of(metaData.GetExtra("gold_result")).ValueOr("执行完成但没有结果"), nil
}
