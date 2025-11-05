package miniohelper

import (
	"os"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	endPointInternal, endPointExternal string
	useSSLInternal, useSSLExternal     bool
	accessKeyID                        = os.Getenv("MINIO_ACCESS_KEY_ID")
	secretAccessKey                    = os.Getenv("MINIO_SECRET_ACCESS_KEY")
	minioClientInternal                *minio.Client
	minioClientExternal                *minio.Client
)

func init() {
	var err error

	if consts.IsTest {
		useSSLInternal = false
		useSSLExternal = true
		endPointInternal = "localhost:19000"
		endPointExternal = "minioapi.kmhomelab.online:2443"
	} else {
		useSSLInternal = false
		useSSLExternal = true
		endPointInternal = "minio:9000"
		endPointExternal = "minioapi.kmhomelab.online:2443"
	}
	minioClientInternal, err = minio.New(endPointInternal, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSLInternal,
	})
	if err != nil {
		logs.L.Panic().Err(err).Msg("MinIO client initialization failed")
	}

	minioClientExternal, err = minio.New(endPointExternal, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSLExternal,
	})
	if err != nil {
		logs.L.Panic().Err(err).Msg("MinIO client initialization failed")
	}
}
