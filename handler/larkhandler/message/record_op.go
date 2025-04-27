package message

import (
	"context"

	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
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
	var imageKeys []string
	if msgType := *event.Event.Message.MessageType; msgType == larkim.MsgTypeImage {
		msg, err := jsonTrans[larkim.MessageImage](*event.Event.Message.Content)
		if err != nil {
			return err
		}
		imageKeys = append(imageKeys, msg.ImageKey)

	} else if msgType == larkim.MsgTypePost {
		msg, err := jsonTrans[tmpPost](*event.Event.Message.Content)
		if err != nil {
			return err
		}

		imageKeys, err = getAllImage(ctx, msg)
		if err != nil {
			return err
		}
	}
	for _, imageKey := range imageKeys {
		err = larkutils.DownloadImageFromMsgWithUpload(
			ctx,
			*event.Event.Message.MessageId,
			larkim.MsgTypeImage,
			imageKey,
		)
		if err != nil {
			return err
		}
	}

	return
}

func jsonTrans[T any](s string) (*T, error) {
	t := new(T)
	err := sonic.UnmarshalString(s, t)
	if err != nil {
		return t, err
	}
	return t, nil
}

type tmpPost struct {
	Title   string           `json:"title"`
	Content [][]*contentData `json:"content"`
}

type contentData struct {
	Tag      string `json:"tag"`
	ImageKey string `json:"image_key"`
}

func getAllImage(ctx context.Context, msg *tmpPost) (imageKeys []string, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	imageKeys = make([]string, 0)

	for _, elements := range msg.Content {
		for _, element := range elements {
			if element.Tag == "img" {
				imageKeys = append(imageKeys, element.ImageKey)
			}
		}
	}
	return
}
