package utility

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// var  cosBucket = "kevinmatt-1303917904"
var (
	COSSecretID        = os.Getenv("COS_SECRET_ID")
	COSSecretKey       = os.Getenv("COS_SECRET_KEY")
	COSBaseURL         = os.Getenv("COS_BASE_URL")
	COSBucketRegionURL = os.Getenv("COS_BUCKET_REGION_URL")
)

// UploadFileToCos 将文件上传到cos
func UploadFileToCos(filePath string) (linkURL string, err error) {
	u, _ := url.Parse(COSBaseURL)

	su, _ := url.Parse(COSBucketRegionURL)

	b := &cos.BaseURL{BucketURL: u, ServiceURL: su}

	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  COSSecretID,
			SecretKey: COSSecretKey,
		},
	})
	uploadName := filepath.Join("betago", filepath.Base(filePath))
	uploadRes, _, err := client.Object.Upload(context.Background(), uploadName, filePath, nil)
	if err != nil {
		return
	}
	return uploadRes.Location, err
}
