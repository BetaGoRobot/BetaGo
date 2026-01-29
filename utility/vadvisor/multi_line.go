package vadvisor

import (
	"context"
	"iter"

	"github.com/BetaGoRobot/BetaGo/cts"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
)

type MultiSeriesLineGraph[X cts.ValidType, Y cts.Numeric] struct {
	context.Context `json:"-"`
	Type            string            `json:"type"`
	Title           *TitleConf        `json:"title,omitempty"`
	Point           *PointConf        `json:"point,omitempty"`
	Line            *LineConf         `json:"line,omitempty"`
	Legends         *LegentConf       `json:"legends,omitempty"`
	DataZoom        *ZoomConf         `json:"dataZoom,omitempty"`
	Data            *DataStruct[X, Y] `json:"data"`
	XField          string            `json:"xField"`
	YField          string            `json:"yField"`
	SeriesField     string            `json:"seriesField"`
	InvalidType     string            `json:"invalidType"`
	Axes            []*AxesStruct     `json:"axes,omitempty"`
	Stack           bool              `json:"stack"`
}
type ZoomConf struct {
	Orient string `json:"orient"`
}
type TitleConf struct {
	Text string `json:"text"`
}
type LineConf struct {
	Style *LineStyle `json:"style,omitempty"`
}

type LineStyle struct {
	CurveType string `json:"curveType"`
}
type PointConf struct {
	Style *PointStyle `json:"style,omitempty"`
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
	Orient    string    `json:"orient"`
	AliasName string    `json:"_alias_name,omitempty"`
	Range     *AxeRange `json:"range,omitempty"`
	Label     *AxeLabel `json:"label,omitempty"`
}

type AxeRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}
type AxeLabel struct {
	AutoHide   bool `json:"autoHide"`
	AutoRotate bool `json:"autoRotate"`
	AutoLimit  bool `json:"autoLimit"`
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

func NewMultiSeriesLineGraph[X cts.ValidType, Y cts.Numeric](ctx context.Context) *MultiSeriesLineGraph[X, Y] {
	return &MultiSeriesLineGraph[X, Y]{
		Context: ctx,
		Type:    "line",
		Title:   &TitleConf{},
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
		DataZoom: &ZoomConf{
			Orient: "bottom",
		},
		XField:      XField,
		YField:      YField,
		SeriesField: SeriesField,
		InvalidType: "link",
		Axes: []*AxesStruct{
			{
				Orient:    "bottom",
				AliasName: "xAxis",
				Label: &AxeLabel{
					AutoHide:   true,
					AutoRotate: false,
					AutoLimit:  true,
				},
			},
		},
	}
}

func (g *MultiSeriesLineGraph[X, Y]) AddData(x X, y Y, seriesField string) *MultiSeriesLineGraph[X, Y] {
	_, span := otel.BetaGoOtelTracer.Start(g, reflecting.GetCurrentFunc())
	defer span.End()

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
	g.Axes = []*AxesStruct{
		{
			Orient:    "bottom",
			AliasName: "xAxis",
			Label: &AxeLabel{
				AutoHide:   true,
				AutoRotate: false,
				AutoLimit:  true,
			},
		},
		{
			Orient:    "left",
			AliasName: "yAxis",
			Range: &AxeRange{
				Min: min,
				Max: max,
			},
			Label: &AxeLabel{
				AutoHide:   true,
				AutoRotate: false,
				AutoLimit:  true,
			},
		},
	}
	return g
}

func (g *MultiSeriesLineGraph[X, Y]) SetStack() *MultiSeriesLineGraph[X, Y] {
	g.Stack = true
	return g
}

func (g *MultiSeriesLineGraph[X, Y]) String() string {
	_, span := otel.BetaGoOtelTracer.Start(g, reflecting.GetCurrentFunc())
	defer span.End()

	s, _ := sonic.MarshalString(g)
	return s
}

type (
	XYSUnit[X, Y cts.ValidType] struct {
		X X      // X轴
		Y Y      // Y轴
		S string // 序列数据、分组
	}
)

// 一般来讲都是适配Y轴的最大最小值，不要处理X轴
func (g *MultiSeriesLineGraph[X, Y]) UpdateMinMax(yVals ...float64) {
	_, span := otel.BetaGoOtelTracer.Start(g, reflecting.GetCurrentFunc())
	defer span.End()

	// 如约定,Y轴是第二根轴
	if len(g.Axes) > 1 && g.Axes[1].Range != nil {
		lastMin, lastMax := g.Axes[1].Range.Min, g.Axes[1].Range.Max
		for _, val := range yVals {
			if val < lastMin {
				lastMin = val
			}
			if val > lastMax {
				lastMax = val
			}
		}
		g.Axes[1].Range.Min = lastMin
		g.Axes[1].Range.Max = lastMax
	}
}

func (g *MultiSeriesLineGraph[X, Y]) AddPointSeries(
	pFunc iter.Seq[XYSUnit[X, Y]],
) *MultiSeriesLineGraph[X, Y] {
	var min, max *Y
	fName := "_pFunc_" + reflecting.GetFunctionName(pFunc)
	for v := range pFunc {
		_, span := otel.BetaGoOtelTracer.Start(g, fName)
		if min == nil || max == nil {
			min, max = new(Y), new(Y)
			*min, *max = v.Y, v.Y
		}
		if *min > v.Y {
			*min = v.Y
		}
		if *max < v.Y {
			*max = v.Y
		}
		g.AddData(v.X, v.Y, v.S)
		span.End()
	}
	if min == nil || max == nil {
		return g
	}
	g.SetRange(float64(*min), float64(*max))
	return g
}
