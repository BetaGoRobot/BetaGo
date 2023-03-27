package gpt3

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	errorsender "github.com/BetaGoRobot/BetaGo/commandHandler/error_sender"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/kook"
	"github.com/patrickmn/go-cache"
	"go.opentelemetry.io/otel/attribute"
)

var chatCache = cache.New(time.Minute*30, time.Minute*1)

func init() {
	go func() {
		for {
			utility.ZapLogger.Info("Syncing chat cache to db...")
			for authorID, messages := range chatCache.Items() {
				m, err := json.Marshal(messages.Object.([]Message))
				if err != nil {
					errorsender.SendErrorInfo("4988093461275944", "", "", err, context.Background())
				}
				table := utility.GetDbConnection().Table("betago.chat_record_logs")
				res := int64(0)
				if table.Where("author_id = ?", authorID).Count(&res); res == 0 {
					utility.GetDbConnection().
						Table("betago.chat_record_logs").
						Create(&utility.ChatRecordLog{
							AuthorID:  authorID,
							RecordStr: string(m),
						})
				} else {
					table.
						Where("author_id = ?", authorID).
						Update("record_str", string(m))
				}
			}
			time.Sleep(time.Minute * 3)
		}
	}()
}

// ClientHandlerStream 1
//
//	@param ctx
//	@param targetID
//	@param quoteID
//	@param authorID
//	@param args
//	@return err
func ClientHandlerStream(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	// defer span.RecordError(err)
	defer span.End()
	spanID := span.SpanContext().TraceID().String()

	cardMessageDupStruct := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "info",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: emoji.Robot.String() + "GPT来帮你",
						Emoji:   false,
					},
				},
				kook.CardMessageSection{
					Mode: kook.CardMessageSectionModeRight,
					Text: kook.CardMessageElementKMarkdown{
						Content: "",
					},
					Accessory: kook.CardMessageElementButton{
						Theme: kook.CardThemeSecondary,
						Value: "GPTTrace:" + spanID,
						Click: "return-val",
						Text:  emoji.StopSign.String() + "Stop",
					},
				},
				&kook.CardMessageDivider{},
				&kook.CardMessageSection{
					Mode: kook.CardMessageSectionModeRight,
					Text: &kook.CardMessageElementKMarkdown{
						Content: "TraceID: `" + spanID + "`",
					},
					Accessory: kook.CardMessageElementButton{
						Theme: kook.CardThemeSuccess,
						Value: "https://jaeger.kevinmatt.top/trace/" + spanID,
						Click: "link",
						Text:  "链路追踪",
					},
				},
			},
		},
	}
	cardMessageStrDup, err := cardMessageDupStruct.BuildMessage()

	resp, err := betagovar.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: targetID,
				Content:  cardMessageStrDup,
				Quote:    quoteID,
			},
		},
	)
	if err != nil {
		return
	}
	curMsgID := resp.MsgID
	msg := strings.Join(args, " ")

	g := &GPTClient{
		Model: "gpt-3.5-turbo",
		Messages: []Message{{
			Role:    "user",
			Content: msg,
		}},
		Stream:    true,
		AsyncChan: make(chan string),
		StopChan:  make(chan string),
	}
	GPTAsyncMap["GPTTrace:"+spanID] = &g.StopChan
	go func(ctx context.Context, curMsgID, quoteID, spanID string, cardMessageDupStruct kook.CardMessage) {
		ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
		defer span.End()

		lastMsg := ""
		for {
			select {
			case s, open := <-g.AsyncChan:
				if !open {
					updateMessage(curMsgID, quoteID, lastMsg, spanID, cardMessageDupStruct, true)
					g.Messages = append(g.Messages, Message{
						Role:    "assistant",
						Content: lastMsg,
					})
					chatCache.Set(authorID, g.Messages, cache.DefaultExpiration)
					return
				}
				lastMsg += s
			}
			updateMessage(curMsgID, quoteID, lastMsg, spanID, cardMessageDupStruct, false)
		}
	}(ctx, curMsgID, quoteID, spanID, cardMessageDupStruct)
	if chatMsg, ok := chatCache.Get(authorID); ok {
		g.Messages = append(chatMsg.([]Message), g.Messages...)
	} else {
		recordLog := struct {
			RecordStr string `json:"record_str"`
		}{}
		utility.GetDbConnection().Table("betago.chat_record_logs").Find(&recordLog, &utility.ChatRecordLog{
			AuthorID: authorID,
		})
		if recordLog.RecordStr != "" {
			oldMessages := make([]Message, 0)
			err = json.Unmarshal([]byte(recordLog.RecordStr), &oldMessages)
			if err != nil {
				return
			}
			g.Messages = append(oldMessages, g.Messages...)
			chatCache.Set(authorID, g.Messages, cache.DefaultExpiration)
		}
	}
	if err = g.PostWithStream(ctx); err != nil {
		return
	}
	return
}

func updateMessage(curMsgID, quoteID, lastMsg, spanID string, cardMessageDupStruct kook.CardMessage, noButton bool) {
	if noButton {
		cardMessageDupStruct[0].Modules[1] = kook.CardMessageSection{
			Text: kook.CardMessageElementKMarkdown{
				Content: lastMsg,
			},
		}
	} else {
		cardMessageDupStruct[0].Modules[1] = kook.CardMessageSection{
			Mode: kook.CardMessageSectionModeRight,
			Text: kook.CardMessageElementKMarkdown{
				Content: lastMsg,
			},
			Accessory: kook.CardMessageElementButton{
				Theme: kook.CardThemeDanger,
				Value: "GPTTrace:" + spanID,
				Click: "return-val",
				Text:  emoji.StopSign.String() + "Stop",
			},
		}
	}

	betagovar.GlobalSession.MessageUpdate(&kook.MessageUpdate{
		MessageUpdateBase: kook.MessageUpdateBase{
			MsgID:   curMsgID,
			Content: cardMessageDupStruct.MustBuildMessage(),
			Quote:   quoteID,
		},
	})
}

// ClientHandler 1
// ! deprecated
// @param targetID 目标ID
// @param quoteID 引用ID
// @param authorID 发送者ID
// @return err 错误信息
func ClientHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	msg := strings.Join(args, " ")
	res, err := CreateChatCompletion(ctx, msg, authorID)
	if err != nil {
		return
	}
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "info",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: emoji.Robot.String() + "GPT来帮你",
						Emoji:   false,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: res,
					},
				},
				&kook.CardMessageDivider{},
				&kook.CardMessageSection{
					Mode: kook.CardMessageSectionModeRight,
					Text: &kook.CardMessageElementKMarkdown{
						Content: "TraceID: `" + span.SpanContext().TraceID().String() + "`",
					},
					Accessory: kook.CardMessageElementButton{
						Theme: kook.CardThemeSuccess,
						Value: "https://jaeger.kevinmatt.top/trace/" + span.SpanContext().TraceID().String(),
						Click: "link",
						Text:  "链路追踪",
					},
				},
			},
		},
	}.BuildMessage()

	_, err = betagovar.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: targetID,
				Content:  cardMessageStr,
				Quote:    quoteID,
			},
		},
	)
	return
}
