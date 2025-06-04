package larkutils

import (
	"context"
	"runtime/debug"

	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type panicPatternStruct struct {
	Config       struct{} `json:"config"`
	I18NElements struct {
		ZhCn []struct {
			Tag       string `json:"tag"`
			Content   string `json:"content,omitempty"`
			TextAlign string `json:"text_align,omitempty"`
			TextSize  string `json:"text_size,omitempty"`
			Actions   []struct {
				Tag  string `json:"tag"`
				Text struct {
					Tag     string `json:"tag"`
					Content string `json:"content"`
				} `json:"text"`
				Type               string `json:"type"`
				ComplexInteraction bool   `json:"complex_interaction"`
				Width              string `json:"width"`
				Size               string `json:"size"`
				MultiURL           struct {
					URL        string `json:"url"`
					PcURL      string `json:"pc_url"`
					IosURL     string `json:"ios_url"`
					AndroidURL string `json:"android_url"`
				} `json:"multi_url"`
			} `json:"actions,omitempty"`
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

func newPattern() *panicPatternStruct {
	p := new(panicPatternStruct)
	err := sonic.UnmarshalString(panicCardPattern, p)
	if err != nil {
		panic(err)
	}
	return p
}

var panicCardPattern = `{
    "config": {},
    "i18n_elements": {
        "zh_cn": [
            {
                "tag": "markdown",
                "content": "{stack}",
                "text_align": "left",
                "text_size": "notation"
            },
            {
                "tag": "action",
                "actions": [
                    {
                        "tag": "button",
                        "text": {
                            "tag": "plain_text",
                            "content": "{buttonText}"
                        },
                        "type": "default",
                        "complex_interaction": true,
                        "width": "default",
                        "size": "medium",
                        "multi_url": {
                            "url": "{buttonURL}",
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
                "content": "{title}"
            },
            "subtitle": {
                "tag": "plain_text",
                "content": "{sub_title}"
            },
            "template": "blue"
        }
    }
}`

func RecoverMsg(ctx context.Context, msgID string) {
	if err := recover(); err != nil {
		SendRecoveredMsg(ctx, err, msgID)
	}
}

func RecoverMsgEvent(ctx context.Context, event *larkim.P2MessageReceiveV1) {
	if err := recover(); err != nil {
		SendRecoveredMsg(ctx, err, *event.Event.Message.MessageId)
	}
}

// SendRecoveredMsg  SendRecoveredMsg
//
//	@param ctx
//	@param msgID
//	@param err
func SendRecoveredMsg(ctx context.Context, err any, msgID string) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, "RecoverMsg")
	defer span.End()

	traceID := span.SpanContext().TraceID().String()
	if e, ok := err.(error); ok {
		span.RecordError(e)
	}
	stack := string(debug.Stack())
	log.Zlog.Error("panic-detected!", zaplog.String("trace_id", traceID), zaplog.Any("panic", err), zaplog.String("msg_id", msgID))
	err = ReplyCardText(ctx, "```go\n"+stack+"\n```", msgID, "", true)
	if err != nil {
		log.Zlog.Error("send error", zaplog.Any("error", err))
	}
}

// SendRecoveredMsgUserID to be filled
//
//	@param ctx context.Context
//	@param err any
//	@param chatID string
//	@param userID string
//	@author kevinmatthe
//	@update 2025-06-04 16:30:33
func SendRecoveredMsgUserID(ctx context.Context, err any, chatID, userID string) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, "RecoverMsg")
	defer span.End()

	traceID := span.SpanContext().TraceID().String()
	if e, ok := err.(error); ok {
		span.RecordError(e)
	}
	stack := string(debug.Stack())

	log.Zlog.Error("panic-detected!", zaplog.String("trace_id", traceID), zaplog.Any("panic", err), zaplog.String("chat_id", chatID))
	err = SendCardText(ctx, "```go\n"+stack+"\n```", chatID, "", true)
	if err != nil {
		log.Zlog.Error("send error", zaplog.Any("error", err))
	}
}
