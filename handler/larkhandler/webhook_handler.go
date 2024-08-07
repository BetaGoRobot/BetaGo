package larkhandler

import (
	"context"
	"fmt"
	"strings"

	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/cardutil"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/bytedance/sonic"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func WebHookHandler(ctx context.Context, cardAction *larkcard.CardAction) (interface{}, error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer larkutils.RecoverMsg(ctx, cardAction.OpenMessageID)
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(cardAction)))
	defer span.End()

	if buttonType, ok := cardAction.Action.Value["type"]; ok {
		if buttonType == "song" {
			if musicID, ok := cardAction.Action.Value["id"]; ok {
				go SendMusicCard(ctx, musicID.(string), cardAction.OpenMessageID, 1)
			}
		} else if buttonType == "album" {
			if albumID, ok := cardAction.Action.Value["id"]; ok {
				_ = albumID
				go SendAlbumCard(ctx, albumID.(string), cardAction.OpenMessageID)
			}
		} else if buttonType == "lyrics" {
			if musicID, ok := cardAction.Action.Value["id"]; ok {
				go HandleFullLyrics(ctx, musicID.(string), cardAction.OpenMessageID)
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

func GetCardMusicByPage(ctx context.Context, musicID string, page int) string {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

	traceID := span.SpanContext().TraceID().String()
	const (
		maxSingleLineLen = 48
		maxPageSize      = 9
	)
	musicURL, err := neteaseapi.NetEaseGCtx.GetMusicURL(ctx, musicID)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return ""
	}

	songDetail := neteaseapi.NetEaseGCtx.GetDetail(ctx, musicID).Songs[0]
	picURL := songDetail.Al.PicURL
	imageKey, ossURL, err := larkutils.UploadPicAllinOne(ctx, picURL, musicID, true)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return ""
	}

	lyrics, lyricsURL := neteaseapi.NetEaseGCtx.GetLyrics(ctx, musicID)
	lyrics = larkutils.TrimLyrics(lyrics)

	artistNameLissst := make([]map[string]string, 0)
	for _, ar := range songDetail.Ar {
		artistNameLissst = append(artistNameLissst, map[string]string{"name": ar.Name})
	}
	artistJSON, err := sonic.MarshalString(artistNameLissst)
	if err != nil {
		log.ZapLogger.Error(err.Error())
	}
	lyricsURL = utility.BuildURL(lyricsURL, musicURL, ossURL, songDetail.Al.Name, songDetail.Name, artistJSON, songDetail.Dt)
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

	lyrics = strings.ReplaceAll(lyrics, "\n", "\n\n")

	return larkutils.NewSheetCardContent(
		larkutils.SingleSongDetailTemplate.TemplateID,
		larkutils.SingleSongDetailTemplate.TemplateVersion,
	).AddVariable("lyrics", lyrics).
		AddVariable("title", songDetail.Name).
		AddVariable("sub_title", songDetail.Ar[0].Name).
		AddVariable("imgkey", imageKey).
		AddVariable("jaeger_trace_info", "JaegerID - "+traceID).
		AddVariable("jaeger_trace_url", "https://jaeger.kmhomelab.cn/"+traceID).
		AddVariable("full_lyrics_button", map[string]string{
			"type": "lyrics",
			"id":   musicID,
		}).String()
}

func SendMusicCard(ctx context.Context, musicID string, msgID string, page int) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

	cardStr := GetCardMusicByPage(ctx, musicID, page)
	fmt.Println(cardStr)
	err := larkutils.ReplyMsgRawContentType(ctx, msgID, larkim.MsgTypeInteractive, cardStr, "_music"+musicID, true)
	if err != nil {
		return
	}
}

func SendAlbumCard(ctx context.Context, albumID string, msgID string) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("albumID").String(albumID))
	defer span.End()

	albumDetails, err := neteaseapi.NetEaseGCtx.GetAlbumDetail(ctx, albumID)
	if err != nil {
		log.ZapLogger.Error(err.Error())
		return
	}
	searchRes := neteaseapi.SearchMusic{Result: *albumDetails}

	result, err := neteaseapi.NetEaseGCtx.AsyncGetSearchRes(ctx, searchRes)
	if err != nil {
		return
	}
	cardContent, err := cardutil.SendMusicListCard(ctx, result, neteaseapi.CommentTypeAlbum)
	if err != nil {
		return
	}
	err = larkutils.ReplyMsgRawContentType(ctx, msgID, larkim.MsgTypeInteractive, cardContent, "_album", true)
	if err != nil {
		return
	}
}

func HandleFullLyrics(ctx context.Context, musicID, msgID string) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("musicID").String(musicID))
	defer span.End()
	traceID := span.SpanContext().TraceID().String()
	songDetail := neteaseapi.NetEaseGCtx.GetDetail(ctx, musicID).Songs[0]

	imgKey, _, err := larkutils.UploadPicAllinOne(ctx, songDetail.Al.PicURL, musicID, true)
	lyric, _ := neteaseapi.NetEaseGCtx.GetLyrics(ctx, musicID)
	lyric = larkutils.TrimLyrics(lyric)
	sp := strings.Split(lyric, "\n")
	left := strings.Join(sp[:len(sp)/2], "\n\n")
	right := strings.Join(sp[len(sp)/2+1:], "\n\n")
	cardStr := larkutils.NewSheetCardContent(
		larkutils.FullLyricsTemplate.TemplateID,
		larkutils.FullLyricsTemplate.TemplateVersion,
	).AddVariable("left_lyrics", left).
		AddVariable("right_lyrics", right).
		AddVariable("title", songDetail.Name).
		AddVariable("sub_title", songDetail.Ar[0].Name).
		AddVariable("imgkey", imgKey).
		AddVariable("jaeger_trace_info", "JaegerID - "+traceID).
		AddVariable("jaeger_trace_url", "https://jaeger.kmhomelab.cn/"+traceID).String()
	err = larkutils.ReplyMsgRawContentType(ctx, msgID, larkim.MsgTypeInteractive, cardStr, "_music", true)
	if err != nil {
		return
	}
}
