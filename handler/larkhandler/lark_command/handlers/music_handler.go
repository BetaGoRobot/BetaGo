package handlers

import (
	"context"
	"errors"

	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi/neteaselark"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/ark/tools"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func MusicSearchHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	defer func() { span.RecordError(err) }()

	argsMap, input := parseArgs(args...)
	searchType, ok := argsMap["type"]
	if !ok {
		// 兼容简易搜索
		searchType = "song"
	}

	keywords := []string{input}

	var cardContent *templates.TemplateCardContent
	if searchType == "album" {
		albumList, err := neteaseapi.NetEaseGCtx.SearchAlbumByKeyWord(ctx, keywords...)
		if err != nil {
			return err
		}

		cardContent, err = neteaselark.BuildMusicListCard(ctx,
			albumList,
			neteaselark.MusicItemTransAlbum,
			neteaseapi.CommentTypeAlbum,
			keywords...,
		)
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
		cardContent, err = neteaselark.BuildMusicListCard(ctx,
			musicList,
			neteaselark.MusicItemNoTrans,
			neteaseapi.CommentTypeSong,
			keywords...,
		)
		if err != nil {
			return err
		}
	} else {
		return errors.New("Unknown search type")
	}

	err = larkutils.ReplyCard(ctx, cardContent, *data.Event.Message.MessageId, "_musicSearch", utility.GetIfInthread(ctx, metaData, env.MusicCardInThread))
	if err != nil {
		return err
	}
	return
}

func init() {
	params := tools.NewParameters("object").
		AddProperty("keywords", &tools.Property{
			Type:        "string",
			Description: "音乐搜索的关键词, 多个关键词之间用空格隔开",
		}).AddRequired("keywords")
	fcu := tools.NewFunctionCallUnit().
		Name("music_search").Desc("根据输入的关键词搜索相关的音乐并发送卡片").Params(params).Func(musicSearchWrap)
	tools.M().Add(fcu)
}

func musicSearchWrap(ctx context.Context, meta *tools.FunctionCallMeta, args string) (any, error) {
	s := struct {
		Keywords string `json:"keywords"`
	}{}
	err := utility.UnmarshallStringPre(args, &s)
	if err != nil {
		return nil, err
	}
	metaData := handlerbase.NewBaseMetaDataWithChatIDUID(ctx, meta.ChatID, meta.UserID)
	return "执行成功", MusicSearchHandler(ctx, meta.LarkData, metaData, s.Keywords)
}
