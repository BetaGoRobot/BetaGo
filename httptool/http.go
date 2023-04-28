package httptool

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/kevinmatthe/zaplog"
)

var (
	zapLogger   = utility.ZapLogger
	sugerLogger = utility.SugerLogger
)

// HTTP Client variables
var (
	HTTPClientWithProxy *http.Client
	HTTPClient          = &http.Client{}
	proxyURL            = os.Getenv("PRIVATE_PROXY")
	ParsedProxyURL      *url.URL
)

func init() {
	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		panic(err)
	}

	ParsedProxyURL = parsedProxyURL
	HTTPClientWithProxy = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

// RequestInfo 请求字段的结构体
type RequestInfo struct {
	URL     string
	Cookies []*http.Cookie
	Params  map[string][]string
	Header  http.Header
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
		if err != nil {
			zapLogger.Error("获取ip失败", zaplog.Error(err))
		} else {
			zapLogger.Error("获取ip失败", zaplog.Int("StatusCode", resp.StatusCode))
		}
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
	paramSlice := make([]string, 0)
	for key, values := range info.Params {
		paramSlice = append(paramSlice, key+"="+strings.Join(values, "%20"))
	}

	url := strings.Join([]string{info.URL, strings.Join(paramSlice, "&")}, "?")
	// 创建client
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}

	// 添加Cookies
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
	paramSlice := make([]string, 0)
	for key, values := range info.Params {
		paramSlice = append(paramSlice, key+"="+strings.Join(values, "%20"))
	}
	paramSlice = append(paramSlice, "timestamp="+fmt.Sprint(time.Now().UnixNano()))
	rawURL := strings.Join([]string{info.URL, strings.Join(paramSlice, "&")}, "?")
	// 创建client
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
	paramSlice := make([]string, 0)
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

	// 创建client
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
	// 添加Cookies
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
	return postWithParamsInner(info, HTTPClient)
}

func postWithParamsInner(info RequestInfo, client *http.Client) (resp *http.Response, err error) {
	paramSlice := make([]byte, 0)
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
	req.Header = info.Header
	// 添加Cookies
	for index := range info.Cookies {
		req.AddCookie(info.Cookies[index])
	}
	resp, _ = client.Do(req)
	if err != nil {
		zapLogger.Error(err.Error())
		return
	}
	return
}

func PostWithParamsWithProxy(info RequestInfo) (resp *http.Response, err error) {
	return postWithParamsInner(info, HTTPClientWithProxy)
}
