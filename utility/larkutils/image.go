package larkutils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"path/filepath"

	"github.com/BetaGoRobot/BetaGo/consts/ct"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	miniohelper "github.com/BetaGoRobot/BetaGo/utility/minio_helper"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

// DownImgFromMsgSync 从Msg中下载附件
//
//	@param ctx context.Context
//	@param msgID string
//	@param fileKey string
//	@param fileType string
//	@return image []byte
//	@return err error
//	@author kevinmatthe
//	@update 2025-04-27 20:15:38
func DownImgFromMsgSync(ctx context.Context, msgID, fileType, fileKey string) (url string, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(
		attribute.Key("msgID").String(msgID),
		attribute.Key("fileKey").String(fileKey),
		attribute.Key("fileType").String(fileType),
	)
	defer span.End()

	req := larkim.NewGetMessageResourceReqBuilder().
		MessageId(msgID).
		FileKey(fileKey).
		Type("image").
		Build()
	// 发起请求
	resp, err := LarkClient.Im.V1.MessageResource.Get(ctx, req)
	// 处理错误
	if err != nil {
		return
	}

	// 服务端错误处理
	if !resp.Success() {
		return "", errors.New(resp.Error())
	}

	reader, contentType, suffix, err := readAndDetectFormat(resp.File)
	if err != nil {
		return
	}

	u, err := miniohelper.
		Client().
		SetContext(ctx).
		SetBucketName("larkchat").
		SetFileFromReader(reader).
		SetObjName(filepath.Join("chat_image", fileType, fileKey+suffix)).
		SetContentType(ct.ContentType(contentType)).
		SetV4().
		Upload()
	if err != nil {
		log.Zlog.Warn("upload pic to minio error",
			zaplog.String("file_key", fileKey),
			zaplog.String("file_type", fileType),
		)
		return
	}
	log.Zlog.Info("upload pic to minio success",
		zaplog.String("file_key", fileKey),
		zaplog.String("file_type", fileType),
		zaplog.String("url", u.String()),
	)
	url = u.String()
	return
}

// DownImgFromMsgAsync 从Msg中下载附件
//
//	@param ctx context.Context
//	@param msgID string
//	@param fileKey string
//	@param fileType string
//	@return image []byte
//	@return err error
//	@author kevinmatthe
//	@update 2025-04-27 20:15:38
func DownImgFromMsgAsync(ctx context.Context, msgID, fileType, fileKey string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(
		attribute.Key("msgID").String(msgID),
		attribute.Key("fileKey").String(fileKey),
		attribute.Key("fileType").String(fileType),
	)
	defer span.End()

	req := larkim.NewGetMessageResourceReqBuilder().
		MessageId(msgID).
		FileKey(fileKey).
		Type(fileType).
		Build()
	// 发起请求
	resp, err := LarkClient.Im.V1.MessageResource.Get(ctx, req)
	// 处理错误
	if err != nil {
		fmt.Println(err)
		return
	}

	// 服务端错误处理
	if !resp.Success() {
		fmt.Printf("logId: %s, error response: \n%s", resp.RequestId(), larkcore.Prettify(resp.CodeError))
		return
	}

	go func() {
		reader, contentType, suffix, err := readAndDetectFormat(resp.File)
		if err != nil {
			return
		}

		// 异步上传
		u, err := miniohelper.
			Client().
			SetContext(ctx).
			SetBucketName("larkchat").
			SetFileFromReader(reader).
			SetObjName(filepath.Join("chat_image", fileType, fileKey+suffix)).
			SetContentType(ct.ContentType(contentType)).
			Upload()
		if err != nil {
			log.Zlog.Warn("upload pic to minio error",
				zaplog.String("file_key", fileKey),
				zaplog.String("file_type", fileType),
			)
			return
		}
		log.Zlog.Info("upload pic to minio success",
			zaplog.String("file_key", fileKey),
			zaplog.String("file_type", fileType),
			zaplog.String("url", u.String()),
		)
	}()

	return
}

// 检测图片格式
func detectImageFormat(header []byte) (string, string, error) {
	// 检查文件头并返回格式
	switch {
	case bytes.HasPrefix(header, []byte{0x89, 0x50, 0x4E, 0x47}): // PNG
		return "image/png", ".png", nil
	case bytes.HasPrefix(header, []byte{0x47, 0x49, 0x46, 0x38}): // GIF
		return "image/gif", ".gif", nil
	case bytes.HasPrefix(header, []byte{0xFF, 0xD8, 0xFF}): // JPEG
		return "image/jpeg", ".jpg", nil
	default:
		return "", "", fmt.Errorf("unknown image format")
	}
}

// 从 io.Reader 中读取完整的字节数据并检测文件头
func readAndDetectFormat(reader io.Reader) (io.ReadCloser, string, string, error) {
	// 读取文件头（例如，读取 8 个字节）
	header := make([]byte, 8)
	_, err := reader.Read(header)
	if err != nil {
		return nil, "", "", fmt.Errorf("error reading file header: %v", err)
	}

	// 根据文件头检测格式
	contentType, suffix, err := detectImageFormat(header)
	if err != nil {
		return nil, "", "", err
	}

	return wrapReaderWithHeader(header, reader), contentType, suffix, nil
}

// 封装一个新的 io.ReadCloser，从头部+原始Reader组成
func wrapReaderWithHeader(header []byte, r io.Reader) io.ReadCloser {
	return &readCloser{
		Reader: io.MultiReader(bytes.NewReader(header), r),
	}
}

