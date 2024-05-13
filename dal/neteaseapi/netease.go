package neteaseapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/consts/ct"
	"github.com/BetaGoRobot/BetaGo/consts/env"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/gotify"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	miniohelper "github.com/BetaGoRobot/BetaGo/utility/minio_helper"
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
		time.Sleep(time.Second * 10) // 等待本地网络启动
	}

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
			err = fmt.Errorf("LoginNetEaseQR error, StatusCode %d", resp.StatusCode())
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
			err = fmt.Errorf("LoginNetEaseQR error, StatusCode %d", resp.StatusCode())
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
	linkURL, err := miniohelper.Client().
		SetContext(ctx).
		SetBucketName("cloudmusic").
		SetFileFromReader(qrImgReadCloser(ctx, neteaseCtx.qrStruct.qrBase64)).
		SetObjName("QRCode/" + strconv.Itoa(int(time.Now().Unix())) + ".png").
		SetContentType(ct.ContentTypeImgPNG).
		SetExpiration(time.Now().Add(time.Hour)).
		Upload()
	if err != nil {
		return err
	}

	gotify.SendMessage(ctx, "网易云登录", fmt.Sprintf("![QRCode](%s)", linkURL.String()), 7)
	go neteaseCtx.checkQRStatus(ctx)
	return
}

func qrImgReadCloser(ctx context.Context, imgBase64 string) (r io.ReadCloser) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	i := strings.Index(imgBase64, ",") // string is img/png;base64,xxx
	d := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgBase64[i+1:]))

	return io.NopCloser(d)
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
		musicIDURL[ID] = URL
		go uploadMusic(ctx, URL, ID)
	}
	return
}

func uploadMusic(ctx context.Context, url string, ID string) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("songID").String(ID))
	defer span.End()
	_, err := miniohelper.Client().
		SetContext(ctx).
		SetBucketName("cloudmusic").
		SetFileFromURL(url).
		SetObjName("music/" + ID + filepath.Ext(url)).
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
	l, err := miniohelper.Client().
		SetContext(ctx).
		SetBucketName("cloudmusic").
		SetFileFromString(lyricJson).
		SetObjName("lyrics/" + songID + ".json").
		SetContentType(ct.ContentTypePlainText).
		Upload()
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

	searchRes := searchMusic{}
	jsoniter.Unmarshal(resp1.Body(), &searchRes)

	sResChan := make(chan *SearchMusicRes, len(searchRes.Result.Songs))

	go neteaseCtx.asyncGetSearchRes(ctx, searchRes, err, sResChan)

	m := asyncUploadPics(ctx, searchRes)
	for res := range sResChan {
		res.ImageKey = m[res.ID]
		result = append(result, res)
	}
	return
}

func asyncUploadPics(ctx context.Context, musicInfos searchMusic) map[string]string {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	var (
		c  = make(chan [2]string)
		wg = &sync.WaitGroup{}
		m  = make(map[string]string, 1)
	)
	go func() {
		defer close(c)
		defer wg.Wait()

		for _, m := range musicInfos.Result.Songs {
			wg.Add(1)
			go uploadPicWorker(ctx, wg, m.Al.PicURL, m.ID, c)
		}
	}()
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

func (neteaseCtx *NetEaseContext) asyncGetSearchRes(ctx context.Context, searchMusic searchMusic, err error, urlChan chan *SearchMusicRes) {
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
		urlChan <- &SearchMusicRes{
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
