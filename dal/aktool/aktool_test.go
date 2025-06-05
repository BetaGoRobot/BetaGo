package aktool

import (
	"context"
	"testing"
)

func TestGetRealtimeGoldPrice(t *testing.T) {
	GetRealtimeGoldPrice(context.Background())
}

func TestGetStockPrice(t *testing.T) {
	GetStockPriceRT(context.TODO(), "600988")
}
