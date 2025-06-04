package neteaselark

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkimg"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/kevinmatthe/zaplog"
)

type musicItemTransFunc[T any] func(*T) *neteaseapi.SearchMusicItem

func MusicItemNoTrans(item *neteaseapi.SearchMusicItem) *neteaseapi.SearchMusicItem {
	return item
}

func MusicItemTransAlbum(album *neteaseapi.Album) *neteaseapi.SearchMusicItem {
	return &neteaseapi.SearchMusicItem{
		ID:         album.IDStr,
		Name:       "[" + album.Type + "] " + album.Name,
		PicURL:     album.PicURL,
		ArtistName: album.Artist.Name,
		ImageKey:   larkimg.UploadPicture2Lark(context.Background(), album.PicURL),
	}
}

func BuildMusicListCard[T any](ctx context.Context, resList []*T, transFunc musicItemTransFunc[T], resourceType neteaseapi.CommentType, keywords ...string) (content *templates.TemplateCardContent, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	res := make([]*neteaseapi.SearchMusicItem, len(resList))
	for i, item := range resList {
		res[i] = transFunc(item)
	}
	lines := make([]map[string]interface{}, len(res))
	var buttonName string
	var buttonType string
	switch resourceType {
	case neteaseapi.CommentTypeSong:
		buttonName = "点击播放"
		buttonType = "song"
	case neteaseapi.CommentTypeAlbum:
		buttonName = "查看专辑"
		buttonType = "album"
	default:
		buttonName = "点击查看"
		buttonType = "null"
	}

	var (
		commentChan = make(chan map[string]interface{}, len(resList))
		wg          = &sync.WaitGroup{}
	)
	go func() {
		defer close(commentChan)
		defer wg.Wait()
		for idx, item := range res {
			wg.Add(1)
			go func(item *neteaseapi.SearchMusicItem) {
				defer wg.Done()
				comment, err := neteaseapi.NetEaseGCtx.GetComment(ctx, resourceType, item.ID)
				if err != nil {
					log.Zlog.Error("GetComment", zaplog.Error(err))
				}
				line := map[string]interface{}{
					"idx":         idx,
					"field_1":     genMusicTitle(item.Name, item.ArtistName),
					"field_2":     map[string]any{"img_key": item.ImageKey},
					"button_info": buttonName,
					"element_id":  item.ID,
					"button_val": map[string]string{
						"type": buttonType,
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
				commentChan <- line
			}(item)
		}
	}()
	for line := range commentChan {
		idx := line["idx"].(int)
		lines[idx] = line
	}
	content = templates.NewCardContent(
		ctx,
		templates.AlbumListTemplate,
	).
		AddVariable("object_list_1", lines).
		AddVariable("query", fmt.Sprintf("[%s]", strings.Join(keywords, " ")))

	return
}

func genMusicTitle(title, artist string) string {
	return fmt.Sprintf("**%s**\n**%s**", title, artist)
}
