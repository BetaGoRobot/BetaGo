package neteaseapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/gotify"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/requests"
	"github.com/bytedance/sonic"
	jsoniter "github.com/json-iterator/go"
	"github.com/kevinmatthe/zaplog"
	"go.opentelemetry.io/otel/attribute"
)

const netEaseQRTmpFile = "/data/tmp"

func init() {
	// 测试环境，使用本地网易云代理
	if consts.IsTest {
		NetEaseAPIBaseURL = "http://192.168.31.74:3335"
	} else if consts.IsCluster {
		NetEaseAPIBaseURL = "http://kubernetes.default:3335"
	}
	time.Sleep(time.Second * 10) // 等待本地网络启动
	startUpCtx := context.Background()
	NetEaseGCtx.TryGetLastCookie(startUpCtx)
	err := NetEaseGCtx.LoginNetEase(startUpCtx)
	if err != nil {
		log.ZapLogger.Info("error in init loginNetease", zaplog.Error(err))
		err = NetEaseGCtx.LoginNetEaseQR(startUpCtx)
		if err != nil {
			log.ZapLogger.Info("error in init loginNeteaseQR", zaplog.Error(err))
		}
	}

	go func() {
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
			time.Sleep(time.Second * 60)
		}
	}()
}

// RefreshLogin 刷新登录
//
//	@receiver ctx
//	@return error
func (neteaseCtx *NetEaseContext) RefreshLogin(ctx context.Context) error {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	log.ZapLogger.Info("RefreshLogin...", zaplog.String("traceID", span.SpanContext().TraceID().String()))

	resp, err := consts.HttpClient.R().
		SetCookies(neteaseCtx.cookies).
		Post(NetEaseAPIBaseURL + "/login/refresh")

	if err != nil || (resp != nil && resp.StatusCode() != 200) {
		log.SugerLogger.Errorf("%s\n", string(resp.Body()))
		return err
	}
	respMap := make(map[string]interface{})
	err = sonic.Unmarshal(resp.Body(), &resp)
	if err != nil {
		return err
	}

	if code, ok := respMap["code"]; ok {
		if code != 200 {
			return fmt.Errorf("RefreshLogin error, with msg %v", respMap["msg"])
		}
	}

	if neteaseCtx.cookies == nil {
		neteaseCtx.cookies = make([]*http.Cookie, 0)
	}
	newCookies := resp.Cookies()
	if len(newCookies) > 0 {
		neteaseCtx.cookies = newCookies
		neteaseCtx.SaveCookie(ctx)
	}

	return err
}

func (neteaseCtx *NetEaseContext) getUniKey(ctx context.Context) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	log.ZapLogger.Info("getUniKey...", zaplog.String("traceID", span.SpanContext().TraceID().String()))

	resp, err := requests.Req().Post(NetEaseAPIBaseURL + "/login/qr/key")
	if err != nil || resp.StatusCode() != 200 {
		if err == nil {
			err = fmt.Errorf("LoginNetEaseQR error, StatusCode %d", resp.StatusCode)
		}
		return
	}
	data := resp.Body()
	respMap := make(map[string]interface{})
	if err = sonic.Unmarshal(data, &respMap); err != nil {
		return
	}
	neteaseCtx.qrStruct.uniKey = respMap["data"].(map[string]interface{})["unikey"].(string)
	return
}

func (neteaseCtx *NetEaseContext) getQRBase64(ctx context.Context) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	log.ZapLogger.Info("getQRBase64...", zaplog.String("traceID", span.SpanContext().TraceID().String()))

	resp, err := requests.
		ReqTimestamp().
		SetFormDataFromValues(
			map[string][]string{
				"key":   {neteaseCtx.qrStruct.uniKey},
				"qrimg": {"1"},
			}).
		Post(NetEaseAPIBaseURL + "/login/qr/create")
	if err != nil || resp.StatusCode() != 200 {
		if err == nil {
			err = fmt.Errorf("LoginNetEaseQR error, StatusCode %d", resp.StatusCode)
		}
		return
	}
	data := (resp.Body())
	respMap := make(map[string]interface{})
	if err = sonic.Unmarshal(data, &respMap); err != nil {
		return
	}
	neteaseCtx.qrStruct.qrBase64 = respMap["data"].(map[string]interface{})["qrimg"].(string)
	return
}

