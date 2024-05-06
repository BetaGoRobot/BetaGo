package httptool

import (
	"io/ioutil"
	"testing"

	"github.com/BetaGoRobot/BetaGo/betagovar/env"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/kevinmatthe/zaplog"
)

const NetEaseAPIBaseURL = "http://192.168.31.74:3335"

func TestPostWithParamsWithTimestamp(t *testing.T) {
	resp, err := PostWithTimestamp(RequestInfo{
		URL: NetEaseAPIBaseURL + "/login/cellphone",
		Params: map[string][]string{
			"email":    {env.NETEASE_EMAIL},
			"password": {env.NETEASE_PASSWORD},
		},
	})
	if err != nil || resp.StatusCode != 200 {
		log.ZapLogger.Error("登录失败", zaplog.Error(err))
	}
	data, _ := ioutil.ReadAll(resp.Body)
	log.ZapLogger.Error("登录成功", zaplog.Error(err), zaplog.String("data", string(data)))
	resp, err = PostWithTimestamp(
		RequestInfo{
			URL:     NetEaseAPIBaseURL + "/login/status",
			Params:  map[string][]string{},
			Cookies: resp.Cookies(),
		},
	)
	if err != nil || resp.StatusCode != 200 {
		log.ZapLogger.Error("获取登录状态失败", zaplog.Error(err))
	}
	data, _ = ioutil.ReadAll(resp.Body)
	log.ZapLogger.Info("获取登录状态成功", zaplog.String("data", string(data)))
}

func TestGet(t *testing.T) {
	resp, err := GetWithParams(RequestInfo{
		URL:    "http://ifconfig.me",
		Params: map[string][]string{},
	})
	if err != nil || resp.StatusCode != 200 {
		log.ZapLogger.Error("获取ip失败", zaplog.Error(err))
	}
	data, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	log.ZapLogger.Info("获取ip成功", zaplog.String("data", string(data)))
}
