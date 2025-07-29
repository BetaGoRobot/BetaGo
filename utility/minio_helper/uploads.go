package miniohelper

import (
	"context"
	"errors"
	"io"
	"net/url"

	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/shorter"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/kevinmatthe/zaplog"
	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel/attribute"
)

func presignObj(ctx context.Context, bucketName, objName string, needAKA bool) (u *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	u, err = presignObjInner(ctx, bucketName, objName)
	span.SetAttributes(attribute.String("presigned_url", u.String()))
	if needAKA {
		u = shortenURL(ctx, u)
	}
	span.SetAttributes(attribute.String("presigned_url_shortened", u.String()))
	log.Zlog.Info("Presined file with url", zaplog.String("presigned_url", u.String()))
	return
}

func presignObjInner(ctx context.Context, bucketName, objName string) (u *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	u, err = minioClientExternal.PresignedGetObject(ctx, bucketName, objName, env.OSS_EXPIRATION_TIME, nil)
	if err != nil {
		log.Zlog.Error(err.Error())
		return
	}
	return
}

func shortenURL(ctx context.Context, u *url.URL) *url.URL {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	if shortenedURL := shorter.GenAKA(u); shortenedURL != nil {
		return shortenedURL
	}
	return u
}

func MinioTryGetFile(ctx context.Context, bucketName, ObjName string, needAKA bool) (url *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	log.Zlog.Info("MinioTryGetFile...", zaplog.String("traceid", span.SpanContext().TraceID().String()))

	if MinioCheckFileExists(ctx, bucketName, ObjName) {
		return presignObj(ctx, bucketName, ObjName, needAKA)
	}
	return nil, errors.New("file not exists")
}

func MinioCheckFileExists(ctx context.Context, bucketName, ObjName string) bool {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	log.Zlog.Info("MinioCheckFileExists...", zaplog.String("traceid", span.SpanContext().TraceID().String()))

	_, err := minioClientInternal.StatObject(ctx, bucketName, ObjName, minio.StatObjectOptions{})
	if err != nil {
		return false
	}
	return true
}

func minioUploadReader(ctx context.Context, bucketName string, file io.ReadCloser, objName string, opts minio.PutObjectOptions) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	log.Zlog.Info("MinioUploadReader...", zaplog.String("traceid", span.SpanContext().TraceID().String()))

	info, err := minioClientInternal.PutObject(ctx,
		bucketName,
		objName,
		file,
		-1,
		opts,
	)
	if err != nil {
		log.Zlog.Error(err.Error())
		return
	}
	defer span.SetAttributes(attribute.Key("path").String(objName), attribute.Key("size").Int64(info.Size))
	log.SLog.Infof("Successfully uploaded %s of size %d\n", objName, info.Size)
	return
}
