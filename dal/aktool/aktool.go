package aktool

import (
	"context"
	"fmt"
	"os"

	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
)

var (
	BaseURL                 = os.Getenv("AKTOOL_BASE_URL")
	PublicAPIURI            = "/api/public/"
	GoldHandlerNameRealtime = "spot_quotations_sge"
	GoldHandlerNameHistory  = "spot_hist_sge"

	StockHandlerNameRealtime = "stock_zh_a_minute"
	StockSingleInfo          = "stock_individual_info_em"
)

type (
	GoldPriceDataRTList []*GoldPriceDataRT
	GoldPriceDataRT     struct {
		Kind       string  `json:"品种"`
		Time       string  `json:"时间"`
		Price      float64 `json:"现价"`
		UpdateTime string  `json:"更新时间"`
	}

	StockPriceDataRTList []*StockPriceDataRT
	StockPriceDataRT     struct {
		DateTime string `json:"day"` // "2025-05-23 10:25:00"
		Open     string `json:"open"`
		High     string `json:"high"`
		Low      string `json:"low"`
		Close    string `json:"close"`
		Volume   string `json:"volume"`
	}
)

func GetRealtimeGoldPrice(ctx context.Context) (res GoldPriceDataRTList, err error) {
	_, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	res = make(GoldPriceDataRTList, 0)
	c, _ := client.NewClient()
	req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
	req.SetRequestURI(BaseURL + PublicAPIURI + GoldHandlerNameRealtime)
	req.SetMethod("GET")

	err = c.Do(ctx, req, resp)
	if err != nil {
		return
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("get gold price failed, status code: %d", resp.StatusCode())
	}

	err = sonic.Unmarshal(resp.Body(), &res)
	if err != nil {
		return
	}
	return
}

type GoldPriceDataHS []struct {
	Date  string  `json:"date"`
	Open  float64 `json:"open"`
	Close float64 `json:"close"`
	Low   float64 `json:"low"`
	High  float64 `json:"high"`
}

func GetHistoryGoldPrice(ctx context.Context) (res GoldPriceDataHS, err error) {
	_, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	res = make(GoldPriceDataHS, 0)
	c, _ := client.NewClient()
	req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
	req.SetRequestURI(BaseURL + PublicAPIURI + GoldHandlerNameHistory)
	req.SetMethod("GET")

	err = c.Do(ctx, req, resp)
	if err != nil {
		return
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("get gold price failed, status code: %d", resp.StatusCode())
	}

	err = sonic.Unmarshal(resp.Body(), &res)
	if err != nil {
		return
	}
	return
}

/*
symbol	str	symbol='000300'; 股票代码
start_date	str	start_date="1979-09-01 09:32:00"; 日期时间; 默认返回所有数据
end_date	str	end_date="2222-01-01 09:32:00"; 日期时间; 默认返回所有数据
period	str	period='5'; choice of {'1', '5', '15', '30', '60'}; 其中 1 分钟数据返回近 5 个交易日数据且不复权
adjust	str	adjust=”; choice of {”, 'qfq', 'hfq'}; ”: 不复权, 'qfq': 前复权, 'hfq': 后复权, 其中 1 分钟数据返回近 5 个交易日数据且不复权
*/
func GetStockPriceRT(ctx context.Context, symbol string) (res StockPriceDataRTList, err error) {
	_, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	res = make(StockPriceDataRTList, 0)
	c, _ := client.NewClient()
	req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
	req.SetRequestURI(BaseURL + PublicAPIURI + StockHandlerNameRealtime)
	req.SetMethod("GET")
	req.SetQueryString(fmt.Sprintf("symbol=sh%s", symbol))

	err = c.Do(ctx, req, resp)
	if err != nil {
		return
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("get gold price failed, status code: %d", resp.StatusCode())
	}

	err = sonic.Unmarshal(resp.Body(), &res)
	if err != nil {
		return
	}
	return
}

func GetStockSymbolInfo(ctx context.Context, symbol string) (stockName string, err error) {
	_, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	res := make([]map[string]any, 0)
	c, _ := client.NewClient()
	req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
	req.SetRequestURI(BaseURL + PublicAPIURI + StockSingleInfo)
	req.SetMethod("GET")
	req.SetQueryString(fmt.Sprintf("symbol=%s", symbol))
	err = c.Do(ctx, req, resp)
	if err != nil {
		return
	}
	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("get stock info failed, status code: %d", resp.StatusCode())
	}
	err = sonic.Unmarshal(resp.Body(), &res)
	if err != nil {
		return
	}
	for _, item := range res {
		if item["item"].(string) == "股票简称" {
			return item["value"].(string), nil
		}
	}
	return
}
