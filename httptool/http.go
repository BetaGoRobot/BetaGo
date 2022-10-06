package httptool

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/heyuhengmatt/zaplog"
)

var (
	zapLogger   = utility.ZapLogger
	sugerLogger = utility.SugerLogger
)

// RequestInfo 请求字段的结构体
type RequestInfo struct {
	URL     string
	Cookies []*http.Cookie
	Params  map[string][]string
}

// GetPubIP 获取公网ip
//
//	@return ip
//	@return err
func GetPubIP() (ip string, err error) {
	resp, err := GetWithParams(RequestInfo{
		URL:    "http://ifconfig.me",
		Params: map[string][]string{},
	})
	if err != nil || resp.StatusCode != 200 {
		zapLogger.Error("获取ip失败", zaplog.Error(err))
	}
	data, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	ip = string(data)
	zapLogger.Info("获取ip成功", zaplog.String("data", ip))
	return
}

// GetWithCookieParams 发送带cookie和params的请求
//
//	@param info 传入的参数、url、cookie信息
//	@return res
//	@return err
func GetWithCookieParams(info RequestInfo) (resp *http.Response, err error) {
	var paramSlice = make([]string, 0)
	for key, values := range info.Params {
		paramSlice = append(paramSlice, key+"="+strings.Join(values, "%20"))
	}

	url := strings.Join([]string{info.URL, strings.Join(paramSlice, "&")}, "?")
	//创建client
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}

	//添加Cookies
	for index := range info.Cookies {
		req.AddCookie(info.Cookies[index])
	}
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	return
}

// GetWithParamsWithTimestamp 发送带cookie和params的请求
//
//	@param info 传入的参数、url、cookie信息
//	@return res
//	@return err
func GetWithParamsWithTimestamp(info RequestInfo) (resp *http.Response, err error) {
	var paramSlice = make([]string, 0)
	for key, values := range info.Params {
		paramSlice = append(paramSlice, key+"="+strings.Join(values, "%20"))
	}
	paramSlice = append(paramSlice, "timestamp="+fmt.Sprint(time.Now().UnixNano()))
	rawURL := strings.Join([]string{info.URL, strings.Join(paramSlice, "&")}, "?")
	//创建client
	resp, err = http.Get(rawURL)
	if err != nil {
		return
	}
	return
}

// GetWithParams 发送带cookie和params的请求
//
//	@param info 传入的参数、url、cookie信息
//	@return res
//	@return err
func GetWithParams(info RequestInfo) (resp *http.Response, err error) {
	var paramSlice = make([]string, 0)
	for key, values := range info.Params {
		for index := range values {
			paramSlice = append(paramSlice, key+"="+values[index])
		}
	}
	var rawURL string
	if len(paramSlice) > 0 {
		rawURL = strings.Join([]string{info.URL, strings.Join(paramSlice, "&")}, "?")
	} else {
		rawURL = info.URL
	}

	//创建client
	resp, err = http.Get(rawURL)
	if err != nil {
		return
	}
	return
}

// PostWithTimestamp 发送带Cookie Params的POST请求
func PostWithTimestamp(info RequestInfo) (resp *http.Response, err error) {
	FormData := url.Values{}
	for key, values := range info.Params {
		for index := range values {
			FormData.Add(key, values[index])
		}
	}
	req, err := http.NewRequest(http.MethodPost, info.URL+"?timestamp="+fmt.Sprint(time.Now().UnixNano()), strings.NewReader(FormData.Encode()))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	//添加Cookies
	for index := range info.Cookies {
		req.AddCookie(info.Cookies[index])
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		zapLogger.Error(err.Error())
		return
	}
	return
}

// PostWithParams WithTimestamp 发送带Cookie Params的POST请求
//
//	@param info
//	@return resp
//	@return err
func PostWithParams(info RequestInfo) (resp *http.Response, err error) {
	var paramSlice = make([]byte, 0)
	req, err := http.NewRequest(http.MethodPost, info.URL, bytes.NewReader([]byte(paramSlice)))
	if err != nil {
		return
	}
	req.PostForm = url.Values{}
	for key, values := range info.Params {
		for index := range values {
			req.PostForm.Add(key, values[index])
		}
	}
	//添加Cookies
	for index := range info.Cookies {
		req.AddCookie(info.Cookies[index])
	}
	resp, _ = http.DefaultClient.Do(req)
	if err != nil {
		zapLogger.Error(err.Error())
		return
	}

	return
}
