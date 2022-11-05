package qqmusicapi

import (
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/httptool"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	if betagovar.IsTest {
		qqmusicBaseURL = "http://localhost:3300"
	} else if betagovar.IsCluster {
		qqmusicBaseURL = "http://qqmusic-api-ix-chart.ix-qqmusic-api:3300"
	}
}
func autoRefreshLogin() {
	for {
		time.Sleep(time.Minute * 5)
		httptool.PostWithTimestamp(httptool.RequestInfo{
			URL: qqmusicBaseURL + "/user/refresh",
		})
	}
}
func init() {
	//获取存储的Cookie
	_, err := httptool.PostWithTimestamp(httptool.RequestInfo{
		URL: qqmusicBaseURL + "/user/cookie",
	})
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
func (ctx *QQmusicContext) SearchMusic(keywords []string) (result []SearchMusicRes, err error) {
	resp, err := httptool.PostWithTimestamp(
		httptool.RequestInfo{
			URL: qqmusicBaseURL + "/search",
			Params: map[string][]string{
				"key":      {strings.Join(keywords, " ")},
				"pageNo":   {"1"},
				"pageSize": {"3"},
				"t":        {"0"},
			},
		},
	)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
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

		songURL, errIn := ctx.GetMusicURLByID(song.Songmid, song.StrMediaMid)
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
func (ctx *QQmusicContext) GetMusicURLByID(mid, mediaMid string) (musicURL string, err error) {

	resp, err := httptool.PostWithTimestamp(
		httptool.RequestInfo{
			URL: qqmusicBaseURL + "/song/url",
			Params: map[string][]string{
				"id":      {mid},
				"type":    {"128"},
				"mediaId": {mediaMid},
			},
		},
	)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	music := &MusicURLId{}
	err = json.Unmarshal(body, music)
	musicURL = music.SongURL
	return
}
func getAlbumPicURL(albumMID string) (picURL string) {
	return qqmusicPicBaseURL + albumMID + ".jpg"
}
