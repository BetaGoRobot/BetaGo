package music

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/dal/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/dal/qqmusicapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/enescakir/emoji"
	"github.com/kevinmatthe/zaplog"
	"github.com/lonelyevil/kook"
	"go.opentelemetry.io/otel/attribute"
)

// SearchMusicByRobot  搜索音乐
//
//	@param targetID
//	@param quoteID
//	@param authorID
//	@return err
func SearchMusicByRobot(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	if len(args) == 0 {
		return fmt.Errorf("搜索关键词不能为空")
	}
	// 使用网易云搜索
	resNetease, err := neteaseapi.NetEaseGCtx.SearchMusicByKeyWord(ctx, args...)
	if err != nil {
		if !neteaseapi.NetEaseGCtx.CheckIfLogin(ctx) {
		}
		return
	}

	// 使用QQ音乐搜索
	// qqmusicCtx := qqmusicapi.QQmusicContext{}
	// resQQmusic, err := qqmusicCtx.SearchMusic(ctx, args)
	// if err != nil {
	// 	return
	// }
	resQQmusic := make([]qqmusicapi.SearchMusicRes, 0)
	var (
		cardMessage   = make(kook.CardMessage, 0)
		modulesNetese = make([]interface{}, 0)
		modulesQQ     = make([]interface{}, 0)
		cardStr       string
	)

	if len(resNetease) != 0 || len(resQQmusic) != 0 {
		tempMap := make(map[string]byte, 0)
		// 添加网易云搜索的结果
		for _, song := range resNetease {
			if _, ok := tempMap[song.Name+" - "+song.ArtistName]; ok {
				continue
			}
			modulesNetese = append(modulesNetese, kook.CardMessageFile{
				Type:  kook.CardMessageFileTypeAudio,
				Src:   strings.Replace(song.SongURL, "https", "http", 1),
				Title: song.Name + " - " + song.ArtistName + " - " + song.ID,
				Cover: song.PicURL,
			})
			tempMap[song.Name+" - "+song.ArtistName] = 0
		}
		if len(resNetease) != 0 {
			modulesNetese = append([]interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: emoji.Headphone.String() + "网易云音乐-搜索结果" + emoji.MagnifyingGlassTiltedLeft.String(),
						Emoji:   false,
					},
				},
			}, modulesNetese...)
			cardMessage = append(
				cardMessage,
				&kook.CardMessageCard{
					Theme: kook.CardThemePrimary,
					Size:  kook.CardSizeLg,
					Modules: append(
						modulesNetese,
						&kook.CardMessageDivider{},
						kook.CardMessageSection{
							Mode: kook.CardMessageSectionModeRight,
							Text: &kook.CardMessageElementKMarkdown{
								Content: fmt.Sprintf("> 音乐无法播放？试试刷新音源\n> 当前音源版本:`%s`", time.Now().Local().Format("01-02T15:04:05")),
							},
							Accessory: kook.CardMessageElementButton{
								Theme: kook.CardThemePrimary,
								Value: "Refresh",
								Click: string(kook.CardMessageElementButtonClickReturnVal),
								Text:  "刷新音源",
							},
						},
						utility.GenerateTraceButtonSection(span.SpanContext().TraceID().String()),
					),
				},
			)
		}
		tempMap = make(map[string]byte)
		// 添加QQ音乐搜索的结果
		for _, song := range resQQmusic {
			if _, ok := tempMap[song.Name+" - "+song.ArtistName]; ok {
				continue
			}
			modulesQQ = append(modulesQQ, kook.CardMessageFile{
				Type:  kook.CardMessageFileTypeAudio,
				Src:   strings.Replace(song.SongURL, "https", "http", 1),
				Title: song.Name + " - " + song.ArtistName,
				Cover: song.PicURL,
			})
			tempMap[song.Name+" - "+song.ArtistName] = 0
		}
		if len(resQQmusic) != 0 {
			modulesQQ = append([]interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: emoji.MusicalNote.String() + "QQ音乐-搜索结果" + emoji.MagnifyingGlassTiltedLeft.String(),
						Emoji:   false,
					},
				},
			}, modulesQQ...)
			cardMessage = append(
				cardMessage,
				&kook.CardMessageCard{
					Theme: kook.CardThemePrimary,
					Size:  kook.CardSizeSm,
					Modules: append(
						modulesQQ,
						&kook.CardMessageDivider{},
						utility.GenerateTraceButtonSection(span.SpanContext().TraceID().String()),
					),
				},
			)
		}
		if len(resNetease) == 0 && len(resQQmusic) == 0 {
			return
		}
		cardStr, err = cardMessage.BuildMessage()
		if err != nil {
			log.Zlog.Error("构建消息失败", zaplog.Error(err))
			return
		}
	} else {
		err = fmt.Errorf("没有找到你要搜索的歌曲, 换一个关键词试试~")
		return
	}
	consts.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: targetID,
				Content:  cardStr,
				Quote:    quoteID,
			},
		},
	)

	return
}
