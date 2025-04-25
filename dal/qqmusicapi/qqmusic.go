package qqmusicapi

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/requests"
	"github.com/BetaGoRobot/go_utils/reflecting"
	jsoniter "github.com/json-iterator/go"
	"go.opentelemetry.io/otel/attribute"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	if consts.IsTest {
		qqmusicBaseURL = "http://192.168.31.74:3300"
	} else if consts.IsCluster {
		qqmusicBaseURL = "http://kubernetes.default:3300"
	} else if consts.IsCompose {
		qqmusicBaseURL = "http://192.168.31.74:3300"
	}
}

func autoRefreshLogin() {
	for {
		time.Sleep(time.Minute * 5)
		requests.Req().Post(qqmusicBaseURL + "/user/refresh")
	}
}

func init() {
	// 获取存储的Cookie
	_, err := requests.Req().Post(qqmusicBaseURL + "/user/cookie")
	if err != nil {
		log.Println(err.Error())
	}
	go autoRefreshLogin()
}

// SearchMusic  搜索音乐
//
//	@receiver ctx
//	@param keywords
//	@return result
//	@return err
func (qqCtx *QQmusicContext) SearchMusic(ctx context.Context, keywords []string) (result []SearchMusicRes, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("keywords").StringSlice(keywords))
	defer span.End()
	resp, err := requests.
		ReqTimestamp().
		SetFormDataFromValues(map[string][]string{
			"key":      {strings.Join(keywords, " ")},
			"pageNo":   {"1"},
			"pageSize": {"3"},
			"t":        {"0"},
		}).Post(qqmusicBaseURL + "/search")
	if err != nil {
		return
	}
	body := resp.Body()
	searchResp := &QQmusicSearchResponse{}
	json.Unmarshal(body, searchResp)
	for _, song := range searchResp.Data.List {
		var ArtistName string
		for _, name := range song.Singer {
			if ArtistName != "" {
				ArtistName += ","
			}
			ArtistName += name.Name
		}

		songURL, errIn := qqCtx.GetMusicURLByID(ctx, song.Songmid, song.StrMediaMid)
		if errIn != nil {
			err = errIn
			return
		}
		if len(songURL) == 0 {
			continue
		}
		result = append(result, SearchMusicRes{
			ID:         song.Songmid,
			Name:       song.Songname,
			ArtistName: ArtistName,
			SongURL:    songURL,
			PicURL:     getAlbumPicURL(song.Albummid),
		})
	}
	return
}

// GetMusicURLByID 获取音乐的url
//
//	@receiver ctx
//	@param IDName
//	@return musicURL
//	@return err
func (qqCtx *QQmusicContext) GetMusicURLByID(ctx context.Context, mid, mediaMid string) (musicURL string, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	resp, err := requests.
		ReqTimestamp().
		SetFormDataFromValues(map[string][]string{
			"id":      {mid},
			"type":    {"128"},
			"mediaId": {mediaMid},
		}).Post(qqmusicBaseURL + "/song/url")
	if err != nil {
		return
	}
	body := resp.Body()
	music := &MusicURLId{}
	err = json.Unmarshal(body, music)
	musicURL = music.SongURL
	return
}

func getAlbumPicURL(albumMID string) (picURL string) {
	return qqmusicPicBaseURL + albumMID + ".jpg"
}
