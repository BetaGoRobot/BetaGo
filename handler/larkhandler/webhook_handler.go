package larkhandler

import (
	"context"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts/ct"
	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/cardutil"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	miniohelper "github.com/BetaGoRobot/BetaGo/utility/minio_helper"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/kevinmatthe/zaplog"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func WebHookHandler(ctx context.Context, cardAction *larkcard.CardAction) (interface{}, error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer larkutils.RecoverMsg(ctx, cardAction.OpenMessageID)
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(cardAction)))
	defer span.End()

	if buttonType, ok := cardAction.Action.Value["type"]; ok {
		switch buttonType {
		case "song":
			if musicID, ok := cardAction.Action.Value["id"]; ok {
				go SendMusicCard(ctx, musicID.(string), cardAction.OpenMessageID, 1)
			}
		case "album":
			if albumID, ok := cardAction.Action.Value["id"]; ok {
				_ = albumID
				go SendAlbumCard(ctx, albumID.(string), cardAction.OpenMessageID)
			}
		case "lyrics":
			if musicID, ok := cardAction.Action.Value["id"]; ok {
				go HandleFullLyrics(ctx, musicID.(string), cardAction.OpenMessageID)
			}
		case "withdraw":
			// 撤回消息
			go HandleWithDraw(ctx, cardAction.OpenMessageID)
		case "refresh":
			if musicID, ok := cardAction.Action.Value["id"]; ok {
				go HandleRefresh(ctx, musicID.(string), cardAction.OpenMessageID)
			}
		}
	}
	// // 处理 cardAction, 这里简单打印卡片内容
	// if musicID, ok := cardAction.Action.Value["show_music"]; ok {
	// 	go SendMusicCard(ctx, musicID.(string), cardAction.OpenMessageID, 1)
	// }
	// if musicID, ok := cardAction.Action.Value["music_id"]; ok {
	// 	go HandleFullLyrics(ctx, musicID.(string), cardAction.OpenMessageID)
	// }
	// 无返回值示例
	return nil, nil
}

func GetCardMusicByPage(ctx context.Context, musicID string, page int) *larkutils.TemplateCardContent {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

	const (
		maxSingleLineLen = 48
		maxPageSize      = 18
	)
	musicURL, err := neteaseapi.NetEaseGCtx.GetMusicURL(ctx, musicID)
	if err != nil {
		log.Zlog.Error(err.Error())
		return nil
	}

	songDetail := neteaseapi.NetEaseGCtx.GetDetail(ctx, musicID).Songs[0]
	picURL := songDetail.Al.PicURL
	imageKey, ossURL, err := larkutils.UploadPicAllinOne(ctx, picURL, musicID, true)
	if err != nil {
		log.Zlog.Error(err.Error())
		return nil
	}

	lyrics, lyricsURL := neteaseapi.NetEaseGCtx.GetLyrics(ctx, musicID)
	lyrics = larkutils.TrimLyrics(lyrics)

	artistNameList := make([]map[string]string, 0)
	for _, ar := range songDetail.Ar {
		artistNameList = append(artistNameList, map[string]string{"name": ar.Name})
	}

	type resultURL struct {
		Title      string
		LyricsURL  string
		MusicURL   string
		PictureURL string
		Album      string
		Artist     []map[string]string
		Duration   int
	}

	targetURL := &resultURL{
		Title:      songDetail.Name,
		LyricsURL:  lyricsURL,
		MusicURL:   musicURL,
		PictureURL: ossURL,
		Album:      songDetail.Al.Name,
		Artist:     artistNameList,
		Duration:   songDetail.Dt,
	}

	u, err := miniohelper.Client().
		SetContext(ctx).
		SetBucketName("cloudmusic").
		SetFileFromString(utility.MustMashal(targetURL)).
		SetObjName("info/" + musicID + ".json").
		SetContentType(ct.ContentTypePlainText).
		Overwrite().
		Upload()
	if err != nil {
		log.Zlog.Error(err.Error())
		return nil
	}

	playerURL := utility.BuildURL(u.String())
	// eg: page = 1
	quotaRemain := maxPageSize
	lyricList := strings.Split(lyrics, "\n")
	newList := make([]string, 0)
	curPage := 1
	for _, l := range lyricList {
		quotaRemain--
		if len(l) > maxSingleLineLen {
			quotaRemain--
		}
		if quotaRemain <= 0 {
			curPage++
			quotaRemain = maxPageSize
			if curPage > page {
				break
			}
		}
		if curPage == page {
			newList = append(newList, l)
		}
	}

	lyrics = strings.Join(newList, "\n")

	template := larkutils.GetTemplate(larkutils.SingleSongDetailTemplate)
	return larkutils.NewSheetCardContent(
		ctx,
		template.TemplateID,
		template.TemplateVersion,
	).
		AddVariable("lyrics", lyrics).
		AddVariable("title", songDetail.Name).
		AddVariable("sub_title", songDetail.Ar[0].Name).
		AddVariable("imgkey", imageKey).
		AddVariable("player_url", playerURL).
		AddVariable("full_lyrics_button", map[string]string{"type": "lyrics", "id": musicID}).
		AddVariable("refresh_id", map[string]string{"type": "refresh", "id": musicID})
}

