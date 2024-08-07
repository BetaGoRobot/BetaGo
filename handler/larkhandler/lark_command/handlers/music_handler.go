package handlers

import (
	"context"
	"errors"

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

		// lines := make([]map[string]interface{}, 0)
		searchRes := []*neteaseapi.SearchMusicRes{}
		for _, album := range albumList {
			searchRes = append(searchRes,
				&neteaseapi.SearchMusicRes{
					ID:         album.IDStr,
					Name:       "[" + album.Type + "] " + album.Name,
					PicURL:     album.PicURL,
					ArtistName: album.Artist.Name,
					ImageKey:   larkutils.UploadPicture2Lark(ctx, album.PicURL),
				},
			)
			// lines = append(lines,
			// 	map[string]interface{}{
			// 		"field_1": fmt.Sprintf("**%s** - **%s**", album.Name, album.Artist.Name),
			// 		"field_2": larkutils.UploadPicture2Lark(ctx, album.PicURL),
			// 		"button_val": map[string]string{
			// 			"type": "album",
			// 			"id":   album.IDStr,
			// 		},
			// 	},
			// )
		}

		// cardContent = larkutils.NewSheetCardContent(
		// 	larkutils.AlbumListTemplate.TemplateID,
		// 	larkutils.AlbumListTemplate.TemplateVersion,
		// ).AddVariable(
		// 	"object_list_1", lines,
		// ).AddVariable("jaeger_trace_info", "JaegerID - "+traceID).
		// 	AddVariable("jaeger_trace_url", "https://jaeger.kmhomelab.cn/"+traceID).String()
		cardContent, err = cardutil.SendMusicListCard(ctx, searchRes, neteaseapi.CommentTypeAlbum)
		if err != nil {
			return err
		}
	} else if searchType == "song" {
		res, err := neteaseapi.NetEaseGCtx.SearchMusicByKeyWord(ctx, keywords...)
		if err != nil {
			return err
		}
		cardContent, err = cardutil.SendMusicListCard(ctx, res, neteaseapi.CommentTypeSong)
		if err != nil {
			return err
		}
	} else {
		return errors.New("Unknown search type")
	}

	err = larkutils.ReplyMsgRawContentType(ctx, *data.Event.Message.MessageId, larkim.MsgTypeInteractive, cardContent, "_musicSearch", true)
	if err != nil {
		return err
	}
	return
}