func (neteaseCtx *NetEaseContext) checkQRStatus(ctx context.Context) (err error) {
	if !neteaseCtx.qrStruct.isOutDated {
		once := &sync.Once{}
		for {

			time.Sleep(time.Second * 1)
			resp, err := consts.HttpClient.R().
				SetFormData(map[string]string{"key": neteaseCtx.qrStruct.uniKey}).
				SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
				SetContext(ctx).
				Post(NetEaseAPIBaseURL + "/login/qr/check")

			if err != nil || resp.StatusCode() != 200 {
				if err == nil {
					return fmt.Errorf("LoginNetEaseQR error, StatusCode %d", resp.StatusCode())
				}
				return err
			}
			data := resp.Body()
			respMap := make(map[string]interface{})
			if err = sonic.Unmarshal(data, &respMap); err != nil {
				return err
			}
			switch respMap["code"].(float64) {
			case 801:
				once.Do(func() { log.ZapLogger.Info("Waiting for scan") })
			case 800:
				once.Do(func() {
					log.ZapLogger.Info("二维码已失效")
					neteaseCtx.qrStruct.isOutDated = true
				})
				return err
			case 802:
				once.Do(func() { log.ZapLogger.Info("扫描未确认") })
			case 803:
				log.ZapLogger.Info("登陆成功！")
				neteaseCtx.cookies = resp.Cookies()
				neteaseCtx.SaveCookie(ctx)
				neteaseCtx.loginType = "qr"
				gotify.SendMessage(ctx, "网易云登录", "登陆成功！", 7)
				return nil
			}
		}
	}
	return
}

// LoginNetEaseQR 通过二维码获取登陆Cookie
//
//	@receiver ctx
//	@return err
func (neteaseCtx *NetEaseContext) LoginNetEaseQR(ctx context.Context) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	log.ZapLogger.Info("LoginNetEaseQR...", zaplog.String("traceID", span.SpanContext().TraceID().String()))
	neteaseCtx.getUniKey(ctx)

	neteaseCtx.getQRBase64(ctx)
	linkURL, err := utility.UploadFileToCos(SaveQRImg(ctx, neteaseCtx.qrStruct.qrBase64))
	if err != nil {
		return err
	}
	gotify.SendMessage(ctx, "网易云登录", fmt.Sprintf("![QRCode](%s)", linkURL), 7)
	go neteaseCtx.checkQRStatus(ctx)
	return
}

// SaveQRImg 保存二维码图片
//
//	@param imgBase64
//	@return filename
func SaveQRImg(ctx context.Context, imgBase64 string) (filename string) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	i := strings.Index(imgBase64, ",")
	d := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgBase64[i+1:]))
	filename = filepath.Join(netEaseQRTmpFile, fmt.Sprintf("qr_%d.jpg", time.Now().Unix()))
	os.MkdirAll(netEaseQRTmpFile, 0o777)
	f, err := os.Create(filename)
	if err != nil {
		log.ZapLogger.Info("error in create qr img", zaplog.Error(err))
		return ""
	}
	defer f.Close()
	_, err = io.Copy(f, d)
	if err != nil {
		log.ZapLogger.Info("error in copy qr img", zaplog.Error(err))
		return ""
	}
	return
}

