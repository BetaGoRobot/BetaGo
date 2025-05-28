package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
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
	argMap, _ := parseArgs(args...)
	log.Zlog.Info("replyHandler", zaplog.Any("args", argMap))
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
		replyType, ok := argMap["reply_type"]
		if !ok {
			replyType = string(consts.ReplyTypeText)
		}

		var reply string

		if replyType == string(consts.ReplyTypeImg) { // 图片类型，需要回复图片
			if data.Event.Message.ParentId == nil {
				return errors.New("reply_type **img** must reply to a image message")
			}
			parentMsg := larkutils.GetMsgFullByID(ctx, *data.Event.Message.ParentId)
			if len(parentMsg.Data.Items) != 0 {
				parentMsgItem := parentMsg.Data.Items[0]
				contentMap := make(map[string]string)
				err := sonic.UnmarshalString(*parentMsgItem.Body.Content, &contentMap)
				if err != nil {
					log.Zlog.Error("repeatMessage", zaplog.Error(err))
					return err
				}
				if *parentMsgItem.MsgType == larkim.MsgTypeSticker {
					imgKey := contentMap["file_key"]
					res, _ := database.FindByCacheFunc(database.StickerMapping{StickerKey: imgKey}, func(r database.StickerMapping) string { return r.StickerKey })
					if len(res) == 0 {
						if stickerFile, err := larkutils.GetMsgImages(ctx, *data.Event.Message.ParentId, contentMap["file_key"], "image"); err != nil {
							log.Zlog.Error("repeatMessage", zaplog.Error(err))
						} else {
							newImgKey := larkutils.UploadPicture2LarkReader(ctx, stickerFile)
							database.GetDbConnection().Clauses(clause.OnConflict{UpdateAll: true}).Create(&database.StickerMapping{
								StickerKey: imgKey,
								ImageKey:   newImgKey,
							})
						}
					}
					reply = imgKey
				} else if *parentMsgItem.MsgType == larkim.MsgTypeImage {
					imageFile, err := larkutils.GetMsgImages(ctx, *data.Event.Message.ParentId, contentMap["image_key"], "image")
					if err != nil {
						return err
					}
					reply = larkutils.UploadPicture2LarkReader(ctx, imageFile)
				} else {
					return errors.New("reply_type **img** must reply to a image message")
				}
			}
		} else {
			reply, ok = argMap["reply"]
			if !ok {
				return errors.New("arg reply is required")
			}
		}

		if result := database.GetDbConnection().
			Create(&database.QuoteReplyMsgCustom{
				GuildID:   *data.Event.Message.ChatId,
				MatchType: consts.WordMatchType(matchType),
				Keyword:   word,
				ReplyNType: database.ReplyNType{
					Reply:     reply,
					ReplyType: consts.ReplyType(replyType),
				},
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
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	argMap, _ := parseArgs(args...)
	log.Zlog.Info("replyGetHandler", zaplog.Any("args", argMap))
	ChatID := *data.Event.Message.ChatId

	lines := make([]map[string]string, 0)
	resListCustom, hitCache := database.FindByCacheFunc(database.QuoteReplyMsgCustom{GuildID: ChatID}, func(r database.QuoteReplyMsgCustom) string { return r.GuildID })
	span.SetAttributes(attribute.Key("hitCache").Bool(hitCache))
	for _, res := range resListCustom {
		if res.GuildID == ChatID {
			if res.ReplyType == consts.ReplyTypeImg {
				if strings.HasPrefix(res.Reply, "img") {
					res.Reply = fmt.Sprintf("![picture](%s)", res.Reply)
				} else {
					res.Reply = fmt.Sprintf("![picture](%s)", getImageKeyByStickerKey(res.Reply))
				}
			}
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
		if res.ReplyType == consts.ReplyTypeImg {
			if strings.HasPrefix(res.Reply, "img") {
				res.Reply = fmt.Sprintf("![picture](%s)", res.Reply)
			} else {
				res.Reply = fmt.Sprintf("![picture](%s)", getImageKeyByStickerKey(res.Reply))
			}
		}
		lines = append(lines, map[string]string{
			"title1": "Global",
			"title2": res.Keyword,
			"title3": res.Reply,
			"title4": string(res.MatchType),
		})
	}
	cardContent := larkutils.NewCardContent(
		ctx,
		larkutils.FourColSheetTemplate,
	).
		AddVariable("title1", "Scope").
		AddVariable("title2", "Keyword").
		AddVariable("title3", "Reply").
		AddVariable("title4", "MatchType").
		AddVariable("table_raw_array_1", lines)

	err := larkutils.ReplyCard(ctx, cardContent, *data.Event.Message.MessageId, "_replyGet", false)
	if err != nil {
		return err
	}
	return nil
}
