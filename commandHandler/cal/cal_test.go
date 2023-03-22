package cal

import (
	"context"
	"testing"
)

func TestDrawPieChartWithAPI(t *testing.T) {
	DrawPieChartWithAPI(GetUserChannelTimeMap(context.Background(), "938697103"), "KevinMatt")
}
