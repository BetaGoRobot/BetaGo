package templates

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
)

type TemplateStru struct {
	TemplateID      string
	TemplateVersion string
}

var (
	FourColSheetTemplate         = database.TemplateVersion{TemplateID: "AAq0LWXpn9FbS"}
	ThreeColSheetTemplate        = database.TemplateVersion{TemplateID: "AAq0LIyUeFhNX"}
	TwoColSheetTemplate          = database.TemplateVersion{TemplateID: "AAq0LPliGGphg"}
	TwoColPicTemplate            = database.TemplateVersion{TemplateID: "AAq0LPJqOoh3s"}
	AlbumListTemplate            = database.TemplateVersion{TemplateID: "AAqdqaEBaxJaf"}
	SingleSongDetailTemplate     = database.TemplateVersion{TemplateID: "AAqdrtjg8g1s8"}
	FullLyricsTemplate           = database.TemplateVersion{TemplateID: "AAq3mcb9ivduh"}
	StreamingReasonTemplate      = database.TemplateVersion{TemplateID: "ONLY_SRC_STERAMING_CARD"}
	NormalCardReplyTemplate      = database.TemplateVersion{TemplateID: "AAqRQtNPSJbsZ"}
	NormalCardGraphReplyTemplate = database.TemplateVersion{TemplateID: "AAqdmx3wt8mit"}
	ChunkMetaTemplate            = database.TemplateVersionV2[ChunkMetaData]{TemplateVersion: database.TemplateVersion{TemplateID: "AAqxfVYYV3Zcr"}}
	WordCountTemplate            = database.TemplateVersionV2[WordCountCardVars]{TemplateVersion: database.TemplateVersion{TemplateID: "AAqx4lG1w3cug"}}
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

func GetTemplateV2[T any](template database.TemplateVersionV2[T]) database.TemplateVersionV2[T] {
	templates, _ := database.FindByCacheFunc(template.TemplateVersion, func(tpl database.TemplateVersion) string {
		return tpl.TemplateID
	})
	if len(templates) > 0 {
		template.TemplateVersion = templates[0]
		return template
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
		TemplateSrc         string                 `json:"template_src"`
	}
)

func NewCardContent(ctx context.Context, template database.TemplateVersion) *TemplateCardContent {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	traceID := span.SpanContext().TraceID().String()
	templateVersion := GetTemplate(template)
	var t *TemplateCardContent
	// 纯template
	t = &TemplateCardContent{
		Type: "template",
		Data: CardData{
			TemplateID:          templateVersion.TemplateID,
			TemplateVersionName: templateVersion.TemplateVersion,
			TemplateVariable:    make(map[string]interface{}),
		},
	}
	if templateVersion.TemplateSrc != "" {
		t.Data.TemplateSrc = templateVersion.TemplateSrc
	}

	// default参数
	t.AddJaegerTraceInfo(traceID)
	t.AddVariable("withdraw_info", "撤回卡片")
	t.AddVariable("withdraw_title", "撤回本条消息")
	t.AddVariable("withdraw_confirm", "你确定要撤回这条消息吗？")
	t.AddVariable("withdraw_object", map[string]string{"type": "withdraw"})
	if srcCmd := ctx.Value(consts.ContextVarSrcCmd); srcCmd != nil {
		t.AddVariable("raw_cmd", srcCmd.(string))
		t.AddVariable("refresh_obj", map[string]string{"type": "refresh_obj", "command": srcCmd.(string)})
	}
	t.AddVariable("refresh_time", time.Now().In(utility.UTCPlus8Loc()).Format(time.DateTime))
	return t
}

func NewCardContentV2[T any](ctx context.Context, template database.TemplateVersionV2[T]) *TemplateCardContent {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	traceID := span.SpanContext().TraceID().String()
	templateVersion := GetTemplateV2(template)
	var t *TemplateCardContent
	// 纯template
	t = &TemplateCardContent{
		Type: "template",
		Data: CardData{
			TemplateID:          templateVersion.TemplateID,
			TemplateVersionName: templateVersion.TemplateVersion.TemplateVersion,
			TemplateVariable:    make(map[string]interface{}),
		},
	}
	if templateVersion.TemplateSrc != "" {
		t.Data.TemplateSrc = templateVersion.TemplateSrc
	}

	// default参数
	v := CardBaseVars{
		JaegerTraceInfo: "Trace",
		JaegerTraceURL:  "https://jaeger.kmhomelab.cn/trace/" + traceID,
		WithdrawInfo:    "撤回卡片",
		WithdrawTitle:   "撤回本条消息",
		WithdrawConfirm: "你确定要撤回这条消息吗？",
		WithdrawObject:  WithDrawObj{Type: "withdraw"},
		RefreshTime:     time.Now().In(utility.UTCPlus8Loc()).Format(time.DateTime),
	}
	if srcCmd := ctx.Value(consts.ContextVarSrcCmd); srcCmd != nil {
		v.RawCmd = utility.Ptr(srcCmd.(string))
		v.RefreshObj = &RefreshObj{Type: "refresh_obj", Command: srcCmd.(string)}
	}
	t.AddVariableStruct(v)

	// 合并
	variables, _ := sonic.Marshal(template.Variables)
	var sourceMap map[string]any
	sonic.Unmarshal(variables, &sourceMap)
	maps.Copy(t.Data.TemplateVariable, sourceMap)
	return t
}

func (c *TemplateCardContent) AddJaegerTraceInfo(traceID string) *TemplateCardContent {
	return c.AddVariable("jaeger_trace_info", "Trace").
		AddVariable("jaeger_trace_url", "https://jaeger.kmhomelab.cn/trace/"+traceID)
}

func (c *TemplateCardContent) AddVariable(key string, value interface{}) *TemplateCardContent {
	c.Data.TemplateVariable[key] = value
	return c
}

func (c *TemplateCardContent) AddVariableStruct(value any) *TemplateCardContent {
	// 结构体转换为map
	variables, _ := sonic.Marshal(value)
	var sourceMap map[string]any
	sonic.Unmarshal(variables, &sourceMap)
	maps.Copy(c.Data.TemplateVariable, sourceMap)
	return c
}

func (c *TemplateCardContent) UpdateVariables(m map[string]interface{}) *TemplateCardContent {
	for k, v := range m {
		c.Data.TemplateVariable[k] = v
	}
	return c
}

func (c *TemplateCardContent) GetVariables() []string {
	s := make([]string, 0, len(c.Data.TemplateVariable))
	for _, v := range c.Data.TemplateVariable {
		s = append(s, fmt.Sprint(v))
	}
	return s
}

func (c *TemplateCardContent) String() string {
	if c == nil {
		return ""
	}
	if c.Data.TemplateSrc == "" {
		res, err := sonic.MarshalString(c)
		if err != nil {
			return ""
		}
		return res
	}
	replacedSrc := c.Data.TemplateSrc
	for k, v := range c.Data.TemplateVariable {
		s, _ := sonic.MarshalString(v)
		s = strings.Trim(s, "\"")
		switch v.(type) {
		case string:
			replacedSrc = strings.ReplaceAll(replacedSrc, "${"+k+"}", s)
		default:
			replacedSrc = strings.ReplaceAll(replacedSrc, "\"${"+k+"}\"", s)
		}
	}
	return replacedSrc
}

func (c *TemplateCardContent) DataString() string {
	if c == nil {
		return ""
	}
	if c.Data.TemplateSrc == "" {
		res, err := sonic.MarshalString(c.Data)
		if err != nil {
			return ""
		}
		return res
	}
	replacedSrc := c.Data.TemplateSrc
	for k, v := range c.Data.TemplateVariable {
		s, _ := sonic.MarshalString(v)
		s = strings.Trim(s, "\"")
		switch v.(type) {
		case string:
			replacedSrc = strings.ReplaceAll(replacedSrc, "${"+k+"}", s)
		default:
			replacedSrc = strings.ReplaceAll(replacedSrc, "\"${"+k+"}\"", s)
		}
	}
	return replacedSrc
}
