package miniohelper

import (
	"context"
	"errors"
	"io"
	"net/url"

	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/shorter"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

func presignObj(ctx context.Context, bucketName, objName string, needAKA bool) (u *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	u, err = presignObjInner(ctx, bucketName, objName)
	span.SetAttributes(attribute.String("presigned_url", u.String()))
	if needAKA {
		u = shortenURL(ctx, u)
	}
	span.SetAttributes(attribute.String("presigned_url_shortened", u.String()))
	logs.L().Ctx(ctx).Info("Presined file with url", zap.String("presigned_url", u.String()))
	return
}

func presignObjInner(ctx context.Context, bucketName, objName string) (u *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	u, err = minioClientExternal.PresignedGetObject(ctx, bucketName, objName, env.OSS_EXPIRATION_TIME, nil)
	if err != nil {
		logs.L().Ctx(ctx).Error("PresignedGetObject failed", zap.Error(err))
		return
	}
	return
}

func shortenURL(ctx context.Context, u *url.URL) *url.URL {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	if shortenedURL := shorter.GenAKA(ctx, u); shortenedURL != nil {
		return shortenedURL
	}
	return u
}

func MinioTryGetFile(ctx context.Context, bucketName, ObjName string, needAKA bool) (url *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	logs.L().Ctx(ctx).Info("MinioTryGetFile...", zap.String("traceid", span.SpanContext().TraceID().String()))

	if MinioCheckFileExists(ctx, bucketName, ObjName) {
		return presignObj(ctx, bucketName, ObjName, needAKA)
	}
	return nil, errors.New("file not exists")
}

func MinioCheckFileExists(ctx context.Context, bucketName, ObjName string) bool {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	logs.L().Ctx(ctx).Info("MinioCheckFileExists...", zap.String("traceid", span.SpanContext().TraceID().String()))

	_, err := minioClientInternal.StatObject(ctx, bucketName, ObjName, minio.StatObjectOptions{})
	if err != nil {
		return false
	}
	return true
}

func minioUploadReader(ctx context.Context, bucketName string, file io.ReadCloser, objName string, opts minio.PutObjectOptions) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	logs.L().Ctx(ctx).Info("MinioUploadReader...", zap.String("traceid", span.SpanContext().TraceID().String()))

	info, err := minioClientInternal.PutObject(ctx,
		bucketName,
		objName,
		file,
		-1,
		opts,
	)
	if err != nil {
		logs.L().Ctx(ctx).Error("Failed to upload file", zap.Error(err))
		return
	}
	defer span.SetAttributes(attribute.Key("path").String(objName), attribute.Key("size").Int64(info.Size))
	logs.L().Ctx(ctx).Info("Successfully uploaded file", zap.String("path", objName), zap.Int64("size", info.Size))
	return
}
