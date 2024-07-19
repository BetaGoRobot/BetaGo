package larkhandler

import (
	"context"
	"fmt"
	"strings"

	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
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

	// 处理 cardAction, 这里简单打印卡片内容
	if musicID, ok := cardAction.Action.Value["show_music"]; ok {
		go SendMusicCard(ctx, musicID.(string), cardAction.OpenMessageID, 1)
	}
	if musicID, ok := cardAction.Action.Value["music_id"]; ok {
		go HandleFullLyrics(ctx, musicID.(string), cardAction.OpenMessageID)
	}
	// 无返回值示例
	return nil, nil
}

func GetCardMusicByPage(ctx context.Context, musicID string, page int) string {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

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
	cardStr := larkutils.GenerateMusicCardByStruct(ctx, imageKey, songDetail.Name, songDetail.Ar[0].Name, lyricsURL, lyrics, musicID)
	return cardStr
}

func SendMusicCard(ctx context.Context, musicID string, msgID string, page int) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("musicID").String(musicID))
	defer span.End()

	cardStr := GetCardMusicByPage(ctx, musicID, page)
	fmt.Println(cardStr)
	req := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().Content(cardStr).MsgType(larkim.MsgTypeInteractive).ReplyInThread(true).Uuid(msgID + musicID).Build(),
	).MessageId(msgID).Build()
	resp, err := larkutils.LarkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		return
	}

	fmt.Println(resp)
}

func HandleFullLyrics(ctx context.Context, musicID, msgID string) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID), attribute.Key("musicID").String(musicID))
	defer span.End()

	songDetail := neteaseapi.NetEaseGCtx.GetDetail(ctx, musicID).Songs[0]

	lyric, _ := neteaseapi.NetEaseGCtx.GetLyrics(ctx, musicID)
	lyric = larkutils.TrimLyrics(lyric)
	sp := strings.Split(lyric, "\n")
	left := strings.Join(sp[:len(sp)/2], "\n\n")
	right := strings.Join(sp[len(sp)/2+1:], "\n\n")
	cardStr := larkutils.GenFullLyricsCard(ctx, songDetail.Name, songDetail.Ar[0].Name, left, right)

	req := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().Content(cardStr).MsgType(larkim.MsgTypeInteractive).ReplyInThread(true).Uuid(msgID).Build(),
	).MessageId(msgID).Build()
	resp, err := larkutils.LarkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		return
	}
	fmt.Println(resp.CodeError.Msg)
}
