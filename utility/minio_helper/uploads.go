package miniohelper

import (
	"context"
	"io"
	"net/url"

	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/shorter"
	"github.com/kevinmatthe/zaplog"
	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel/attribute"
)

func presignObj(ctx context.Context, bucketName, objName string) (u *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	u, err = minioClient.PresignedGetObject(ctx, bucketName, objName, env.OSS_EXPIRATION_TIME, nil)
	if err != nil {
		span.SetAttributes(attribute.Key("hit_cache").Bool(false))
		log.ZapLogger.Error(err.Error())
		return
	}
	span.SetAttributes(attribute.Key("hit_cache").Bool(true))

	newURL := shorter.GenAKA(u)
	if newURL != nil {
		u = newURL
	}

	log.ZapLogger.Info("Presined file with url", zaplog.String("presigned_url", u.String()))
	return
}

func minioTryGetFile(ctx context.Context, bucketName, ObjName string) (url *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	log.ZapLogger.Info("MinioTryGetFile...", zaplog.String("traceid", span.SpanContext().TraceID().String()))

	_, err = minioClient.StatObject(ctx, bucketName, ObjName, minio.StatObjectOptions{})
	if err != nil {
		return
	}
	return presignObj(ctx, bucketName, ObjName)
}

func minioUploadReader(ctx context.Context, bucketName string, file io.ReadCloser, objName string, opts minio.PutObjectOptions) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	log.ZapLogger.Info("MinioUploadReader...", zaplog.String("traceid", span.SpanContext().TraceID().String()))

	info, err := minioClient.PutObject(ctx,
		bucketName,
		objName,
		file,
		-1,
		opts,
	)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return
	}
	defer span.SetAttributes(attribute.Key("path").String(objName), attribute.Key("size").Int64(info.Size))
	log.SugerLogger.Infof("Successfully uploaded %s of size %d\n", objName, info.Size)
	return
}
