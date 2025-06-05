package vadvisor

import (
	"cmp"
	"context"
	"iter"
	"slices"

	"github.com/BetaGoRobot/BetaGo/cts"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
)

type PieChartsGraph[X cts.ValidType, Y cts.Numeric] struct {
	Type   string                      `json:"type"`
	Player *PieChartPlayerConfig[X, Y] `json:"player,omitempty"`
	Series []*PieChartSerieItem[X, Y]  `json:"series"`

	valueMap   map[X]*PieChartDataUnit[X, Y] `json:"-"`
	playerType string                        `json:"-"`
}

type PieChartPlayerConfig[X cts.ValidType, Y cts.Numeric] struct {
	Auto      bool                   `json:"auto"`
	Loop      bool                   `json:"loop"`
	Alternate bool                   `json:"alternate"`
	Interval  int                    `json:"interval"`
	Width     int                    `json:"width"`
	Position  string                 `json:"position"`
	Type      string                 `json:"type"`
	Specs     []*PieChartSpecs[X, Y] `json:"specs,omitempty"`
}

type PieChartSpecs[X cts.ValidType, Y cts.Numeric] struct {
	Data *PieChartDataUnit[X, Y] `json:"data,omitempty"`
}

type PieChartSerieItem[X cts.ValidType, Y cts.Numeric] struct {
	Type        string                  `json:"type"`
	Data        *PieChartDataUnit[X, Y] `json:"data"` // ==> 对应specs[0].Data
	DataKey     string                  `json:"dataKey"`
	OuterRadius float64                 `json:"outerRadius"`
	InnerRadius float64                 `json:"innerRadius"`
	Label       *PieChartLabelConfig    `json:"label,omitempty"`
	ValueField  string                  `json:"valueField"`
	SeriesField string                  `json:"seriesField"`
}

type PieChartDataUnit[X cts.ValidType, Y cts.Numeric] struct {
	ID     string                     `json:"id"`
	Values []*PieChartValueUnit[X, Y] `json:"values,omitempty"`
}

type PieChartValueUnit[X cts.ValidType, Y cts.Numeric] struct {
	XAxis       X      `json:"xAxis"`
	SeriesField string `json:"seriesField"`
	Value       Y      `json:"value"`
}

type PieChartLabelConfig struct {
	Visible  bool                 `json:"visible"`
	Position string               `json:"position"`
	Line     *PieChartsLineConfig `json:"line,omitempty"`
}

type PieChartsLineConfig struct {
	Visible bool `json:"visible"`
}

type PieChartValue[Y cts.Numeric] struct {
	YAxis       Y
	SeriesField string
}

func NewPieChartsGraph[X cts.ValidType, Y cts.Numeric]() *PieChartsGraph[X, Y] {
	return &PieChartsGraph[X, Y]{
		Type:   "pie",
		Player: &PieChartPlayerConfig[X, Y]{},
		Series: []*PieChartSerieItem[X, Y]{},

		valueMap:   make(map[X]*PieChartDataUnit[X, Y]),
		playerType: "continuous",
	}
}

func (h *PieChartsGraph[X, Y]) NewSerie(data *PieChartDataUnit[X, Y]) *PieChartSerieItem[X, Y] {
	return &PieChartSerieItem[X, Y]{
		Type:        "pie",
		DataKey:     "seriesField",
		Data:        data,
		OuterRadius: 0.81,
		InnerRadius: 0.5,
		Label: &PieChartLabelConfig{
			Visible:  true,
			Position: "outside",
			Line: &PieChartsLineConfig{
				Visible: false,
			},
		},
		ValueField:  "value",
		SeriesField: "seriesField",
	}
}

func (h *PieChartsGraph[X, Y]) SetContinuous() *PieChartsGraph[X, Y] {
	h.playerType = "continuous"
	return h
}

func (h *PieChartsGraph[X, Y]) SetDiscrete() *PieChartsGraph[X, Y] {
	h.playerType = "discrete"
	return h
}

func (h *PieChartsGraph[X, Y]) NewPlayer(data []*PieChartSpecs[X, Y]) *PieChartPlayerConfig[X, Y] {
	return &PieChartPlayerConfig[X, Y]{
		Auto:      true,
		Loop:      false,
		Alternate: true,
		Interval:  500,
		Width:     500,
		Position:  "middle",
		Type:      h.playerType,
		Specs:     data,
	}
}

func (h *PieChartsGraph[X, Y]) AddData(values ...*PieChartValueUnit[X, Y]) *PieChartsGraph[X, Y] {
	for _, v := range values {
		if unit, ok := h.valueMap[v.XAxis]; ok {
			unit.Values = append(unit.Values, v)
			h.valueMap[v.XAxis] = unit
		} else {
			unit = &PieChartDataUnit[X, Y]{
				ID:     "data",
				Values: []*PieChartValueUnit[X, Y]{v},
			}
			h.valueMap[v.XAxis] = unit
		}
	}
	return h
}

func (h *PieChartsGraph[X, Y]) AddDataSeq(xAxis X, seq iter.Seq[*PieChartValueUnit[X, Y]]) *PieChartsGraph[X, Y] {
	values := slices.Collect(seq)
	for _, v := range values {
		if unit, ok := h.valueMap[v.XAxis]; ok {
			unit.Values = append(unit.Values, v)
			h.valueMap[v.XAxis] = unit
		} else {
			unit = &PieChartDataUnit[X, Y]{
				ID:     "data",
				Values: []*PieChartValueUnit[X, Y]{v},
			}
			h.valueMap[v.XAxis] = unit
		}
	}
	return h
}

func (h *PieChartsGraph[X, Y]) BuildPlayer(ctx context.Context) *PieChartsGraph[X, Y] {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	// Merge Data
	datas := make([]*PieChartSpecs[X, Y], 0)
	for _, v := range h.valueMap {
		slices.SortFunc(v.Values, func(l, r *PieChartValueUnit[X, Y]) int {
			return cmp.Compare(l.SeriesField, r.SeriesField)
		})
		datas = append(datas, &PieChartSpecs[X, Y]{Data: v})
	}
	slices.SortFunc(datas, func(i, j *PieChartSpecs[X, Y]) int {
		return cmp.Compare(i.Data.Values[0].XAxis, j.Data.Values[0].XAxis)
	})
	if len(datas) > 0 {
		// 构建specs数据
		h.Player = h.NewPlayer(datas)

		// 构建series数据
		h.Series = append(h.Series,
			h.NewSerie(datas[0].Data),
		)
	}
	return h
}
