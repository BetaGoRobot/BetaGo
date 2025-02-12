package neteaseapi

import (
	"cmp"
	"context"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/consts/ct"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	miniohelper "github.com/BetaGoRobot/BetaGo/utility/minio_helper"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/requests"
	"github.com/bytedance/sonic"
	"github.com/dlclark/regexp2"
	jsoniter "github.com/json-iterator/go"
	"github.com/kevinmatthe/zaplog"
	"go.opentelemetry.io/otel/attribute"
)

const netEaseQRTmpFile = "/data/tmp"

type CommentType string

const (
	CommentTypeSong  CommentType = "0"
	CommentTypeAlbum CommentType = "3"
)

func init() {
	// 测试环境，使用本地网易云代理
	if consts.IsTest {
		NetEaseAPIBaseURL = "http://192.168.31.74:3336"
	} else if consts.IsCluster {
		NetEaseAPIBaseURL = "http://kubernetes.default:3335"
		time.Sleep(time.Second * 10) // 等待本地网络启动
	} else if consts.IsCompose {
		NetEaseAPIBaseURL = "http://netease-api:3335"
	}

	startUpCtx := context.Background()
	NetEaseGCtx.TryGetLastCookie(startUpCtx)

	go func() {
		err := NetEaseGCtx.LoginNetEase(startUpCtx)
		if err != nil {
			log.ZapLogger.Info("error in init loginNetease", zaplog.Error(err))
			err = NetEaseGCtx.LoginNetEaseQR(startUpCtx)
			if err != nil {
				log.ZapLogger.Info("error in init loginNeteaseQR", zaplog.Error(err))
			}
		}
		for {
			if NetEaseGCtx.loginType == "qr" {
				if !NetEaseGCtx.CheckIfLogin(startUpCtx) {
					NetEaseGCtx.LoginNetEaseQR(startUpCtx)
				}
			} else {
				NetEaseGCtx.RefreshLogin(startUpCtx)
				if NetEaseGCtx.CheckIfLogin(startUpCtx) {
					NetEaseGCtx.SaveCookie(startUpCtx)
				} else {
					log.ZapLogger.Info("error in refresh login")
				}
			}
			time.Sleep(time.Second * 300)
		}
	}()
}

// GetDailyRecommendID 获取当前账号日推
//
//	@receiver ctx
//	@return musicIDs
//	@return err
func (neteaseCtx *NetEaseContext) GetDailyRecommendID() (musicIDs map[string]string, err error) {
	musicIDs = make(map[string]string)

	resp, err := requests.
		ReqTimestamp().
		SetCookies(neteaseCtx.cookies).
		Post(NetEaseAPIBaseURL + "/recommend/songs")
	if err != nil || resp.StatusCode() != 200 {
		return
	}

	body := resp.Body()
	music := dailySongs{}
	sonic.Unmarshal(body, &music)
	for index := range music.Data.DailySongs {
		musicIDs[strconv.Itoa(music.Data.DailySongs[index].ID)] = music.Data.DailySongs[index].Name
	}
	return
}

// GetMusicURLByIDs 依据ID获取URL/Name
//
//	@receiver ctx
//	@param IDName
//	@return InfoList
//	@return err
func (neteaseCtx *NetEaseContext) GetMusicURLByIDs(ctx context.Context, musicIDs []string) (musicIDURL map[string]string, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	musicIDURL = make(map[string]string)
	var fullIDs []string

	for _, ID := range musicIDs {
		if ID == "" {
			continue
		}
		fullIDs = append(fullIDs, ID)
	}

	r, err := consts.HttpClient.R().SetQueryParams(
		map[string]string{
			"id":        strings.Join(fullIDs, ","),
			"level":     "standard",
			"timestamp": fmt.Sprint(time.Now().UnixNano()),
		},
	).SetCookies(neteaseCtx.cookies).Post(NetEaseAPIBaseURL + "/song/url/v1")
	music := &musicList{}
	sonic.Unmarshal(r.Body(), music)

	for index := range music.Data {
		ID := strconv.Itoa(music.Data[index].ID)
		URL := music.Data[index].URL
		if URL == "" {
			log.ZapLogger.Warn("[PreUploadMusic] Get minio url failed...", zaplog.Error(err))
			continue
		}
		musicIDURL[ID] = URL
		go uploadMusic(ctx, URL, ID)
	}
	return
}