func SendMusicCard(ctx context.Context, musicID string, msgID string, page int) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

	card := GetCardMusicByPage(ctx, musicID, page)
	err := larkutils.ReplyCard(ctx, card, msgID, "_music"+musicID, env.MusicCardInThread)
	if err != nil {
		return
	}
}

func SendAlbumCard(ctx context.Context, albumID string, msgID string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("albumID").String(albumID))
	defer span.End()

	albumDetails, err := neteaseapi.NetEaseGCtx.GetAlbumDetail(ctx, albumID)
	if err != nil {
		log.Zlog.Error(err.Error())
		return
	}
	searchRes := neteaseapi.SearchMusic{Result: *albumDetails}

	result, err := neteaseapi.NetEaseGCtx.AsyncGetSearchRes(ctx, searchRes)
	if err != nil {
		return
	}
	cardContent, err := cardutil.BuildMusicListCard(ctx, result, cardutil.MusicItemNoTrans, neteaseapi.CommentTypeSong)
	if err != nil {
		return
	}
	err = larkutils.ReplyCard(ctx, cardContent, msgID, "_album", env.MusicCardInThread)
	if err != nil {
		return
	}
}

func HandleFullLyrics(ctx context.Context, musicID, msgID string) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("musicID").String(musicID))
	defer span.End()
	songDetail := neteaseapi.NetEaseGCtx.GetDetail(ctx, musicID).Songs[0]

	imgKey, _, err := larkutils.UploadPicAllinOne(ctx, songDetail.Al.PicURL, musicID, true)
	lyric, _ := neteaseapi.NetEaseGCtx.GetLyrics(ctx, musicID)
	lyric = larkutils.TrimLyrics(lyric)
	sp := strings.Split(lyric, "\n")
	left := strings.Join(sp[:len(sp)/2], "\n")
	right := strings.Join(sp[len(sp)/2+1:], "\n")

	template := larkutils.GetTemplate(larkutils.FullLyricsTemplate)
	cardContent := larkutils.NewSheetCardContent(
		ctx,
		template.TemplateID,
		template.TemplateVersion,
	).
		AddVariable("left_lyrics", left).
		AddVariable("right_lyrics", right).
		AddVariable("title", songDetail.Name).
		AddVariable("sub_title", songDetail.Ar[0].Name).
		AddVariable("imgkey", imgKey)
	err = larkutils.ReplyCard(ctx, cardContent, msgID, "_music", env.MusicCardInThread)
	if err != nil {
		return
	}
}

func HandleWithDraw(ctx context.Context, msgID string) {
	// 撤回消息
	resp, err := larkutils.LarkClient.Im.Message.Delete(ctx, larkim.NewDeleteMessageReqBuilder().MessageId(msgID).Build())
	if err != nil {
		log.Zlog.Error(err.Error())
		return
	}
	if !resp.Success() {
		log.Zlog.Error("delete message error", zaplog.String("error", resp.Error()))
	}
	return
}

func HandleRefresh(ctx context.Context, musicID, msgID string) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("musicID").String(musicID))
	defer span.End()

	card := GetCardMusicByPage(ctx, musicID, 1)
	resp, err := larkutils.LarkClient.Im.V1.Message.Patch(
		ctx, larkim.NewPatchMessageReqBuilder().
			MessageId(msgID).
			Body(
				larkim.NewPatchMessageReqBodyBuilder().
					Content(card.String()).
					Build(),
			).
			Build(),
	)
	if err != nil {
		return
	}
	if !resp.Success() {
		log.Zlog.Error("refresh music card error", zaplog.Error(err))
		return
	}
	return
}
