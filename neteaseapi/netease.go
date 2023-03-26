package neteaseapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/httptool"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/gotify"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	jsoniter "github.com/json-iterator/go"
	"go.opentelemetry.io/otel/attribute"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const netEaseQRTmpFile = "/data/tmp"

func init() {
	// 测试环境，使用本地网易云代理
	if betagovar.IsTest {
		NetEaseAPIBaseURL = "http://localhost:3335"
	} else if betagovar.IsCluster {
		NetEaseAPIBaseURL = "http://kubernetes.default:3335"
	}
	NetEaseGCtx.TryGetLastCookie()
	err := NetEaseGCtx.LoginNetEase()
	err = NetEaseGCtx.LoginNetEaseQR()
	if err != nil {
		log.Println("error in init loginNetease", err)
	}
	go func() {
		for {
			time.Sleep(time.Second * 15)
			if NetEaseGCtx.loginType == "qr" {
				// if !NetEaseGCtx.CheckIfLogin() {
				// 	NetEaseGCtx.LoginNetEaseQR()
				// }
			} else {
				NetEaseGCtx.RefreshLogin()
				if NetEaseGCtx.CheckIfLogin() {
					NetEaseGCtx.SaveCookie()
				} else {
					log.Println("error in refresh login")
					// NetEaseGCtx.LoginNetEaseQR()
				}
			}
		}
	}()
}

// RefreshLogin 刷新登录
//
//	@receiver ctx
//	@return error
func (neteaseCtx *NetEaseContext) RefreshLogin() error {
	resp, err := httptool.PostWithParams(
		httptool.RequestInfo{
			URL:     NetEaseAPIBaseURL + "/login/refresh",
			Params:  map[string][]string{},
			Cookies: neteaseCtx.cookies,
		},
	)
	if err != nil || (resp != nil && resp.StatusCode != 200) {
		log.Printf("%#v", resp)
		return err
	}
	neteaseCtx.cookies = make([]*http.Cookie, 0)
	neteaseCtx.cookies = resp.Cookies()
	neteaseCtx.SaveCookie()
	return err
}

func (neteaseCtx *NetEaseContext) getUniKey() (err error) {
	resp, err := httptool.PostWithTimestamp(
		httptool.RequestInfo{
			URL: NetEaseAPIBaseURL + "/login/qr/key",
		},
	)
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			err = fmt.Errorf("LoginNetEaseQR error, StatusCode %d", resp.StatusCode)
		}
		return
	}
	data, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	respMap := make(map[string]interface{})
	if err = json.Unmarshal(data, &respMap); err != nil {
		return
	}
	neteaseCtx.qrStruct.uniKey = respMap["data"].(map[string]interface{})["unikey"].(string)
	return
}

func (neteaseCtx *NetEaseContext) getQRBase64() (err error) {
	resp, err := httptool.PostWithTimestamp(
		httptool.RequestInfo{
			URL: NetEaseAPIBaseURL + "/login/qr/create",
			Params: map[string][]string{
				"key":   {neteaseCtx.qrStruct.uniKey},
				"qrimg": {"1"},
			},
		},
	)
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			err = fmt.Errorf("LoginNetEaseQR error, StatusCode %d", resp.StatusCode)
		}
		return
	}
	data, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	respMap := make(map[string]interface{})
	if err = json.Unmarshal(data, &respMap); err != nil {
		return
	}
	neteaseCtx.qrStruct.qrBase64 = respMap["data"].(map[string]interface{})["qrimg"].(string)
	return
}

