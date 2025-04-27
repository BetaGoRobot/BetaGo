package larkutils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/BetaGoRobot/BetaGo/consts/ct"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	miniohelper "github.com/BetaGoRobot/BetaGo/utility/minio_helper"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

// DownloadImageFromMsgWithUpload 从Msg中下载附件
//
//	@param ctx context.Context
//	@param msgID string
//	@param fileKey string
//	@param fileType string
//	@return image []byte
//	@return err error
//	@author kevinmatthe
//	@update 2025-04-27 20:15:38
func DownloadImageFromMsgWithUpload(ctx context.Context, msgID, fileType, fileKey string) (err error) {
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
		u, err := miniohelper.Client().
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
