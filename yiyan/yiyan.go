package yiyan

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/BetaGoRobot/BetaGo/httptool"
)

const yiyanURL = "https://api.fanlisky.cn/niuren/getSen"
const yiyanPoemURL = "https://v1.jinrishici.com/all.json"

// GetSen  获取最新一言
//  @return res
func GetSen() (res string) {
	resp, err := httptool.GetWithParams(httptool.RequestInfo{
		URL: yiyanURL,
	})
	if err != nil {
		log.Println("error with req to ", yiyanURL, err.Error())
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	resMap := make(map[string]interface{})
	if err := json.Unmarshal(body, &resMap); err != nil {
		log.Println("error when unmarshal into map", err.Error())
		return
	}
	return resMap["data"].(string)
}

// GetPoem 获取最新一言
//  @return poem
func GetPoem() (poemMap map[string]interface{}) {
	resp, err := httptool.GetWithParams(httptool.RequestInfo{
		URL: yiyanPoemURL,
	})
	if err != nil {
		log.Println("error with req to ", yiyanURL, err.Error())
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	resMap := make(map[string]interface{})
	if err := json.Unmarshal(body, &resMap); err != nil {
		log.Println("error when unmarshal into map", err.Error())
		return
	}
	return resMap
}
