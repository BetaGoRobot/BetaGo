package cal

import (
	"context"
	"testing"
)

func TestDrawPieChartWithAPI(t *testing.T) {
	DrawPieChartWithAPI(context.Background(), GetUserChannelTimeMap(context.Background(), "938697103", "3757937292559087"), "KevinMatt")
}
