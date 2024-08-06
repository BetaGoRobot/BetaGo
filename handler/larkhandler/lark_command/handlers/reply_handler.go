package handlers

import (
	"context"
	"errors"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm/clause"
)

// ReplyAddHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
//	@author heyuhengmatt
//	@update 2024-08-06 08:27:18
func ReplyAddHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	argMap := parseArgs(args...)
	log.ZapLogger.Info("replyHandler", zaplog.Any("args", argMap))
	if len(argMap) > 0 {
		word, ok := argMap["word"]
		if !ok {
			return errors.New("arg word is required")
		}
		matchType, ok := argMap["type"]
		if !ok {
			return errors.New("arg type(substr, regex, full) is required")
		}
		if matchType != string(consts.MatchTypeSubStr) && matchType != string(consts.MatchTypeRegex) && matchType != string(consts.MatchTypeFull) {
			return errors.New("type must be substr, regex or full")
		}
		reply, ok := argMap["reply"]
		if !ok {
			return errors.New("arg reply is required")
		}
		if result := database.GetDbConnection().
			Clauses(clause.OnConflict{UpdateAll: true}).
			Create(&database.QuoteReplyMsgCustom{
				GuildID:   *data.Event.Message.ChatId,
				MatchType: consts.WordMatchType(matchType),
				Reply:     reply,
				Keyword:   word,
			}); result.Error != nil {
			return result.Error
		}
		larkutils.ReplyMsgText(ctx, "回复语句添加成功", *data.Event.Message.MessageId, "_replyAdd", false)
		return nil
	}
	return consts.ErrArgsIncompelete
}

// ReplyGetHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
func ReplyGetHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	argMap := parseArgs(args...)
	log.ZapLogger.Info("replyGetHandler", zaplog.Any("args", argMap))
	ChatID := *data.Event.Message.ChatId

	lines := make([]map[string]string, 0)
	resListCustom, hitCache := database.FindByCacheFunc(database.QuoteReplyMsgCustom{GuildID: ChatID}, func(r database.QuoteReplyMsgCustom) string { return r.GuildID })
	span.SetAttributes(attribute.Key("hitCache").Bool(hitCache))
	for _, res := range resListCustom {
		if res.GuildID == ChatID {
			lines = append(lines, map[string]string{
				"title1": "Custom",
				"title2": res.Keyword,
				"title3": res.Reply,
				"title4": string(res.MatchType),
			})
		}
	}
	resListGlobal, hitCache := database.FindByCacheFunc(database.QuoteReplyMsg{}, func(r database.QuoteReplyMsg) string { return "" })
	span.SetAttributes(attribute.Key("hitCache").Bool(hitCache))
	for _, res := range resListGlobal {
		lines = append(lines, map[string]string{
			"title1": "Global",
			"title2": res.Keyword,
			"title3": res.Reply,
			"title4": string(res.MatchType),
		})
	}

	cardContent := larkutils.NewSheetCardContent(
		larkutils.FourColSheetTemplate.TemplateID,
		larkutils.FourColSheetTemplate.TemplateVersion,
	).
		AddVariable("title1", "Scope").
		AddVariable("title2", "Keyword").
		AddVariable("title3", "Reply").
		AddVariable("title4", "MatchType").
		AddVariable("table_raw_array_1", lines).String()

	err := larkutils.ReplyMsgRawContentType(ctx, *data.Event.Message.MessageId, larkim.MsgTypeInteractive, cardContent, "_replyGet", false)
	if err != nil {
		return err
	}
	return nil
}
