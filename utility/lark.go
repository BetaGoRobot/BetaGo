package utility

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/kevinmatthe/zaplog"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

var larkClient *lark.Client = lark.NewClient(os.Getenv("LARK_CLIENT_ID"), os.Getenv("LARK_SECRET"))

func UploadPicAllinOne(ctx context.Context, imageURL, musicID string, uploadOSS bool) (key string, ossURL string, err error) { // also minio
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, GetCurrentFunc())
	span.SetAttributes(attribute.Key("imgURL").String(imageURL))
	defer span.End()

	err, resp := Upload2Lark(ctx, imageURL)
	if err != nil {
		log.ZapLogger.Error("upload pic to lark error", zaplog.Error(err))
		return
	}
	if uploadOSS {
		u, err := MinioUploadFileFromURL(ctx, "cloudmusic", imageURL, "picture/"+musicID+filepath.Ext(imageURL), "image/jpeg")
		if err != nil {
			log.ZapLogger.Warn("upload pic to minio error", zaplog.String("imageURL", imageURL), zaplog.String("imageKey", *resp.Data.ImageKey))
			err = nil
		}
		if u != nil {
			ossURL = u.String()
		}
	}

	return *resp.Data.ImageKey, ossURL, err
}

func Upload2Lark(ctx context.Context, imageURL string) (error, *larkim.CreateImageResp) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, GetCurrentFunc())
	span.SetAttributes(attribute.Key("imgURL").String(imageURL))
	defer span.End()

	picResp, err := http.Get(imageURL)

	req := larkim.NewCreateImageReqBuilder().
		Body(
			larkim.NewCreateImageReqBodyBuilder().
				ImageType(larkim.ImageTypeMessage).
				Image(picResp.Body).
				Build(),
		).
		Build()
	resp, err := larkClient.Im.Image.Create(ctx, req)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return nil, nil
	}
	return err, resp
}

func UploadPicBatch(ctx context.Context, sourceURLIDs map[string]int) chan [2]string {
	var (
		c  = make(chan [2]string)
		wg = &sync.WaitGroup{}
	)
	defer close(c)
	defer wg.Wait()

	for url, musicID := range sourceURLIDs {
		go func(url string, musicID int) {
			_, _, err := UploadPicAllinOne(ctx, url, strconv.Itoa(musicID), true)
			if err != nil {
				log.ZapLogger.Error("upload pic to lark error", zaplog.Error(err))
				return
			}
			c <- [2]string{url, strconv.Itoa(musicID)}
		}(url, musicID)
	}

	return c
}
