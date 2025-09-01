package message

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkimg"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"gorm.io/gorm/clause"
)

var _ Op = &RepeatMsgOperator{}

// RecordMsgOperator  RepeatMsg Op
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:51
type RecordMsgOperator struct {
	OpBase
}

// PreRun Repeat
//
//	@receiver r *RepeatMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:35
func (r *RecordMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	return
}

// Run Repeat
//
//	@receiver r *RepeatMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:41
func (r *RecordMsgOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	imgSeq, err := larkimg.GetAllImageFromMsgEvent(ctx, event.Event.Message)
	if err != nil {
		return
	}
	if imgSeq != nil {
		for imageKey := range imgSeq {
			err = larkimg.DownImgFromMsgAsync(
				ctx,
				*event.Event.Message.MessageId,
				larkim.MsgTypeImage,
				imageKey,
			)
			if err != nil {
				return err
			}
		}
	}
	msg := event.Event.Message
	if msg != nil && *msg.MessageType == larkim.MsgTypeSticker {
		contentMap := make(map[string]string)
		err := sonic.UnmarshalString(*msg.Content, &contentMap)
		if err != nil {
			log.Zlog.Error("repeatMessage", zaplog.Error(err))
			return err
		}
		stickerKey := contentMap["file_key"]
		// 表情包为全局file_key，可以直接存下
		if result := database.GetDbConnection().Clauses(clause.OnConflict{DoNothing: true}).
			Create(&database.ReactImageMeterial{GuildID: *msg.ChatId, FileID: stickerKey, Type: consts.LarkResourceTypeSticker}); result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 {
			return nil
		}
	}

	return
}
