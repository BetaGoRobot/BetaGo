package vadvisor

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/cts"
)

type WordCloudChartsWithPlayer[X cts.ValidType, Y cts.Numeric] struct {
	*BaseChartsGraphWithPlayer[X, Y]

	ValueField string `json:"valueField"`
	NameField  string `json:"nameField"`
}

// func  NewWordCloudChartsGraphWithPlayer
//
//	@update 2025-06-05 19:51:25
func NewWordCloudChartsGraphWithPlayer[X cts.ValidType, Y cts.Numeric]() *WordCloudChartsWithPlayer[X, Y] {
	base := NewBaseChartsGraph[X, Y]()
	base.Type = "wordCloud"
	return &WordCloudChartsWithPlayer[X, Y]{
		BaseChartsGraphWithPlayer: base,
		NameField:                 "xField",
		ValueField:                "yField",
	}
}

func (w *WordCloudChartsWithPlayer[X, Y]) AddData(groupKey string, values ...*ValueUnit[X, Y]) *WordCloudChartsWithPlayer[X, Y] {
	w.BaseChartsGraphWithPlayer.AddData(groupKey, values...)
	return w
}

func (w *WordCloudChartsWithPlayer[X, Y]) BuildPlayer(ctx context.Context) *WordCloudChartsWithPlayer[X, Y] {
	w.BaseChartsGraphWithPlayer.BuildWithPlayer(ctx)
	return w
}

func (w *WordCloudChartsWithPlayer[X, Y]) Build(ctx context.Context) *WordCloudChartsWithPlayer[X, Y] {
	w.BaseChartsGraphWithPlayer.Build(ctx)
	return w
}