func (neteaseCtx *NetEaseContext) checkQRStatus() (err error) {
	if !neteaseCtx.qrStruct.isOutDated {
		once := &sync.Once{}
		for {

			time.Sleep(time.Second)
			resp, err := httptool.PostWithTimestamp(
				httptool.RequestInfo{
					URL: NetEaseAPIBaseURL + "/login/qr/check",
					Params: map[string][]string{
						"key": {neteaseCtx.qrStruct.uniKey},
					},
				},
			)
			if err != nil || resp.StatusCode != 200 {
				if err == nil {
					return fmt.Errorf("LoginNetEaseQR error, StatusCode %d", resp.StatusCode)
				}
				return err
			}
			data, _ := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			respMap := make(map[string]interface{})
			if err = json.Unmarshal(data, &respMap); err != nil {
				return err
			}
			switch respMap["code"].(float64) {
			case 801:
				once.Do(func() { log.Println("Waiting for scan") })
			case 800:
				once.Do(func() {
					log.Println("二维码已失效")
					neteaseCtx.qrStruct.isOutDated = true
				})
				return err
			case 802:
				once.Do(func() { log.Println("扫描未确认") })
			case 803:
				log.Println("登陆成功！")
				neteaseCtx.cookies = resp.Cookies()
				neteaseCtx.SaveCookie()
				neteaseCtx.loginType = "qr"
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
func (neteaseCtx *NetEaseContext) LoginNetEaseQR() (err error) {
	neteaseCtx.getUniKey()

	neteaseCtx.getQRBase64()
	linkURL, err := utility.UploadFileToCos(SaveQRImg(neteaseCtx.qrStruct.qrBase64))
	if err != nil {
		return err
	}
	gotify.SendMessage("网易云登录", fmt.Sprintf("![QRCode](%s)", linkURL), 7)
	go neteaseCtx.checkQRStatus()
	return
}

// SaveQRImg 保存二维码图片
//
//	@param imgBase64
//	@return filename
func SaveQRImg(imgBase64 string) (filename string) {
	i := strings.Index(imgBase64, ",")
	d := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgBase64[i+1:]))
	filename = filepath.Join(netEaseQRTmpFile, fmt.Sprintf("qr_%d.jpg", time.Now().Unix()))
	os.MkdirAll(netEaseQRTmpFile, 0o777)
	f, err := os.Create(filename)
	if err != nil {
		log.Println("error in create qr img", err)
		return ""
	}
	defer f.Close()
	_, err = io.Copy(f, d)
	if err != nil {
		log.Println("error in copy qr img", err)
		return ""
	}
	return
}

// LoginNetEase 获取登陆Cookie
//
//	@receiver ctx
//	@return err
func (neteaseCtx *NetEaseContext) LoginNetEase() (err error) {
	if len(neteaseCtx.cookies) > 0 {
		return
	}
	if phoneNum, password := os.Getenv("NETEASE_PHONE"), os.Getenv("NETEASE_PASSWORD"); phoneNum == "" && password == "" {
		log.Println("Empty NetEase account and password")
		return
	}
	var resp *http.Response
	// !Step1:检查登陆状态
	if neteaseCtx.CheckIfLogin() {
		// 已登陆，刷新登陆
		err = neteaseCtx.RefreshLogin()
		return
	}
	// !Step2:未登陆，启动登陆
	resp, err = httptool.PostWithTimestamp(
		httptool.RequestInfo{
			URL: NetEaseAPIBaseURL + "/login/cellphone",
			Params: map[string][]string{
				"phone":    {os.Getenv("NETEASE_PHONE")},
				"password": {os.Getenv("NETEASE_PASSWORD")},
			},
		},
	)
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			err = fmt.Errorf("LoginNetEase error, StatusCode %d", resp.StatusCode)
		}
		return
	}
	neteaseCtx.cookies = resp.Cookies()
	neteaseCtx.SaveCookie()
	return
}

