package handlers

import (
	"context"
	"errors"

	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/cardutil"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func MusicSearchHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	argsMap, input := parseArgs(args...)
	searchType, ok := argsMap["type"]
	if !ok {
		// 兼容简易搜索
		searchType = "song"
	}

	keywords := []string{input}

	var cardContent string
	if searchType == "album" {
		albumList, err := neteaseapi.NetEaseGCtx.SearchAlbumByKeyWord(ctx, keywords...)
		if err != nil {
			return err
		}

		cardContent, err = cardutil.SendMusicListCard(ctx, albumList, cardutil.MusicItemTransAlbum, neteaseapi.CommentTypeAlbum)
		if err != nil {
			return err
		}
	} else if searchType == "artist" {
	} else if searchType == "playlist" {
	} else if searchType == "song" {
		musicList, err := neteaseapi.NetEaseGCtx.SearchMusicByKeyWord(ctx, keywords...)
		if err != nil {
			return err
		}
		cardContent, err = cardutil.SendMusicListCard(ctx, musicList, cardutil.MusicItemNoTrans, neteaseapi.CommentTypeSong)
		if err != nil {
			return err
		}
	} else {
		return errors.New("Unknown search type")
	}

	err = larkutils.ReplyMsgRawContentType(ctx, *data.Event.Message.MessageId, larkim.MsgTypeInteractive, cardContent, "_musicSearch", env.MusicCardInThread)
	if err != nil {
		return err
	}
	return
}
