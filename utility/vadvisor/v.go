package vadvisor

import (
	"iter"

	"github.com/BetaGoRobot/BetaGo/cts"
	"github.com/bytedance/sonic"
)

type MultiSeriesLineGraph[X cts.ValidType, Y cts.Numeric] struct {
	Type  string `json:"type"`
	Title struct {
		Text string `json:"text"`
	} `json:"title"`
	Point       *PointConf        `json:"point"`
	Line        *LineConf         `json:"line"`
	Legends     *LegentConf       `json:"legends"`
	Data        *DataStruct[X, Y] `json:"data"`
	XField      string            `json:"xField"`
	YField      string            `json:"yField"`
	SeriesField string            `json:"seriesField"`
	InvalidType string            `json:"invalidType"`
	Axes        []*AxesStruct     `json:"axes"`
	Stack       bool              `json:"stack"`
}
type LineConf struct {
	Style *LineStyle `json:"style"`
}

type LineStyle struct {
	CurveType string `json:"curveType"`
}
type PointConf struct {
	Style *PointStyle `json:"style"`
}
type PointStyle struct {
	Size int `json:"size"`
}
type LegentConf struct {
	Type    string `json:"type"`
	Visible bool   `json:"visible"`
	Orient  string `json:"orient"`
}
type AxesStruct struct {
	Orient    string `json:"orient"`
	AliasName string `json:"_alias_name"`
	Range     struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	} `json:"range"`
}

type PagerStruct struct {
	Type string `json:"type"`
}

type DataStruct[X, Y cts.ValidType] struct {
	Values []*DataValue[X, Y] `json:"values"`
}

type DataValue[X, Y cts.ValidType] struct {
	SeriesField string `json:"seriesField"`
	XField      X      `json:"xField"`
	YField      Y      `json:"yField"`
}

const (
	SeriesField = "seriesField"
	XField      = "xField"
	YField      = "yField"
)

func NewMultiSeriesLineGraph[X cts.ValidType, Y cts.Numeric]() *MultiSeriesLineGraph[X, Y] {
	return &MultiSeriesLineGraph[X, Y]{
		Type: "line",
		Title: struct {
			Text string `json:"text"`
		}{},
		Point: &PointConf{
			&PointStyle{
				Size: 0,
			},
		},
		Line: &LineConf{
			Style: &LineStyle{
				CurveType: "monotone",
			},
		},
		Legends: &LegentConf{
			Type:    "discrete",
			Visible: true,
			Orient:  "right",
		},
		Data: &DataStruct[X, Y]{
			make([]*DataValue[X, Y], 0),
		},
		XField:      XField,
		YField:      YField,
		SeriesField: SeriesField,
		InvalidType: "link",
		Axes:        make([]*AxesStruct, 0),
	}
}

func (g *MultiSeriesLineGraph[X, Y]) AddData(x X, y Y, seriesField string) *MultiSeriesLineGraph[X, Y] {
	g.Data.Values = append(g.Data.Values, &DataValue[X, Y]{
		YField:      y,
		XField:      x,
		SeriesField: seriesField,
	})
	return g
}

func (g *MultiSeriesLineGraph[X, Y]) SetTitle(title string) *MultiSeriesLineGraph[X, Y] {
	g.Title.Text = title
	return g
}

func (g *MultiSeriesLineGraph[X, Y]) SetRange(min, max float64) *MultiSeriesLineGraph[X, Y] {
	g.Axes = append(g.Axes, &AxesStruct{
		Orient:    "left",
		AliasName: "yAxis",
		Range: struct {
			Min float64 `json:"min"`
			Max float64 `json:"max"`
		}{
			Min: min,
			Max: max,
		},
	})
	return g
}

func (g *MultiSeriesLineGraph[X, Y]) SetStack() *MultiSeriesLineGraph[X, Y] {
	g.Stack = true
	return g
}

func (g *MultiSeriesLineGraph[X, Y]) String() string {
	s, _ := sonic.MarshalString(g)
	return s
}

type (
	XYSUnit[X, Y cts.ValidType] struct {
		XField      X
		YField      Y
		SeriesField string
	}
)

func (g *MultiSeriesLineGraph[X, Y]) AddPointSeries(
	pFunc iter.Seq[XYSUnit[X, Y]],
) *MultiSeriesLineGraph[X, Y] {
	var min, max *Y
	for v := range pFunc {
		if min == nil || max == nil {
			min, max = new(Y), new(Y)
			*min, *max = v.YField, v.YField
		}
		if *min > v.YField {
			*min = v.YField
		}
		if *max < v.YField {
			*max = v.YField
		}
		g.AddData(v.XField, v.YField, v.SeriesField)
	}
	if min == nil || max == nil {
		return g
	}
	g.SetRange(float64(*min), float64(*max))
	return g
}