func uploadMusic(ctx context.Context, URL string, ID string) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("songID").String(ID))
	defer span.End()
	parsedURL, err := url.Parse(URL)
	if err != nil {
		log.ZapLogger.Warn("[PreUploadMusic] parsedURL failed...", zaplog.Error(err))
		return
	}
	_, err = miniohelper.Client().
		SetContext(ctx).
		SetBucketName("cloudmusic").
		SetFileFromURL(URL).
		SetObjName("music/" + ID + path.Ext(path.Base(parsedURL.Path))).
		SetContentType(ct.ContentTypeAudio).
		Upload()
	if err != nil {
		log.ZapLogger.Warn("[PreUploadMusic] Get minio url failed...", zaplog.Error(err))
	}
}

// GetMusicURLByID 依据ID获取URL/Name //TODO: replace this method more generic
//
//	@receiver ctx
//	@param IDName
//	@return InfoList
//	@return err
func (neteaseCtx *NetEaseContext) GetMusicURLByID(ctx context.Context, musicIDName []*MusicIDName) (InfoList []MusicInfo, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	var id string
	for _, m := range musicIDName {
		if id != "" {
			id += ","
		}
		id += m.ID
	}
	r, err := consts.HttpClient.R().SetQueryParams(
		map[string]string{
			"id":        id,
			"level":     "standard",
			"timestamp": fmt.Sprint(time.Now().UnixNano()),
		},
	).SetCookies(neteaseCtx.cookies).Post(NetEaseAPIBaseURL + "/song/url/v1")
	body := r.Body()
	music := musicList{}
	sonic.ConfigStd.Unmarshal(body, &music)

	for index := range music.Data {
		ID := strconv.Itoa(music.Data[index].ID)
		URL := music.Data[index].URL
		musicURL, err := miniohelper.Client().
			SetContext(ctx).
			SetBucketName("cloudmusic").
			SetFileFromURL(URL).
			SetObjName("music/" + ID + filepath.Ext(music.Data[index].URL)).
			SetContentType(ct.ContentTypeAudio).
			Upload()
		if err != nil {
			log.ZapLogger.Error("Get minio url failed, will use raw url", zaplog.Error(err))
		} else {
			URL = musicURL.String()
		}
		InfoList = append(InfoList, MusicInfo{
			ID:   ID,
			Name: musicIDName[index].Name,
			URL:  URL,
		})
	}
	return
}

func (neteaseCtx *NetEaseContext) GetMusicURL(ctx context.Context, ID string) (url string, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("songID").String(ID))
	defer span.End()

	r, err := consts.HttpClient.R().SetQueryParams(
		map[string]string{
			"id":        ID,
			"level":     "standard",
			"timestamp": fmt.Sprint(time.Now().UnixNano()),
		},
	).SetCookies(neteaseCtx.cookies).Post(NetEaseAPIBaseURL + "/song/url/v1")
	body := r.Body()
	music := musicList{}
	err = sonic.Unmarshal(body, &music)
	if err != nil {
		return "", err
	}
	URL := music.Data[0].URL
	u, err := miniohelper.Client().
		SetContext(ctx).
		SetBucketName("cloudmusic").
		SetFileFromURL(URL).
		SetObjName("music/" + ID + filepath.Ext(URL)).
		SetContentType(ct.ContentTypeAudio).
		Upload()
	if err != nil {
		log.ZapLogger.Error("Get minio url failed, will use raw url", zaplog.Error(err))
	} else {
		URL = u.String()
	}
	return URL, err
}

