package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/copywriting"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm/clause"
)

// ImageAddHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
//	@author heyuhengmatt
//	@update 2024-08-06 08:27:13
func ImageAddHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	var imgKey string
	argMap, _ := parseArgs(args...)
	log.ZapLogger.Info("wordAddHandler", zaplog.Any("args", argMap))
	if len(argMap) > 0 {
		// by url
		picURL, ok := argMap["url"]
		if ok {
			imgKey = larkutils.UploadPicture2Lark(ctx, picURL)
		}
		// by img_key
		inputImgKey, ok := argMap["img_key"]
		if ok {
			imgKey = inputImgKey
		}
	} else if data.Event.Message.ParentId != nil {
		parentMsg := larkutils.GetMsgFullByID(ctx, *data.Event.Message.ParentId)
		if len(parentMsg.Data.Items) != 0 {
			parentMsgItem := parentMsg.Data.Items[0]
			contentMap := make(map[string]string)
			err := sonic.UnmarshalString(*parentMsgItem.Body.Content, &contentMap)
			if err != nil {
				log.ZapLogger.Error("repeatMessage", zaplog.Error(err))
				return err
			}
			if *parentMsgItem.MsgType == larkim.MsgTypeSticker {
				// 表情包为全局file_key，可以直接存下
				imgKey = contentMap["file_key"] // 其实是StickerKey

				// Save Mapping
				// 检查存在性
				res, _ := database.FindByCacheFunc(database.StickerMapping{StickerKey: imgKey}, func(r database.StickerMapping) string { return r.StickerKey })
				if len(res) == 0 {
					if stickerFile, err := larkutils.GetMsgImages(ctx, *data.Event.Message.ParentId, contentMap["file_key"], "image"); err != nil {
						log.ZapLogger.Error("repeatMessage", zaplog.Error(err))
					} else {
						newImgKey := larkutils.UploadPicture2LarkReader(ctx, stickerFile)
						database.GetDbConnection().Clauses(clause.OnConflict{UpdateAll: true}).Create(&database.StickerMapping{
							StickerKey: imgKey,
							ImageKey:   newImgKey,
						})
					}
				}
				if result := database.GetDbConnection().Clauses(clause.OnConflict{DoNothing: true}).
					Create(&database.ReactImageMeterial{GuildID: *data.Event.Message.ChatId, FileID: imgKey, Type: consts.LarkResourceTypeSticker}); result.Error != nil {
					return err
				} else {
					if result.RowsAffected == 0 {
						return errors.New(copywriting.GetSampleCopyWritings(ctx, *data.Event.Message.ChatId, copywriting.ImgAddRespAlreadyAdd))
					}
				}
			} else if *parentMsgItem.MsgType == larkim.MsgTypeImage {
				imageFile, err := larkutils.GetMsgImages(ctx, *data.Event.Message.ParentId, contentMap["image_key"], "image")
				if err != nil {
					return err
				}
				imgKey = larkutils.UploadPicture2LarkReader(ctx, imageFile)
				if result := database.GetDbConnection().Clauses(clause.OnConflict{DoNothing: true}).
					Create(&database.ReactImageMeterial{GuildID: *data.Event.Message.ChatId, FileID: imgKey, Type: consts.LarkResourceTypeImage}); result.Error != nil {
					return err
				} else {
					if result.RowsAffected == 0 {
						return errors.New(copywriting.GetSampleCopyWritings(ctx, *data.Event.Message.ChatId, copywriting.ImgAddRespAlreadyAdd))
					}
				}

			} else {
				return errors.New(copywriting.GetSampleCopyWritings(ctx, *data.Event.Message.ChatId, copywriting.ImgNotStickerOrIMG))
			}
		} else {
			return errors.New(copywriting.GetSampleCopyWritings(ctx, *data.Event.Message.ChatId, copywriting.ImgQuoteNoParent))
		}
	} else {
		return errors.New(copywriting.GetSampleCopyWritings(ctx, *data.Event.Message.ChatId, copywriting.ImgNotAnyValidArgs))
	}
	successCopywriting := copywriting.GetSampleCopyWritings(ctx, *data.Event.Message.ChatId, copywriting.ImgAddRespAddSuccess)
	larkutils.ReplyMsgText(ctx, successCopywriting, *data.Event.Message.MessageId, "_imgAdd", false)
	return nil
}

// ImageGetHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
func ImageGetHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	argMap, _ := parseArgs(args...)
	log.ZapLogger.Info("replyGetHandler", zaplog.Any("args", argMap))
	ChatID := *data.Event.Message.ChatId

	lines := make([]map[string]string, 0)
	resList, hitCache := database.FindByCacheFunc(database.ReactImageMeterial{GuildID: ChatID}, func(r database.ReactImageMeterial) string { return r.GuildID })
	span.SetAttributes(attribute.Key("hitCache").Bool(hitCache))
	for _, res := range resList {
		if res.GuildID == ChatID {
			lines = append(lines, map[string]string{
				"title1": res.Type,
				"title2": fmt.Sprintf("![picture](%s)", getImageKeyByStickerKey(res.FileID)),
			})
		}
	}
	template := larkutils.GetTemplate(larkutils.TwoColSheetTemplate)
	cardContent := larkutils.NewSheetCardContent(
		ctx,
		template.TemplateID,
		template.TemplateVersion,
	).
		AddVariable("title1", "Type").
		AddVariable("title2", "Picture").
		AddVariable("table_raw_array_1", lines).String()

	err := larkutils.ReplyMsgRawContentType(ctx, *data.Event.Message.MessageId, larkim.MsgTypeInteractive, cardContent, "_replyGet", false)
	if err != nil {
		return err
	}
	return nil
}

func getImageKeyByStickerKey(stickerKey string) string {
	res, _ := database.FindByCacheFunc(database.StickerMapping{StickerKey: stickerKey}, func(r database.StickerMapping) string { return r.StickerKey })
	if len(res) == 0 {
		return stickerKey
	}
	return res[0].ImageKey
}
