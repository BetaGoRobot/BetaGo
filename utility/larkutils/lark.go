package larkutils

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/BetaGoRobot/BetaGo/consts/ct"
	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	miniohelper "github.com/BetaGoRobot/BetaGo/utility/minio_helper"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/requests"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/kevinmatthe/zaplog"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

var LarkClient *lark.Client = lark.NewClient(env.LarkAppID, env.LarkAppSecret)

func GetAndResizePicFromURL(ctx context.Context, imageURL string) (res []byte, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("imgURL").String(imageURL))
	defer span.End()

	picResp, err := requests.Req().SetDoNotParseResponse(true).Get(imageURL)
	if err != nil {
		log.ZapLogger.Error("get pic from url error", zaplog.Error(err))
		return
	}

	res = utility.ResizeIMGFromReader(ctx, picResp.RawBody())
	return
}

func checkDBCache(ctx context.Context, musicID string) (imgKey string, err error) {
	larkImgs := make([]*database.LarkImg, 0)

	err = database.GetDbConnection().
		Table("betago.lark_imgs").
		Find(&database.LarkImg{SongID: musicID}).
		First(&larkImgs).Error
	if err != nil {
		log.ZapLogger.Error("get lark img from db error", zaplog.Error(err))
		return
	}
	return larkImgs[0].ImgKey, err
}

func UploadPicAllinOne(ctx context.Context, imageURL, musicID string, uploadOSS bool) (key string, ossURL string, err error) { // also minio
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("imgURL").String(imageURL))
	defer span.End()

	imgKey, err := checkDBCache(ctx, musicID)
	if err != nil {
		log.ZapLogger.Warn("get lark img from db error", zaplog.Error(err))
		// db 缓存未找到，准备resize上传
		var picData []byte
		picData, err = GetAndResizePicFromURL(ctx, imageURL)
		if err != nil {
			log.ZapLogger.Error("resize pic from url error", zaplog.Error(err))
			return
		}

		imgKey, err = Upload2Lark(ctx, musicID, io.NopCloser(bytes.NewReader(picData)))
		if err != nil {
			log.ZapLogger.Error("upload pic to lark error", zaplog.Error(err))
			return
		}
		if uploadOSS {
			u, err := miniohelper.Client().
				SetContext(ctx).
				SetBucketName("cloudmusic").
				SetFileFromReader(io.NopCloser(bytes.NewReader(picData))).
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
	}
	u, err := miniohelper.MinioTryGetFile(ctx, "cloudmusic", "picture/"+musicID+filepath.Ext(imageURL), true)
	if err != nil {
		log.ZapLogger.Warn("get pic from minio error", zaplog.Error(err))
		err = nil
	}
	if u != nil {
		ossURL = u.String()
	}
	return imgKey, ossURL, err
}

func Upload2Lark(ctx context.Context, musicID string, bodyReader io.ReadCloser) (imgKey string, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	req := larkim.NewCreateImageReqBuilder().
		Body(
			larkim.NewCreateImageReqBodyBuilder().
				ImageType(larkim.ImageTypeMessage).
				Image(bodyReader).
				Build(),
		).
		Build()
	resp, err := LarkClient.Im.Image.Create(ctx, req)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return "", nil
	}
	if resp.Err != nil {
		return "", errors.New("error with code" + strconv.Itoa(resp.Code))
	}
	imgKey = *resp.Data.ImageKey
	err = database.GetDbConnection().
		Table("betago.lark_imgs").
		Find(&database.LarkImg{SongID: musicID}).
		FirstOrCreate(&database.LarkImg{SongID: musicID, ImgKey: imgKey}).Error
	if err != nil {
		log.ZapLogger.Warn("create lark img in db error", zaplog.Error(err))
		return imgKey, nil
	}

	return
}

func UploadPicture2LarkReader(ctx context.Context, picture io.Reader) (imgKey string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	req := larkim.NewCreateImageReqBuilder().
		Body(
			larkim.NewCreateImageReqBodyBuilder().
				ImageType(larkim.ImageTypeMessage).
				Image(picture).
				Build(),
		).
		Build()

	resp, err := LarkClient.Im.Image.Create(ctx, req)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return
	}
	if resp.Err != nil {
		log.ZapLogger.Error("error with code" + strconv.Itoa(resp.Code))
		return
	}
	imgKey = *resp.Data.ImageKey
	return imgKey
}

func UploadPicture2Lark(ctx context.Context, URL string) (imgKey string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	picData, err := GetAndResizePicFromURL(ctx, URL)
	if err != nil {
		log.ZapLogger.Error("resize pic from url error", zaplog.Error(err))
	}

	req := larkim.NewCreateImageReqBuilder().
		Body(
			larkim.NewCreateImageReqBodyBuilder().
				ImageType(larkim.ImageTypeMessage).
				Image(bytes.NewReader(picData)).
				Build(),
		).
		Build()

	resp, err := LarkClient.Im.Image.Create(ctx, req)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return
	}
	if resp.Err != nil {
		log.ZapLogger.Error("error with code" + strconv.Itoa(resp.Code))
		return
	}
	imgKey = *resp.Data.ImageKey
	return imgKey
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

func GetUserMapFromChatID(ctx context.Context, chatID string) (memberMap map[string]*larkim.ListMember, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	memberMap = make(map[string]*larkim.ListMember)
	hasMore := true
	pageToken := ""
	for hasMore {
		builder := larkim.
			NewGetChatMembersReqBuilder().
			MemberIdType(`open_id`).
			ChatId(chatID).
			PageSize(20)
		if pageToken != "" {
			builder.PageToken(pageToken)
		}
		resp, err := LarkClient.Im.ChatMembers.Get(ctx, builder.Build())
		if err != nil {
			return memberMap, err
		}
		if resp.CodeError.Code != 0 {
			err = errors.New(resp.Error())
			return memberMap, err
		}
		for _, item := range resp.Data.Items {
			memberMap[*item.MemberId] = item
		}
		hasMore = *resp.Data.HasMore
		pageToken = *resp.Data.PageToken
	}
	return
}

func GetUserMemberFromChat(ctx context.Context, chatID, openID string) (member *larkim.ListMember, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	memberMap, err := GetUserMapFromChatID(ctx, chatID)
	if err != nil {
		return
	}
	return memberMap[openID], err
}

func GetChatName(ctx context.Context, chatID string) (chatName string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	resp, err := LarkClient.Im.V1.Chat.Get(ctx, larkim.NewGetChatReqBuilder().ChatId(chatID).Build())
	if err != nil {
		return
	}
	if resp == nil || resp.CodeError.Code != 0 {
		err = errors.New(resp.Error())
		return
	}
	chatName = *resp.Data.Name
	return
}

func GetChatIDFromMsgID(ctx context.Context, msgID string) (chatID string, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	resp := GetMsgFullByID(ctx, msgID)
	if resp == nil || resp.CodeError.Code != 0 {
		err = errors.New(resp.Error())
		return
	}
	chatID = *resp.Data.Items[0].ChatId
	return
}