func (neteaseCtx *NetEaseContext) GetDetail(ctx context.Context, musicID string) (musicDetail *MusicDetail) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("songID").String(musicID))
	defer span.End()

	resp, err := consts.HttpClient.R().
		SetFormDataFromValues(
			map[string][]string{
				"ids": {musicID},
			},
		).
		SetCookies(neteaseCtx.cookies).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().UnixNano())).
		Post(NetEaseAPIBaseURL + "/song/detail")
	if err != nil {
		log.ZapLogger.Info(err.Error())
	}
	musicDetail = &MusicDetail{}
	err = jsoniter.Unmarshal(resp.Body(), musicDetail)
	if err != nil {
		log.ZapLogger.Info(err.Error())
	}
	if len(musicDetail.Songs) == 0 {
		return nil
	}
	for _, song := range musicDetail.Songs {
		picURL := song.Al.PicURL
		_, err = miniohelper.Client().
			SetContext(ctx).
			SetBucketName("cloudmusic").
			SetFileFromURL(picURL).
			SetObjName("picture/" + musicID + filepath.Ext(picURL)).
			SetContentType(ct.ContentTypeImgJPEG).
			Upload()
		if err != nil {
			log.ZapLogger.Error(err.Error())
		}
	}
	return
}

func (neteaseCtx *NetEaseContext) GetLyrics(ctx context.Context, songID string) (lyrics string, lyricsURL string) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("songID").String(songID))
	defer span.End()

	resp, err := consts.HttpClient.R().
		SetFormDataFromValues(
			map[string][]string{
				"id": {songID},
			},
		).
		SetCookies(neteaseCtx.cookies).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().UnixNano())).
		Post(NetEaseAPIBaseURL + "/lyric")
	if err != nil {
		log.ZapLogger.Info(err.Error())
	}
	searchLyrics := &SearchLyrics{}
	body := string(resp.Body())
	err = sonic.UnmarshalString(body, searchLyrics)
	if err != nil {
		log.ZapLogger.Info(err.Error())
		return
	}
	l, err := miniohelper.Client().
		SetContext(ctx).
		SetBucketName("cloudmusic").
		SetFileFromString(body).
		SetObjName("lyrics/" + songID + ".json").
		SetContentType(ct.ContentTypePlainText).
		Upload()
	if err != nil {
		log.ZapLogger.Error("upload lyrics failed", zaplog.Error(err))
		return
	}
	lyricsMerged := mergeLyrics(searchLyrics.Lrc.Lyric, searchLyrics.Tlyric.Lyric)
	return lyricsMerged, l.String()
}

var lyricsRepattern = regexp2.MustCompile(`\[(?P<time>.*)\](?P<line>.*)`, regexp2.RE2)

func mergeLyrics(lyrics, translatedLyrics string) string {
	lyricsMap := map[string]string{}
	lines := strings.Split(lyrics, "\n")
	for _, line := range lines {
		match, err := lyricsRepattern.FindStringMatch(line)
		if err != nil {
			panic(err)
		}
		if match != nil {
			if lyric := match.GroupByName("line").String(); lyric != "" {
				lyricsMap[match.GroupByName("time").String()] = lyric + "\n"
			}
		}
	}
	for _, translatedLine := range strings.Split(translatedLyrics, "\n") {
		match, err := lyricsRepattern.FindStringMatch(translatedLine)
		if err != nil {
			panic(err)
		}
		if match != nil {
			if lyric := match.GroupByName("line").String(); lyric != "" {
				lyricsMap[match.GroupByName("time").String()] += lyric + "\n"
			}
		}
	}
	resStr := ""
	type lineStruct struct {
		time string
		line string
	}
	lyricsLines := make([]*lineStruct, 0)
	for time, line := range lyricsMap {
		if line == "" {
			continue
		}
		lyricsLines = append(lyricsLines, &lineStruct{
			time, line,
		})
	}
	slices.SortFunc(lyricsLines, func(i, j *lineStruct) int {
		return cmp.Compare(i.time, j.time)
	})
	for _, line := range lyricsLines {
		resStr += line.line + "\n"
	}
	return resStr
}

func (neteaseCtx *NetEaseContext) AsyncGetSearchRes(ctx context.Context, searchRes SearchMusic) (result []*SearchMusicItem, err error) {
	sResChan := make(chan *SearchMusicItem, len(searchRes.Result.Songs))

	go neteaseCtx.asyncGetSearchRes(ctx, searchRes, err, sResChan)

	m := asyncUploadPics(ctx, searchRes)
	for res := range sResChan {
		res.ImageKey = m[res.ID]
		result = append(result, res)
	}
	return
}