// LoginNetEase 获取登陆Cookie
//
//	@receiver ctx
//	@return err
func (neteaseCtx *NetEaseContext) LoginNetEase(ctx context.Context) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	log.ZapLogger.Info("LoginNetEase...", zaplog.String("traceID", span.SpanContext().TraceID().String()))
	if len(neteaseCtx.cookies) > 0 {
		return
	}
	if phoneNum, password := env.NETEASE_EMAIL, env.NETEASE_PASSWORD; phoneNum == "" && password == "" {
		log.ZapLogger.Info("Empty NetEase account and password")
		return
	}
	// !Step1:检查登陆状态
	if neteaseCtx.CheckIfLogin(ctx) {
		log.ZapLogger.Info("Already login")
		// 已登陆，刷新登陆
		err = neteaseCtx.RefreshLogin(ctx)
		return
	}

	// !Step2:未登陆，启动登陆
	resp, err := consts.HttpClient.R().
		SetCookies(neteaseCtx.cookies).
		SetFormData(
			map[string]string{
				"email":    env.NETEASE_EMAIL,
				"password": env.NETEASE_PASSWORD,
			},
		).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().Unix())).
		Post(NetEaseAPIBaseURL + "/login")
	if err != nil || resp.StatusCode() != 200 {
		if err == nil {
			err = fmt.Errorf("LoginNetEase error, with msg %v, StatusCode %d", string(resp.Body()), resp.StatusCode())
		}
		return
	}
	neteaseCtx.cookies = resp.Cookies()
	neteaseCtx.SaveCookie(ctx)
	return
}

// CheckIfLogin 检查是否登陆
//
//	@receiver ctx
//	@return bool
func (neteaseCtx *NetEaseContext) CheckIfLogin(ctx context.Context) bool {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	log.ZapLogger.Info("ChekIfLogin...", zaplog.String("traceID", span.SpanContext().TraceID().String()))

	resp, err := consts.HttpClient.R().
		SetCookies(neteaseCtx.cookies).
		SetContext(ctx).
		SetQueryParam("timestamp", fmt.Sprint(time.Now().UnixNano())).
		Get(NetEaseAPIBaseURL + "/login/status")
	if err != nil || resp.StatusCode() != 200 {
		log.SugerLogger.Errorf("%#v\n", resp)
		return false
	}
	data := resp.Body()
	loginStatus := LoginStatusStruct{}
	if err = sonic.Unmarshal(data, &loginStatus); err != nil {
		log.ZapLogger.Info("error in unmarshal loginStatus", zaplog.Error(err))
	} else {
		if loginStatus.Data.Account != nil {
			return true
		}
		return false
	}
	return false
}

// TryGetLastCookie 获取初始化Cookie
//
//	@receiver ctx
func (neteaseCtx *NetEaseContext) TryGetLastCookie(ctx context.Context) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	f, err := os.Open("/data/last_cookie.json")
	if err != nil {
		log.ZapLogger.Info("error in open last_cookie.json", zaplog.Error(err))
		return
	}
	defer f.Close()
	cookieData := make([]byte, 0)
	cookieData, err = io.ReadAll(f)
	if len(cookieData) == 0 {
		log.ZapLogger.Info("No cookieData, skip json marshal")
		return
	}
	cookie := make(map[string]string)

	if err = sonic.Unmarshal(cookieData, &cookie); err != nil {
		log.ZapLogger.Info("error in unmarshal cookieData", zaplog.Error(err))
	}
	for k, v := range cookie {
		neteaseCtx.cookies = append(neteaseCtx.cookies, &http.Cookie{Name: k, Value: v})
	}
	neteaseCtx.loginType = "qr"
}

// SaveCookie 保存Cookie
//
//	@receiver ctx
func (neteaseCtx *NetEaseContext) SaveCookie(ctx context.Context) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	if neteaseCtx.cookies == nil && len(neteaseCtx.cookies) == 0 {
		return
	}
	f, err := os.OpenFile("/data/last_cookie.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		log.ZapLogger.Info("error in open last_cookie.json", zaplog.Error(err))
		return
	}
	defer f.Close()

	toWriteMap := make(map[string]string)
	for _, cookie := range neteaseCtx.cookies {
		toWriteMap[cookie.Name] = cookie.Value
	}
	cookieData, err := sonic.Marshal(toWriteMap)
	if err != nil {
		log.ZapLogger.Error(err.Error())
	}
	f.Write(cookieData)
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

