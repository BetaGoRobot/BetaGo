package utility

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/kevinmatthe/zaplog"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	endPoint        string
	useSSL          bool
	accessKeyID     = os.Getenv("MINIO_ACCESS_KEY_ID")
	secretAccessKey = os.Getenv("MINIO_SECRET_ACCESS_KEY")
	minioClient     *minio.Client
)

func init() {
	var err error
	// if betagovar.IsTest {
	// 	endPoint = "192.168.31.74:29000"
	// 	useSSL = false
	// } else {
	endPoint = "minioapi.kmhomelab.cn"
	useSSL = true
	if betagovar.IsCluster {
		endPoint = "192.168.31.74:29000"
		useSSL = false
	}
	// }
	minioClient, err = minio.New(endPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.ZapLogger.Panic(err.Error())
	}
}

func MinioUploadFile(ctx context.Context, bucketName, filePath, objName, contentType string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, GetCurrentFunc())
	defer span.End()
	log.ZapLogger.Info("MinioUploadFile...", zaplog.String("traceid", span.SpanContext().TraceID().String()))

	// Upload the test file
	// Change the value of filePath if the file is in another location

	// Upload the test file with FPutObject
	info, err := minioClient.FPutObject(ctx, bucketName, objName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return
	}

	log.SugerLogger.Infof("Successfully uploaded %s of size %d\n", objName, info.Size)
	return
}

func downloadFile(ctx context.Context, filepath string, url string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, GetCurrentFunc())
	defer span.End()
	log.ZapLogger.Info("downloadFile...", zaplog.String("traceid", span.SpanContext().TraceID().String()))

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func MinioTryGetFile(ctx context.Context, bucketName, ObjName string) (url *url.URL, err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, GetCurrentFunc())
	defer span.End()
	log.ZapLogger.Info("MinioTryGetFile...", zaplog.String("traceid", span.SpanContext().TraceID().String()))

	_, err = minioClient.StatObject(ctx, bucketName, ObjName, minio.StatObjectOptions{})
	if err != nil {
		return
	}
	return PresignObj(ctx, bucketName, ObjName)
}

func MinioUploadFileFromURL(ctx context.Context, bucketName, fileURL, objName, contentType string) (u *url.URL, err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, GetCurrentFunc())
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

	err = downloadFile(ctx, "/tmp/"+objName, fileURL)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return
	}

	err = MinioUploadFile(ctx, bucketName, "/tmp/"+objName, objName, contentType)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return
	}

	return PresignObj(ctx, bucketName, objName)
}

func PresignObj(ctx context.Context, bucketName, objName string) (u *url.URL, err error) {
	u, err = minioClient.PresignedGetObject(ctx, bucketName, objName, time.Hour, nil)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return
	}
	return
}
