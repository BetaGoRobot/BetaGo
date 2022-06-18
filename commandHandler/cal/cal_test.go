package cal

import (
	"testing"
)

func TestGetUserChannelTimeMap(t *testing.T) {
	DrawPieChart(GetUserChannelTimeMap("938697103"), "KevinMatt")
}
