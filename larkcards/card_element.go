package larkcards

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type SearchListCard struct {
	Config       struct{} `json:"config"`
	I18NElements struct {
		ZhCn []*ZhcnItem `json:"zh_cn"`
	} `json:"i18n_elements"`
	I18NHeader i18nHeader `json:"i18n_header"`
}

type i18nHeader struct {
	ZhCn struct {
		Title    *i18nHeaderTitle    `json:"title,omitempty"`
		Subtitle *i18nHeaderSubtitle `json:"subtitle,omitempty"`
		Template string              `json:"template,omitempty"`
	} `json:"zh_cn,omitempty"`
}

type i18nHeaderTitle struct {
	Tag     string `json:"tag,omitempty"`
	Content string `json:"content,omitempty"`
}

type i18nHeaderSubtitle struct {
	Tag     string `json:"tag,omitempty"`
	Content string `json:"content,omitempty"`
}

type ZhcnItem struct {
	Tag             string             `json:"tag,omitempty"`
	FlexMode        string             `json:"flex_mode,omitempty"`
	BackgroundStyle string             `json:"background_style,omitempty"`
	Columns         []SearchListColumn `json:"columns,omitempty"`
	Margin          string             `json:"margin,omitempty"`
	Actions         []ActionItem       `json:"actions,omitempty"`
}

type ActionItem struct {
	Tag  string `json:"tag"`
	Text struct {
		Tag     string `json:"tag"`
		Content string `json:"content"`
	} `json:"text"`
	Type               string         `json:"type"`
	ComplexInteraction bool           `json:"complex_interaction"`
	Width              string         `json:"width"`
	Size               string         `json:"size"`
	MultiURL           MultiURLStruct `json:"multi_url"`
}

type MultiURLStruct struct {
	URL        string `json:"url"`
	PcURL      string `json:"pc_url"`
	IosURL     string `json:"ios_url"`
	AndroidURL string `json:"android_url"`
}
type SearchListColumn struct {
	Tag           string              `json:"tag"`
	Width         string              `json:"width"`
	VerticalAlign string              `json:"vertical_align"`
	Elements      []SearchListElement `json:"elements"`
	Weight        int                 `json:"weight,omitempty"`
}
type SearchListElement struct {
	Tag                string            `json:"tag"`
	Content            string            `json:"content,omitempty"`
	TextAlign          string            `json:"text_align,omitempty"`
	TextSize           string            `json:"text_size,omitempty"`
	Text               *Text             `json:"text,omitempty"`
	Type               string            `json:"type,omitempty"`
	ComplexInteraction bool              `json:"complex_interaction,omitempty"`
	Width              string            `json:"width,omitempty"`
	Size               string            `json:"size,omitempty"`
	Value              map[string]string `json:"value,omitempty"`
	ImgKey             string            `json:"img_key,omitempty"`
	ScaleType          string            `json:"scale_type,omitempty"`
	Alt                *Alt              `json:"alt,omitempty"`
	Preview            bool              `json:"preview,omitempty"`
}

type Alt struct {
	Tag     string `json:"tag,omitempty"`
	Content string `json:"content"`
}
type Text struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

func NewSearchListCard() *SearchListCard {
	return &SearchListCard{
		Config: struct{}{},
		I18NElements: struct {
			ZhCn []*ZhcnItem "json:\"zh_cn\""
		}{},
		I18NHeader: struct{}{},
	}
}

func (c *SearchListCard) AddColumn(ctx context.Context, imgKey, title, artist, musicID string) {
	ctx, span := jaeger_client.LarkRobotTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

	columnTextButton := &SearchListColumn{
		Tag:           "column",
		Width:         "weighted",
		VerticalAlign: "top",
		Elements: []SearchListElement{
			{
				Tag:       "markdown",
				Content:   GenMusicTitle(title, artist),
				TextAlign: "left",
				TextSize:  "normal",
			},
			{
				Tag: "button",
				Text: &Text{
					Tag:     "plain_text",
					Content: "选择歌曲",
				},
				Type:               "default",
				ComplexInteraction: true,
				Width:              "default",
				Size:               "medium",
				Value: map[string]string{
					"show_music": musicID,
				},
			},
		},
		Weight: 1,
	}
	columnImg := &SearchListColumn{
		Tag:           "column",
		Width:         "auto",
		VerticalAlign: "top",
		Elements: []SearchListElement{
			{
				Tag:                "img",
				ComplexInteraction: true,
				Size:               "medium",
				Preview:            true,
				ScaleType:          "crop_center",
				ImgKey:             imgKey,
				Alt: &Alt{
					Tag:     "plain_text",
					Content: "",
				},
			},
		},
	}

	c.I18NElements.ZhCn = append(c.I18NElements.ZhCn, &ZhcnItem{
		Tag:             "column_set",
		FlexMode:        "none",
		BackgroundStyle: "default",
		Columns: []SearchListColumn{
			*columnTextButton, *columnImg,
		},
		Margin: "16px 0px 0px 0px",
	})
}

func (c *SearchListCard) AddTitleColumn(ctx context.Context, searchKeyword string) {
	c.I18NHeader = i18nHeader{
		ZhCn: struct {
			Title    *i18nHeaderTitle    "json:\"title,omitempty\""
			Subtitle *i18nHeaderSubtitle "json:\"subtitle,omitempty\""
			Template string              "json:\"template,omitempty\""
		}{
			Title: &i18nHeaderTitle{
				Tag:     "plain_text",
				Content: searchKeyword,
			},
		},
	}
}

func (c *SearchListCard) AddJaegerTracer(ctx context.Context, span trace.Span) {
	ctx, span = jaeger_client.LarkRobotTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	c.I18NElements.ZhCn = append(c.I18NElements.ZhCn, &ZhcnItem{
		Tag: "action",
		Actions: []ActionItem{{
			Tag: "button",
			Text: struct {
				Tag     string "json:\"tag\""
				Content string "json:\"content\""
			}{
				"plain_text", "Jaeger Tracer - " + span.SpanContext().TraceID().String(),
			},
			Type:               "primary_filled",
			ComplexInteraction: true,
			Width:              "default",
			Size:               "tiny",
			MultiURL: MultiURLStruct{
				URL:        "https://jaeger.kmhomelab.cn/trace/" + span.SpanContext().TraceID().String(),
				PcURL:      "",
				IosURL:     "",
				AndroidURL: "",
			},
		}},
	})
}
