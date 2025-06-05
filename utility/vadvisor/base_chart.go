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

type BaseChartsGraphWithPlayer[X cts.ValidType, Y cts.Numeric] struct {
	Type        string              `json:"type"`
	Player      *PlayerConfig[X, Y] `json:"player,omitempty"`
	Data        *DataUnit[X, Y]     `json:"data"` // ==> 对应specs[0].Data
	OuterRadius float64             `json:"outerRadius"`
	InnerRadius float64             `json:"innerRadius"`
	Label       *LabelConfig        `json:"label,omitempty"`
	Direction   string              `json:"direction"`

	AnimationUpdate *Animation `json:"animationUpdate,omitempty"`
	AnimationEnter  *Animation `json:"animationEnter,omitempty"`
	AnimationExit   *Animation `json:"animationExit,omitempty"`

	valueMap     map[string]*DataUnit[X, Y]                       `json:"-"`
	customTagSet map[X]struct{}                                   `json:"-"`
	playerType   string                                           `json:"-"`
	sortFunc     func(a *ValueUnit[X, Y], b *ValueUnit[X, Y]) int `json:"-"`
}

type Animation struct {
	Bar  []*AnimationBar `json:"bar,omitempty"`
	Axis *AnimationAxis  `json:"axis,omitempty"`
}

type AnimationAxis struct {
	Duration int    `json:"duration"`
	Easing   string `json:"easing"`
}

type AnimationBar struct {
	Type     string               `json:"type,omitempty"`
	Options  *AnimationBarOptions `json:"options,omitempty"`
	Easing   string               `json:"easing"`
	Duration int                  `json:"duration"`
	Channel  []string             `json:"channel,omitempty"`
}

type AnimationBarOptions struct {
	ExcludeChannels []string `json:"excludeChannels,omitempty"`
}

type PlayerConfig[X cts.ValidType, Y cts.Numeric] struct {
	Auto      bool           `json:"auto"`
	Loop      bool           `json:"loop"`
	Alternate bool           `json:"alternate"`
	Interval  int            `json:"interval"`
	Width     int            `json:"width"`
	Position  string         `json:"position"`
	Type      string         `json:"type"`
	Specs     []*Specs[X, Y] `json:"specs,omitempty"`

	Controller *ControllerConf `json:"controller,omitempty"`
}

type ControllerConf struct {
	Backward *ControllerBackward `json:"backward,omitempty"`
	Forward  *ControllerForward  `json:"forward,omitempty"`
	Start    *ControllerStart    `json:"start,omitempty"`
}

type ControllerBackward struct {
	Style struct {
		Size int `json:"size"`
	} `json:"style"`
}

type ControllerForward struct {
	Style struct {
		Size int `json:"size"`
	} `json:"style"`
}

type ControllerStart struct {
	Order    int    `json:"order"`
	Position string `json:"position"`
}

type Specs[X cts.ValidType, Y cts.Numeric] struct {
	Data *DataUnit[X, Y] `json:"data,omitempty"`
}

type DataUnit[X cts.ValidType, Y cts.Numeric] struct {
	ID     string             `json:"id"`
	Values []*ValueUnit[X, Y] `json:"values,omitempty"`
}

type GroupKeyWrap[T any] struct {
	GroupKey string
	Value    T
}
type ValueUnit[X cts.ValidType, Y cts.Numeric] struct {
	XField      X      `json:"xField"`
	YField      Y      `json:"yField"`
	SeriesField string `json:"seriesField"`
}

type LabelConfig struct {
	Visible  bool        `json:"visible"`
	Position string      `json:"position"`
	Line     *LineConfig `json:"line,omitempty"`
}

type LineConfig struct {
	Visible bool `json:"visible"`
}

type Value[Y cts.Numeric] struct {
	YAxis       Y
	SeriesField string
}

