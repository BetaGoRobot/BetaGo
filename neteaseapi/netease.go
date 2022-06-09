package neteaseapi

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/BetaGoRobot/BetaGo/httptool"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	IsTest = os.Getenv("IS_TEST")
)

func init() {
	if IsTest == "true" {
		NetEaseAPIBaseURL = "http://localhost:3335"
	}
}

// LoginNetEase 返回cookie
//  @receiver ctx
//  @return err
func (ctx *NetEaseContext) LoginNetEase() (err error) {
	if phoneNum, password := os.Getenv("NETEASE_PHONE"), os.Getenv("NETEASE_PASSWORD"); phoneNum == "" && password == "" {
		log.Println("Empty NetEase account and password")
		return
	}

	resp, err := httptool.PostWithParams(
		httptool.RequestInfo{
			URL: NetEaseAPIBaseURL + "/login/cellphone",
			Params: map[string][]string{
				"phone":    {os.Getenv("NETEASE_PHONE")},
				"password": {os.Getenv("NETEASE_PASSWORD")},
			},
		},
	)
	if err != nil || resp.StatusCode != 200 {
		log.Printf("%#v", resp)
	}
	ctx.cookies = resp.Cookies()
	return
}

// GetDailyRecommendID 获取当前账号日推
//  @receiver ctx
//  @return musicIDs
//  @return err
func (ctx *NetEaseContext) GetDailyRecommendID() (musicIDs map[string]string, err error) {
	musicIDs = make(map[string]string)
	resp, err := httptool.PostWithParamsWithTimestamp(
		httptool.RequestInfo{
			URL:     NetEaseAPIBaseURL + "/recommend/songs",
			Cookies: ctx.cookies,
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
	music := dailySongs{}
	json.Unmarshal(body, &music)
	for index := range music.Data.DailySongs {
		musicIDs[strconv.Itoa(music.Data.DailySongs[index].ID)] = music.Data.DailySongs[index].Name
	}
	return
}

// GetMusicURLByID 依据ID获取URL/Name
//  @receiver ctx
//  @param IDName
//  @return InfoList
//  @return err
func (ctx *NetEaseContext) GetMusicURLByID(IDName map[string]string) (InfoList []MusicInfo, err error) {
	var id string
	for key := range IDName {
		if id != "" {
			id += ","
		}
		id += key
	}
	resp, err := httptool.PostWithParamsWithTimestamp(
		httptool.RequestInfo{
			URL:     NetEaseAPIBaseURL + "/song/url",
			Cookies: ctx.cookies,
			Params:  map[string][]string{"id": {id}, "br": {"128000"}},
		})
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
	music := musicList{}
	json.Unmarshal(body, &music)
	for index := range music.Data {
		ID := strconv.Itoa(music.Data[index].ID)
		InfoList = append(InfoList, MusicInfo{
			ID:   ID,
			Name: IDName[ID],
			URL:  music.Data[index].URL,
		})
	}
	return
}

// SearchMusicByKeyWord 通过关键字搜索歌曲
//  @receiver ctx
//  @param keywords
//  @return result
//  @return err
func (ctx *NetEaseContext) SearchMusicByKeyWord(keywords []string) (result []SearchMusicRes, err error) {
	resp, err := httptool.PostWithParamsWithTimestamp(
		httptool.RequestInfo{
			URL:     NetEaseAPIBaseURL + "/cloudsearch",
			Cookies: ctx.cookies,
			Params: map[string][]string{
				"limit":    {"3"},
				"type":     {"1"},
				"keywords": {strings.Join(keywords, " ")},
			},
		})
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
	searchMusic := searchMusic{}
	jsoniter.Unmarshal(body, &searchMusic)
	for _, song := range searchMusic.Result.Songs {
		var ArtistName string
		for _, name := range song.Ar {
			if ArtistName != "" {
				ArtistName += ","
			}
			ArtistName += name.Name
		}
		SongURL, errIn := ctx.GetMusicURLByID(map[string]string{strconv.Itoa(song.ID): song.Name})
		if errIn != nil {
			err = errIn
			return
		}
		if len(SongURL) == 0 {
			continue
		}
		result = append(result, SearchMusicRes{
			ID:         strconv.Itoa(song.ID),
			Name:       song.Name,
			ArtistName: ArtistName,
			PicURL:     song.Al.PicURL,
			SongURL:    SongURL[0].URL,
		})
	}
	return
}

// GetNewRecommendMusic 获得新的推荐歌曲
//  @receiver ctx
//  @return res
//  @return err
func (ctx *NetEaseContext) GetNewRecommendMusic() (res []SearchMusicRes, err error) {
	resp, err := httptool.PostWithParamsWithTimestamp(
		httptool.RequestInfo{
			URL: NetEaseAPIBaseURL + "/personalized/newsong",
			Params: map[string][]string{
				"limit": {"3"},
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
	music := &GlobRecommendMusicRes{}
	json.Unmarshal(body, music)
	for _, result := range music.Result {
		var ArtistName string
		for _, name := range result.Song.Artists {
			if ArtistName != "" {
				ArtistName += ","
			}
			ArtistName += name.Name
		}
		SongURL, errIn := ctx.GetMusicURLByID(map[string]string{strconv.Itoa(result.Song.ID): result.Song.Name})
		if err != nil {
			err = errIn
			return
		}
		if len(SongURL) == 0 {
			continue
		}
		res = append(res, SearchMusicRes{
			ID:         strconv.Itoa(result.Song.ID),
			Name:       result.Song.Name,
			ArtistName: ArtistName,
			PicURL:     result.PicURL,
			SongURL:    SongURL[0].URL,
		})
	}
	return
}
