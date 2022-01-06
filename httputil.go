package main

import (
	"fmt"
	"net/http"
	"strings"
)

// GetRequestInfo Get请求字段的结构体
type GetRequestInfo struct {
	URL     string
	cookies []*http.Cookie
	params  map[string]string
}

// GetWithCookieParams 发送带cookie和params的请求
//  @param info 传入的参数、url、cookie信息
//  @return res
//  @return err
func GetWithCookieParams(info GetRequestInfo) (resp *http.Response, err error) {
	var paramSlice = make([]string, 0)

	//处理参数
	for key := range info.params {
		paramSlice = append(paramSlice, strings.Join([]string{key, info.params[key]}, "="))
	}
	url := strings.Join([]string{info.URL, strings.Join(paramSlice, "&")}, "?")

	//创建client
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}

	//添加Cookies
	for index := range info.cookies {
		req.AddCookie(info.cookies[index])
	}
	fmt.Println("--------", string(req.URL.String()))
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	return
}

// PostWithParamsCookie 发送带Cookie Params的POST请求
func PostWithParamsCookie(info GetRequestInfo) (resp *http.Response, err error) {
	//TODO 实现POST请求
	return
}
