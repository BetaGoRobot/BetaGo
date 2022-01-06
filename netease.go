package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// loginNetEase 返回cookie
//  @return cookies
func (ctx *NetEaseContext) loginNetEase() (err error) {
	if phoneNum, password := os.Getenv("NETEASE_PHONE"), os.Getenv("NETEASE_PASSWORD"); phoneNum == "" && password == "" {
		log.Println("Empty NetEase account and password")
		return
	}
	resp, err := GetWithCookieParams(
		GetRequestInfo{
			URL:     NetEaseAPIBaseURL + "/login/cellphone",
			cookies: nil,
			params: map[string]string{
				"phone":    os.Getenv("NETEASE_PHONE"),
				"password": os.Getenv("NETEASE_PASSWORD"),
			},
		},
	)
	if err != nil {
		return
	}
	ctx.cookies = resp.Cookies()
	return
}

// getDailyRecommendID 获取当前账号日推
func (ctx *NetEaseContext) getDailyRecommendID() (musicIDs map[string]string, err error) {
	musicIDs = make(map[string]string)
	resp, err := GetWithCookieParams(
		GetRequestInfo{
			URL:     NetEaseAPIBaseURL + "/recommend/songs",
			cookies: ctx.cookies,
		},
	)
	if err != nil {
		return
	}
	defer resp.Body.Close()

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

// getMusicURLByID 依据ID获取URL/Name
func (ctx *NetEaseContext) getMusicURLByID(IDName map[string]string) (InfoList []musicInfo, err error) {
	var id string
	for key := range IDName {
		if id != "" {
			id += ","
		}
		id += key
	}
	resp, err := GetWithCookieParams(GetRequestInfo{
		URL:     NetEaseAPIBaseURL + "/song/url",
		cookies: ctx.cookies,
		params:  map[string]string{"id": id, "br": "320000"},
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
		InfoList = append(InfoList, musicInfo{ID: ID, Name: IDName[ID], URL: music.Data[index].URL})
	}
	return
}

func (ctx *NetEaseContext) searchMusicByKeyWord(keywords []string) (result []searchMusicRes, err error) {
	resp, err := GetWithCookieParams(GetRequestInfo{
		URL:     NetEaseAPIBaseURL + "/cloudsearch",
		cookies: ctx.cookies,
		params: map[string]string{
			"keywords": strings.Join(keywords, " "),
			"limit":    "3",
			"type":     "1",
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
	json.Unmarshal(body, &searchMusic)
	for _, song := range searchMusic.Result.Songs {
		var ArtistName string
		for _, name := range song.Ar {
			if ArtistName != "" {
				ArtistName += ","
			}
			ArtistName += name.Name
		}
		SongURL, errIn := ctx.getMusicURLByID(map[string]string{strconv.Itoa(song.ID): song.Name})
		if errIn != nil {
			err = errIn
			return
		}
		if len(SongURL) == 0 {
			continue
		}
		result = append(result, searchMusicRes{
			ID:         strconv.Itoa(song.ID),
			Name:       song.Name,
			ArtistName: ArtistName,
			PicURL:     song.Al.PicURL,
			SongURL:    SongURL[0].URL,
		})
	}
	return
}
