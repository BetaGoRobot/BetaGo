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

func MinioUploadFileFromReadCloser(ctx context.Context, file io.ReadCloser, bucketName, objName string, opts minio.PutObjectOptions) (u *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	log.ZapLogger.Info("MinioUploadFileFromURL...", zaplog.String("traceid", span.SpanContext().TraceID().String()))

	shareURL, err := MinioTryGetFile(ctx, bucketName, objName)
	if err != nil {
		if e, ok := err.(minio.ErrorResponse); ok {
			err = nil
			log.ZapLogger.Warn(e.Error())
		} else {
			log.ZapLogger.Error(err.Error())
			return
		}
	}
	if shareURL != nil {
		u = shareURL
		return
	}

	err = MinioUploadReader(ctx, bucketName, file, objName, opts)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return
	}

	return PresignObj(ctx, bucketName, objName)
}

func PresignObj(ctx context.Context, bucketName, objName string) (u *url.URL, err error) {
	u, err = minioClient.PresignedGetObject(ctx, bucketName, objName, env.OSS_EXPIRATION_TIME, nil)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return
	}

	newURL := shorter.GenAKA(u)
	if newURL != nil {
		u = newURL
	}

	log.ZapLogger.Info("Presined file with url", zaplog.String("presigned_url", u.String()))
	return
}

func MinioTryGetFile(ctx context.Context, bucketName, ObjName string) (url *url.URL, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	log.ZapLogger.Info("MinioTryGetFile...", zaplog.String("traceid", span.SpanContext().TraceID().String()))

	_, err = minioClient.StatObject(ctx, bucketName, ObjName, minio.StatObjectOptions{})
	if err != nil {
		return
	}
	return PresignObj(ctx, bucketName, ObjName)
}

func MinioUploadReader(ctx context.Context, bucketName string, file io.ReadCloser, objName string, opts minio.PutObjectOptions) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	log.ZapLogger.Info("MinioUploadReader...", zaplog.String("traceid", span.SpanContext().TraceID().String()))

	// Upload the test file
	// Change the value of filePath if the file is in another location

	// Upload the test file with FPutObject
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
