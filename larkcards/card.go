package larkcards

import (
	"cmp"
	"fmt"
	"log"
	"strings"

	"github.com/bytedance/sonic"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/mdaverde/jsonpath"
)

type MusicCardMsg struct {
	Config       struct{} `json:"config"`
	I18NElements struct {
		ZhCn []struct {
			Tag               string `json:"tag"`
			FlexMode          string `json:"flex_mode"`
			HorizontalSpacing string `json:"horizontal_spacing"`
			BackgroundStyle   string `json:"background_style"`
			Columns           []struct {
				Tag      string `json:"tag"`
				Elements []struct {
					Tag  string `json:"tag"`
					Text struct {
						Tag       string `json:"tag"`
						Content   string `json:"content"`
						TextSize  string `json:"text_size"`
						TextAlign string `json:"text_align"`
						TextColor string `json:"text_color"`
					} `json:"text"`
				} `json:"elements"`
				Width  string `json:"width"`
				Weight int    `json:"weight"`
			} `json:"columns"`
			HorizontalAlign string `json:"horizontal_align,omitempty"`
			Margin          string `json:"margin,omitempty"`
		} `json:"zh_cn"`
	} `json:"i18n_elements"`
	I18NHeader struct {
		ZhCn struct {
			Title struct {
				Tag     string `json:"tag"`
				Content string `json:"content"`
			} `json:"title"`
			Subtitle struct {
				Tag     string `json:"tag"`
				Content string `json:"content"`
			} `json:"subtitle"`
			Template string `json:"template"`
		} `json:"zh_cn"`
	} `json:"i18n_header"`
}

var (
	botOpenID        = "ou_8817f540f718affd21718f415b81597f"
	MusicCardPattern = `{
    "config": {},
    "i18n_elements": {
        "zh_cn": [
            {
                "tag": "column_set",
                "flex_mode": "none",
                "horizontal_spacing": "default",
                "background_style": "default",
                "columns": [
                    {
                        "tag": "column",
                        "elements": [
                            {
                                "tag": "div",
                                "text": {
                                    "tag": "plain_text",
                                    "content": "",
                                    "text_size": "normal",
                                    "text_align": "left",
                                    "text_color": "default"
                                }
                            }
                        ],
                        "width": "weighted",
                        "weight": 1
                    }
                ]
            },
            {
                "tag": "column_set",
                "flex_mode": "none",
                "background_style": "default",
                "horizontal_spacing": "8px",
                "horizontal_align": "left",
                "columns": [
                    {
                        "tag": "column",
                        "width": "weighted",
                        "vertical_align": "top",
                        "elements": [
                            {
                                "tag": "markdown",
                                "content": "",
                                "text_align": "left",
                                "text_size": "normal"
                            },
                            {
                                "tag": "img",
                                "img_key": "%s",
                                "preview": false,
                                "transparent": false,
                                "scale_type": "fit_horizontal",
                                "alt": {
                                    "tag": "plain_text",
                                    "content": ""
                                },
                                "corner_radius": "70%"
                            }
                        ],
                        "weight": 1
                    },
                    {
                        "tag": "column",
                        "width": "weighted",
                        "vertical_align": "top",
                        "elements": [
                            {
                                "tag": "button",
                                "text": {
                                    "tag": "plain_text",
                                    "content": "Play"
                                },
                                "type": "primary_filled",
                                "complex_interaction": true,
                                "width": "default",
                                "size": "small",
                                "multi_url": {
                                    "url": "%s",
                                    "pc_url": "",
                                    "ios_url": "",
                                    "android_url": ""
                                }
                            },
                            {
                                "tag": "markdown",
                                "content": "%s",
                                "text_align": "left",
                                "text_size": "notation"
                            }
                        ],
                        "weight": 1
                    }
                ],
                "margin": "16px 0px 0px 0px"
            }
        ]
    },
    "i18n_header": {
        "zh_cn": {
            "title": {
                "tag": "plain_text",
                "content": "%s"
            },
            "subtitle": {
                "tag": "plain_text",
                "content": "%s"
            },
            "template": "blue"
        }
    }
}`
)

func GenMusicTitle(title, artist string) string {
	return fmt.Sprintf("**%s** - **%s**", title, artist)
}

func GenerateMusicCardByStruct(imgKey, title, artist, playURL, lyrics string) string {
	var jsonData interface{}
	err := sonic.UnmarshalString(MusicCardPattern, &jsonData)
	if err != nil {
		log.Println(err.Error())
	}
	jsonpath.Set(jsonData, "i18n_elements.zh_cn[1].columns[0].elements[1].img_key", imgKey)
	jsonpath.Set(jsonData, "i18n_header.zh_cn.title.content", title)
	jsonpath.Set(jsonData, "i18n_header.zh_cn.subtitle.content", artist)
	jsonpath.Set(jsonData, "i18n_elements.zh_cn[1].columns[1].elements[0].multi_url.url", playURL)
	jsonpath.Set(jsonData, "i18n_elements.zh_cn[1].columns[1].elements[1].content", lyrics)
	var s string
	s, err = sonic.MarshalString(jsonData)
	if err != nil {
		log.Println(err)
	}
	return s
}

func GenerateMusicCard(imgKey, title, artist, playURL, lyrics string) string {
	return fmt.Sprintf(MusicCardPattern, imgKey, playURL, lyrics, title, artist)
}

func TrimLyrics(lyrics string) string {
	lyricsList := strings.Split(lyrics, "\n")
	for index, lyric := range lyricsList {
		right := strings.Index(lyric, "]")
		lyricsList[index] = lyric[right+1:]
	}
	return strings.Join(lyricsList, "\n")
}

type getItemFunc[T cmp.Ordered] func(item any) T

func defaultCheck[T cmp.Ordered](item any) T {
	return item.(T)
}

func IsMentioned(mentions []*larkim.MentionEvent) bool {
	for _, mention := range mentions {
		if *mention.Id.OpenId == botOpenID {
			return true
		}
	}
	return false
}

// func InSlice[T cmp.Ordered](needCheck T, slice any, f getItemFunc[T]) bool {
// 	if f == nil {
// 		f = defaultCheck[T]
// 	}
// 	for _, item := range slice {
// 		if needCheck == defaultCheck[T](item) {
// 			return true
// 		}
// 	}
// 	return false
// }
