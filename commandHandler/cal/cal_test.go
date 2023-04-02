package cal

import (
	"context"
	"testing"
)

func TestDrawPieChartWithAPI(t *testing.T) {
	DrawPieChartWithLocal(context.Background(), GetUserChannelTimeMap(context.Background(), "938697103", "3757937292559087"), "KevinMatt")
}

func BenchmarkDrawPieChartWithLocal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawPieChartWithLocal(context.Background(), GetUserChannelTimeMap(context.Background(), "938697103", "3757937292559087"), "KevinMatt")
	}
}