// SearchMusicByKeyWord 通过关键字搜索歌曲
//
//	@receiver neteaseCtx
//	@param ctx
//	@param keywords
//	@return result
//	@return err
func (neteaseCtx *NetEaseContext) SearchMusicByKeyWord(ctx context.Context, keywords ...string) (result []*SearchMusicItem, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("keywords").StringSlice(keywords))
	defer span.End()

	resp1, err := consts.HttpClient.R().
		SetFormDataFromValues(
			map[string][]string{
				"limit":    {"5"},
				"type":     {"1"},
				"keywords": {strings.Join(keywords, " ")},
			},
		).
		SetCookies(neteaseCtx.cookies).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().UnixNano())).
		Post(NetEaseAPIBaseURL + "/cloudsearch")
	if err != nil {
		log.ZapLogger.Info(err.Error())
	}

	searchRes := SearchMusic{}
	jsoniter.Unmarshal(resp1.Body(), &searchRes)

	result, err = neteaseCtx.AsyncGetSearchRes(ctx, searchRes)
	if err != nil {
		return
	}

	return
}

func (neteaseCtx *NetEaseContext) SearchPlaylistByKeyWord(ctx context.Context, keywords ...string) {
	return
}

// SearchAlbumByKeyWord  通过关键字搜索歌曲
//
//	@receiver neteaseCtx *NetEaseContext
//	@param ctx context.Context
//	@param keywords ...string
//	@return result []*Album
//	@return err error
//	@author heyuhengmatt
//	@update 2024-08-07 08:46:58
func (neteaseCtx *NetEaseContext) SearchAlbumByKeyWord(ctx context.Context, keywords ...string) (result []*Album, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("keywords").StringSlice(keywords))
	defer span.End()

	resp1, err := consts.HttpClient.R().
		SetFormDataFromValues(
			map[string][]string{
				"limit":    {"5"},
				"type":     {"10"},
				"keywords": {strings.Join(keywords, " ")},
			},
		).
		SetCookies(neteaseCtx.cookies).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().UnixNano())).
		Post(NetEaseAPIBaseURL + "/cloudsearch")
	if err != nil {
		log.ZapLogger.Info(err.Error())
	}

	searchRes := searchAlbumResult{}
	err = sonic.Unmarshal(resp1.Body(), &searchRes)
	if err != nil {
		log.ZapLogger.Info(err.Error())
	}
	result = searchRes.Result.Albums
	return
}

// GetAlbumDetail 通过关键字搜索歌曲
//
//	@receiver neteaseCtx *NetEaseContext
//	@param ctx context.Context
//	@param albumID
//	@return result []*Album
//	@return err error
func (neteaseCtx *NetEaseContext) GetAlbumDetail(ctx context.Context, albumID string) (result *AlbumDetail, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("albumID").String(albumID))
	defer span.End()

	resp1, err := consts.HttpClient.R().
		SetFormDataFromValues(
			map[string][]string{
				"id": {albumID},
			},
		).
		SetCookies(neteaseCtx.cookies).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().UnixNano())).
		Post(NetEaseAPIBaseURL + "/album")
	if err != nil {
		log.ZapLogger.Info(err.Error())
	}

	searchRes := AlbumDetail{}
	err = sonic.Unmarshal(resp1.Body(), &searchRes)
	if err != nil {
		log.ZapLogger.Info(err.Error())
	}

	return &searchRes, err
}

func asyncUploadPics(ctx context.Context, musicInfos SearchMusic) map[string]string {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	var (
		c  = make(chan [2]string)
		wg = &sync.WaitGroup{}
		m  = make(map[string]string, 1)
	)
	go func(ctx context.Context) {
		defer close(c)
		defer wg.Wait()

		for _, m := range musicInfos.Result.Songs {
			wg.Add(1)
			go uploadPicWorker(ctx, wg, m.Al.PicURL, m.ID, c)
		}
	}(ctx)
	for res := range c {
		m[res[1]] = res[0]
	}
	return m
}

