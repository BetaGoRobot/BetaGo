package handlers

import (
	"context"
	"errors"
	"strconv"

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

// WordAddHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
//	@author heyuhengmatt
//	@update 2024-08-06 08:27:09
func WordAddHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	if len(args) < 2 {
		return errors.ErrUnsupported
	}
	argMap, _ := parseArgs(args...)
	log.ZapLogger.Info("wordAddHandler", zaplog.Any("args", argMap))

	word, ok := argMap["word"]
	if !ok {
		return errors.New("word is required")
	}
	rate, ok := argMap["rate"]
	if !ok {
		return errors.New("rate is required")
	}

	ChatID := *data.Event.Message.ChatId
	return database.GetDbConnection().Debug().Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&database.RepeatWordsRateCustom{
		GuildID: ChatID,
		Word:    word,
		Rate:    utility.MustAtoI(rate),
	}).Error
}

// WordGetHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
//	@author heyuhengmatt
//	@update 2024-08-06 08:27:07
func WordGetHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	argMap, _ := parseArgs(args...)
	log.ZapLogger.Info("wordGetHandler", zaplog.Any("args", argMap))
	ChatID := *data.Event.Message.ChatId

	lines := make([]map[string]string, 0)
	resListCustom, hitCache := database.FindByCacheFunc(database.RepeatWordsRateCustom{GuildID: ChatID}, func(r database.RepeatWordsRateCustom) string { return r.GuildID })
	span.SetAttributes(attribute.Key("hitCache").Bool(hitCache))
	for _, res := range resListCustom {
		if res.GuildID == ChatID {
			lines = append(lines, map[string]string{
				"title1": "Custom",
				"title2": res.Word,
				"title3": strconv.Itoa(res.Rate),
			})
		}
	}
	resListGlobal, hitCache := database.FindByCacheFunc(database.RepeatWordsRate{}, func(r database.RepeatWordsRate) string { return "" })
	span.SetAttributes(attribute.Key("hitCache").Bool(hitCache))
	for _, res := range resListGlobal {
		lines = append(lines, map[string]string{
			"title1": "Global",
			"title2": res.Word,
			"title3": strconv.Itoa(res.Rate),
		})
	}
	template := larkutils.GetTemplate(larkutils.ThreeColSheetTemplate)
	cardContent := larkutils.NewSheetCardContent(
		template.TemplateID,
		template.TemplateVersion,
	).
		AddVariable("title1", "Scope").
		AddVariable("title2", "Keyword").
		AddVariable("title3", "Rate").
		AddVariable("table_raw_array_1", lines).String()

	err := larkutils.ReplyMsgRawContentType(ctx, *data.Event.Message.MessageId, larkim.MsgTypeInteractive, cardContent, "_wordGet", false)
	if err != nil {
		return err
	}
	return nil
}
