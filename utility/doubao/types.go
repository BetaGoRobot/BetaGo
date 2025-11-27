package doubao

import (
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
)

type FunctionCallMeta struct {
	ChatID string
	UserID string
}

type ReplyUnit struct {
	ID      string
	Content string
}

const (
	ToolSearchHistory = "search_history"
)

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

type SearchArgs struct {
	Keywords  []string `json:"keywords"`
	TopK      int      `json:"top_k"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	UserID    string   `json:"user_id"`
}

func (p *Parameters) JSON() []byte {
	return []byte(utility.MustMashal(p))
}

func Tools() []*responses.ResponsesTool {
	p := &Parameters{
		Type: "object",
		Properties: map[string]*Property{
			"keywords": {
				Type:        "array",
				Description: "需要检索的关键词列表",
				Items: []*Property{
					{
						Type:        "string",
						Description: "关键词",
					},
				},
			},
			"user_id": {
				Type:        "string",
				Description: "用户ID",
			},
			"start_time": {
				Type:        "string",
				Description: "开始时间，格式为YYYY-MM-DD HH:MM:SS",
			},
			"end_time": {
				Type:        "string",
				Description: "结束时间，格式为YYYY-MM-DD HH:MM:SS",
			},
			"top_k": {
				Type:        "number",
				Description: "返回的结果数量",
			},
		},
		Required: []string{"keywords"},
	}
	return []*responses.ResponsesTool{
		{
			Union: &responses.ResponsesTool_ToolWebSearch{
				ToolWebSearch: &responses.ToolWebSearch{
					Type:  responses.ToolType_web_search,
					Limit: utility.Ptr[int64](10),
				},
			},
		},
		{
			Union: &responses.ResponsesTool_ToolFunction{
				ToolFunction: &responses.ToolFunction{
					Name:        ToolSearchHistory,
					Type:        responses.ToolType_function,
					Description: utility.Ptr("根据输入的关键词搜索相关的历史对话记录"),
					Parameters:  &responses.Bytes{Value: p.JSON()},
				},
			},
		},
	}
}
