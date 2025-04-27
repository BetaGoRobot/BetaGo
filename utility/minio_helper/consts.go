package miniohelper

import (
	"os"

	"github.com/BetaGoRobot/BetaGo/utility/log"
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
	// }
	minioClient, err = minio.New(endPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Zlog.Panic(err.Error())
	}
}
