package larkcards

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/bytedance/sonic"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/mdaverde/jsonpath"
	"go.opentelemetry.io/otel/attribute"
)

var (
	BotOpenID             = "ou_8817f540f718affd21718f415b81597f"
	FullLyricsCardPattern = `{
        "config": {},
        "i18n_elements": {
            "zh_cn": [
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
                            "vertical_spacing": "8px",
                            "background_style": "default",
                            "elements": [
                                {
                                    "tag": "column_set",
                                    "flex_mode": "none",
                                    "background_style": "default",
                                    "columns": [
                                        {
                                            "tag": "column",
                                            "width": "weighted",
                                            "vertical_align": "top",
                                            "elements": [
                                                {
                                                    "tag": "markdown",
                                                    "content": "**🗳工单来源：**\n报事报修",
                                                    "text_align": "left"
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
                                                    "tag": "markdown",
                                                    "content": "**📝工单类型：**\n客服工单",
                                                    "text_align": "left"
                                                }
                                            ],
                                            "weight": 1
                                        }
                                    ],
                                    "margin": "0px 0px 0px 0px"
                                }
                            ],
                            "weight": 1
                        }
                    ],
                    "margin": "16px 0px 0px 0px"
                },
                {
                    "tag": "action",
                    "actions": [
                        {
                            "tag": "button",
                            "text": {
                                "tag": "plain_text",
                                "content": "Jaeger Trace"
                            },
                            "type": "primary_filled",
                            "complex_interaction": true,
                            "width": "default",
                            "size": "medium",
                            "multi_url": {
                                "url": "te",
                                "pc_url": "",
                                "ios_url": "",
                                "android_url": ""
                            }
                        }
                    ]
                }
            ]
        },
        "i18n_header": {
            "zh_cn": {
                "title": {
                    "tag": "plain_text",
                    "content": "示例标题"
                },
                "subtitle": {
                    "tag": "plain_text",
                    "content": "示例文本"
                },
                "template": "blue"
            }
        }
    }`
	MusicCardPattern = `{
        "config": {"update_multi":true},
        "i18n_elements": {
            "zh_cn": [
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
                            "vertical_spacing": "8px",
                            "background_style": "default",
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
                                        "url": "music_url",
                                        "pc_url": "",
                                        "ios_url": "",
                                        "android_url": ""
                                    }
                                },
                                {
                                    "tag": "markdown",
                                    "content": "Lyrics",
                                    "text_align": "center",
                                    "text_size": "notation"
                                },
                                {
                                    "tag": "markdown",
                                    "content": "",
                                    "text_align": "left",
                                    "text_size": "normal"
                                }
                            ],
                            "weight": 1
                        },
                        {
                            "tag": "column",
                            "width": "weighted",
                            "vertical_align": "top",
                            "vertical_spacing": "8px",
                            "background_style": "default",
                            "elements": [
                                {
                                    "tag": "img",
                                    "img_key": "img_v3_02ae_dd4a1dac-f6fd-4c4d-a701-0b87d8eb36ag",
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
                        }
                    ],
                    "margin": "16px 0px 0px 0px"
                },
                {
                    "tag": "action",
                    "layout": "default",
                    "actions": [
                        {
                            "tag": "button",
                            "text": {
                                "tag": "plain_text",
                                "content": "完整歌词"
                            },
                            "type": "default",
                            "complex_interaction": true,
                            "width": "default",
                            "size": "medium",
                            "value": "last_page_x",
                            "disabled": false
                        }
                    ]
                },
                {
                    "tag": "action",
                    "actions": [
                        {
                            "tag": "button",
                            "text": {
                                "tag": "plain_text",
                                "content": "Jaeger Trace"
                            },
                            "type": "primary_filled",
                            "complex_interaction": true,
                            "width": "default",
                            "size": "tiny",
                            "multi_url": {
                                "url": "test",
                                "pc_url": "",
                                "ios_url": "",
                                "android_url": ""
                            }
                        }
                    ]
                }
            ]
        },
        "i18n_header": {
            "zh_cn": {
                "title": {
                    "tag": "plain_text",
                    "content": "示例标题"
                },
                "subtitle": {
                    "tag": "plain_text",
                    "content": "示例文本"
                },
                "template": "blue"
            }
        }
    }`
)

func GenFullLyricsCard(ctx context.Context, title, artist, leftLyrics, rightLyrics string) string {
	ctx, span := jaeger_client.LarkRobotTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	var jsonData interface{}
	err := sonic.UnmarshalString(FullLyricsCardPattern, &jsonData)
	if err != nil {
		log.Println(err.Error())
	}
	jsonpath.Set(jsonData, "i18n_elements.zh_cn[0].columns[0].elements[0].columns[0].elements[0].content", leftLyrics)
	jsonpath.Set(jsonData, "i18n_elements.zh_cn[0].columns[0].elements[0].columns[1].elements[0].content", rightLyrics)

	jsonpath.Set(jsonData, "i18n_header.zh_cn.title.content", title)
	jsonpath.Set(jsonData, "i18n_header.zh_cn.subtitle.content", artist)
	jsonpath.Set(jsonData, "i18n_elements.zh_cn[1].actions[0].text.content", "Jaeger Tracer - "+span.SpanContext().TraceID().String())
	jsonpath.Set(jsonData, "i18n_elements.zh_cn[1].actions[0].multi_url.url", "https://jaeger.kevinmatt.top/trace/"+span.SpanContext().TraceID().String())

	var s string
	s, err = sonic.MarshalString(jsonData)
	if err != nil {
		log.Println(err)
	}
	return s
}

func GenMusicTitle(title, artist string) string {
	return fmt.Sprintf("**%s** - **%s**", title, artist)
}

func GenerateMusicCardByStruct(ctx context.Context, imgKey, title, artist, playURL, lyrics, musicID string) string {
	ctx, span := jaeger_client.LarkRobotTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

	var jsonData interface{}
	err := sonic.UnmarshalString(MusicCardPattern, &jsonData)
	if err != nil {
		log.Println(err.Error())
	}
	jsonpath.Set(jsonData, "i18n_elements.zh_cn[0].columns[1].elements[0].img_key", imgKey)
	jsonpath.Set(jsonData, "i18n_header.zh_cn.title.content", title)
	jsonpath.Set(jsonData, "i18n_header.zh_cn.subtitle.content", artist)
	jsonpath.Set(jsonData, "i18n_elements.zh_cn[0].columns[0].elements[0].multi_url.url", playURL)
	jsonpath.Set(jsonData, "i18n_elements.zh_cn[0].columns[0].elements[1].content", lyrics)
	jsonpath.Set(jsonData, "i18n_elements.zh_cn[1].actions[0].value", map[string]interface{}{"music_id": musicID})
	jsonpath.Set(jsonData, "i18n_elements.zh_cn[2].actions[0].text.content", "Jaeger Tracer - "+span.SpanContext().TraceID().String())
	jsonpath.Set(jsonData, "i18n_elements.zh_cn[2].actions[0].multi_url.url", "https://jaeger.kevinmatt.top/trace/"+span.SpanContext().TraceID().String())

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
		if *mention.Id.OpenId == BotOpenID {
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