func uploadPicWorker(ctx context.Context, wg *sync.WaitGroup, url string, musicID int, c chan [2]string) bool {
	defer wg.Done()
	imgKey, _, err := larkutils.UploadPicAllinOne(ctx, url, strconv.Itoa(musicID), true)
	if err != nil {
		log.ZapLogger.Error("upload pic to lark error", zaplog.Error(err))
		return true
	}
	c <- [2]string{imgKey, strconv.Itoa(musicID)}
	return false
}

func (neteaseCtx *NetEaseContext) asyncGetSearchRes(ctx context.Context, searchMusic SearchMusic, err error, urlChan chan *SearchMusicItem) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	defer close(urlChan)

	songMap := make(map[string]*Song)
	songList := make([]string, len(searchMusic.Result.Songs))
	for i, song := range searchMusic.Result.Songs {
		songMap[strconv.Itoa(song.ID)] = &searchMusic.Result.Songs[i]
		songList[i] = strconv.Itoa(song.ID)
	}

	musicIDURL, err := neteaseCtx.GetMusicURLByIDs(ctx, songList)
	for _, ID := range songList {
		songInfo := songMap[ID]
		urlChan <- &SearchMusicItem{
			// Index:      index,
			ID:         ID,
			Name:       songInfo.Name,
			ArtistName: genArtistName(songInfo),
			PicURL:     songInfo.Al.PicURL,
			SongURL:    musicIDURL[ID],
		}

	}

	return
}

func (neteaseCtx *NetEaseContext) GetComment(ctx context.Context, commentType CommentType, id string) (res *CommentResult, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("commentType").String(string(commentType)))
	defer span.End()

	resp, err := consts.HttpClient.R().
		SetCookies(neteaseCtx.cookies).
		SetFormDataFromValues(
			map[string][]string{
				"id":       {id},
				"pageSize": {"1"},
				"pageNo":   {"1"},
				"sortType": {"2"},
				"type":     {string(commentType)},
			},
		).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().UnixNano())).
		Post(NetEaseAPIBaseURL + "/comment/new/")
	if err != nil {
		log.ZapLogger.Info(err.Error())
	}
	res = &CommentResult{}
	err = sonic.Unmarshal(resp.Body(), res)
	if err != nil {
		return
	}
	return
}

func genArtistName(song *Song) (artistName string) {
	artistList := make([]string, 0, len(song.Ar))
	for _, s := range song.Ar {
		if s.Name == "" {
			continue
		}
		artistList = append(artistList, s.Name)
	}
	return strings.Join(artistList, ", ")
}

// GetNewRecommendMusic 获得新的推荐歌曲
//
//	@receiver ctx
//	@return res
//	@return err
func (neteaseCtx *NetEaseContext) GetNewRecommendMusic() (res []SearchMusicItem, err error) {
	resp, err := consts.HttpClient.R().SetFormDataFromValues(
		map[string][]string{
			"limit": {"5"},
		},
	).Post(NetEaseAPIBaseURL + "/personalized/newsong")
	if err != nil {
		return
	}

	music := &GlobRecommendMusicRes{}
	sonic.Unmarshal(resp.Body(), music)
	for _, result := range music.Result {
		var ArtistName string
		for _, name := range result.Song.Artists {
			if ArtistName != "" {
				ArtistName += ","
			}
			ArtistName += name.Name
		}
		SongURL, errIn := neteaseCtx.GetMusicURLByID(context.Background(), []*MusicIDName{
			{strconv.Itoa(result.Song.ID), result.Song.Name},
		})
		if errIn != nil {
			err = errIn
			return
		}
		if len(SongURL) == 0 {
			continue
		}
		res = append(res, SearchMusicItem{
			ID:         strconv.Itoa(result.Song.ID),
			Name:       result.Song.Name,
			ArtistName: ArtistName,
			PicURL:     result.PicURL,
			SongURL:    SongURL[0].URL,
		})
	}
	return
}
