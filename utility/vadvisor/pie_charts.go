package vadvisor

import (
	"context"
	"iter"

	"github.com/BetaGoRobot/BetaGo/cts"
)

type PieChartsGraphWithPlayer[X cts.ValidType, Y cts.Numeric] struct {
	*BaseChartsGraphWithPlayer[X, Y]

	ValueField  string `json:"valueField"`
	SeriesField string `json:"seriesField"`
}

// func  NewPieChartsGraphWithPlayer
//
//	@update 2025-06-05 19:51:25
func NewPieChartsGraphWithPlayer[X cts.ValidType, Y cts.Numeric]() *PieChartsGraphWithPlayer[X, Y] {
	base := NewBaseChartsGraph[X, Y]()
	base.Type = "pie"
	return &PieChartsGraphWithPlayer[X, Y]{
		BaseChartsGraphWithPlayer: base,
		ValueField:                "yField",
		SeriesField:               "seriesField",
	}
}

func (b *PieChartsGraphWithPlayer[X, Y]) SetContinuous() *PieChartsGraphWithPlayer[X, Y] {
	b.BaseChartsGraphWithPlayer.SetContinuous()
	return b
}

func (b *PieChartsGraphWithPlayer[X, Y]) SetDiscrete() *PieChartsGraphWithPlayer[X, Y] {
	b.BaseChartsGraphWithPlayer.SetDiscrete()
	return b
}

func (b *PieChartsGraphWithPlayer[X, Y]) AddData(groupKey string, values ...*ValueUnit[X, Y]) *PieChartsGraphWithPlayer[X, Y] {
	b.BaseChartsGraphWithPlayer.AddData(groupKey, values...)
	return b
}

func (b *PieChartsGraphWithPlayer[X, Y]) AddDataSeq(xAxis X, seq iter.Seq[GroupKeyWrap[*ValueUnit[X, Y]]]) *PieChartsGraphWithPlayer[X, Y] {
	b.BaseChartsGraphWithPlayer.AddDataSeq(xAxis, seq)
	return b
}

func (b *PieChartsGraphWithPlayer[X, Y]) BuildPlayer(ctx context.Context) *PieChartsGraphWithPlayer[X, Y] {
	b.BaseChartsGraphWithPlayer.BuildWithPlayer(ctx)
	return b
}

func (h *PieChartsGraphWithPlayer[X, Y]) SetSortFunc(cmp func(a *ValueUnit[X, Y], b *ValueUnit[X, Y]) int) *PieChartsGraphWithPlayer[X, Y] {
	h.BaseChartsGraphWithPlayer.SetSortFunc(cmp)
	return h
}
