package larkutils

import (
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/bytedance/sonic"
)

type TemplateStru struct {
	TemplateID      string
	TemplateVersion string
}

var (
	// 兜底的版本
	FourColSheetTemplate     = database.TemplateVersion{TemplateID: "AAq0LWXpn9FbS", TemplateVersion: "1.0.0"}
	ThreeColSheetTemplate    = database.TemplateVersion{TemplateID: "AAq0LIyUeFhNX", TemplateVersion: "1.0.2"}
	TwoColSheetTemplate      = database.TemplateVersion{TemplateID: "AAq0LPliGGphg", TemplateVersion: "1.0.2"}
	TwoColPicTemplate        = database.TemplateVersion{TemplateID: "AAq0LPJqOoh3s", TemplateVersion: "1.0.0"}
	AlbumListTemplate        = database.TemplateVersion{TemplateID: "AAq0bN2vGqhvl", TemplateVersion: "1.0.14"}
	SingleSongDetailTemplate = database.TemplateVersion{TemplateID: "AAqke9FChxpYj", TemplateVersion: "1.0.3"}
	FullLyricsTemplate       = database.TemplateVersion{TemplateID: "AAq3mcb9ivduh", TemplateVersion: "1.0.4"}
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

func NewSheetCardContent(templateID, templateVersion string) *TemplateCardContent {
	return &TemplateCardContent{
		Type: "template",
		Data: CardData{
			TemplateID:          templateID,
			TemplateVersionName: templateVersion,
			TemplateVariable:    make(map[string]interface{}),
		},
	}
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
