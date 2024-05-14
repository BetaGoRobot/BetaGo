package larkutils

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/BetaGoRobot/BetaGo/consts/ct"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	miniohelper "github.com/BetaGoRobot/BetaGo/utility/minio_helper"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/requests"
	"github.com/kevinmatthe/zaplog"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

var larkClient *lark.Client = lark.NewClient(os.Getenv("LARK_CLIENT_ID"), os.Getenv("LARK_SECRET"))

func UploadPicAllinOne(ctx context.Context, imageURL, musicID string, uploadOSS bool) (key string, ossURL string, err error) { // also minio
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("imgURL").String(imageURL))
	defer span.End()

	err, imgKey := Upload2Lark(ctx, musicID, imageURL)
	if err != nil {
		log.ZapLogger.Error("upload pic to lark error", zaplog.Error(err))
		return
	}
	if uploadOSS {
		u, err := miniohelper.Client().
			SetContext(ctx).
			SetBucketName("cloudmusic").
			SetFileFromURL(imageURL).
			SetObjName("picture/" + musicID + filepath.Ext(imageURL)).
			SetContentType(ct.ContentTypeImgJPEG).
			Upload()
		if err != nil {
			log.ZapLogger.Warn("upload pic to minio error", zaplog.String("imageURL", imageURL), zaplog.String("imageKey", imgKey))
			err = nil
		}
		if u != nil {
			ossURL = u.String()
		}
	}

	return imgKey, ossURL, err
}

func Upload2Lark(ctx context.Context, musicID, imageURL string) (err error, imgKey string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("imgURL").String(imageURL))
	defer span.End()

	larkImgs := make([]*database.LarkImg, 0)
	err = database.GetDbConnection().
		Table("betago.lark_imgs").
		Find(&database.LarkImg{SongID: musicID}).
		First(&larkImgs).Error
	if err != nil {
		log.ZapLogger.Error("get lark img from db error", zaplog.Error(err))

		picResp, err := requests.Req().SetDoNotParseResponse(true).Get(imageURL)
		if err != nil {
			log.ZapLogger.Error("get pic from url error", zaplog.Error(err))
			return err, imgKey
		}

		bodyReader := utility.ResizeIMGFromReader(ctx, picResp.RawBody())
		req := larkim.NewCreateImageReqBuilder().
			Body(
				larkim.NewCreateImageReqBodyBuilder().
					ImageType(larkim.ImageTypeMessage).
					Image(bodyReader).
					Build(),
			).
			Build()
		resp, err := larkClient.Im.Image.Create(ctx, req)
		if err != nil {
			log.ZapLogger.Error(err.Error())
			return nil, ""
		}
		if resp.Err != nil {
			return errors.New("error with code" + strconv.Itoa(resp.Code)), ""
		}
		imgKey := *resp.Data.ImageKey
		err = database.GetDbConnection().
			Table("betago.lark_imgs").
			Find(&database.LarkImg{SongID: musicID}).
			FirstOrCreate(&database.LarkImg{SongID: musicID, ImgKey: imgKey}).Error
		if err != nil {
			log.ZapLogger.Warn("create lark img in db error", zaplog.Error(err))
			return nil, imgKey
		}

		return err, *resp.Data.ImageKey
	}
	return nil, larkImgs[0].ImgKey
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
