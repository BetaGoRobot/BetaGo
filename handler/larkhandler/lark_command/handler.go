package larkcommand

import (
	"context"
	"errors"
	"strings"

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

const (
	getIDText      = "Quoted Msg OpenID is "
	getGroupIDText = "Current ChatID is "
)

// getIDHandler get ID Handler
//
//	@param ctx
//	@param data
//	@param args
//	@return error
func getIDHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	if data.Event.Message.ParentId == nil {
		return errors.New("No parent Msg Quoted")
	}

	err := larkutils.ReplyMsgText(ctx, getIDText+*data.Event.Message.ParentId, *data.Event.Message.MessageId, "_getID", false)
	if err != nil {
		log.ZapLogger.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
		return err
	}
	return nil
}

// getIDHandler get ID Handler
//
//	@param ctx
//	@param data
//	@param args
//	@return error
func getGroupIDHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	chatID := data.Event.Message.ChatId
	if chatID != nil {
		err := larkutils.ReplyMsgText(ctx, getGroupIDText+*chatID, *data.Event.Message.MessageId, "_getGroupID", false)
		if err != nil {
			log.ZapLogger.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
			return err
		}
	}

	return nil
}

// getIDHandler get ID Handler
//
//	@param ctx
//	@param data
//	@param args
//	@return error
func tryPanicHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	panic("try panic!")
}

func getTraceURLMD(traceID string) string {
	return strings.Join([]string{"[Trace-", traceID[:8], "]", "(https://jaeger.kmhomelab.cn/trace/", traceID, ")"}, "")
}

func getTraceFromMsgID(ctx context.Context, msgID string) ([]string, error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	traceLogs, hitCache := database.FindByCacheFunc(database.MsgTraceLog{MsgID: msgID}, func(d database.MsgTraceLog) string {
		return d.MsgID
	})
	span.SetAttributes(attribute.Bool("MsgTraceLog hitCache", hitCache))
	if len(traceLogs) == 0 {
		return nil, errors.New("No trace log found for the message qouted")
	}
	traceIDs := make([]string, 0)
	for _, traceLog := range traceLogs {
		traceIDs = append(traceIDs, getTraceURLMD(traceLog.TraceID))
	}
	return traceIDs, nil
}

func traceHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	if data.Event.Message.ThreadId != nil { // 话题模式，找到所有的traceID
		resp, err := larkutils.LarkClient.Im.Message.List(ctx, larkim.NewListMessageReqBuilder().ContainerIdType("thread").ContainerId(*data.Event.Message.ThreadId).Build())
		if err != nil {
			return err
		}
		traceIDs := make([]string, 0)
		for _, msg := range resp.Data.Items {
			if *msg.Sender.Id == larkutils.BotAppID {
				traceIDsTmp, err := getTraceFromMsgID(ctx, *msg.MessageId)
				if err != nil {
					return err
				}
				traceIDs = append(traceIDs, traceIDsTmp...)
			}
		}
		traceIDStr := "TraceIDs:\\n" + strings.Join(traceIDs, "\\n")
		err = larkutils.ReplyMsgText(ctx, traceIDStr, *data.Event.Message.MessageId, "_trace", false)
		if err != nil {
			log.ZapLogger.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
			return err
		}
	} else if data.Event.Message.ParentId != nil {
		traceIDs, err := getTraceFromMsgID(ctx, *data.Event.Message.MessageId)
		if err != nil {
			return err
		}
		traceIDStr := "TraceIDs:\\n" + strings.Join(traceIDs, "\\n")
		err = larkutils.ReplyMsgText(ctx, traceIDStr, *data.Event.Message.MessageId, "_trace", false)
		if err != nil {
			log.ZapLogger.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
			return err
		}
	}
	return nil
}

// getIDHandler get ID Handler
//
//	@param ctx
//	@param data
//	@param args
//	@return error
func wordAddHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	if len(args) < 2 {
		return errors.ErrUnsupported
	}
	argMap := parseArgs(args...)
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

func parseArgs(args ...string) map[string]string {
	resMap := make(map[string]string)
	for _, arg := range args {
		if argKV := strings.Split(arg, "="); len(argKV) == 2 {
			resMap[strings.TrimLeft(argKV[0], "--")] = argKV[1]
		}
	}
	return resMap
}

func imageAddHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	var imgKey string
	argMap := parseArgs(args...)
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
				imgKey = contentMap["file_key"]
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

func replyAddHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
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
