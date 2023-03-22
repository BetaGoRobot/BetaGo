package music

import (
	"context"
	"fmt"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/qqmusicapi"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/enescakir/emoji"
	"github.com/kevinmatthe/zaplog"
	"github.com/lonelyevil/kook"
	"go.opentelemetry.io/otel/attribute"
)

var (
	zapLogger   = utility.ZapLogger
	sugerLogger = utility.SugerLogger
)

// SearchMusicByRobot  搜索音乐
//
//	@param targetID
//	@param quoteID
//	@param authorID
//	@return err
func SearchMusicByRobot(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.End()

	if len(args) == 0 {
		return fmt.Errorf("搜索关键词不能为空")
	}
	// 使用网易云搜索
	resNetease, err := neteaseapi.NetEaseGCtx.SearchMusicByKeyWord(args)
	if err != nil {
		if !neteaseapi.NetEaseGCtx.CheckIfLogin() {
		}
		return
	}

	// 使用QQ音乐搜索
	qqmusicCtx := qqmusicapi.QQmusicContext{}
	resQQmusic, err := qqmusicCtx.SearchMusic(args)
	if err != nil {
		return
	}

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
				Src:   song.SongURL,
				Title: song.Name + " - " + song.ArtistName,
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
					Size:  kook.CardSizeSm,
					Modules: append(
						modulesNetese,
						&kook.CardMessageSection{
							Mode: kook.CardMessageSectionModeLeft,
							Text: &kook.CardMessageElementKMarkdown{
								Content: "TraceID: `" + span.SpanContext().TraceID().String() + "`",
							},
						},
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
				Src:   song.SongURL,
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
						&kook.CardMessageSection{
							Mode: kook.CardMessageSectionModeLeft,
							Text: &kook.CardMessageElementKMarkdown{
								Content: "TraceID: `" + span.SpanContext().TraceID().String() + "`",
							},
						},
					),
				},
			)
		}
		if len(resNetease) == 0 && len(resQQmusic) == 0 {
			return
		}
		cardStr, err = cardMessage.BuildMessage()
		if err != nil {
			zapLogger.Error("构建消息失败", zaplog.Error(err))
			return
		}
	} else {
		err = fmt.Errorf("没有找到你要搜索的歌曲, 换一个关键词试试~")
		return
	}
	betagovar.GlobalSession.MessageCreate(
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