// 自定义 ReadCloser
type readCloser struct {
	io.Reader
}

func (rc *readCloser) Close() error {
	// 如果原始 r 是 ReadCloser，可以在这里关闭底层流
	// 这里为了简单，假设不用关闭底层流或者由外部管理
	return nil
}

type postData struct {
	Title   string           `json:"title"`
	Content [][]*contentData `json:"content"`
}

type contentData struct {
	Tag      string `json:"tag"`
	ImageKey string `json:"image_key"`
}

// GetAllImgTagFromMsg 从消息事件中获取所有图片
//
//	@param event *larkim.P2MessageReceiveV1
//	@author kevinmatthe
//	@update 2025-04-28 19:47:21
func GetAllImgTagFromMsg(ctx context.Context, message *larkim.Message) (imageKeys iter.Seq[string], err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("message").String(larkcore.Prettify(message)))
	defer span.End()

	if msgType := *message.MsgType; msgType == larkim.MsgTypeImage {
		var msg *larkim.MessageImage
		msg, err = jsonTrans[larkim.MessageImage](*message.Body.Content)
		if err != nil {
			return
		}
		return func(yield func(string) bool) {
			if !yield(msg.ImageKey) {
				return
			}
		}, nil
	} else if msgType == larkim.MsgTypePost {
		var msg *postData
		msg, err = jsonTrans[postData](*message.Body.Content)
		if err != nil {
			return
		}
		return func(yield func(string) bool) {
			for key := range getAllImage(ctx, msg) {
				if !yield(key) {
					return
				}
			}
		}, nil
	}
	return
}

// GetAllImageFromMsgEvent 从消息事件中获取所有图片
//
//	@param event *larkim.P2MessageReceiveV1
//	@author kevinmatthe
//	@update 2025-04-28 19:47:21
func GetAllImageFromMsgEvent(ctx context.Context, message *larkim.EventMessage) (imageKeys iter.Seq[string], err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("message").String(larkcore.Prettify(message)))
	defer span.End()

	if msgType := *message.MessageType; msgType == larkim.MsgTypeImage {
		var msg *larkim.MessageImage
		msg, err = jsonTrans[larkim.MessageImage](*message.Content)
		if err != nil {
			return
		}
		return func(yield func(string) bool) {
			if !yield(msg.ImageKey) {
				return
			}
		}, nil
	} else if msgType == larkim.MsgTypePost {
		var msg *postData
		msg, err = jsonTrans[postData](*message.Content)
		if err != nil {
			return
		}
		return func(yield func(string) bool) {
			for key := range getAllImage(ctx, msg) {
				if !yield(key) {
					return
				}
			}
		}, nil
	}
	return
}

func getAllImage(ctx context.Context, msg *postData) iter.Seq[string] {
	return func(yield func(string) bool) {
		_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
		defer span.End()
		for _, elements := range msg.Content {
			for _, element := range elements {
				if element.Tag == "img" {
					if !yield(element.ImageKey) {
						return
					}
				}
			}
		}
	}
}

func jsonTrans[T any](s string) (*T, error) {
	t := new(T)
	err := sonic.UnmarshalString(s, t)
	if err != nil {
		return t, err
	}
	return t, nil
}

func GetAllImgURLFromMsg(ctx context.Context, msgID string) (iter.Seq[string], error) {
	resp := GetMsgFullByID(ctx, msgID)
	msg := resp.Data.Items[0]
	if msg == nil {
		return nil, errors.New("No message found")
	}
	if msg.Sender.Id == nil {
		return nil, errors.New("Message is not sent by bot")
	}
	seq, err := GetAllImgTagFromMsg(ctx, msg)
	if err != nil {
		return nil, err
	}
	if seq == nil {
		return nil, err
	}
	return func(yield func(string) bool) {
		for imageKey := range seq {
			url, err := DownImgFromMsgSync(ctx, *msg.MessageId, *msg.MsgType, imageKey)
			if err != nil {
				return
			}
			if !yield(url) {
				return
			}
		}
	}, nil
}

func GetAllImgURLFromParent(ctx context.Context, data *larkim.P2MessageReceiveV1) (iter.Seq[string], error) {
	if data.Event.Message.ThreadId != nil {
		// 话题模式 找图片
		resp, err := LarkClient.Im.Message.List(ctx,
			larkim.NewListMessageReqBuilder().ContainerIdType("thread").ContainerId(*data.Event.Message.ThreadId).Build())
		if err != nil {
			return nil, err
		}
		if !resp.Success() {
			return nil, errors.New(resp.Error())
		}
		return func(yield func(string) bool) {
			for _, msg := range resp.Data.Items {
				if msg.MsgType == nil || (*msg.MsgType != larkim.MsgTypeImage && *msg.MsgType != larkim.MsgTypePost) {
					continue
				}
				seq, err := GetAllImgURLFromMsg(ctx, *msg.MessageId)
				if err != nil {
					return
				}
				if seq != nil {
					for url := range seq {
						if !yield(url) {
							return
						}
					}
				}
			}
		}, nil
	} else if data.Event.Message.ParentId != nil {
		return func(yield func(string) bool) {
			seq, err := GetAllImgURLFromMsg(ctx, *data.Event.Message.ParentId)
			if err != nil {
				return
			}
			if seq != nil {
				for url := range seq {
					if !yield(url) {
						return
					}
				}
			}
		}, nil
	}
	return nil, nil
}
