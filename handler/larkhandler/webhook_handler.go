package larkhandler

import (
	"context"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts/ct"
	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi/neteaselark"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler/message"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	miniohelper "github.com/BetaGoRobot/BetaGo/utility/minio_helper"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher/callback"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func WebHookHandler(ctx context.Context, cardAction *callback.CardActionTriggerEvent) (*callback.CardActionTriggerResponse, error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer larkutils.RecoverMsg(ctx, cardAction.Event.Context.OpenMessageID)
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(cardAction)))
	defer span.End()

	// 记录一下操作记录
	defer larkutils.RecordCardAction2Opensearch(ctx, cardAction)
	if buttonType, ok := cardAction.Event.Action.Value["type"]; ok {
		switch buttonType {
		case "song":
			if musicID, ok := cardAction.Event.Action.Value["id"]; ok {
				go SendMusicCard(ctx, musicID.(string), cardAction.Event.Context.OpenMessageID, 1)
			}
		case "album":
			if albumID, ok := cardAction.Event.Action.Value["id"]; ok {
				_ = albumID
				go SendAlbumCard(ctx, albumID.(string), cardAction.Event.Context.OpenMessageID)
			}
		case "lyrics":
			if musicID, ok := cardAction.Event.Action.Value["id"]; ok {
				go HandleFullLyrics(ctx, musicID.(string), cardAction.Event.Context.OpenMessageID)
			}
		case "withdraw":
			// 撤回消息
			go HandleWithDraw(ctx, cardAction.Event.Context.OpenMessageID)
		case "refresh":
			if musicID, ok := cardAction.Event.Action.Value["id"]; ok {
				go HandleRefreshMusic(ctx, musicID.(string), cardAction.Event.Context.OpenMessageID)
			}
		case "refresh_obj":
			// 通用的卡片刷新结构，重点是记录触发的command重新触发？
			go HandleRefreshObj(ctx, cardAction.Event.Action.Value["command"].(string), cardAction.Event.Context.OpenMessageID)
		}
	}
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

	return larkutils.NewCardContent(
		ctx,
		larkutils.SingleSongDetailTemplate,
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
	cardContent, err := neteaselark.BuildMusicListCard(ctx,
		result,
		neteaselark.MusicItemNoTrans,
		neteaseapi.CommentTypeSong,
	)
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

	cardContent := larkutils.NewCardContent(
		ctx,
		larkutils.FullLyricsTemplate,
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

func HandleRefreshMusic(ctx context.Context, musicID, msgID string) {
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

func HandleRefreshObj(ctx context.Context, srcCmd, msgID string) {
	data := new(larkim.P2MessageReceiveV1)
	data.Event = new(larkim.P2MessageReceiveV1Data)
	data.Event.Message = new(larkim.EventMessage)
	data.Event.Message.MessageId = utility.StrPointer(msgID)

	err := message.ExecuteFromRawCommand(
		ctx,
		data,
		&handlerbase.BaseMetaData{
			Refresh: true,
		},
		srcCmd,
	)
	if err != nil {
		log.Zlog.Error("refresh obj error", zaplog.Error(err))
	}
}
