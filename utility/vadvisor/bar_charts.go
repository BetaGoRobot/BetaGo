package vadvisor

import (
	"context"
	"iter"

	"github.com/BetaGoRobot/BetaGo/cts"
)

type BarChartsGraphWithPlayer[X cts.ValidType, Y cts.Numeric] struct {
	*BaseChartsGraphWithPlayer[X, Y]

	XField      string `json:"xField"`
	YField      string `json:"yField"`
	SeriesField string `json:"seriesField"`
}

func NewBarChartsGraphWithPlayer[X cts.ValidType, Y cts.Numeric]() *BarChartsGraphWithPlayer[X, Y] {
	base := NewBaseChartsGraph[X, Y]()
	base.Type = "bar"
	return &BarChartsGraphWithPlayer[X, Y]{
		BaseChartsGraphWithPlayer: base,
		SeriesField:               "seriesField",
		XField:                    "xField",
		YField:                    "yField",
	}
}

func (h *BarChartsGraphWithPlayer[X, Y]) SetDirection(direction string) *BarChartsGraphWithPlayer[X, Y] {
	h.Direction = direction
	return h
}

func (b *BarChartsGraphWithPlayer[X, Y]) ReverseAxis() *BarChartsGraphWithPlayer[X, Y] {
	b.XField, b.YField = b.YField, b.XField
	return b
}

func (b *BarChartsGraphWithPlayer[X, Y]) SetContinuous() *BarChartsGraphWithPlayer[X, Y] {
	b.BaseChartsGraphWithPlayer.SetContinuous()
	return b
}

func (b *BarChartsGraphWithPlayer[X, Y]) SetDiscrete() *BarChartsGraphWithPlayer[X, Y] {
	b.BaseChartsGraphWithPlayer.SetDiscrete()
	return b
}

func (b *BarChartsGraphWithPlayer[X, Y]) AddData(groupKey string, values ...*ValueUnit[X, Y]) *BarChartsGraphWithPlayer[X, Y] {
	b.BaseChartsGraphWithPlayer.AddData(groupKey, values...)
	return b
}

func (b *BarChartsGraphWithPlayer[X, Y]) AddDataSeq(xAxis X, seq iter.Seq[GroupKeyWrap[*ValueUnit[X, Y]]]) *BarChartsGraphWithPlayer[X, Y] {
	b.BaseChartsGraphWithPlayer.AddDataSeq(xAxis, seq)
	return b
}

func (b *BarChartsGraphWithPlayer[X, Y]) BuildPlayer(ctx context.Context) *BarChartsGraphWithPlayer[X, Y] {
	b.BaseChartsGraphWithPlayer.BuildWithPlayer(ctx)
	return b
}

func (b *BarChartsGraphWithPlayer[X, Y]) SetSortFunc(cmp func(a *ValueUnit[X, Y], b *ValueUnit[X, Y]) int) *BarChartsGraphWithPlayer[X, Y] {
	b.BaseChartsGraphWithPlayer.SetSortFunc(cmp)
	return b
}