func NewBaseChartsGraph[X cts.ValidType, Y cts.Numeric]() *BaseChartsGraphWithPlayer[X, Y] {
	return &BaseChartsGraphWithPlayer[X, Y]{
		Player:      &PlayerConfig[X, Y]{},
		valueMap:    make(map[string]*DataUnit[X, Y]),
		playerType:  "continuous",
		OuterRadius: 0.81,
		InnerRadius: 0.5,
		Label: &LabelConfig{
			Visible:  true,
			Position: "outside",
			Line: &LineConfig{
				Visible: false,
			},
		},
	}
}

func (h *BaseChartsGraphWithPlayer[X, Y]) NewPlayer(data []*Specs[X, Y]) *PlayerConfig[X, Y] {
	return &PlayerConfig[X, Y]{
		Auto:      true,
		Loop:      false,
		Alternate: true,
		Interval:  500,
		Width:     500,
		Position:  "middle",
		Type:      h.playerType,
		Specs:     data,
		Controller: &ControllerConf{
			&ControllerBackward{
				Style: struct {
					Size int "json:\"size\""
				}{
					12,
				},
			},
			&ControllerForward{
				Style: struct {
					Size int "json:\"size\""
				}{
					12,
				},
			},
			&ControllerStart{
				1,
				"end",
			},
		},
	}
}

func (h *BaseChartsGraphWithPlayer[X, Y]) SetContinuous() *BaseChartsGraphWithPlayer[X, Y] {
	h.playerType = "continuous"
	return h
}

func (h *BaseChartsGraphWithPlayer[X, Y]) SetDiscrete() *BaseChartsGraphWithPlayer[X, Y] {
	h.playerType = "discrete"
	return h
}

func (h *BaseChartsGraphWithPlayer[X, Y]) AddData(groupKey string, values ...*ValueUnit[X, Y]) *BaseChartsGraphWithPlayer[X, Y] {
	for _, v := range values {
		if unit, ok := h.valueMap[groupKey]; ok {
			unit.Values = append(unit.Values, v)
			h.valueMap[groupKey] = unit
		} else {
			unit = &DataUnit[X, Y]{
				ID:     "data",
				Values: []*ValueUnit[X, Y]{v},
			}
			h.valueMap[groupKey] = unit
		}
	}
	return h
}

func (h *BaseChartsGraphWithPlayer[X, Y]) AddDataSeq(xAxis X, seq iter.Seq[GroupKeyWrap[*ValueUnit[X, Y]]]) *BaseChartsGraphWithPlayer[X, Y] {
	values := slices.Collect(seq)
	for _, v := range values {
		if unit, ok := h.valueMap[v.GroupKey]; ok {
			unit.Values = append(unit.Values, v.Value)
			h.valueMap[v.GroupKey] = unit
		} else {
			unit = &DataUnit[X, Y]{
				ID:     "data",
				Values: []*ValueUnit[X, Y]{v.Value},
			}
			h.valueMap[v.GroupKey] = unit
		}
	}
	return h
}

func (h *BaseChartsGraphWithPlayer[X, Y]) SetSortFunc(cmp func(a *ValueUnit[X, Y], b *ValueUnit[X, Y]) int) *BaseChartsGraphWithPlayer[X, Y] {
	h.sortFunc = cmp
	return h
}

func (h *BaseChartsGraphWithPlayer[X, Y]) BuildPlayer(ctx context.Context) *BaseChartsGraphWithPlayer[X, Y] {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	// Merge Data
	datas := make([]*Specs[X, Y], 0)
	for _, v := range h.valueMap {
		if h.sortFunc == nil {
			slices.SortFunc(v.Values, func(l, r *ValueUnit[X, Y]) int {
				return cmp.Compare(l.SeriesField, r.SeriesField)
			})
		} else {
			slices.SortFunc(v.Values, h.sortFunc)
		}

		datas = append(datas, &Specs[X, Y]{Data: v})
	}
	slices.SortFunc(datas, func(i, j *Specs[X, Y]) int {
		return cmp.Compare(i.Data.Values[0].XField, j.Data.Values[0].XField)
	})

	if len(datas) > 0 {
		// 构建specs数据
		h.Player = h.NewPlayer(datas)
		h.Data = datas[0].Data
	}
	return h
}