// GetMusicURLByID 依据ID获取URL/Name
//
//	@receiver ctx
//	@param IDName
//	@return InfoList
//	@return err
func (neteaseCtx *NetEaseContext) GetMusicURLByID(ctx context.Context, IDName map[string]string) (InfoList []MusicInfo, err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	var id string
	for key := range IDName {
		if id != "" {
			id += ","
		}
		id += key
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
	sonic.Unmarshal(body, &music)
	for index := range music.Data {
		ID := strconv.Itoa(music.Data[index].ID)
		URL := music.Data[index].URL
		musicUrl, err := utility.MinioUploadFileFromURL(
			ctx,
			"cloudmusic",
			music.Data[index].URL,
			"music/"+ID+filepath.Ext(music.Data[index].URL),
			"audio/mpeg;charset=UTF-8",
		)
		if err != nil {
			log.ZapLogger.Error("Get minio url failed, will use raw url", zaplog.Error(err))
		} else {
			URL = musicUrl.String()
		}
		InfoList = append(InfoList, MusicInfo{
			ID:   ID,
			Name: IDName[ID],
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
	u, err := utility.MinioUploadFileFromURL(ctx, "cloudmusic", URL, "music/"+ID+filepath.Ext(URL), "audio/mpeg;charset=UTF-8")
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
		_, err = utility.MinioUploadFileFromURL(
			ctx,
			"cloudmusic",
			picURL,
			"picture/"+musicID+filepath.Ext(picURL),
			"image/jpg",
		)
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
	err = jsoniter.Unmarshal(resp.Body(), searchLyrics)
	if err != nil {
		log.ZapLogger.Info(err.Error())
	}
	// lyricJson, err := utility.ExtractLyrics(searchLyrics.Lrc.Lyric)
	// if err != nil {
	// 	log.ZapLogger.Warn("extract lyrics error", zaplog.Error(err))
	// }
	// lyricJson, err := sonic.MarshalString(map[string]interface{}{"lrc": map[string]interface{}{
	// 	"lyric": searchLyrics.Lrc.Lyric,
	// }})
	lyricJson := string(resp.Body())
	if err != nil {
		log.ZapLogger.Error("marshal lyrics error", zaplog.Error(err))
	}
	l, err := utility.MinioUploadTextFile(ctx, "cloudmusic", lyricJson, "lyrics/"+songID+".json", "text/plain")
	if err != nil {
		log.ZapLogger.Error("upload lyrics failed", zaplog.Error(err))
		return
	}

	return searchLyrics.Lrc.Lyric, l.String()
}

// SearchMusicByKeyWord 通过关键字搜索歌曲
//
//	@receiver neteaseCtx
//	@param ctx
//	@param keywords
//	@return result
//	@return err
func (neteaseCtx *NetEaseContext) SearchMusicByKeyWord(ctx context.Context, keywords ...string) (result []*SearchMusicRes, err error) {
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

	searchMusic := searchMusic{}
	jsoniter.Unmarshal(resp1.Body(), &searchMusic)

	urlChan := make(chan *SearchMusicRes)
	wg := &sync.WaitGroup{}
	go func() {
		for i, song := range searchMusic.Result.Songs {
			wg.Add(1)
			go func(index int, song Song) {
				defer wg.Done()
				var ArtistName string
				for _, name := range song.Ar {
					if ArtistName != "" {
						ArtistName += ","
					}
					ArtistName += name.Name
				}
				SongURL, errIn := neteaseCtx.GetMusicURLByID(ctx, map[string]string{strconv.Itoa(song.ID): song.Name})
				if errIn != nil {
					err = errIn
					return
				}
				if len(SongURL) == 0 {
					return
				}
				urlChan <- &SearchMusicRes{
					Index:      index,
					ID:         strconv.Itoa(song.ID),
					Name:       song.Name,
					ArtistName: ArtistName,
					PicURL:     song.Al.PicURL,
					SongURL:    SongURL[0].URL,
				}
			}(i, song)
		}
		wg.Wait()
		close(urlChan)
	}()
	for res := range urlChan {
		result = append(result, res)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Index < result[j].Index
	})
	return
}

// GetNewRecommendMusic 获得新的推荐歌曲
//
//	@receiver ctx
//	@return res
//	@return err
func (neteaseCtx *NetEaseContext) GetNewRecommendMusic() (res []SearchMusicRes, err error) {
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
		SongURL, errIn := neteaseCtx.GetMusicURLByID(context.Background(), map[string]string{strconv.Itoa(result.Song.ID): result.Song.Name})
		if errIn != nil {
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
