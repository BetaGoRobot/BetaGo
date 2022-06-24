package music

import (
	"fmt"
	"log"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/neteaseapi"
	"github.com/BetaGoRobot/BetaGo/qqmusicapi"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/khl"
)

// SearchMusicByRobot  搜索音乐
//  @param targetID
//  @param quoteID
//  @param authorID
//  @return err
func SearchMusicByRobot(targetID, quoteID, authorID string, args ...string) (err error) {
	if len(args) == 0 {
		return fmt.Errorf("搜索关键词不能为空")
	}
	// 使用网易云搜索
	resNetease, err := neteaseapi.NetEaseGCtx.SearchMusicByKeyWord(args)
	if err != nil {
		return
	}

	// 使用QQ音乐搜索
	qqmusicCtx := qqmusicapi.QQmusicContext{}
	resQQmusic, err := qqmusicCtx.SearchMusic(args)
	if err != nil {
		return
	}

	var (
		cardMessage   = make(khl.CardMessage, 0)
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
			modulesNetese = append(modulesNetese, khl.CardMessageFile{
				Type:  khl.CardMessageFileTypeAudio,
				Src:   song.SongURL,
				Title: song.Name + " - " + song.ArtistName,
				Cover: song.PicURL,
			})
			tempMap[song.Name+" - "+song.ArtistName] = 0
		}
		modulesNetese = append([]interface{}{
			khl.CardMessageHeader{
				Text: khl.CardMessageElementText{
					Content: emoji.Headphone.String() + "网易云音乐-搜索结果" + emoji.MagnifyingGlassTiltedLeft.String(),
					Emoji:   false,
				},
			},
		}, modulesNetese...)

		tempMap = make(map[string]byte)
		// 添加QQ音乐搜索的结果
		for _, song := range resQQmusic {
			if _, ok := tempMap[song.Name+" - "+song.ArtistName]; ok {
				continue
			}
			modulesQQ = append(modulesQQ, khl.CardMessageFile{
				Type:  khl.CardMessageFileTypeAudio,
				Src:   song.SongURL,
				Title: song.Name + " - " + song.ArtistName,
				Cover: song.PicURL,
			})
			tempMap[song.Name+" - "+song.ArtistName] = 0
		}
		modulesQQ = append([]interface{}{
			khl.CardMessageHeader{
				Text: khl.CardMessageElementText{
					Content: emoji.MusicalNote.String() + "QQ音乐-搜索结果" + emoji.MagnifyingGlassTiltedLeft.String(),
					Emoji:   false,
				},
			},
		}, modulesQQ...)

		cardMessage = append(cardMessage,
			&khl.CardMessageCard{
				Theme:   khl.CardThemePrimary,
				Size:    khl.CardSizeSm,
				Modules: modulesNetese,
			},
			&khl.CardMessageCard{
				Theme:   khl.CardThemePrimary,
				Size:    khl.CardSizeSm,
				Modules: modulesQQ,
			},
		)
		cardStr, err = cardMessage.BuildMessage()
		if err != nil {
			log.Println("-------------", err.Error())
			return
		}
	} else {
		err = fmt.Errorf("没有找到你要搜索的歌曲, 换一个关键词试试~")
		return
	}
	betagovar.GlobalSession.MessageCreate(
		&khl.MessageCreate{
			MessageCreateBase: khl.MessageCreateBase{
				Type:     khl.MessageTypeCard,
				TargetID: targetID,
				Content:  cardStr,
				Quote:    quoteID,
			},
		},
	)

	return
}
