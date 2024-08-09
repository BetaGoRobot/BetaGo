package larkutils

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
)

type TemplateStru struct {
	TemplateID      string
	TemplateVersion string
}

var (
	FourColSheetTemplate     = database.TemplateVersion{TemplateID: "AAq0LWXpn9FbS"}
	ThreeColSheetTemplate    = database.TemplateVersion{TemplateID: "AAq0LIyUeFhNX"}
	TwoColSheetTemplate      = database.TemplateVersion{TemplateID: "AAq0LPliGGphg"}
	TwoColPicTemplate        = database.TemplateVersion{TemplateID: "AAq0LPJqOoh3s"}
	AlbumListTemplate        = database.TemplateVersion{TemplateID: "AAq0bN2vGqhvl"}
	SingleSongDetailTemplate = database.TemplateVersion{TemplateID: "AAqke9FChxpYj"}
	FullLyricsTemplate       = database.TemplateVersion{TemplateID: "AAq3mcb9ivduh"}
)

func GetTemplate(template database.TemplateVersion) database.TemplateVersion {
	templates, _ := database.FindByCacheFunc(template, func(tpl database.TemplateVersion) string {
		return tpl.TemplateID
	})
	if len(templates) > 0 {
		return templates[0]
	}
	return template
}

type (
	TemplateCardContent struct {
		Type string   `json:"type"` // must be template
		Data CardData `json:"data"`
	}
	CardData struct {
		TemplateID          string                 `json:"template_id"`
		TemplateVersionName string                 `json:"template_version_name"`
		TemplateVariable    map[string]interface{} `json:"template_variable"`
	}
)

func NewSheetCardContent(ctx context.Context, templateID, templateVersion string) *TemplateCardContent {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	traceID := span.SpanContext().TraceID().String()
	t := &TemplateCardContent{
		Type: "template",
		Data: CardData{
			TemplateID:          templateID,
			TemplateVersionName: templateVersion,
			TemplateVariable:    make(map[string]interface{}),
		},
	}
	// default参数
	t.AddJaegerTraceInfo(traceID)
	t.AddVariable("withdraw_info", "撤回卡片")
	t.AddVariable("withdraw_title", "撤回本条消息")
	t.AddVariable("withdraw_confirm", "你确定要撤回这条消息吗？")
	t.AddVariable("withdraw_object", map[string]string{"type": "withdraw"})
	return t
}

func (c *TemplateCardContent) AddJaegerTraceInfo(traceID string) *TemplateCardContent {
	return c.AddVariable("jaeger_trace_info", "JaegerID - "+traceID).
		AddVariable("jaeger_trace_url", "https://jaeger.kmhomelab.cn/"+traceID)
}

func (c *TemplateCardContent) AddVariable(key string, value interface{}) *TemplateCardContent {
	c.Data.TemplateVariable[key] = value
	return c
}

func (c *TemplateCardContent) UpdateVariables(m map[string]interface{}) *TemplateCardContent {
	for k, v := range m {
		c.Data.TemplateVariable[k] = v
	}
	return c
}

func (c *TemplateCardContent) String() string {
	res, err := sonic.MarshalString(c)
	if err != nil {
		return ""
	}
	return res
}
