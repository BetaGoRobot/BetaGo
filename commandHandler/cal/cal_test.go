package cal

import (
	"testing"
)

func TestDrawPieChartWithAPI(t *testing.T) {
	DrawPieChartWithAPI(GetUserChannelTimeMap("938697103"), "KevinMatt")
}
