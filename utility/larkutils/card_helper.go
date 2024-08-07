package larkutils

import "github.com/bytedance/sonic"

type TemplateStru struct {
	TemplateID      string
	TemplateVersion string
}

var (
	FourColSheetTemplate     = TemplateStru{"AAq0LWXpn9FbS", "1.0.0"}
	ThreeColSheetTemplate    = TemplateStru{"AAq0LIyUeFhNX", "1.0.2"}
	TwoColSheetTemplate      = TemplateStru{"AAq0LPliGGphg", "1.0.2"}
	TwoColPicTemplate        = TemplateStru{"AAq0LPJqOoh3s", "1.0.0"}
	AlbumListTemplate        = TemplateStru{"AAq0bN2vGqhvl", "1.0.11"}
	SingleSongDetailTemplate = TemplateStru{"AAqke9FChxpYj", "1.0.1"}
	FullLyricsTemplate       = TemplateStru{"AAq3mcb9ivduh", "1.0.4"}
)

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
