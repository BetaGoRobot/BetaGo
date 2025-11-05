package handlers

import (
	"context"
	"fmt"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/dal/lark"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility/copywriting"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkconsts"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkimg"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
func ImageAddHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	defer func() { span.RecordError(err) }()

	argMap, _ := parseArgs(args...)
	logs.L.Info(ctx, "word add handler", "args", argMap)
	if len(argMap) > 0 {
		var imgKey string
		// by url
		picURL, ok := argMap["url"]
		if ok {
			imgKey = larkimg.UploadPicture2Lark(ctx, picURL)
		}
		// by img_key
		inputImgKey, ok := argMap["img_key"]
		if ok {
			imgKey = inputImgKey
		}
		err := createImage(ctx, *data.Event.Message.MessageId, *data.Event.Message.ChatId, imgKey, consts.LarkResourceTypeImage)
		if err != nil {
			return err
		}

	} else if data.Event.Message.ThreadId != nil {
		// 找到话题中的所有图片
		var combinedErr error
		resp, err := lark.LarkClient.Im.Message.List(ctx, larkim.NewListMessageReqBuilder().ContainerIdType("thread").ContainerId(*data.Event.Message.ThreadId).Build())
		if err != nil {
			return err
		}
		for _, msg := range resp.Data.Items {
			if *msg.Sender.Id != larkconsts.BotAppID {
				if imgKey := getImageKey(msg); imgKey != "" {
					err := createImage(ctx, *msg.MessageId, *msg.ChatId, imgKey, *msg.MsgType)
					if err != nil {
						if combinedErr == nil {
							span.RecordError(err)
							combinedErr = err
						} else {
							combinedErr = errors.Wrapf(combinedErr, "%v", err)
						}
					} else {
						larkutils.AddReactionAsync(ctx, "JIAYI", *msg.MessageId)
					}
				}
			}
		}
		if combinedErr != nil {
			span.SetStatus(codes.Error, "addImage not complete with some error")
			return errors.New("addImage not complete with some error")
		}
	} else if data.Event.Message.ParentId != nil {
		parentMsg := larkutils.GetMsgFullByID(ctx, *data.Event.Message.ParentId)
		if len(parentMsg.Data.Items) != 0 {
			msg := parentMsg.Data.Items[0]
			imgKey := getImageKey(msg)
			err := createImage(ctx, *msg.MessageId, *data.Event.Message.ChatId, imgKey, *msg.MsgType)
			if err != nil {
				span.SetStatus(codes.Error, "addImage not complete with some error")
				span.RecordError(err)
				return err
			}
			larkutils.AddReactionAsync(ctx, "JIAYI", *msg.MessageId)
		} else {
			return errors.New(copywriting.GetSampleCopyWritings(ctx, *data.Event.Message.ChatId, copywriting.ImgQuoteNoParent))
		}
	} else {
		return errors.New(copywriting.GetSampleCopyWritings(ctx, *data.Event.Message.ChatId, copywriting.ImgNotAnyValidArgs))
	}
	// successCopywriting := copywriting.GetSampleCopyWritings(ctx, *data.Event.Message.ChatId, copywriting.ImgAddRespAddSuccess)
	larkutils.AddReactionAsync(ctx, "DONE", *data.Event.Message.MessageId)
	// larkutils.ReplyMsgText(ctx, successCopywriting, *data.Event.Message.MessageId, "_imgAdd", false)
	return nil
}

// ImageGetHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
func ImageGetHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	defer func() { span.RecordError(err) }()
	argMap, _ := parseArgs(args...)
	logs.L.Info(ctx, "reply get handler", "args", argMap)
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
	cardContent := templates.NewCardContent(
		ctx,
		templates.TwoColSheetTemplate,
	).
		AddVariable("title1", "Type").
		AddVariable("title2", "Picture").
		AddVariable("table_raw_array_1", lines)

	err = larkutils.ReplyCard(ctx, cardContent, *data.Event.Message.MessageId, "_replyGet", false)
	if err != nil {
		return err
	}
	return nil
}

// ImageDelHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
func ImageDelHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	defer func() { span.RecordError(err) }()
	defer span.RecordError(err)

	argMap, _ := parseArgs(args...)
	logs.L.Info(ctx, "reply del handler", "args", argMap)

	if data.Event.Message.ThreadId != nil {
		// 找到话题中的所有图片
		var combinedErr error
		resp, err := lark.LarkClient.Im.Message.List(ctx, larkim.NewListMessageReqBuilder().ContainerIdType("thread").ContainerId(*data.Event.Message.ThreadId).Build())
		if err != nil {
			return err
		}
		for _, msg := range resp.Data.Items {
			if imgKey := getImageKey(msg); imgKey != "" {
				err = deleteImage(ctx, *msg.MessageId, *msg.ChatId, imgKey, *msg.MsgType)
				if err != nil {
					span.RecordError(err)
					if combinedErr == nil {
						combinedErr = err
					} else {
						combinedErr = errors.Wrapf(combinedErr, "%v", err)
					}
				} else {
					larkutils.AddReactionAsync(ctx, "GeneralDoNotDisturb", *msg.MessageId)
				}
			}
		}
		if combinedErr != nil {
			span.SetStatus(codes.Error, "delImage not complete with some error")
			return errors.New("delImage not complete with some error")
		}
	} else if data.Event.Message.ParentId != nil {
		parentMsgResp := larkutils.GetMsgFullByID(ctx, *data.Event.Message.ParentId)
		if len(parentMsgResp.Data.Items) != 0 {
			msg := parentMsgResp.Data.Items[0]
			if *msg.Sender.Id == larkconsts.BotAppID {
				if imgKey := getImageKey(msg); imgKey != "" {
					err := deleteImage(ctx, *msg.MessageId, *msg.ChatId, imgKey, *msg.MsgType)
					if err != nil {
						span.SetStatus(codes.Error, "delImage not complete with some error")
						span.RecordError(err)
						return err
					}
					larkutils.AddReactionAsync(ctx, "GeneralDoNotDisturb", *msg.MessageId)
				}
			}

		}

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

func getImageKey(msg *larkim.Message) string {
	if *msg.MsgType == larkim.MsgTypeSticker || *msg.MsgType == larkim.MsgTypeImage {
		contentMap := make(map[string]string)
		err := sonic.UnmarshalString(*msg.Body.Content, &contentMap)
		if err != nil {
			logs.L.Error(context.Background(), "repeat message error", "error", err)
			return ""
		}
		switch *msg.MsgType {
		case larkim.MsgTypeImage:
			return contentMap["image_key"]
		case larkim.MsgTypeSticker:
			return contentMap["file_key"]
		default:
			return ""
		}
	}
	return ""
}

func deleteImage(ctx context.Context, msgID, chatID, imgKey, msgType string) error {
	switch msgType {
	case "image":
		// 检查存在性
		if result := database.GetDbConnection().
			Delete(&database.ReactImageMeterial{GuildID: chatID, FileID: imgKey, Type: consts.LarkResourceTypeImage}); result.Error != nil {
			return result.Error
		} else {
			if result.RowsAffected == 0 {
				return fmt.Errorf("img_key %s not exists\n", imgKey)
			}
		}
	case "sticker":
		// 表情包为全局file_key，可以直接存下
		if result := database.GetDbConnection().
			Delete(&database.ReactImageMeterial{GuildID: chatID, FileID: imgKey, Type: consts.LarkResourceTypeSticker}); result.Error != nil {
			return result.Error
		} else {
			if result.RowsAffected == 0 {
				return fmt.Errorf("img_key %s not exists\n", imgKey)
			}
		}
	default:
		// do nothing
	}
	return nil
}

func createImage(ctx context.Context, msgID, chatID, imgKey, msgType string) error {
	switch msgType {
	case "image":
		// 检查存在性
		if result := database.GetDbConnection().Clauses(clause.OnConflict{DoNothing: true}).
			Create(&database.ReactImageMeterial{GuildID: chatID, FileID: imgKey, Type: consts.LarkResourceTypeImage}); result.Error != nil {
			return result.Error
		} else {
			if result.RowsAffected == 0 {
				return errors.New(copywriting.GetSampleCopyWritings(ctx, chatID, copywriting.ImgAddRespAlreadyAdd))
			}
		}
	case "sticker":
		// 表情包为全局file_key，可以直接存下
		if result := database.GetDbConnection().Clauses(clause.OnConflict{DoNothing: true}).
			Create(&database.ReactImageMeterial{GuildID: chatID, FileID: imgKey, Type: consts.LarkResourceTypeSticker}); result.Error != nil {
			return result.Error
		} else {
			if result.RowsAffected == 0 {
				return errors.New(copywriting.GetSampleCopyWritings(ctx, chatID, copywriting.ImgAddRespAlreadyAdd))
			}
		}
	default:
		// do nothing
	}
	return nil
}
