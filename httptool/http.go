package httptool

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestInfo 请求字段的结构体
type RequestInfo struct {
	URL     string
	Cookies []*http.Cookie
	Params  map[string][]string
}

// GetWithCookieParams 发送带cookie和params的请求
//  @param info 传入的参数、url、cookie信息
//  @return res
//  @return err
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
	fmt.Println("--------", string(req.URL.String()))
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	return
}

// GetWithParams 发送带cookie和params的请求
//  @param info 传入的参数、url、cookie信息
//  @return res
//  @return err
func GetWithParams(info RequestInfo) (resp *http.Response, err error) {
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
	// data, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(data))
	return
}

// PostWithParamsWithTimestamp 发送带Cookie Params的POST请求
func PostWithParamsWithTimestamp(info RequestInfo) (resp *http.Response, err error) {
	params := url.Values{}
	for key, values := range info.Params {
		for index := range values {
			params.Add(key, values[index])
		}
	}
	params.Set("timestamp", fmt.Sprintf("%d", time.Now().UnixNano()))

	// body := ioutil.NopCloser(strings.NewReader(params.Encode()))
	resp, err = http.PostForm(info.URL+"?timestamp="+fmt.Sprint(time.Now().UnixNano()), params)
	if err != nil {
		log.Println(err.Error())
		return
	}

	return
}

// PostWithParams WithTimestamp 发送带Cookie Params的POST请求
//  @param info
//  @return resp
//  @return err
func PostWithParams(info RequestInfo) (resp *http.Response, err error) {
	params := url.Values{}
	for key, values := range info.Params {
		for index := range values {
			params.Add(key, values[index])
		}
	}

	// body := ioutil.NopCloser(strings.NewReader(params.Encode()))
	resp, err = http.PostForm(info.URL+"?timestamp="+fmt.Sprint(time.Now().UnixNano()), params)
	if err != nil {
		log.Println(err.Error())
		return
	}

	return
}
