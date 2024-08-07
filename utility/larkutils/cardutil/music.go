package cardutil

import (
	"context"
	"fmt"

	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/kevinmatthe/zaplog"
)

func SendMusicListCard(ctx context.Context, res []*neteaseapi.SearchMusicRes, resourceType neteaseapi.CommentType) (content string, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	traceID := span.SpanContext().TraceID().String()

	lines := make([]map[string]interface{}, 0)
	var buttonName string
	switch resourceType {
	case neteaseapi.CommentTypeSong:
		buttonName = "点击播放"
	case neteaseapi.CommentTypeAlbum:
		buttonName = "查看专辑"
	default:
		buttonName = "点击查看"
	}
	for _, item := range res {
		comment, err := neteaseapi.NetEaseGCtx.GetComment(ctx, resourceType, item.ID)
		if err != nil {
			log.ZapLogger.Error("GetComment", zaplog.Error(err))
		}

		line := map[string]interface{}{
			"field_1":     genMusicTitle(item.Name, item.ArtistName),
			"field_2":     item.ImageKey,
			"button_info": buttonName,
			"button_val": map[string]string{
				"type": "song",
				"id":   item.ID,
			},
		}
		if len(comment.Data.Comments) != 0 {
			line["field_3"] = comment.Data.Comments[0].Content
			if runeSlice := []rune(comment.Data.Comments[0].Content); len(runeSlice) > 50 {
				line["field_3"] = string(runeSlice[:50]) + "..."
			}
			line["comment_time"] = comment.Data.Comments[0].TimeStr
		}
		if resourceType == neteaseapi.CommentTypeSong && item.SongURL == "" { // 无效歌曲
			line["button_info"] = "歌曲无效"
		}
		lines = append(lines, line)
	}

	content = larkutils.NewSheetCardContent(
		larkutils.AlbumListTemplate.TemplateID,
		larkutils.AlbumListTemplate.TemplateVersion,
	).AddVariable(
		"object_list_1", lines,
	).AddVariable("jaeger_trace_info", "JaegerID - "+traceID).
		AddVariable("jaeger_trace_url", "https://jaeger.kmhomelab.cn/"+traceID).String()

	return
}

func genMusicTitle(title, artist string) string {
	return fmt.Sprintf("**%s** - **%s**", title, artist)
}
