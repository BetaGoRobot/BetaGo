package vadvisor

import (
	"github.com/bytedance/sonic"
	"golang.org/x/exp/constraints"
)

type Numeric interface {
	constraints.Integer | constraints.Float | string
}

type MultiSeriesLineGraph[X, Y Numeric] struct {
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
	Axes        []*AxesStruct[Y]  `json:"axes"`
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
}
type AxesStruct[Y Numeric] struct {
	Orient    string `json:"orient"`
	AliasName string `json:"_alias_name"`
	Range     struct {
		Min Y `json:"min"`
		Max Y `json:"max"`
	} `json:"range"`
}

type DataStruct[X, Y Numeric] struct {
	Values []*DataValue[X, Y] `json:"values"`
}

type DataValue[X, Y Numeric] struct {
	SeriesField string `json:"seriesField"`
	XField      X      `json:"xField"`
	YField      Y      `json:"yField"`
}

const (
	SeriesField = "seriesField"
	XField      = "xField"
	YField      = "yField"
)

func NewMultiSeriesLineGraph[X, Y Numeric]() *MultiSeriesLineGraph[X, Y] {
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
		},
		Data: &DataStruct[X, Y]{
			make([]*DataValue[X, Y], 0),
		},
		XField:      XField,
		YField:      YField,
		SeriesField: SeriesField,
		InvalidType: "link",
		Axes:        make([]*AxesStruct[Y], 0),
	}
}

func (g *MultiSeriesLineGraph[X, Y]) AddData(x X, y Y, seriesField string) {
	g.Data.Values = append(g.Data.Values, &DataValue[X, Y]{
		YField:      y,
		XField:      x,
		SeriesField: seriesField,
	})
}

func (g *MultiSeriesLineGraph[X, Y]) SetTitle(title string) {
	g.Title.Text = title
}

func (g *MultiSeriesLineGraph[X, Y]) SetRange(min, max Y) {
	g.Axes = append(g.Axes, &AxesStruct[Y]{
		Orient:    "left",
		AliasName: "yAxis",
		Range: struct {
			Min Y `json:"min"`
			Max Y `json:"max"`
		}{
			Min: min,
			Max: max,
		},
	})
}

func (g *MultiSeriesLineGraph[X, Y]) String() string {
	s, _ := sonic.MarshalString(g)
	return s
}