// CheckIfLogin 检查是否登陆
//
//	@receiver ctx
//	@return bool
func (neteaseCtx *NetEaseContext) CheckIfLogin() bool {
	resp, err := httptool.PostWithTimestamp(
		httptool.RequestInfo{
			URL:     NetEaseAPIBaseURL + "/login/status",
			Params:  map[string][]string{},
			Cookies: neteaseCtx.cookies,
		},
	)
	if err != nil || resp.StatusCode != 200 {
		log.Printf("%#v", resp)
		return false
	}
	data, _ := ioutil.ReadAll(resp.Body)
	loginStatus := LoginStatusStruct{}
	if err = json.Unmarshal(data, &loginStatus); err != nil {
		log.Println("error in unmarshal loginStatus", err)
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
func (neteaseCtx *NetEaseContext) TryGetLastCookie() {
	f, err := os.OpenFile("/data/last_cookie.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		log.Println("error in open last_cookie.json", err)
		return
	}
	defer f.Close()
	cookieData := make([]byte, 0)
	cookieData, err = ioutil.ReadAll(f)
	if len(cookieData) == 0 {
		utility.ZapLogger.Info("No cookieData, skip json marshal")
		return
	}
	if err = json.Unmarshal(cookieData, &neteaseCtx.cookies); err != nil {
		log.Println("error in unmarshal cookieData", err)
	}
}

// SaveCookie 保存Cookie
//
//	@receiver ctx
func (neteaseCtx *NetEaseContext) SaveCookie() {
	f, err := os.OpenFile("/data/last_cookie.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		log.Println("error in open last_cookie.json", err)
		return
	}
	defer f.Close()
	cookieData, _ := json.Marshal(neteaseCtx.cookies)
	f.Write(cookieData)
}

// GetDailyRecommendID 获取当前账号日推
//
//	@receiver ctx
//	@return musicIDs
//	@return err
func (neteaseCtx *NetEaseContext) GetDailyRecommendID() (musicIDs map[string]string, err error) {
	musicIDs = make(map[string]string)
	resp, err := httptool.PostWithTimestamp(
		httptool.RequestInfo{
			URL:     NetEaseAPIBaseURL + "/recommend/songs",
			Cookies: neteaseCtx.cookies,
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
//
//	@receiver ctx
//	@param IDName
//	@return InfoList
//	@return err
func (neteaseCtx *NetEaseContext) GetMusicURLByID(IDName map[string]string) (InfoList []MusicInfo, err error) {
	var id string
	for key := range IDName {
		if id != "" {
			id += ","
		}
		id += key
	}
	resp, err := httptool.PostWithTimestamp(
		httptool.RequestInfo{
			URL:     NetEaseAPIBaseURL + "/song/url",
			Cookies: neteaseCtx.cookies,
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
//
//	@receiver neteaseCtx
//	@param ctx
//	@param keywords
//	@return result
//	@return err
func (neteaseCtx *NetEaseContext) SearchMusicByKeyWord(ctx context.Context, keywords []string) (result []SearchMusicRes, err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("keywords").StringSlice(keywords))
	defer span.End()

	resp1, err := betagovar.HttpClient.R().
		SetFormDataFromValues(
			map[string][]string{
				"limit":    {"3"},
				"type":     {"1"},
				"keywords": {strings.Join(keywords, " ")},
			},
		).
		SetCookies(neteaseCtx.cookies).
		Post(NetEaseAPIBaseURL + "/cloudsearch")
	if err != nil {
		log.Println(err.Error())
	}

	searchMusic := searchMusic{}
	jsoniter.Unmarshal(resp1.Body(), &searchMusic)
	for _, song := range searchMusic.Result.Songs {
		var ArtistName string
		for _, name := range song.Ar {
			if ArtistName != "" {
				ArtistName += ","
			}
			ArtistName += name.Name
		}
		SongURL, errIn := neteaseCtx.GetMusicURLByID(map[string]string{strconv.Itoa(song.ID): song.Name})
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
//
//	@receiver ctx
//	@return res
//	@return err
func (neteaseCtx *NetEaseContext) GetNewRecommendMusic() (res []SearchMusicRes, err error) {
	resp, err := betagovar.HttpClient.R().SetFormDataFromValues(
		map[string][]string{
			"limit": {"5"},
		},
	).Post(NetEaseAPIBaseURL + "/personalized/newsong")
	if err != nil {
		return
	}

	music := &GlobRecommendMusicRes{}
	json.Unmarshal(resp.Body(), music)
	for _, result := range music.Result {
		var ArtistName string
		for _, name := range result.Song.Artists {
			if ArtistName != "" {
				ArtistName += ","
			}
			ArtistName += name.Name
		}
		SongURL, errIn := neteaseCtx.GetMusicURLByID(map[string]string{strconv.Itoa(result.Song.ID): result.Song.Name})
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
