package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type getHeader struct {
	headerName  string
	headerValue string
}

func sendGetWithHeader(link, api string, header getHeader, params map[string][]string) string {
	req, err := http.NewRequest(http.MethodGet, link+"/"+api, nil)
	client := http.Client{}
	if err != nil {
		log.Println(err.Error())
	}
	req.Header.Add(header.headerName, header.headerValue)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
	}
	fmt.Println(string(b))
	// req.URL.RawQuery = url.Values(params).Encode()
	// parsedURL := req.URL.String()
	// fmt.Println(parsedURL)
	// resp, err := http.Get(parsedURL)
	// defer resp.Body.Close()
	// if err != nil {
	// 	log.Println(err.Error())
	// }
	// b, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Println(err.Error())
	// }
	// fmt.Println(string(b))
	return ""
}
func sendGetReq(location, api string, params map[string]string) string {
	var url string
	for key, value := range params {
		if url != "" {
			url += "&"
		}
		url += key + "=" + value
	}
	url = location + "/" + api + "?" + url
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	// res := loginJSON{}
	// _ = json.Unmarshal(body, &res)
	// fmt.Println(res.Code)
	return string(body)
}
func loginNetEase() {
	resp, err := http.Get("http://localhost:3333/login/cellphone?phone=18681655914&password=heyuheng1.22.3")
	if err != nil {
		fmt.Println(err.Error())
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(body))
}
