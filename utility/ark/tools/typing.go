package tools

import (
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/bytedance/gg/gstd/gsync"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
)

var m *Manager

func init() {
	m = &Manager{FunctionMap: gsync.Map[string, *FunctionCallUnit]{}}
	m.SetWebSearch()
	registerHistorySearch()
}

func M() *Manager {
	return m
}

func registerHistorySearch() {
	params := NewParameters("object").
		AddProperty("keywords", &Property{
			Type:        "array",
			Description: "需要检索的关键词列表,英文逗号隔开",
		}).
		AddProperty("user_id", &Property{
			Type:        "string",
			Description: "用户ID",
		}).
		AddProperty("start_time", &Property{
			Type:        "string",
			Description: "开始时间，格式为YYYY-MM-DD HH:MM:SS",
		}).
		AddProperty("end_time", &Property{
			Type:        "string",
			Description: "结束时间，格式为YYYY-MM-DD HH:MM:SS",
		}).
		AddProperty("top_k", &Property{
			Type:        "number",
			Description: "返回的结果数量",
		}).
		AddRequired("keywords")
	fcu := NewFunctionCallUnit().
		Name(ToolSearchHistory).Desc("根据输入的关键词搜索相关的历史对话记录").
		Params(params).Func(HybridSearch)
	M().Add(fcu)
}

func NewFunctionCallUnit() *FunctionCallUnit {
	return &FunctionCallUnit{}
}

type Manager struct {
	FunctionMap   gsync.Map[string, *FunctionCallUnit]
	WebsearchTool *responses.ResponsesTool
}

type FunctionCallUnit struct {
	FunctionName string
	Description  string
	Parameters   *Parameters
	Function     fcFunc
}

func (h *FunctionCallUnit) Name(name string) *FunctionCallUnit {
	h.FunctionName = name
	return h
}

func (h *FunctionCallUnit) Desc(desc string) *FunctionCallUnit {
	h.Description = desc
	return h
}

func (h *FunctionCallUnit) Params(params *Parameters) *FunctionCallUnit {
	h.Parameters = params
	return h
}

func (h *FunctionCallUnit) Func(f fcFunc) *FunctionCallUnit {
	h.Function = f
	return h
}

func (h *Manager) Add(unit *FunctionCallUnit) *Manager {
	if unit.FunctionName == "" || unit.Description == "" || unit.Function == nil {
		panic("FunctionRegisterHelper: missing field")
	}
	h.FunctionMap.Store(unit.FunctionName, unit)
	return h
}

func (h *Manager) SetWebSearch() *Manager {
	h.WebsearchTool = &responses.ResponsesTool{
		Union: &responses.ResponsesTool_ToolWebSearch{
			ToolWebSearch: &responses.ToolWebSearch{
				Type:  responses.ToolType_web_search,
				Limit: utility.Ptr[int64](10),
			},
		},
	}
	return h
}

func (h *Manager) Get(name string) (*FunctionCallUnit, bool) {
	return h.FunctionMap.Load(name)
}

func (h *Manager) Tools() []*responses.ResponsesTool {
	tools := make([]*responses.ResponsesTool, 0)
	h.FunctionMap.Range(
		func(key string, unit *FunctionCallUnit) bool {
			tools = append(tools, &responses.ResponsesTool{
				Union: &responses.ResponsesTool_ToolFunction{
					ToolFunction: &responses.ToolFunction{
						Name:        unit.FunctionName,
						Type:        responses.ToolType_function,
						Description: utility.Ptr(unit.Description),
						Parameters:  &responses.Bytes{Value: unit.Parameters.JSON()},
					},
				},
			})
			return true
		})
	if h.WebsearchTool != nil {
		tools = append(tools, h.WebsearchTool)
	}
	return tools
}

type Property struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Items       []*Property `json:"items,omitempty"`
}
type Parameters struct {
	Type       string               `json:"type"`
	Properties map[string]*Property `json:"properties"`
	Required   []string             `json:"required,omitempty"`
}

func NewParameters(typ string) *Parameters {
	return &Parameters{
		Type:       typ,
		Properties: make(map[string]*Property),
		Required:   make([]string, 0),
	}
}

func (p *Parameters) JSON() []byte {
	return []byte(utility.MustMashal(p))
}

func (p *Parameters) AddProperty(name string, prop *Property) *Parameters {
	p.Properties[name] = prop
	return p
}

func (p *Parameters) AddRequired(name string) *Parameters {
	p.Required = append(p.Required, name)
	return p
}
