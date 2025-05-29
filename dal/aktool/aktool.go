package aktool

import (
	"context"
	"fmt"
	"os"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
)

var (
	BaseURL                 = os.Getenv("AKTOOL_BASE_URL")
	PublicAPIURI            = "/api/public/"
	GoldHandlerNameRealtime = "spot_quotations_sge"
	GoldHandlerNameHistory  = "spot_hist_sge"
)

type (
	GoldPriceDataRTList []*GoldPriceDataRT
	GoldPriceDataRT     struct {
		Kind       string  `json:"品种"`
		Time       string  `json:"时间"`
		Price      float64 `json:"现价"`
		UpdateTime string  `json:"更新时间"`
	}
)

func GetRealtimeGoldPrice(ctx context.Context) (res GoldPriceDataRTList, err error) {
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
